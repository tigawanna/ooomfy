package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Redis     RedisConfig     `yaml:"redis"`
	S3        S3Config        `yaml:"s3"`
	SMTP      SMTPConfig      `yaml:"smtp"`
	Dashboard DashboardConfig `yaml:"dashboard"`
	Persist   PersistConfig   `yaml:"persist"`
}

type RedisConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type S3Config struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type SMTPConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type DashboardConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type PersistConfig struct {
	Enabled bool               `yaml:"enabled"`
	Dir     string             `yaml:"directory"`
	Redis   RedisPersistConfig `yaml:"redis"`
	S3      S3PersistConfig    `yaml:"s3"`
	SMTP    SMTPPersistConfig  `yaml:"smtp"`
}

type RedisPersistConfig struct {
	Enabled bool   `yaml:"enabled"`
	File    string `yaml:"file"`
}

type S3PersistConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Directory string `yaml:"directory"`
}

type SMTPPersistConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Directory string `yaml:"directory"`
}

var defaultConfig = Config{
	Redis: RedisConfig{
		Port: 6379,
		Host: "127.0.0.1",
	},
	S3: S3Config{
		Port: 9000,
		Host: "127.0.0.1",
	},
	SMTP: SMTPConfig{
		Port: 1025,
		Host: "127.0.0.1",
	},
	Dashboard: DashboardConfig{
		Port: 8080,
		Host: "127.0.0.1",
	},
	Persist: PersistConfig{
		Enabled: true,
		Dir:     ".ooomfs",
		Redis: RedisPersistConfig{
			Enabled: true,
			File:    "redis.dump",
		},
		S3: S3PersistConfig{
			Enabled:   true,
			Directory: "s3-data",
		},
		SMTP: SMTPPersistConfig{
			Enabled:   true,
			Directory: "smtp-emails",
		},
	},
}

func Load(configPath string) (*Config, error) {
	cfg := defaultConfig

	homeDir, _ := os.UserHomeDir()
	configPaths := []string{
		configPath,
		"ooomfs.yaml",
		"ooomfs.yml",
		".ooomfs.yaml",
		".ooomfs.yml",
		filepath.Join(homeDir, ".ooomfs", "config.yaml"),
		filepath.Join(homeDir, ".ooomfs", "config.yml"),
	}

	for _, path := range configPaths {
		if path == "" {
			continue
		}
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config from %s: %w", path, err)
			}
			break
		}
	}

	applyEnvOverrides(&cfg)

	resolvePaths(&cfg, homeDir)

	return &cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if port := getEnv("OOOMFS_REDIS_PORT", "DEVSTACK_REDIS_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Redis.Port)
	}
	if host := getEnv("OOOMFS_REDIS_HOST", "DEVSTACK_REDIS_HOST"); host != "" {
		cfg.Redis.Host = host
	}
	if port := getEnv("OOOMFS_S3_PORT", "DEVSTACK_S3_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.S3.Port)
	}
	if host := getEnv("OOOMFS_S3_HOST", "DEVSTACK_S3_HOST"); host != "" {
		cfg.S3.Host = host
	}
	if port := getEnv("OOOMFS_SMTP_PORT", "DEVSTACK_SMTP_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.SMTP.Port)
	}
	if host := getEnv("OOOMFS_SMTP_HOST", "DEVSTACK_SMTP_HOST"); host != "" {
		cfg.SMTP.Host = host
	}
	if port := getEnv("OOOMFS_DASHBOARD_PORT", "DEVSTACK_DASHBOARD_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Dashboard.Port)
	}
	if host := getEnv("OOOMFS_DASHBOARD_HOST", "DEVSTACK_DASHBOARD_HOST"); host != "" {
		cfg.Dashboard.Host = host
	}
	if dir := getEnv("OOOMFS_PERSIST_DIR", "DEVSTACK_PERSIST_DIR"); dir != "" {
		cfg.Persist.Dir = dir
	}
	if enabled := getEnv("OOOMFS_PERSIST_ENABLED", "DEVSTACK_PERSIST_ENABLED"); enabled != "" {
		cfg.Persist.Enabled = strings.ToLower(enabled) == "true"
	}
}

func getEnv(primary, fallback string) string {
	if val := os.Getenv(primary); val != "" {
		return val
	}
	return os.Getenv(fallback)
}

func resolvePaths(cfg *Config, homeDir string) {
	resolvePath := func(path string) string {
		if strings.HasPrefix(path, "~/") {
			return filepath.Join(homeDir, path[2:])
		}
		return path
	}

	cfg.Persist.Dir = resolvePath(cfg.Persist.Dir)
	cfg.Persist.Redis.File = filepath.Join(cfg.Persist.Dir, cfg.Persist.Redis.File)
	cfg.Persist.S3.Directory = filepath.Join(cfg.Persist.Dir, cfg.Persist.S3.Directory)
	cfg.Persist.SMTP.Directory = filepath.Join(cfg.Persist.Dir, cfg.Persist.SMTP.Directory)
}

func (c *Config) GetPersistDir() string {
	return c.Persist.Dir
}
