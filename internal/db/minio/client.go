package minio

import (
	"backend/pkg/logs"
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOConfig MinIO配置
type MinIOConfig struct {
	Endpoint string `yaml:"endpoint"`
	AccessKeyID string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	UseSSL bool `yaml:"use_ssl"`
	Region string `yaml:"region"`
}

// MinIOClient MinIO客户端结构
type MinIOClient struct {
    Client *minio.Client
}

// GetMinIOClient 获取MinIO客户端
func GetMinIOClient(config MinIOConfig) (*MinIOClient, error) {
    client, err := InitMinIOClient(config)
    if err != nil {
        return nil, err
    }
    return &MinIOClient{Client: client}, nil
}

// InitMinIOClient 初始化MinIO客户端
func InitMinIOClient(config MinIOConfig) (*minio.Client, error) {
    // 创建MinIO客户端
    client, err := minio.New(config.Endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
        Secure: config.UseSSL,
        Region: config.Region,
    })
    if err != nil {
        return nil, fmt.Errorf("创建MinIO客户端失败: %v", err)
    }

    // 测试连接
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err = client.ListBuckets(ctx)
    if err != nil {
        return nil, fmt.Errorf("MinIO连接测试失败: %v", err)
    }

    logs.L().Info("MinIO客户端初始化成功")
    return client, nil
}

