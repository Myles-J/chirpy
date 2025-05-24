package utils

import (
	"os"

	"github.com/Myles-J/chirpy/internal/logger"
)

func MustGetenv(key string) string {
	logger := logger.NewLogger()
	val := os.Getenv(key)
	if val == "" {
		logger.Error("Environment variable is not set", "key", key)
		os.Exit(1)
	}
	return val
}
