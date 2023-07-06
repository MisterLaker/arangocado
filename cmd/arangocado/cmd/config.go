package cmd

import (
	"github.com/spf13/viper"
)

type ArangoDB struct {
	Host        string
	Port        string
	User        string
	Password    string
	Database    string
	Collections []string
	Directory   string
}

type S3 struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
	Workers   int
}

type Config struct {
	Arango ArangoDB
	S3     S3
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
