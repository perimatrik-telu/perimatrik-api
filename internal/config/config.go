// Package config loads application settings from the environment (and an
// optional .env file). Variable names match the Laravel app so the same .env
// values can be reused.
package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	TLS      string // auto, true, or false
}

type CloudinaryConfig struct {
	CloudName string
	APIKey    string
	APISecret string
	Folder    string
}

type Config struct {
	AppEnv      string
	AppDebug    bool
	AppURL      string // public base URL of this API, used for stored file URLs
	Port        string
	Location    *time.Location
	FrontendURL string

	DB DBConfig

	// StoragePath is the root of the public disk (Laravel: storage/app/public).
	StoragePath   string
	StorageDriver string
	Cloudinary    CloudinaryConfig

	BcryptCost  int
	TokenPrefix string // SANCTUM_TOKEN_PREFIX

	CORSOrigins        []string
	CORSOriginPatterns []*regexp.Regexp
}

var defaultCORSOrigins = []string{
	"http://localhost:3000",
	"http://localhost:5173",
	"http://news.localhost:5173",
	"http://dashboard.localhost:5173",
	"https://perimatrik.org",
	"https://news.perimatrik.org",
	"https://dashboard.perimatrik.org",
	"https://beta.perimatrik.org",
	"https://perimatrik.vercel.app",
}

var defaultCORSPatterns = []string{
	`^https?://.*\.perimatrik\.org$`,
	`^https?://.*\.localhost:5173$`,
}

func Load() (*Config, error) {
	_ = godotenv.Load() // .env is optional; real environment variables win

	tz := env("APP_TIMEZONE", "UTC")
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("invalid APP_TIMEZONE %q: %w", tz, err)
	}

	cost, err := strconv.Atoi(env("BCRYPT_ROUNDS", "12"))
	if err != nil || cost < 4 || cost > 31 {
		return nil, fmt.Errorf("invalid BCRYPT_ROUNDS")
	}

	cfg := &Config{
		AppEnv:      env("APP_ENV", "production"),
		AppDebug:    envBool("APP_DEBUG", false),
		AppURL:      strings.TrimRight(env("APP_URL", "http://localhost:8000"), "/"),
		Port:        env("PORT", "8000"),
		Location:    loc,
		FrontendURL: env("FRONTEND_URL", "http://localhost:5173"),
		DB: DBConfig{
			Host:     env("DB_HOST", "127.0.0.1"),
			Port:     env("DB_PORT", "3306"),
			Name:     env("DB_DATABASE", "perimatrik"),
			User:     env("DB_USERNAME", "root"),
			Password: env("DB_PASSWORD", ""),
			TLS:      env("DB_TLS", "auto"),
		},
		StoragePath:   env("STORAGE_PATH", "./storage/app/public"),
		StorageDriver: env("STORAGE_DRIVER", "local"),
		Cloudinary: CloudinaryConfig{
			CloudName: env("CLOUDINARY_CLOUD_NAME", ""),
			APIKey:    env("CLOUDINARY_API_KEY", ""),
			APISecret: env("CLOUDINARY_API_SECRET", ""),
			Folder:    env("CLOUDINARY_FOLDER", "perimatrik"),
		},
		BcryptCost:  cost,
		TokenPrefix: env("SANCTUM_TOKEN_PREFIX", ""),
		CORSOrigins: envList("CORS_ALLOWED_ORIGINS", defaultCORSOrigins),
	}

	for _, p := range envList("CORS_ALLOWED_ORIGIN_PATTERNS", defaultCORSPatterns) {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("invalid CORS pattern %q: %w", p, err)
		}
		cfg.CORSOriginPatterns = append(cfg.CORSOriginPatterns, re)
	}

	return cfg, nil
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func envList(key string, def []string) []string {
	v, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return def
	}
	var out []string
	for _, s := range strings.Split(v, ",") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, strings.TrimRight(s, "/"))
		}
	}
	return out
}
