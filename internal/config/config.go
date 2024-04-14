package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env            string `yaml:"env" env_default:"local"`
	HTTPServer     `yaml:"http_server"`
	PostgresServer `yaml:"postgres_server"`
}

type HTTPServer struct {
	Address                 string        `yaml:"address" env-default:"localhost:8085"`
	ReadTimeout             time.Duration `yaml:"read_timeout" env-default:"5s"`
	WriteTimeout            time.Duration `yaml:"write_timeout" env-default:"5s"`
	IdleTimeout             time.Duration `yaml:"idle_timeout" env-default:"60s"`
	GracefulShutdownTimeout time.Duration `yaml:"graceful_shutdown_timeout" env-default:"10s"`
}

type PostgresServer struct {
	Host         string        `yaml:"host" env-default:"localhost"`
	Port         int           `yaml:"port" env-default:"5432"`
	Username     string        `yaml:"username" env-default:"postgres"`
	DBname       string        `yaml:"db_name" env-default:"postgres"`
	SSLmode      string        `yaml:"ssl_mode" env-default:"disable"`
	MaxOpenConns int           `yaml:"max_open_conns" env-default:"100"`
	MaxIdleConns int           `yaml:"max_idle_conns" env-default:"2"`
	MaxLifetime  time.Duration `yaml:"max_lifetme" env-default:"1h"`
	DriverName   string        `yaml:"driver_name" env-default:"postgres"`
}

type Secret struct {
	PostgresPassword string `env:"DB_PASSWORD" env-required:"true"`
}

func MustLoad() (*Config, *Secret) {
	configPath := fetchConfigPath()
	if configPath == "" {
		log.Fatal("Config path is empty")
	}

	d, _ := os.Getwd()
	fmt.Println(d)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file %s does not exist", configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	scr := &Secret{}
	if err := cleanenv.ReadEnv(scr); err != nil {
		log.Fatalf("failed to get secret env")
	}

	return &cfg, scr
}

func fetchConfigPath() string {
	var configPath, envPath string

	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.StringVar(&envPath, "env", "", "path to env file")
	flag.Parse()

	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Env file %s does not exist", envPath)
	}

	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
	}

	return configPath
}
