package influxdb

import (
	"backend/pkg/logs"
	"context"
	"fmt"
	"time"

	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
)

// InfluxDBCli InfluxDB客户端实例
var InfluxDBCli *influxdb3.Client

type InfluxDBClient struct {
    Client *influxdb3.Client
}

// InfluxDBConfig InfluxDB配置结构
type InfluxDBConfig struct {
    Host     string  `yaml:"host"`
    Token    string  `yaml:"token"`
    Database string  `yaml:"database"`
}

// 获取Influxdb实例
func GetInfluxDBClient(config InfluxDBConfig) (*InfluxDBClient, error) {
    if InfluxDBCli != nil {
        return &InfluxDBClient{Client: InfluxDBCli}, nil
    }
    client, err := InitInfluxDBClient(config)
    if err != nil {
        return nil, err
    }
    InfluxDBCli = client
    return &InfluxDBClient{Client: client}, nil
}

// InitInfluxDBClient 初始化InfluxDB客户端
func InitInfluxDBClient(config InfluxDBConfig) (*influxdb3.Client, error) {
    // 创建InfluxDB客户端
    client, err := influxdb3.New(influxdb3.ClientConfig{
        Host:     config.Host,
        Token:    config.Token,
        Database: config.Database,
    })
    if err != nil {
        return nil, fmt.Errorf("创建InfluxDB客户端失败: %v", err)
    }

    // 测试连接
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // 可以通过执行一个简单的查询来测试连接
    query := "SELECT 1"
    _, err = client.Query(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("InfluxDB连接测试失败: %v", err)
    }

    logs.L().Info("InfluxDB客户端初始化成功")
    return client, nil
}

func (c *InfluxDBClient) Close() error {
    if c.Client != nil {
        return c.Client.Close()
    }
    return nil
}