package utils

import (
	"os"
	"strings"
)

var privateEnvKeys = map[string]struct{}{
	"MYSQL_PASSWORD":         {},
	"MYSQL_USER_PRIVATE":     {},
	"MYSQL_PASSWORD_PRIVATE": {},
	"REDIS_PASSWORD":         {},
	"QDRANT_API_KEY":         {},
	"RABBITMQ_PASSWORD":      {},
	"THIRD_PARTY_API_KEY":    {},
	"SLIDER_ACCESS_KEY":      {},
}

func GetPublicEnvVars() []string {
	envVars := make([]string, 0, len(os.Environ()))
	for _, item := range os.Environ() {
		key, _, ok := strings.Cut(item, "=")
		if !ok {
			continue
		}
		if _, private := privateEnvKeys[key]; private {
			continue
		}
		envVars = append(envVars, item)
	}
	return envVars
}
