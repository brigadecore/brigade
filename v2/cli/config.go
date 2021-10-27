package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/brigadecore/brigade-foundations/file"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

type config struct {
	APIAddress       string `json:"apiAddress"`
	APIToken         string `json:"apiToken"`
	IgnoreCertErrors bool   `json:"ignoreCertErrors"`
}

var getHomeDir = homedir.Dir

func getConfig() (config, error) {
	cfg := config{}

	configFile, err := getConfigPath()
	if err != nil {
		return cfg, errors.Wrapf(err, "error finding configuration file path")
	}
	exists, err := file.Exists(configFile)
	if err != nil {
		return cfg, errors.Wrapf(err, "error finding configuration file")
	}
	if !exists {
		return cfg, errors.Errorf(
			"no configuration file was found at %s; please use `brig login` to "+
				"continue\n",
			configFile,
		)
	}

	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return cfg, errors.Wrapf(
			err,
			"error reading configuration file at %s",
			configFile,
		)
	}

	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		return cfg, errors.Wrapf(
			err,
			"error parsing configuration file at %s",
			configFile,
		)
	}

	return cfg, nil
}

func saveConfig(config config) error {
	brigadeHome, err := getBrigadeHome()
	if err != nil {
		return errors.Wrapf(err, "error locating brigade home directory")
	}
	if _, err = os.Stat(brigadeHome); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(
				err,
				"error checking for existence of brigade home at %s",
				brigadeHome,
			)
		}
		// The directory doesn't exist-- create it
		if err = os.MkdirAll(brigadeHome, 0755); err != nil {
			return errors.Wrapf(
				err,
				"error creating brigade home at %s",
				brigadeHome,
			)
		}
	}
	configFile := path.Join(brigadeHome, "config")

	configBytes, err := json.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "error marshaling config")
	}
	if err :=
		ioutil.WriteFile(configFile, configBytes, 0600); err != nil {
		return errors.Wrapf(err, "error writing to %s", configFile)
	}
	return nil
}

func deleteConfig() error {
	configFile, err := getConfigPath()
	if err != nil {
		return errors.Wrapf(err, "error finding configuration")
	}
	if err := os.Remove(configFile); err != nil {
		return errors.Wrap(err, "error deleting configuration")
	}
	return nil
}

func getBrigadeHome() (string, error) {
	homeDir, err := getHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "error locating user's home directory")
	}
	return path.Join(homeDir, ".brigade"), nil
}

func getConfigPath() (string, error) {
	brigadeHome, err := getBrigadeHome()
	if err != nil {
		return "", errors.Wrap(err, "error locating brigade home directory")
	}
	return path.Join(brigadeHome, "config"), nil
}
