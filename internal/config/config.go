package config

import "github.com/spf13/viper"

type HTTPServerConfig struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
}

type PostgresConfig struct {
    DSN string `mapstructure:"dsn"`
    AutoMigrate bool   `mapstructure:"auto_migrate"`
}

type LoggingConfig struct {
    Level string `mapstructure:"level"`
}

type Config struct {
    HTTPServer HTTPServerConfig `mapstructure:"http_server"`
    Postgres   PostgresConfig   `mapstructure:"postgres"`
    Logging    LoggingConfig    `mapstructure:"logging"`
}

func Load() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.AutomaticEnv() 

    viper.SetDefault("http_server.host", "0.0.0.0")
    viper.SetDefault("http_server.port", 8080)
    viper.SetDefault("logging.level", "info")
    viper.SetDefault("postgres.auto_migrate", false)

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
