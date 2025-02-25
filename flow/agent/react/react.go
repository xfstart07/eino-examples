/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"

	"github.com/cloudwego/eino-examples/flow/agent/react/tools"
	"github.com/cloudwego/eino-examples/internal/logs"
)

func init() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 未能加载 .env 文件: %v", err)
	}
}

func main() {
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	openAIBaseURL := os.Getenv("OPENAI_BASE_URL")
	openAIModelName := os.Getenv("OPENAI_MODEL_NAME")

	ctx := context.Background()

	// prepare chat model
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: openAIBaseURL, // if using Azure or other models conforming to OpenAI protocol, set this to your model's base URL
		APIKey:  openAIAPIKey,
		Model:   openAIModelName,
		// ByAzure:    true, // if using Azure's OpenAI API, set this to true
		// APIVersion: "2024-06-01",
	})
	if err != nil {
		logs.Errorf("failed to create chat model: %v", err)
		return
	}

	// prepare tools
	restaurantTool := tools.GetRestaurantTool() // 查询餐厅信息的工具
	dishTool := tools.GetDishTool()             // 查询餐厅菜品信息的工具

	// prepare persona (system prompt) (optional)
	persona := `# Character:
你是一个帮助用户推荐餐厅和菜品的助手，根据用户的需要，查询餐厅信息并推荐，查询餐厅的菜品并推荐。
`

	ragent, err := react.NewAgent(ctx, &react.AgentConfig{
		Model: chatModel,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{restaurantTool, dishTool},
		},

		MessageModifier: react.NewPersonaModifier(persona),
	})
	if err != nil {
		logs.Errorf("failed to create agent: %v", err)
		return
	}

	// if you want ping/pong, use Generate
	msg, err := ragent.Generate(ctx, []*schema.Message{
		{
			Role:    schema.User,
			Content: "我在海淀区，给我推荐一些菜，需要有口味辣一点的菜，至少推荐有 2 家餐厅",
		},
	}, agent.WithComposeOptions(compose.WithCallbacks(&LoggerCallback{})))
	if err != nil {
		log.Printf("failed to generate: %v\n", err)
		return
	}
	fmt.Println(msg.String())

	// sr, err := ragent.Stream(ctx, []*schema.Message{
	// 	{
	// 		Role:    schema.User,
	// 		Content: "我在徐汇区，给我推荐一些菜，需要有口味辣一点的菜，至少推荐有 2 家餐厅",
	// 	},
	// }, agent.WithComposeOptions(compose.WithCallbacks(&LoggerCallback{})))
	// if err != nil {
	// 	logs.Errorf("failed to stream: %v", err)
	// 	return
	// }

	// defer sr.Close() // remember to close the stream

	// logs.Infof("\n\n===== start streaming =====\n\n")

	// for {
	// 	msg, err := sr.Recv()
	// 	if err != nil {
	// 		if errors.Is(err, io.EOF) {
	// 			// finish
	// 			break
	// 		}
	// 		// error
	// 		logs.Infof("failed to recv: %v", err)
	// 		return
	// 	}

	// 	// 打字机打印
	// 	logs.Tokenf("%v", msg.Content)
	// }

	logs.Infof("\n\n===== finished =====\n")

}

type LoggerCallback struct {
	callbacks.HandlerBuilder // 可以用 callbacks.HandlerBuilder 来辅助实现 callback
}

func (cb *LoggerCallback) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	fmt.Println("==================")
	inputStr, _ := json.MarshalIndent(input, "", "  ") // nolint: byted_s_returned_err_check
	fmt.Printf("[OnStart] %s\n", string(inputStr))
	return ctx
}

func (cb *LoggerCallback) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	fmt.Println("=========[OnEnd]=========")
	outputStr, _ := json.MarshalIndent(output, "", "  ") // nolint: byted_s_returned_err_check
	fmt.Println(string(outputStr))
	return ctx
}

func (cb *LoggerCallback) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	fmt.Println("=========[OnError]=========")
	fmt.Println(err)
	return ctx
}

func (cb *LoggerCallback) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo,
	output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {

	var graphInfoName = "PregelGraph"

	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("[OnEndStream] panic err:", err)
			}
		}()

		defer output.Close() // remember to close the stream in defer

		fmt.Println("=========[OnEndStream]=========")
		for {
			frame, err := output.Recv()
			if errors.Is(err, io.EOF) {
				// finish
				break
			}
			if err != nil {
				fmt.Printf("internal error: %s\n", err)
				return
			}

			s, err := json.Marshal(frame)
			if err != nil {
				fmt.Printf("internal error: %s\n", err)
				return
			}

			if info.Name == graphInfoName { // 仅打印 graph 的输出, 否则每个 stream 节点的输出都会打印一遍
				fmt.Printf("%s: %s\n", info.Name, string(s))
			}
		}

	}()
	return ctx
}

func (cb *LoggerCallback) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo,
	input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	defer input.Close()
	return ctx
}
