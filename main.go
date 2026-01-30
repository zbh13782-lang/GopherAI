package main

import (
	"GopherAI/common/mysql"
	"GopherAI/common/rabbitmq"
	"GopherAI/common/redis"
	"GopherAI/config"
	"GopherAI/router"
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("没有找到 .env 文件: %v", err)
	}

	if err := config.InitConfig(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	if err := mysql.InitMysql(); err != nil {
		log.Fatalf("初始化MySQL失败: %v", err)
	}

	redis.Init()
	// 在 redis.Init() 之后添加
	rabbitmq.InitRabbitMQ()

	r := router.InitRouter()

	conf := config.GetConfig()
	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	fmt.Printf("服务器正在%s启动...\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
