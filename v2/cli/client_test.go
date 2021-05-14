package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetClient(t *testing.T) {
	testHome, err := ioutil.TempDir("", "home")
	require.NoError(t, err)
	getHomeDir = func() (string, error) {
		return testHome, nil
	}
	testConfig := config{
		APIAddress: "http://localhost:8080",
		APIToken:   "thisisafaketoken",
	}
	brigadeHome := path.Join(testHome, ".brigade")
	err = os.Mkdir(brigadeHome, 0755)
	require.NoError(t, err)
	configFile := path.Join(testHome, ".brigade", "config")
	configBytes, err := json.Marshal(testConfig)
	require.NoError(t, err)
	err = ioutil.WriteFile(configFile, configBytes, 0644)
	require.NoError(t, err)
	cfg, err := getConfig()
	require.NoError(t, err)
	require.Equal(t, testConfig, cfg)
	client, err := getClient()
	require.NoError(t, err)
	require.NotNil(t, client)
}
