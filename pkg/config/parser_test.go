package config

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"
)

func TestParseConfig(t *testing.T) {
	runParseConfig := func(rawYaml string, config *Config) error {
		tmpFile, err := ioutil.TempFile("", "kloudConfigTest")
		if err != nil {
			return err
		}

		tmpFile.WriteString(rawYaml)
		tmpFile.Close()

		if err != nil {
			return err
		}
		defer os.Remove(tmpFile.Name())

		return parseConfig(tmpFile.Name(), config)
	}

	equal := func(actual, expected interface{}) {
		if expected != actual {
			t.Errorf("expected %v, got %v\n", expected, actual)
		}
	}

	t.Run("Valid YAML with root dest", func(t *testing.T) {
		rawYaml := `server: https://cloud.domain.com
share: XXXX`

		var config Config
		if err := runParseConfig(rawYaml, &config); err != nil {
			t.Error(err)
		}

		equal(config.Server, "https://cloud.domain.com")
		equal(config.ShareID, "XXXX")
	})

	t.Run("Invalid YAML", func(t *testing.T) {
		rawYaml := `server = cloud.domain.com`

		var config Config
		if err := runParseConfig(rawYaml, &config); err == nil {
			t.Errorf("Expected error but got nil")
		}
	})

	t.Run("File does not exist", func(t *testing.T) {
		var config Config
		err := parseConfig("/i/dont/exist", &config)
		equal(err != nil, true)
	})
}

func TestValidateConfig(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "kloudConfigTest")
	if err != nil {
		t.Error(err)
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	config := Config{"https://cloud.domain.com", "XXX"}

	equal := func(err, target error) {
		if errors.Is(err, target) == false {
			t.Errorf("expected %v but got %v", target, err)
		}
	}

	equal(validateConfig(config), nil)

	config.Server = "cloud.domain.com"
	equal(validateConfig(config), ErrMissingScheme)
}
