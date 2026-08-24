package config

import (
	"fmt"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		HTTP     HTTPConfig
		Postgres PostgresConfig
		Redis    RedisConfig
		MinIO    MinIOConfig
		JWT      JWTConfig
	}

	HTTPConfig struct {
		Port         string        `env:"HTTP_PORT" env-default:"8080"`
		ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"5s"`
		WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"5s"`
	}

	PostgresConfig struct {
		Host            string        `env:"POSTGRES_HOST" env-default:"localhost"`
		Port            string        `env:"POSTGRES_PORT" env-default:"5432"`
		User            string        `env:"POSTGRES_USER" env-default:"postgres"`
		Password        string        `env:"POSTGRES_PASSWORD" env-default:"postgres"`
		DBName          string        `env:"POSTGRES_DB" env-default:"autodb"`
		SSLMode         string        `env:"POSTGRES_SSLMODE" env-default:"disable"`
		MaxConns        int32         `env:"POSTGRES_MAX_CONNS" env-default:"25"`
		MinConns        int32         `env:"POSTGRES_MIN_CONNS" env-default:"5"`
		MaxConnLifetime time.Duration `env:"POSTGRES_MAX_CONN_LIFETIME" env-default:"1h"`
		MaxConnIdleTime time.Duration `env:"POSTGRES_MAX_CONN_IDLE_TIME" env-default:"30m"`
	}

	RedisConfig struct {
		Host     string `env:"REDIS_HOST" env-default:"localhost"`
		Port     string `env:"REDIS_PORT" env-default:"6379"`
		Password string `env:"REDIS_PASSWORD" env-default:""`
	}

	MinIOConfig struct {
		Endpoint     string `env:"MINIO_ENDPOINT" env-default:"localhost:9000"`
		RootUser     string `env:"MINIO_ROOT_USER" env-default:"minioadmin"`
		RootPassword string `env:"MINIO_ROOT_PASSWORD" env-default:"minioadmin"`
		BucketName   string `env:"MINIO_BUCKET_NAME" env-default:"auto-images"`
		UseSSL       bool   `env:"MINIO_USE_SSL" env-default:"false"`
	}

	JWTConfig struct {
		SecretKey  string        `env:"JWT_SECRET_KEY" env-required:"true"`
		AccessTTL  time.Duration `env:"JWT_ACCESS_TTL" env-default:"15m"`
		RefreshTTL time.Duration `env:"JWT_REFRESH_TTL" env-default:"720h"`
	}
)

var (
	instance *Config
	once     sync.Once
)

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.DBName, p.SSLMode,
	)
}

func GetConfig() (*Config, error) {
	var err error
	once.Do(func() {
		cfg := &Config{}
		if err = cleanenv.ReadConfig(".env", cfg); err != nil {
			err = cleanenv.ReadEnv(cfg)
		}
		if err == nil {
			instance = cfg
		}
	})

	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return instance, nil
}
