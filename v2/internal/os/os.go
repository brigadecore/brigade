package os

import (
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

func GetEnvVar(name, defaultValue string) string {
	val := os.Getenv(name)
	if val == "" {
		return defaultValue
	}
	return val
}

func GetRequiredEnvVar(name string) (string, error) {
	val := os.Getenv(name)
	if val == "" {
		return "", errors.Errorf(
			"value not found for required environment variable %s",
			name,
		)
	}
	return val, nil
}

func GetIntFromEnvVar(name string, defaultValue int) (int, error) {
	valStr := os.Getenv(name)
	if valStr == "" {
		return defaultValue, nil
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return 0, errors.Errorf(
			"value %q for environment variable %s was not parsable as an int",
			valStr,
			name,
		)
	}
	return val, nil
}

func GetBoolFromEnvVar(name string, defaultValue bool) (bool, error) {
	valStr := os.Getenv(name)
	if valStr == "" {
		return defaultValue, nil
	}
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return false, errors.Errorf(
			"value %q for environment variable %s was not parsable as a bool",
			valStr,
			name,
		)
	}
	return val, nil
}

func GetDurationFromEnvVar(
	name string,
	defaultValue time.Duration,
) (time.Duration, error) {
	valStr := os.Getenv(name)
	if valStr == "" {
		return defaultValue, nil
	}
	val, err := time.ParseDuration(valStr)
	if err != nil {
		return 0, errors.Errorf(
			"value %q for environment variable %s was not parsable as a duration",
			valStr,
			name,
		)
	}
	return val, nil
}
