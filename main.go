package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/sa7mon/s3scanner/bucket"
	"github.com/sa7mon/s3scanner/db"
	log2 "github.com/sa7mon/s3scanner/log"
	"github.com/sa7mon/s3scanner/mq"
	"github.com/sa7mon/s3scanner/provider"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"os"
	"reflect"
	"strings"
	"sync"
	"text/tabwriter"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func printResult(b *bucket.Bucket) {
	if args.json {
		log.WithField("bucket", b).Info()
		return
	}

	if b.Exists == bucket.BucketNotExist {
		log.Infof("not_exist | %s", b.Name)
		return
	}

	result := fmt.Sprintf("exists    | %v | %v | %v", b.Name, b.Region, b.String())
	if b.ObjectsEnumerated {
		result = fmt.Sprintf("%v | %v objects (%v)", result, len(b.Objects), humanize.Bytes(b.BucketSize))
	}
	log.Info(result)
}

func work(wg *sync.WaitGroup, buckets chan bucket.Bucket, provider provider.StorageProvider, enumerate bool, writeToDB bool) {
	defer wg.Done()
	for b1 := range buckets {
		b, existsErr := provider.BucketExists(&b1)
		if existsErr != nil {
			log.Errorf("error     | %s | %s", b.Name, existsErr.Error())
			continue
		}

		if b.Exists == bucket.BucketNotExist {
			printResult(b)
			continue
		}

		// Scan permissions
		scanErr := provider.Scan(b, false)
		if scanErr != nil {
			log.WithFields(log.Fields{"bucket": b}).Error(scanErr)
		}

		if enumerate && b.PermAllUsersRead == bucket.PermissionAllowed {
			log.WithFields(log.Fields{"method": "main.work()",
				"bucket_name": b.Name, "region": b.Region}).Debugf("enumerating objects...")
			enumErr := provider.Enumerate(b)
			if enumErr != nil {
				log.Errorf("Error enumerating bucket '%s': %v\nEnumerated objects: %v", b.Name, enumErr, len(b.Objects))
				continue
			}
		}
		printResult(b)

		if writeToDB {
			dbErr := db.StoreBucket(b)
			if dbErr != nil {
				log.Error(dbErr)
			}
		}
	}
}

func mqwork(threadId int, wg *sync.WaitGroup, conn *amqp.Connection, provider provider.StorageProvider, queue string, threads int,
	doEnumerate bool, writeToDB bool) {
	_, once := os.LookupEnv("TEST_MQ") // If we're being tested, exit after one bucket is scanned
	defer wg.Done()

	// Wrap the whole thing in a for (while) loop so if the mq server kills the channel, we start it up again
	for {
		ch, chErr := mq.Connect(conn, queue, threads, threadId)
		if chErr != nil {
			failOnError(chErr, "couldn't connect to message queue")
		}

		msgs, consumeErr := ch.Consume(queue, fmt.Sprintf("%s_%v", queue, threadId), false, false, false, false, nil)
		if consumeErr != nil {
			log.Error(fmt.Errorf("failed to register a consumer: %w", consumeErr))
			return
		}

		for j := range msgs {
			bucketToScan := bucket.Bucket{}

			unmarshalErr := json.Unmarshal(j.Body, &bucketToScan)
			if unmarshalErr != nil {
				log.Error(unmarshalErr)
			}

			if !bucket.IsValidS3BucketName(bucketToScan.Name) {
				log.Info(fmt.Sprintf("invalid   | %s", bucketToScan.Name))
				failOnError(j.Ack(false), "failed to ack")
				continue
			}

			b, existsErr := provider.BucketExists(&bucketToScan)
			if existsErr != nil {
				log.WithFields(log.Fields{"bucket": b.Name, "step": "checkExists"}).Error(existsErr)
				failOnError(j.Reject(false), "failed to reject")
			}
			if b.Exists == bucket.BucketNotExist {
				// ack the message and skip to the next
				log.Infof("not_exist | %s", b.Name)
				failOnError(j.Ack(false), "failed to ack")
				continue
			}

			scanErr := provider.Scan(b, false)
			if scanErr != nil {
				log.WithFields(log.Fields{"bucket": b}).Error(scanErr)
				failOnError(j.Reject(false), "failed to reject")
				continue
			}

			if doEnumerate {
				if b.PermAllUsersRead != bucket.PermissionAllowed {
					printResult(&bucketToScan)
					failOnError(j.Ack(false), "failed to ack")
					if writeToDB {
						dbErr := db.StoreBucket(&bucketToScan)
						if dbErr != nil {
							log.Error(dbErr)
						}
					}
					continue
				}

				log.WithFields(log.Fields{"method": "main.mqwork()",
					"bucket_name": b.Name, "region": b.Region}).Debugf("enumerating objects...")

				enumErr := provider.Enumerate(b)
				if enumErr != nil {
					log.Errorf("Error enumerating bucket '%s': %v\nEnumerated objects: %v", b.Name, enumErr, len(b.Objects))
					failOnError(j.Reject(false), "failed to reject")
				}
			}

			printResult(&bucketToScan)
			ackErr := j.Ack(false)
			if ackErr != nil {
				// Acknowledge mq message. May fail if we've taken too long and the server has closed the channel
				// If it has, we break and start at the top of the outer for-loop again which re-establishes a new
				// channel
				log.WithFields(log.Fields{"bucket": b}).Error(ackErr)
				break
			}

			// Write to database
			if writeToDB {
				dbErr := db.StoreBucket(&bucketToScan)
				if dbErr != nil {
					log.Error(dbErr)
				}
			}
			if once {
				return
			}
		}
	}
}

type flagSetting struct {
	indentLevel int
	category    int
}

type argCollection struct {
	bucketFile   string
	bucketName   string
	doEnumerate  bool
	json         bool
	providerFlag string
	threads      int
	useMq        bool
	verbose      bool
	version      bool
	writeToDB    bool
}

func (args argCollection) Validate() error {
	// Validate: only 1 input flag is provided
	numInputFlags := 0
	if args.useMq {
		numInputFlags += 1
	}
	if args.bucketName != "" {
		numInputFlags += 1
	}
	if args.bucketFile != "" {
		numInputFlags += 1
	}
	if numInputFlags != 1 {
		return errors.New("exactly one of: -bucket, -bucket-file, -mq required")
	}

	return nil
}

/*
validateConfig checks that the config file contains all necessary keys according to the args specified
*/
func validateConfig(args argCollection) error {
	expectedKeys := []string{}
	configFileRequired := false
	if args.providerFlag == "custom" {
		configFileRequired = true
		expectedKeys = append(expectedKeys, []string{"providers.custom.insecure", "providers.custom.endpoint_format", "providers.custom.regions", "providers.custom.address_style"}...)
	}
	if args.writeToDB {
		configFileRequired = true
		expectedKeys = append(expectedKeys, []string{"db.uri"}...)
	}
	if args.useMq {
		configFileRequired = true
		expectedKeys = append(expectedKeys, []string{"mq.queue_name", "mq.uri"}...)
	}
	// User didn't give any arguments that require the config file
	if !configFileRequired {
		return nil
	}

	// Try to find and read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Error("config file not found")
			os.Exit(1)
		} else {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	// Verify all expected keys are in the config file
	for _, k := range expectedKeys {
		if !viper.IsSet(k) {
			return fmt.Errorf("config file missing key: %s", k)
		}
	}
	return nil
}

const (
	CategoryInput   int = 0
	CategoryOutput  int = 1
	CategoryOptions int = 2
	CategoryDebug   int = 3
)

var configPaths = []string{".", "/etc/s3scanner/", "$HOME/.s3scanner/"}

var version = "dev"
var args = argCollection{}

func main() {
	// https://twin.sh/articles/39/go-concurrency-goroutines-worker-pools-and-throttling-made-simple
	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws#AnonymousCredentials

	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yml")    // REQUIRED if the config file does not have the extension in the name
	for _, p := range configPaths {
		viper.AddConfigPath(p)
	}

	flagSettings := make(map[string]flagSetting, 11)
	flag.StringVar(&args.providerFlag, "provider", "aws", fmt.Sprintf(
		"Object storage provider: %s - custom requires config file.",
		strings.Join(provider.AllProviders, ", ")))
	flagSettings["provider"] = flagSetting{category: CategoryOptions}
	flag.StringVar(&args.bucketName, "bucket", "", "Name of bucket to check.")
	flagSettings["bucket"] = flagSetting{category: CategoryInput}
	flag.StringVar(&args.bucketFile, "bucket-file", "", "File of bucket names to check.")
	flagSettings["bucket-file"] = flagSetting{category: CategoryInput}
	flag.BoolVar(&args.useMq, "mq", false, "Connect to RabbitMQ to get buckets. Requires config file key \"mq\".")
	flagSettings["mq"] = flagSetting{category: CategoryInput}

	flag.BoolVar(&args.writeToDB, "db", false, "Save results to a Postgres database. Requires config file key \"db.uri\".")
	flagSettings["db"] = flagSetting{category: CategoryOutput}
	flag.BoolVar(&args.json, "json", false, "Print logs to stdout in JSON format instead of human-readable.")
	flagSettings["json"] = flagSetting{category: CategoryOutput}

	flag.BoolVar(&args.doEnumerate, "enumerate", false, "Enumerate bucket objects (can be time-consuming).")
	flagSettings["enumerate"] = flagSetting{category: CategoryOptions}
	flag.IntVar(&args.threads, "threads", 4, "Number of threads to scan with.")
	flagSettings["threads"] = flagSetting{category: CategoryOptions}
	flag.BoolVar(&args.verbose, "verbose", false, "Enable verbose logging.")
	flagSettings["verbose"] = flagSetting{category: CategoryDebug}
	flag.BoolVar(&args.version, "version", false, "Print version")
	flagSettings["version"] = flagSetting{category: CategoryDebug}

	flag.Usage = func() {
		bufferCategoryInput := new(bytes.Buffer)
		bufferCategoryOutput := new(bytes.Buffer)
		bufferCategoryOptions := new(bytes.Buffer)
		bufferCategoryDebug := new(bytes.Buffer)
		categoriesWriters := map[int]*tabwriter.Writer{
			CategoryInput:   tabwriter.NewWriter(bufferCategoryInput, 0, 0, 2, ' ', 0),
			CategoryOutput:  tabwriter.NewWriter(bufferCategoryOutput, 0, 0, 2, ' ', 0),
			CategoryOptions: tabwriter.NewWriter(bufferCategoryOptions, 0, 0, 2, ' ', 0),
			CategoryDebug:   tabwriter.NewWriter(bufferCategoryDebug, 0, 0, 2, ' ', 0),
		}
		flag.VisitAll(func(f *flag.Flag) {
			setting, ok := flagSettings[f.Name]
			if !ok {
				log.Errorf("flag is missing category: %s", f.Name)
				os.Exit(1)
			}
			writer := categoriesWriters[setting.category]

			fmt.Fprintf(writer, "%s  -%s\t", strings.Repeat("   ", setting.indentLevel), f.Name) // Two spaces before -; see next two comments.
			name, usage := flag.UnquoteUsage(f)
			fmt.Fprintf(writer, " %s\t", name)
			fmt.Fprint(writer, usage)
			if !reflect.ValueOf(f.DefValue).IsZero() {
				fmt.Fprintf(writer, " Default: %q", f.DefValue)
			}
			fmt.Fprint(writer, "\n")
		})

		// Output all the categories
		categoriesWriters[CategoryInput].Flush()
		categoriesWriters[CategoryOutput].Flush()
		categoriesWriters[CategoryOptions].Flush()
		categoriesWriters[CategoryDebug].Flush()
		fmt.Fprint(flag.CommandLine.Output(), "INPUT: (1 required)\n", bufferCategoryInput.String())
		fmt.Fprint(flag.CommandLine.Output(), "\nOUTPUT:\n", bufferCategoryOutput.String())
		fmt.Fprint(flag.CommandLine.Output(), "\nOPTIONS:\n", bufferCategoryOptions.String())
		fmt.Fprint(flag.CommandLine.Output(), "\nDEBUG:\n", bufferCategoryDebug.String())

		// Add config file description
		quotedPaths := ""
		for i, b := range configPaths {
			if i != 0 {
				quotedPaths += " "
			}
			quotedPaths += fmt.Sprintf("\"%s\"", b)
		}

		fmt.Fprintf(flag.CommandLine.Output(), "\nIf config file is required these locations will be searched for config.yml: %s\n",
			quotedPaths)
	}
	flag.Parse()

	if args.version {
		fmt.Println(version)
		os.Exit(0)
	}

	argsErr := args.Validate()
	if argsErr != nil {
		log.Error(argsErr)
		os.Exit(1)
	}

	// Configure logging
	log.SetLevel(log.InfoLevel)
	if args.verbose {
		log.SetLevel(log.DebugLevel)
	}
	log.SetOutput(os.Stdout)
	if args.json {
		log.SetFormatter(&log2.NestedJSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
	}

	var p provider.StorageProvider
	var err error
	configErr := validateConfig(args)
	if configErr != nil {
		log.Error(configErr)
		os.Exit(1)
	}
	if args.providerFlag == "custom" {
		if viper.IsSet("providers.custom") {
			log.Debug("found custom provider")
			p, err = provider.NewCustomProvider(
				viper.GetString("providers.custom.address_style"),
				viper.GetBool("providers.custom.insecure"),
				viper.GetStringSlice("providers.custom.regions"),
				viper.GetString("providers.custom.endpoint_format"))
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}
		}
	} else {
		p, err = provider.NewProvider(args.providerFlag)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
	}

	// Setup database connection
	if args.writeToDB {
		dbConfig := viper.GetString("db.uri")
		log.Debugf("using database URI from config: %s", dbConfig)
		dbErr := db.Connect(dbConfig, true)
		if dbErr != nil {
			log.Error(dbErr)
			os.Exit(1)
		}
	}

	var wg sync.WaitGroup

	if !args.useMq {
		buckets := make(chan bucket.Bucket)

		for i := 0; i < args.threads; i++ {
			wg.Add(1)
			go work(&wg, buckets, p, args.doEnumerate, args.writeToDB)
		}

		if args.bucketFile != "" {
			err := bucket.ReadFromFile(args.bucketFile, buckets)
			close(buckets)
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}
		} else if args.bucketName != "" {
			if !bucket.IsValidS3BucketName(args.bucketName) {
				log.Info(fmt.Sprintf("invalid   | %s", args.bucketName))
				os.Exit(0)
			}
			c := bucket.NewBucket(strings.ToLower(args.bucketName))
			buckets <- c
			close(buckets)
		}

		wg.Wait()
		os.Exit(0)
	}

	// Setup mq connection and spin off consumers
	mqUri := viper.GetString("mq.uri")
	mqName := viper.GetString("mq.queue_name")
	conn, err := amqp.Dial(mqUri)
	failOnError(err, fmt.Sprintf("failed to connect to AMQP URI '%s'", mqUri))
	defer conn.Close()

	for i := 0; i < args.threads; i++ {
		wg.Add(1)
		go mqwork(i, &wg, conn, p, mqName, args.threads, args.doEnumerate, args.writeToDB)
	}
	log.Printf("Waiting for messages. To exit press CTRL+C")
	wg.Wait()
}
