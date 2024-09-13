package s3scanner

import (
	"errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestArgCollection_Validate(t *testing.T) {
	goodInputs := []ArgCollection{
		{
			BucketName: "asdf",
			BucketFile: "",
			UseMq:      false,
		},
		{
			BucketName: "",
			BucketFile: "buckets.txt",
			UseMq:      false,
		},
		{
			BucketName: "",
			BucketFile: "",
			UseMq:      true,
		},
	}
	tooManyInputs := []ArgCollection{
		{
			BucketName: "asdf",
			BucketFile: "asdf",
			UseMq:      false,
		},
		{
			BucketName: "adsf",
			BucketFile: "",
			UseMq:      true,
		},
		{
			BucketName: "",
			BucketFile: "asdf.txt",
			UseMq:      true,
		},
	}

	for _, v := range goodInputs {
		err := v.Validate()
		if err != nil {
			t.Errorf("%v: %e", v, err)
		}
	}
	for _, v := range tooManyInputs {
		err := v.Validate()
		if err == nil {
			t.Errorf("expected error but did not find one: %v", v)
		}
	}
}

func TestValidateConfig(t *testing.T) {
	a := ArgCollection{
		DoEnumerate:  false,
		JSON:         false,
		ProviderFlag: "custom",
		UseMq:        true,
		WriteToDB:    true,
	}
	viper.AddConfigPath("../../")
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yml")    // REQUIRED if the config file does not have the extension in the name
	err := validateConfig(a)
	assert.Nil(t, err)
}

func TestValidateConfig_NotFound(t *testing.T) {
	a := ArgCollection{
		DoEnumerate:  false,
		JSON:         false,
		ProviderFlag: "custom",
		UseMq:        true,
		WriteToDB:    true,
	}
	viper.SetConfigName("asdf") // won't be found
	viper.SetConfigType("yml")
	err := validateConfig(a)
	assert.Equal(t, errors.New("config file not found"), err)
}
