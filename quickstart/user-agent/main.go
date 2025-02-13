package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino-examples/quickstart/chatmsg/pkg/models"
	"github.com/cloudwego/eino-examples/quickstart/user-agent/pkg/agent"
	"github.com/joho/godotenv"
)

func init() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 未能加载 .env 文件: %v", err)
	}
}

func main() {
	ctx := context.Background()

	// 获取环境变量
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY 环境变量未设置")
	}

	model := models.CreateOpenAIChatModel(ctx)

	userAgent := agent.NewUserAgent(model)

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("欢迎使用员工信息 Agent, 支持用户信息的增删改查，输入 'exit' 退出程序。")
	inputTips := "\n请输入操作: "
	for {
		fmt.Print(inputTips)
		if !scanner.Scan() {
			fmt.Println("读取输入失败，程序退出。")
			break
		}

		input := scanner.Text()

		switch strings.ToLower(input) {
		case "exit":
			fmt.Println("欢迎再次使用，再见。")
			return
		default:
			userAgent.Invoke(ctx, strings.Replace(input, inputTips, "", 1))
		}
	}
}
