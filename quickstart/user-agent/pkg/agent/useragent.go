package agent

import (
    "context"
    "github.com/cloudwego/eino-examples/quickstart/user-agent/pkg/tool"
    "github.com/cloudwego/eino/components/model"
    "github.com/cloudwego/eino/compose"
    "github.com/cloudwego/eino/schema"
    "log"
)

type UserAgent struct {
    runnable compose.Runnable[[]*schema.Message, []*schema.Message]
}

func NewUserAgent(chatModel model.ChatModel) *UserAgent {
    ctx := context.Background()

    tools, toolInfos := tool.UserTools(ctx)
    err := chatModel.BindTools(toolInfos)
    if err != nil {
        log.Fatal("bind tools failed: ", err)
        return nil
    }

    // 创建 tools 节点
    todoToolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{Tools: tools})
    if err != nil {
        log.Fatal("create tools node failed: ", err)
        return nil
    }

    // 构建完整的处理链
    chain := compose.NewChain[[]*schema.Message, []*schema.Message]()
    chain.
        AppendChatModel(chatModel, compose.WithNodeName("chat_model")).
        AppendToolsNode(todoToolsNode, compose.WithNodeName("tools"))

    // 编译并运行 chain
    agent, err := chain.Compile(ctx)
    if err != nil {
        log.Fatal(err)
    }

    return &UserAgent{runnable: agent}
}

func (ua *UserAgent) Invoke(ctx context.Context, content string) {
    output, err := ua.runnable.Invoke(ctx, []*schema.Message{createSchemeMessage(content)})
    if err != nil {
        log.Fatal("run failed: ", err)
        return
    }
    log.Printf("output: %+v\n", output)
}

func createSchemeMessage(content string) *schema.Message {
    return &schema.Message{
        Role:    schema.System,
        Content: content,
    }
}
