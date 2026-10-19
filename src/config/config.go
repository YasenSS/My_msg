package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type TomlConfig struct {
	Common CommonConfig `toml:"common"`
	MySQL  MySQLConfig  `toml:"mysql"`
}
type MySQLConfig struct {
	Host                   string `toml:"host"`
	Port                   int    `toml:"port"`
	User                   string `toml:"user"`
	Password               string `toml:"password"`
	DBName                 string `toml:"db_name"`
	MaxOpenConns           int    `toml:"max_open_conns"`
	MaxIdleConns           int    `toml:"max_idle_conns"`
	ConnMaxLifetimeSeconds int    `toml:"conn_max_lifetime_seconds"`
	ConnectTimeoutSeconds  int    `toml:"connect_timeout_seconds"`
}
type CommonConfig struct {
	Port int `toml:"port"`
}

var Conf *TomlConfig

func (c TomlConfig) validate() error {
	if c.Common.Port < 1 || c.Common.Port > 65535 {
		return fmt.Errorf("common.port 必须在 1 ~ 65535 之间")
	}

	m := c.MySQL
	if m.Host == "" || m.User == "" || m.DBName == "" {
		return fmt.Errorf("mysql.host、user 和 db_name 不能为空")
	}
	if m.Port < 1 || m.Port > 65535 {
		return fmt.Errorf("mysql.port 必须在 1~65535 之间")
	}
	if m.MaxOpenConns <= 0 ||
		m.MaxIdleConns < 0 ||
		m.MaxIdleConns > m.MaxOpenConns {
		return fmt.Errorf("MySQL 连接池参数不合法")
	}
	if m.ConnectTimeoutSeconds <= 0 || m.ConnMaxLifetimeSeconds < 0 {
		return fmt.Errorf("MySQL 连接超时必须大于 0,连接寿命不能为负数")
	}
	return nil
}
func Init(path string) error {
	var cfg TomlConfig
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return fmt.Errorf("读取配置失败：%w", err)
	}
	if err := cfg.validate(); err != nil {
		return err
	}
	Conf = &cfg
	return nil
}
