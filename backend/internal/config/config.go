package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv             string
	Port               string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPass             string
	DBName             string
	DBSSLMode          string
	RedisURL           string
	JWTSecret          string
	JWTExpiry          time.Duration
	RefreshTokenExpiry time.Duration
	SMSProvider        string
	SMSAPIKey          string
	ZaloAppID          string
	ZaloAppSecret      string
	ZaloZNSTemplateID  string
	ZaloRefreshToken   string
	CloudinaryURL      string
	FirebaseCredPath   string
	MinIOEndpoint      string
	MinIOAccessKey     string
	MinIOSecretKey     string
	MinIOBucket        string
	MinIOUseSSL        bool
	MinIOPublicURL     string
	RateLimitRPS       int
	RateLimitBurst     int
	CORSOrigins        string
	RequestTimeout     time.Duration
	PhoneEncryptKey    string
	CookieDomain       string
	CookieSecure       bool
	SepayAPIKey        string
}

func Load() *Config {
	return &Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		Port:               getEnv("PORT", "8080"),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5435"),
		DBUser:             getEnv("DB_USER", "rice_user"),
		DBPass:             getEnv("DB_PASSWORD", "rice_secret_dev"),
		DBName:             getEnv("DB_NAME", "rice_marketplace"),
		DBSSLMode:          getEnv("DB_SSL_MODE", "disable"),
		RedisURL:           getEnv("REDIS_URL", "redis://:r1c3_r3d1s_s3cur3_d3v@localhost:6381"),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		JWTExpiry:          parseDuration(getEnv("JWT_EXPIRY", "15m")),
		RefreshTokenExpiry: parseDuration(getEnv("REFRESH_TOKEN_EXPIRY", "720h")),
		SMSProvider:        getEnv("SMS_PROVIDER", "mock"),
		SMSAPIKey:          getEnv("SMS_API_KEY", ""),
		ZaloAppID:          getEnv("ZALO_APP_ID", ""),
		ZaloAppSecret:      getEnv("ZALO_APP_SECRET", ""),
		ZaloZNSTemplateID:  getEnv("ZALO_ZNS_TEMPLATE_ID", ""),
		ZaloRefreshToken:   getEnv("ZALO_REFRESH_TOKEN", ""),
		CloudinaryURL:      getEnv("CLOUDINARY_URL", ""),
		FirebaseCredPath:   getEnv("FIREBASE_CRED_PATH", ""),
		MinIOEndpoint:      getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey:     getEnv("MINIO_ACCESS_KEY", "rice_minio"),
		MinIOSecretKey:     getEnv("MINIO_SECRET_KEY", "rice_minio_secret_dev"),
		MinIOBucket:        getEnv("MINIO_BUCKET", "rice-images"),
		MinIOUseSSL:        getEnv("MINIO_USE_SSL", "false") == "true",
		MinIOPublicURL:     getEnv("MINIO_PUBLIC_URL", "http://localhost:9000"),
		RateLimitRPS:       parseInt(getEnv("RATE_LIMIT_RPS", "10")),
		RateLimitBurst:     parseInt(getEnv("RATE_LIMIT_BURST", "20")),
		CORSOrigins:        getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:8080"),
		RequestTimeout:     parseDuration(getEnv("REQUEST_TIMEOUT", "30s")),
		PhoneEncryptKey:    getEnv("PHONE_ENCRYPT_KEY", ""),
		CookieDomain:       getEnv("COOKIE_DOMAIN", ""),
		CookieSecure:       getEnv("COOKIE_SECURE", "false") == "true",
		SepayAPIKey:        getEnv("SEPAY_API_KEY", ""),
	}
}

func (c *Config) Validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if c.PhoneEncryptKey == "" {
		return fmt.Errorf("PHONE_ENCRYPT_KEY is required")
	}
	if len(c.PhoneEncryptKey) != 64 {
		return fmt.Errorf("PHONE_ENCRYPT_KEY must be exactly 64 hex characters (32 bytes)")
	}
	if c.AppEnv == "production" {
		if c.CORSOrigins == "" || c.CORSOrigins == "*" {
			return fmt.Errorf("CORS_ORIGINS must be explicitly set in production (not empty or '*')")
		}
		if c.DBPass == "rice_secret_dev" {
			return fmt.Errorf("DB_PASSWORD must be changed from default in production")
		}
		if c.MinIOSecretKey == "rice_minio_secret_dev" {
			return fmt.Errorf("MINIO_SECRET_KEY must be changed from default in production")
		}
		if strings.Contains(c.RedisURL, "r1c3_r3d1s_s3cur3_d3v") {
			return fmt.Errorf("REDIS_URL must be changed from default in production")
		}
		// DB_SSL_MODE=disable is OK for internal Docker network (no TLS on postgres container)
	}
	return nil
}

func (c *Config) DBDSN() string {
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" user=" + c.DBUser +
		" password=" + c.DBPass +
		" dbname=" + c.DBName +
		" sslmode=" + c.DBSSLMode
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}

func parseInt(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
