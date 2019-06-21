package utils

import (
	"log"
	"os"
)

func GetRequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Panicf("the %s environment varialbe is required", key)
	}
	return value
}