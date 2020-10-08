package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetConfig(t *testing.T) {
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
}

func TestSaveConfig(t *testing.T) {
	testHome, err := ioutil.TempDir("", "home")
	require.NoError(t, err)
	getHomeDir = func() (string, error) {
		return testHome, nil
	}
	testConfig := config{
		APIAddress: "http://localhost:8080",
		APIToken:   "thisisafaketoken",
	}
	err = saveConfig(testConfig)
	require.NoError(t, err)
	brigadeHome := path.Join(testHome, ".brigade")
	info, err := os.Stat(brigadeHome)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0755), info.Mode().Perm())
	configFile := path.Join(brigadeHome, "config")
	info, err = os.Stat(configFile)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0644), info.Mode().Perm())
	configBytes, err := ioutil.ReadFile(configFile)
	require.NoError(t, err)
	cfg := config{}
	err = json.Unmarshal(configBytes, &cfg)
	require.NoError(t, err)
	require.Equal(t, testConfig, cfg)
}

func TestDeleteConfig(t *testing.T) {
	testHome, err := ioutil.TempDir("", "home")
	require.NoError(t, err)
	getHomeDir = func() (string, error) {
		return testHome, nil
	}
	brigadeHome := path.Join(testHome, ".brigade")
	err = os.Mkdir(brigadeHome, 0755)
	require.NoError(t, err)
	configFile := path.Join(testHome, ".brigade", "config")
	err = ioutil.WriteFile(configFile, []byte("config!"), 0644)
	require.NoError(t, err)
	err = deleteConfig()
	require.NoError(t, err)
}

func TestGetBrigadeHome(t *testing.T) {
	testHome, err := ioutil.TempDir("", "home")
	require.NoError(t, err)
	getHomeDir = func() (string, error) {
		return testHome, nil
	}
	brigadeHome, err := getBrigadeHome()
	require.NoError(t, err)
	require.Equal(t, path.Join(testHome, ".brigade"), brigadeHome)
}

func TestGetConfigPath(t *testing.T) {
	testHome, err := ioutil.TempDir("", "home")
	require.NoError(t, err)
	getHomeDir = func() (string, error) {
		return testHome, nil
	}
	configPath, err := getConfigPath()
	require.NoError(t, err)
	require.Equal(t, path.Join(testHome, ".brigade", "config"), configPath)
}
