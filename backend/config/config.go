package config

import (
	"backend/internal/db/influxdb"
	"backend/internal/db/minio"
	"backend/internal/db/mysql"
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)


type Config struct{
	InfluxDB influxdb.InfluxDBConfig `yaml:"influxdb"`
	MinIO    minio.MinIOConfig    `yaml:"minio"`
	Mysql    mysql.MysqlCOnfig    `yaml:"mysql"`
}

func InitConfig(path string) (*Config, error) {
	v := viper.New()

	if path == "" {
		path = "./config/dev.yaml"
	}
	v.SetConfigFile(path) // 绝对/相对路径都可
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config failed: %v", err)
	}

	var cfg Config
	// 让 viper 用 yaml 标签解码（否则用 mapstructure 标签）
	if err := v.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "yaml"
	}); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %v", err)	
	}
	return &cfg, nil
}