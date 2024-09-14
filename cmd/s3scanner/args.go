package s3scanner

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
)

type ArgCollection struct {
	BucketFile   string
	BucketName   string
	DoEnumerate  bool
	JSON         bool
	ProviderFlag string
	Threads      int
	UseMq        bool
	Verbose      bool
	Version      bool
	WriteToDB    bool
}

func (args ArgCollection) Validate() error {
	// Validate: only 1 input flag is provided
	numInputFlags := 0
	if args.UseMq {
		numInputFlags++
	}
	if args.BucketName != "" {
		numInputFlags++
	}
	if args.BucketFile != "" {
		numInputFlags++
	}
	if numInputFlags != 1 {
		return errors.New("exactly one of: -bucket, -bucket-file, -mq required")
	}

	return nil
}

/*
validateConfig checks that the config file contains all necessary keys according to the args specified
*/
func validateConfig(args ArgCollection) error {
	expectedKeys := []string{}
	configFileRequired := false
	if args.ProviderFlag == "custom" {
		configFileRequired = true
		expectedKeys = append(expectedKeys, []string{"providers.custom.insecure", "providers.custom.endpoint_format", "providers.custom.regions", "providers.custom.address_style"}...)
	}
	if args.WriteToDB {
		configFileRequired = true
		expectedKeys = append(expectedKeys, []string{"db.uri"}...)
	}
	if args.UseMq {
		configFileRequired = true
		expectedKeys = append(expectedKeys, []string{"mq.queue_name", "mq.uri"}...)
	}
	// User didn't give any arguments that require the config file
	if !configFileRequired {
		return nil
	}

	// Try to find and read config file
	if err := viper.ReadInConfig(); err != nil {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if ok {
			return errors.New("config file not found")
		}
		return err
	}

	// Verify all expected keys are in the config file
	for _, k := range expectedKeys {
		if !viper.IsSet(k) {
			return fmt.Errorf("config file missing key: %s", k)
		}
	}
	return nil
}
