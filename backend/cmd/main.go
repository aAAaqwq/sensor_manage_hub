package main

import (
	"backend/config"
	"backend/internal/db/influxdb"
	"backend/internal/db/minio"
	"backend/internal/db/mysql"
	"backend/pkg/logs"
	"fmt"
	"log"
)

var AppClient *Client


type Client struct {
	InfluxDB *influxdb.InfluxDBClient
	MinIO    *minio.MinIOClient
	Mysql    *mysql.MysqlClient
}

func InitClient()error{
	// 初始化配置
	cfg, err := config.InitConfig("./config/dev.yaml")
	if err != nil {
		return err
	}
	fmt.Println(cfg)
	// 初始化Client
	AppClient = &Client{}
	// 初始化InfluxDB客户端
	AppClient.InfluxDB, err = influxdb.GetInfluxDBClient(cfg.InfluxDB)
	if err != nil {
		return err
	}
	// 初始化MinIO客户端
	AppClient.MinIO, err = minio.GetMinIOClient(cfg.MinIO)
	if err != nil {
		return err
	}
	// 初始化Mysql客户端
	AppClient.Mysql, err = mysql.GetMysqlClient(cfg.Mysql)
	if err != nil {
		return err
	}
	return nil
}

func CloseClient(){
	// 关闭InfluxDB客户端
	AppClient.InfluxDB.Close()
	// 关闭MinIO客户端,一般自动关闭

	// 关闭Mysql客户端
	AppClient.Mysql.Close()
	logs.L().Info("客户端关闭成功")
}

func main() {
	// 初始化客户端
	if err := InitClient(); err != nil {
		log.Fatalf("初始化客户端失败: %v", err)
	}
	defer CloseClient()

	
}