package models

import (
	"context"
	"log"
	"os"

	"github.com/cloudwego/eino-examples/internal/gptr"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

func CreateOpenAIChatModel(ctx context.Context) model.ChatModel {
	url := os.Getenv("OPENAI_BASE_URL")
	key := os.Getenv("OPENAI_API_KEY")
	model := os.Getenv("OPENAI_MODEL")
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     url,
		Model:       model,
		APIKey:      key,
		Temperature: gptr.Of(float32(0.7)),
	})
	if err != nil {
		log.Fatalf("create openai chat model failed, err=%v", err)
	}
	return chatModel
}
