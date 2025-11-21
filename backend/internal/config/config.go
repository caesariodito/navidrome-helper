package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	Port             string
	DataDir          string
	TempDir          string
	NavidromePath    string
	ConcurrentJobs   int
	EnableDownloads  bool
	DownloadTimeout  time.Duration
	AmazonAPIBaseURL string
}

// Load reads environment variables and returns a Config with defaults applied.
func Load() Config {
	cfg := Config{
		Port:             getEnv("PORT", "8080"),
		DataDir:          getEnv("DATA_DIR", "data"),
		TempDir:          getEnv("TEMP_DIR", "tmp"),
		NavidromePath:    getEnv("NAVIDROME_MUSIC_PATH", "navidrome_music"),
		ConcurrentJobs:   getInt("CONCURRENT_JOBS", 2),
		EnableDownloads:  getBool("ENABLE_DOWNLOADS", false),
		DownloadTimeout:  getDuration("DOWNLOAD_TIMEOUT", 10*time.Minute),
		AmazonAPIBaseURL: getEnv("AMAZON_API_BASE_URL", ""),
	}

	// Ensure key directories exist.
	_ = os.MkdirAll(cfg.DataDir, 0755)
	_ = os.MkdirAll(cfg.TempDir, 0755)
	_ = os.MkdirAll(cfg.NavidromePath, 0755)

	// Normalize directories to absolute paths for clearer logging.
	cfg.DataDir = absOrDefault(cfg.DataDir)
	cfg.TempDir = absOrDefault(cfg.TempDir)
	cfg.NavidromePath = absOrDefault(cfg.NavidromePath)
	return cfg
}

func getEnv(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func getInt(key string, def int) int {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	if n, err := strconv.Atoi(val); err == nil {
		return n
	}
	return def
}

func getBool(key string, def bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	if v, err := strconv.ParseBool(val); err == nil {
		return v
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	if d, err := time.ParseDuration(val); err == nil {
		return d
	}
	return def
}

func absOrDefault(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}
