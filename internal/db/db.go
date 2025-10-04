package db

import "fmt"

type DatabaseConfig struct {
	Host 			string `yaml:"host"`
	Port			int		 `yaml:"port"`
	User			string `yaml:"user"`
	Password  string `yaml:"password"`
	Database	string `yaml:"database"`
}

func GenerateConnString(config DatabaseConfig) (string) {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", config.User, config.Password, config.Host, config.Port, config.Database)
}
