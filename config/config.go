package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	PublicHost              string
	Port                    string
	DBPort                  string
	DBUser                  string
	DBHost                  string
	DBPassword              string
	DBAddress               string
	DBName                  string
	DBsslMode               string
	CookiesAuthSecret       string
	CookiesAuthAgeInSeconds int
	CookiesAuthIsSecure     bool
	CookiesAuthIsHttpOnly   bool
	GoogleKey               string
	GoogleSecret            string
	DiscordKey              string
	DiscordSecret           string
	GithubKey               string
	GithubSecret            string
}

const (
	twoDaysInSeconds = 60 * 60 * 24 * 2
)

var Envs = initConfig()

func initConfig() Config {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file")
	}
	return Config{
		PublicHost: getEnv("HOST", "http://localhost"),
		Port:       getEnv("PORT", "8000"),

		DBPort:     getEnv("DB_PORT", "5432"),
		DBHost:     getEnv("DB_HOST", "host"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", "mypassword"),
		DBAddress:  fmt.Sprintf("%s:%s", getEnv("DB_HOST", "127.0.0.1"), getEnv("DB_PORT", "5432")),
		DBName:     getEnv("DB_NAME", "jobsList"),
		DBsslMode:  getEnv("DB_SSL_MODE", "require"),

		CookiesAuthSecret:       getEnv("COOKIES_AUTH_SECRET", "secret_cookie"),
		CookiesAuthAgeInSeconds: getEnvAsInt("COOKIES_AUTH_AGE", twoDaysInSeconds),
		CookiesAuthIsSecure:     getEnvAsBool("COOKIES_AUTH_IS_SECURE", true),
		CookiesAuthIsHttpOnly:   getEnvAsBool("COOKIES_AUTH_IS_HTTP_ONLY", true),
		GoogleKey:               getEnvOrError("GOOGLE_KEY"),
		GoogleSecret:            getEnvOrError("GOOGLE_SECRET"),
		DiscordKey:              getEnvOrError("DISCORD_KEY"),
		DiscordSecret:           getEnvOrError("DISCORD_SECRET"),
		GithubKey:               getEnvOrError("GITHUB_KEY"),
		GithubSecret:            getEnvOrError("GITHUB_SECRET"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) > 0 {
		return value
	}

	return fallback
}

func getEnvOrError(key string) string {
	value := os.Getenv(key)
	if len(value) > 0 {
		return value
	}

	panic(fmt.Sprintf("Environment variable %s is not set", key))
}

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	i, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return i
}

func getEnvAsBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	b, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return b
}
