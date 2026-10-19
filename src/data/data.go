package data

import (
	"context"
	"database/sql"
	"fmt"
	"my-msg/src/config"
	"time"

	gormmysql "gorm.io/driver/mysql"

	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type Data struct {
	db    *gorm.DB
	sqlDB *sql.DB
}

var data *Data

func GetData() *Data {
	return data
}

func (d *Data) GetDB() *gorm.DB {
	return d.db
}

func (d *Data) Close() error {
	return d.sqlDB.Close()
}

func NewData(cfg *config.TomlConfig) (*Data, error) {
	m := cfg.MySQL
	timeout := time.Duration(m.ConnectTimeoutSeconds) * time.Second
	driverCfg := mysqldriver.NewConfig()
	driverCfg.Net = "tcp"
	driverCfg.Addr = fmt.Sprintf("%s:%d", m.Host, m.Port)
	driverCfg.User = m.User
	driverCfg.Passwd = m.Password
	driverCfg.DBName = m.DBName

	//让驱动把 MySQL 的 DATE、DATETIME 等日期时间值解析为 Go 的 time.Time，
	// 方便映射到模型里的时间字段。
	driverCfg.ParseTime = true
	driverCfg.Timeout = timeout // 建立网络连接时的超时时间
	if err := driverCfg.Apply(mysqldriver.Charset("utf8mb4", "")); err != nil {
		return nil, fmt.Errorf("配置 MySQL 字符集失败：%w", err)
	}
	connector, err := mysqldriver.NewConnector(driverCfg)
	if err != nil {
		return nil, fmt.Errorf("创建 MySQL 连接器失败：%w", err)
	}

	// OpenDB 创建连接池，实际连接检查在 PingContext 中完成。
	sqlDB := sql.OpenDB(connector)
	sqlDB.SetMaxOpenConns(m.MaxOpenConns)
	sqlDB.SetMaxIdleConns(m.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(
		time.Duration(m.ConnMaxLifetimeSeconds) * time.Second,
	)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// 让连接池获取连接，必要时建立新连接，然后检查数据库是否可访问。
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("检查 MySQL 连接失败：%w", err)
	}
	db, err := gorm.Open(gormmysql.New(gormmysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DisableAutomaticPing: true,
	})
	if err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("初始化 GORM 失败：%w", err)
	}
	d := &Data{
		db:    db,
		sqlDB: sqlDB,
	}
	data = d
	return d, nil
}
