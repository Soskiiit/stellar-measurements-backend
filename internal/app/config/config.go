package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type JWTConfig struct {
	Token     string
	ExpiresIn time.Duration
}

type RedisConfig struct {
	Host        string
	Password    string
	Port        int
	User        string
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

type Config struct {
	ServiceHost string
	ServicePort int
	JWT         JWTConfig
	Redis       RedisConfig
}

const (
	envRedisHost = "REDIS_HOST"
	envRedisPort = "REDIS_PORT"
	envRedisUser = "REDIS_USER"
	envRedisPass = "REDIS_PASSWORD"
	envJWTSecret = "JWT_SECRET"
	envJWTExpire = "JWT_EXPIRES_IN"
)

func NewConfig() (*Config, error) {
	var err error

	configName := "config"
	_ = godotenv.Load()
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	if h := os.Getenv(envRedisHost); h != "" {
		cfg.Redis.Host = h
	}
	if p := os.Getenv(envRedisPort); p != "" {
		if port, errPort := strconv.Atoi(p); errPort == nil {
			cfg.Redis.Port = port
		}
	}
	if cfg.Redis.Port == 0 {
		cfg.Redis.Port = 6379
	}
	if u := os.Getenv(envRedisUser); u != "" {
		cfg.Redis.User = u
	}
	if pass := os.Getenv(envRedisPass); pass != "" {
		cfg.Redis.Password = pass
	}

	if secret := os.Getenv(envJWTSecret); secret != "" {
		cfg.JWT.Token = secret
	}
	if cfg.JWT.Token == "" {
		cfg.JWT.Token = "stellar-secret-token-key-2026"
	}
	if exp := os.Getenv(envJWTExpire); exp != "" {
		if d, errExp := time.ParseDuration(exp); errExp == nil {
			cfg.JWT.ExpiresIn = d
		}
	}
	if cfg.JWT.ExpiresIn == 0 {
		cfg.JWT.ExpiresIn = 24 * time.Hour
	}

	if cfg.Redis.DialTimeout == 0 {
		cfg.Redis.DialTimeout = 10 * time.Second
	}
	if cfg.Redis.ReadTimeout == 0 {
		cfg.Redis.ReadTimeout = 10 * time.Second
	}

	log.Info("config parsed")

	return cfg, nil
}
