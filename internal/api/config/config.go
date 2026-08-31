package config

import (
	"time"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port         string `env:"PORT" envDefault:"8080"`
	FrontendURL  string `env:"FRONTEND_URL,required"`
	IsProduction bool   `env:"IS_PRODUCTION" envDefault:"false"`

	// Database
	DBConn string `env:"DB_URL,required"`

	// Redis
	RedisURL      string `env:"REDIS_URL" envDefault:""`
	RedisURI      string `env:"REDIS_URI" envDefault:""`
	RedisHost     string `env:"REDIS_HOST" envDefault:"localhost"`
	RedisPort     string `env:"REDIS_PORT" envDefault:"6379"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB       int    `env:"REDIS_DB" envDefault:"0"`

	// JWT & Cookies
	JWTAccessSecret  string        `env:"JWT_ACCESS_SECRET,required"`
	JWTRefreshSecret string        `env:"JWT_REFRESH_SECRET,required"`
	JWTAccessTTL     time.Duration `env:"JWT_ACCESS_TTL" envDefault:"15m"`
	JWTRefreshTTL    time.Duration `env:"JWT_REFRESH_TTL" envDefault:"168h"`
	CookieDomain     string        `env:"COOKIE_DOMAIN" envDefault:""`

	// Mailer
	MailHost     string `env:"SMTP_HOST,required"`
	MailPort     int    `env:"SMTP_PORT" envDefault:"587"`
	MailUsername string `env:"SMTP_USER,required"`
	MailPassword string `env:"SMTP_PASSWORD,required"`
	MailFrom     string `env:"SMTP_FROM,required"`
}

func NewConfig() (*Config, error) {
	_ = godotenv.Load()

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
