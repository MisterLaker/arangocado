package cmd

import (
	"time"

	"github.com/spf13/viper"
)

type Backup struct {
	Name        string
	Host        string
	Port        string
	User        string
	Password    string
	Database    string
	Collections []string
	HistorySize int
	CacheDir    string
}

type S3 struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
	Workers   int
}

type Scheduler struct {
	Backup             `mapstructure:",squash"`
	Schedule           string
	TriggerImmediately bool
}

type Config struct {
	S3            S3
	Backups       []Scheduler
	CheckInterval time.Duration
	CacheDir      string
}

func (c *Config) GetBackup(name string) *Backup {
	for _, b := range c.Backups {
		if b.Backup.Name == "" {
			b.Backup.Name = "arangocado"
		}

		if b.Backup.Name == name {
			return &b.Backup
		}
	}

	return nil
}

func loadConfig(path string) (*Config, error) {
	var conf Config

	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&conf); err != nil {
		return nil, err
	}

	return &conf, nil
}
