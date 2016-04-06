package config

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Addr           string    `toml:"addr"`
	LogPath        string    `toml:"log_path"`
	LogLevel       string    `toml:"log_level"`
	DatabaseConfig *DBConfig `toml:"storage_db"`
}

type DBConfig struct {
	Host         string `toml:"mysql_host"`
	Port         int    `toml:"mysql_port"`
	User         string `toml:"user"`
	Password     string `toml:"password"`
	DBName       string `toml:"db_name"`
	MaxIdleConns int    `toml:"max_idle_conns"`
}

func ParseConfigFile(fileName string) (*Config, error) {
	var cfg Config

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	_, err = toml.Decode(string(data), &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
