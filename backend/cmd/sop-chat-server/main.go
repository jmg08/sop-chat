package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"sop-chat/internal/api"
	"sop-chat/internal/client"
	"sop-chat/internal/config"

	"github.com/joho/godotenv"
)

func main() {
	// 解析命令行参数
	var configPath string
	var showHelp bool
	flag.StringVar(&configPath, "config", "", "配置文件路径（默认: config.yaml 或 CONFIG_PATH 环境变量）")
	flag.StringVar(&configPath, "c", "", "配置文件路径（-config 的简写）")
	flag.BoolVar(&showHelp, "help", false, "显示帮助信息")
	flag.BoolVar(&showHelp, "h", false, "显示帮助信息")
	flag.Usage = func() {
		log.Printf("SOP Chat API Server\n\n")
		log.Printf("用法: %s [选项]\n\n", os.Args[0])
		log.Printf("选项:\n")
		flag.PrintDefaults()
		log.Printf("\n示例:\n")
		log.Printf("  %s -config /path/to/config.yaml\n", os.Args[0])
		log.Printf("  %s -c ./config.yaml\n", os.Args[0])
		log.Printf("  %s --config /etc/sop-chat/config.yaml\n", os.Args[0])
		log.Printf("\n注意: 端口配置请在 config.yaml 的 global.port 中设置\n")
	}
	flag.Parse()

	// 显示帮助信息
	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// 首先加载当前目录的 .env 文件
	// 如果文件不存在，忽略错误（使用系统环境变量）
	if err := godotenv.Load(); err != nil {
		log.Println("提示: 未找到 .env 文件，将使用系统环境变量")
	}

	// 确定配置文件路径（优先级: 命令行参数 > 环境变量 > 默认值）
	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
		if configPath == "" {
			configPath = "config.yaml"
		}
	} else {
		// 如果通过命令行指定了配置文件路径，设置到环境变量中
		os.Setenv("CONFIG_PATH", configPath)
		log.Printf("使用配置文件: %s", configPath)
	}

	// 加载统一配置以获取端口配置
	var finalPort int
	unifiedConfig, actualPath, err := config.LoadConfig(configPath)
	if err == nil {
		log.Printf("加载配置文件: %s", actualPath)
		// 从配置文件读取端口
		finalPort = unifiedConfig.GetPort()
	} else {
		log.Printf("警告: 无法加载配置文件 %s: %v，将使用环境变量或默认值", configPath, err)
		// 如果统一配置文件不存在，从环境变量或默认值读取
		portStr := os.Getenv("PORT")
		if portStr != "" {
			if parsedPort, err := strconv.Atoi(portStr); err == nil {
				finalPort = parsedPort
			} else {
				log.Fatalf("环境变量 PORT 包含无效的端口号: %s", portStr)
			}
		} else {
			finalPort = 8080
		}
	}

	// 加载客户端配置
	clientConfig, err := client.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 启动 API 服务器
	server, err := api.NewServer(clientConfig, unifiedConfig)
	if err != nil {
		log.Fatalf("初始化服务器失败: %v", err)
	}
	log.Printf("启动 API 服务器，监听端口 %d", finalPort)
	if err := server.Run(fmt.Sprintf(":%d", finalPort)); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
