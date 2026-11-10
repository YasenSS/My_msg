package main

import (
	"flag"
	"fmt"
	"log"
	"my-msg/src/config"
	"my-msg/src/data"
	"my-msg/src/initialize"

	"github.com/gin-gonic/gin"
)

func run(configPath string) error {
	if err := config.Init(configPath); err != nil {
		return fmt.Errorf("初始化配置失败：%w", err)
	}
	d, err := data.NewData(config.Conf)
	if err != nil {
		return fmt.Errorf("初始化数据库失败：%w", err)
	}
	defer func() {
		if err := d.Close(); err != nil {
			log.Printf("关闭数据库连接池失败：%v", err)
		}
	}()
	log.Print("MySQL 初始化成功")

	router := gin.Default()
	initialize.RegisterRouter(router)
	addr := fmt.Sprintf(":%d", config.Conf.Common.Port)
	if err := router.Run(addr); err != nil {
		return fmt.Errorf("启动 HTTP 服务失败：%w", err)
	}
	return nil
}

func main() {
	configPath := flag.String(
		"config",
		"./config/config-test.toml",
		"配置文件路径",
	)
	flag.Parse()

	if err := run(*configPath); err != nil {
		log.Fatal(err)
	}
}
