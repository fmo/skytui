package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

const (
	defaultFocusDuration      = 25 * time.Minute
	defaultShortBreakDuration = 5 * time.Minute
)

type Config struct {
	viper *viper.Viper
	path  string
}

func New(appDir string) (*Config, error) {
	config := viper.New()
	config.SetConfigName("config")
	config.SetConfigType("yaml")
	config.AddConfigPath(appDir)

	config.SetDefault("default-duration", defaultFocusDuration.String())
	config.SetDefault("short-break-duration", defaultShortBreakDuration.String())
	config.SetDefault("active-project-id", "")
	config.SetDefault("notifications-enabled", true)
	configPath := filepath.Join(appDir, "config.yaml")

	if err := config.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("read config: %w", err)
		}

		if err := config.SafeWriteConfigAs(configPath); err != nil {
			return nil, fmt.Errorf("create config: %w", err)
		}
	}

	return &Config{viper: config, path: configPath}, nil
}

func (c *Config) LoadActiveProjectID() string {
	return c.viper.GetString("active-project-id")
}

func (c *Config) LoadNotificationsEnabled() bool {
	return c.viper.GetBool("notifications-enabled")
}

func (c *Config) SaveActiveProjectID(id string) error {
	c.viper.Set("active-project-id", id)
	if err := c.viper.WriteConfigAs(c.path); err != nil {
		return fmt.Errorf("save active-project-id: %w", err)
	}

	return nil
}

func (c *Config) LoadDefaultFocusDuration() (time.Duration, error) {
	duration, err := time.ParseDuration(c.viper.GetString("default-duration"))
	if err != nil {
		return 0, fmt.Errorf("parse default-duration: %w", err)
	}
	if duration < time.Second || duration%time.Second != 0 {
		return 0, fmt.Errorf("default-duration should be at least 1 second and use whole seconds: %v", duration)
	}

	return duration, nil
}

func (c *Config) LoadShortBreakDuration() (time.Duration, error) {
	duration, err := time.ParseDuration(c.viper.GetString("short-break-duration"))
	if err != nil {
		return 0, fmt.Errorf("parse short-break-duration: %w", err)
	}

	if duration < time.Second || duration%time.Second != 0 {
		return 0, fmt.Errorf("short break duration should be at least 1 second and use whole seconds: %v", duration)
	}

	return duration, nil
}
