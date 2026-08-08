package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Log      LogConfig      `mapstructure:"log"`
}

type AuthConfig struct {
	AdminRegistrationKey string `mapstructure:"admin_registration_key"`
}

type ServerConfig struct {
	Address        string   `mapstructure:"address"`
	Mode           string   `mapstructure:"mode"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

type DatabaseConfig struct {
	Driver                string        `mapstructure:"driver"`
	DSN                   string        `mapstructure:"dsn"`
	MaxIdleConnections    int           `mapstructure:"max_idle_connections"`
	MaxOpenConnections    int           `mapstructure:"max_open_connections"`
	ConnectionMaxLifetime time.Duration `mapstructure:"connection_max_lifetime"`
}

type JWTConfig struct {
	Secret    string        `mapstructure:"secret"`
	Issuer    string        `mapstructure:"issuer"`
	ExpiresIn time.Duration `mapstructure:"expires_in"`
}

type RedisConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

func Load(path string) (Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("FORUM")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if cfg.JWT.Secret == "" {
		return Config{}, fmt.Errorf("jwt.secret 不能为空")
	}
	if cfg.JWT.Issuer == "" {
		return Config{}, fmt.Errorf("jwt.issuer 不能为空")
	}
	if cfg.JWT.ExpiresIn <= 0 {
		return Config{}, fmt.Errorf("过期时间必须大于0")
	}
	if cfg.Server.Mode == "release" && (len(cfg.JWT.Secret) < 32 || strings.Contains(cfg.JWT.Secret, "change-me") || strings.Contains(cfg.JWT.Secret, "replace-with")) {
		return Config{}, fmt.Errorf("release mode requires a non-placeholder jwt.secret of at least 32 characters")
	}
	if cfg.Server.Mode == "release" && cfg.Auth.AdminRegistrationKey != "" && (len(cfg.Auth.AdminRegistrationKey) < 32 || strings.Contains(cfg.Auth.AdminRegistrationKey, "change-me") || strings.Contains(cfg.Auth.AdminRegistrationKey, "replace-with")) {
		return Config{}, fmt.Errorf("release mode requires admin_registration_key to be empty or a non-placeholder value of at least 32 characters")
	}
	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.address", "127.0.0.1:8080")
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.allowed_origins", []string{"*"})
	v.SetDefault("auth.admin_registration_key", "")
	v.SetDefault("database.driver", "mysql")
	v.SetDefault("database.dsn", "root:root@tcp(127.0.0.1:3306)/forum?charset=utf8mb4&parseTime=True&loc=Local")
	v.SetDefault("database.max_idle_connections", 5)
	v.SetDefault("database.max_open_connections", 20)
	v.SetDefault("database.connection_max_lifetime", "1h")
	v.SetDefault("jwt.secret", "change-me-in-production")
	v.SetDefault("jwt.issuer", "forum-api")
	v.SetDefault("jwt.expires_in", "2h")
	v.SetDefault("redis.enabled", false)
	v.SetDefault("redis.address", "127.0.0.1:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.file", "")
}
