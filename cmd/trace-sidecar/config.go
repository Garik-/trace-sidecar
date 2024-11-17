package main

import (
	"os"
	"strings"
)

const (
	defaultAddr         = ":8001"
	defaultInternalAddr = ":2223"
	defaultTargetURL    = "http://localhost:8000"
	serviceName         = "sidecar"
)

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// set from ldflags.
var (
	Version = ""
)

type service struct {
	instanceID string
	namespace  string
	name       string
	version    string
}

type config struct {
	addr         string
	internalAddr string
	targetURL    string
	service      service
}

func newConfig() *config {
	prefix := strings.ToUpper(serviceName) + "_"

	return &config{
		addr:         getEnv(prefix+"ADDR", defaultAddr),
		internalAddr: getEnv(prefix+"INTERNAL_ADDR", defaultInternalAddr),
		targetURL:    getEnv(prefix+"TARGET", defaultTargetURL),
		service: service{
			name:       serviceName,
			version:    Version,
			namespace:  getEnv(prefix+"NAMESPACE", "local"),
			instanceID: getEnv(prefix+"INSTANCE_ID", ""),
		},
	}
}
