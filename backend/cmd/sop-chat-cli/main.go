package main

import (
	"log"

	"sop-chat/internal/cli"

	"github.com/joho/godotenv"
)

func main() {
	// 加载当前目录的 .env 文件
	// 如果文件不存在，忽略错误（使用系统环境变量）
	if err := godotenv.Load(); err != nil {
		// .env 文件不存在时，使用系统环境变量，不报错
		log.Println("提示: 未找到 .env 文件，将使用系统环境变量")
	}

	cli.Execute()
}
