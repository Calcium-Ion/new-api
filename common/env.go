package common

import (
	"fmt"
	"os"
	"strconv"
)

func GetEnvOrDefault(env string, defaultValue int) int {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	num, err := strconv.Atoi(os.Getenv(env))
	if err != nil {
		SysError(fmt.Sprintf("failed to parse %s: %s, using default value: %d", env, err.Error(), defaultValue))
		return defaultValue
	}
	return num
}

func GetEnvOrDefaultString(env string, defaultValue string) string {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	return os.Getenv(env)
}
