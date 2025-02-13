package models

import (
    "context"
    "github.com/cloudwego/eino-ext/components/model/ollama"
    "github.com/cloudwego/eino/components/model"
    "log"
)

func CreateOllamaChatModel(ctx context.Context) model.ChatModel {
    chatModel, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
        BaseURL: "http://localhost:11434", // Ollama 服务地址
        Model:   "deepseek-r1:1.5b",       // 模型名称
    })
    if err != nil {
        log.Fatalf("create ollama chat model failed: %v", err)
    }
    return chatModel
}
