package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ServerConfig  `yaml:"http_server"`
	StorageConfig `yaml:"storage"`
}

type ServerConfig struct {
	Addres      string        `yaml:"addres" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"10s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type StorageConfig struct {
	StorageHost     string `yaml:"storage_host" nv-required:"true" env:"PG_HOST_NAME"`
	StoragePort     string `yaml:"storage_port" env-default:"5432" env:"PG_PORT"`
	StorageDB       string `yaml:"storage_db" nv-required:"true" env:"PG_DB"`
	StorageUser     string `yaml:"storage_user" nv-required:"true" env:"PG_USER_NAME"`
	StoragePassword string `yaml:"storage_password" nv-required:"true" env:"PG_USER_PASSWORD"`
}

func LoadConfig() *Config {
	path := os.Getenv("CONFIG_PATH")

	if len(path) == 0 {
		log.Fatal("path is empty CONFIG_PATH not set")
	}

	//check file exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("config file %s is not exist", path)
	}

	//load File
	var result Config

	if err := cleanenv.ReadConfig(path, &result); err != nil {
		log.Fatalf("read file error %s", err)
	}

	return &result

}
