package config

import "github.com/garyburd/redigo/redis"

type Config struct {
	General struct {
		LogLevel     int    `mapstructure:"log_level"`
		LogPath      string `mapstructure:"log_path"`
		Name         string `mapstructure:"name"`
		SnapshotPath string `mapstructure:"snapshot_path"`
		Addr         string `mapstructure:"addr"`
		Username     string `mapstructure:"username"`
		Password     string `mapstructure:"password"`
	}

	Camera struct {
		RPCServer    string `mapstructure:"rpc_server"`
		MQTTServer   string `mapstructure:"mqtt_server"`
		MQTTUsername string `mapstructure:"mqtt_username"`
		MQTTPassword string `mapstructure:"mqtt_password"`
	} `mapstructure:"camera"`

	File struct {
		URL       string `mapstructure:"url"`
	} `mapstructure:"file_server"`

	Redis struct {
		URL       string `mapstructure:"url"`
		MaxIdle   int    `mapstructure:"max_idle"`
		MaxActive int    `mapstructure:"max_active"`
		Pool      *redis.Pool
	}
}

// C holds the global configuration.
var C Config
