package mysql

import (
	"backend/pkg/logs"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var MysqlCli *MysqlClient

type MysqlClient struct{
	Client *sql.DB
}

// MysqlCOnfig Mysql配置
type MysqlCOnfig struct{
	Host string `yaml:"host"`
	Port int `yaml:"port"`
	User string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Charset string `yaml:"charset"`
	MaxOpenConns int `yaml:"max_open_conns"`
	MaxIdleConns int `yaml:"max_idle_conns"`
	MaxLifetime int `yaml:"max_lifetime"`
}

// GetMysqlClient 获取Mysql客户端
func GetMysqlClient(config MysqlCOnfig) (*MysqlClient, error) {
	if MysqlCli != nil {
		return MysqlCli, nil
	}
	db, err := InitMysqlClient(config)
	if err != nil {
		return nil, err
	}
	MysqlCli = &MysqlClient{Client: db}
	return MysqlCli, nil
}

// InitMysqlClient 初始化Mysql客户端
func InitMysqlClient(config MysqlCOnfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		config.User, config.Password, config.Host, config.Port, config.Database, config.Charset)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil,fmt.Errorf("failed to open mysql connect: %v",err)
	}
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(config.MaxLifetime) * time.Minute)

	// 测试连接是否成功
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping mysql server: %v", err)
	}
	logs.L().Info("Mysql客户端初始化成功")
	return db, nil
}

// Close 关闭Mysql客户端
func (db *MysqlClient) Close() error {
	if db != nil {
		return db.Client.Close()
	}
	return nil
}
