package provider

import (
	"fmt"
	"github.com/johannesboyne/gofakes3"
	"os"
	"testing"
)

func Setup(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func PrintBuckets(b gofakes3.Backend) error {
	buckets, err := b.ListBuckets()
	if err != nil {
		return err
	}
	for _, bucket := range buckets {
		fmt.Println("* " + bucket.Name)
		objects, oerr := b.ListBucket(bucket.Name, nil, gofakes3.ListBucketPage{
			Marker:    "",
			HasMarker: false,
			MaxKeys:   100,
		})
		if oerr != nil {
			return oerr
		}
		for _, o := range objects.Contents {
			fmt.Println("   * " + o.Key)
		}
	}
	return nil
}

func TestSomething(t *testing.T) {
	//// Create a new bucket using the CreateBucket call.
	//_, err := client.CreateBucket(context.TODO(), cparams)
	//if err != nil {
	//	// Message from an error.
	//	fmt.Println(err.Error())
	//	return
	//}
	//
	//// Upload a new object "testobject" with the string "Hello World!" to our "newbucket".
	//_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
	//	Body:   strings.NewReader(`{"configuration": {"main_color": "#333"}, "screens": []}`),
	//	Bucket: aws.String("public-read"),
	//	Key:    aws.String("test.txt"),
	//})

	//// DEBUG: Print all buckets and their objects
	//berr := PrintBuckets(backend)
	//if err != nil {
	//	panic(berr)
	//}
}

//func TestCustomProvider_BucketExists(t *testing.T) {
//	t.Parallel()
//
//	p := providers["custom"]
//	var tests = []struct {
//		name   string
//		b      bucket.Bucket
//		exists uint8
//	}{
//		{name: "exists, access denied", b: bucket.NewBucket("assets"), exists: bucket.BucketExists},
//		{name: "exists, open", b: bucket.NewBucket("nurse-virtual-assistants"), exists: bucket.BucketExists},
//		{name: "no such bucket", b: bucket.NewBucket("s3scanner-no-exist"), exists: bucket.BucketNotExist},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t2 *testing.T) {
//			gb, err := p.BucketExists(&tt.b)
//			assert.Nil(t2, err)
//			assert.Equal(t2, tt.exists, gb.Exists)
//		})
//	}
//}
