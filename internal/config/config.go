package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Site  SiteConfig  `mapstructure:"site"`
	Theme ThemeConfig `mapstructure:"theme"`
	Build BuildConfig `mapstructure:"build"`
}

type SiteConfig struct {
	Title       string `mapstructure:"title"`
	Subtitle    string `mapstructure:"subtitle"`
	Description string `mapstructure:"description"`
	BaseURL     string `mapstructure:"base_url"`
	Author      string `mapstructure:"author"`
}

type ThemeConfig struct {
	Name string `mapstructure:"name"`
	// Mode string `mapstructure:"mode"`
}

type BuildConfig struct {
	ContentDir string `mapstructure:"content"`
	OutputDir  string `mapstructure:"output"`
	Drafts     bool   `mapstructure:"drafts"`
}

func NewConfig() *Config {
	return &Config{}
}

func Setup(config *Config) error {
	viper.SetConfigName("ink")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	err = viper.Unmarshal(config)
	if err != nil {
		return err
	}

	return nil
}
