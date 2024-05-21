package smpp

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	SMSCHost string
	SMSCPort int
	SystemID string
	Password string
}

func LoadConfig() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		return Config{}, err
	}

	smscHost, ok := os.LookupEnv("SMSC_HOST")
	if !ok {
		return Config{}, fmt.Errorf("SMSC_HOST not found in environment variables")
	}

	smscPort, err := getEnvInt("SMSC_PORT")
	if err != nil {
		return Config{}, err
	}

	systemID, ok := os.LookupEnv("SYSTEM_ID")
	if !ok {
		return Config{}, fmt.Errorf("SYSTEM_ID not found in environment variables")
	}

	password, ok := os.LookupEnv("PASSWORD")
	if !ok {
		return Config{}, fmt.Errorf("PASSWORD not found in environment variables")
	}

	return Config{
		SMSCHost: smscHost,
		SMSCPort: smscPort,
		SystemID: systemID,
		Password: password,
	}, nil
}

func getEnvInt(key string) (int, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return 0, fmt.Errorf("%s not found in environment variables", key)
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("failed to convert %s to int: %v", key, err)
	}

	return intValue, nil
}
