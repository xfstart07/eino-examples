package main

import (
    "context"
    "github.com/cloudwego/eino-examples/quickstart/chatmsg/pkg/models"
    "io"
    "log"
    "strings"

    "github.com/cloudwego/eino/components/model"
    "github.com/cloudwego/eino/components/prompt"
    "github.com/cloudwego/eino/schema"
)

func main() {
    ctx := context.Background()

    // 使用模版创建messages
    log.Printf("===create messages===\n")
    messages := createMessagesFromTemplate()
    log.Printf("messages: %+v\n\n", messages)

    // 创建llm
    log.Printf("===create llm===\n")
    cm := models.CreateOpenAIChatModel(ctx)
    //cm := createOllamaChatModel(ctx)
    log.Printf("create llm success\n\n")

    // log.Printf("===llm generate===\n")
    // result := generate(ctx, cm, messages)
    // log.Printf("result: %+v\n\n", result)

    log.Printf("===llm stream generate===\n")
    streamResult := stream(ctx, cm, messages)
    reportStream(streamResult)
}

func createTemplate() prompt.ChatTemplate {
    // 创建模板，使用 FString 格式
    return prompt.FromMessages(schema.FString,
        // 系统消息模板
        schema.SystemMessage("你是一个{role}。你需要用{style}的语气回答问题。你的目标是帮助程序员解答程序开发问题，提供技术建议。"),

        // 插入需要的对话历史（新对话的话这里不填）
        schema.MessagesPlaceholder("chat_history", true),

        // 用户消息模板
        schema.UserMessage("问题: {question}"),
    )
}

func createMessagesFromTemplate() []*schema.Message {
    template := createTemplate()

    // 使用模板生成消息
    messages, err := template.Format(context.Background(), map[string]any{
        "role":     "程序大师",
        "style":    "专业",
        "question": "介绍下golang的泛型",
        // 对话历史（这个例子里模拟两轮对话历史）
        "chat_history": []*schema.Message{
            schema.UserMessage("你好"),
            schema.AssistantMessage("嘿！我是你的程序员程序大师,记住，每个优秀的程序员都是从 Debug 中成长起来的。有什么我可以帮你的吗？", nil),
        },
    })
    if err != nil {
        log.Fatalf("format template failed: %v\n", err)
    }
    return messages
}

func generate(ctx context.Context, llm model.ChatModel, in []*schema.Message) *schema.Message {
    result, err := llm.Generate(ctx, in)
    if err != nil {
        log.Fatalf("llm generate failed: %v", err)
    }
    return result
}

func stream(ctx context.Context, llm model.ChatModel, in []*schema.Message) *schema.StreamReader[*schema.Message] {
    result, err := llm.Stream(ctx, in)
    if err != nil {
        log.Fatalf("llm generate failed: %v", err)
    }
    return result
}

func reportStream(sr *schema.StreamReader[*schema.Message]) {
    defer sr.Close()

    var sb strings.Builder

    i := 0
    for {
        message, err := sr.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatalf("recv failed: %v", err)
        }
        // log.Printf("message[%d]: %+v\n", i, message)
        sb.WriteString(message.Content)
        i++
    }

    log.Printf("report: %v\n", sb.String())
}
