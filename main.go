package main

import (
	"GopherAI/common/mysql"
	"GopherAI/common/redis"
	"GopherAI/config"
	"GopherAI/router"
	"fmt"
	"log"
)

func main() {
	
	if err := config.InitConfig(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	
	if err := mysql.InitMysql(); err != nil {
		log.Fatalf("初始化MySQL失败: %v", err)
	}

	redis.Init()

	r := router.InitRouter()

	conf := config.GetConfig()
	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	fmt.Printf("服务器正在%s启动...\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
