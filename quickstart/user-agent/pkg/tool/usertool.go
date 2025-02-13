package tool

import (
    "context"
    "encoding/json"
    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/components/tool/utils"
    "github.com/cloudwego/eino/schema"
    "log"
)

type UserInfo struct {
    Username   string `json:"username" jsonschema:"description=username of the employee"`
    Age        int    `json:"age" jsonschema:"description=age of the employee"`
    Department string `json:"department" jsonschema:"description=department of the employee"`
}

func UserTools(ctx context.Context) ([]tool.BaseTool, []*schema.ToolInfo) {
    addTool, _ := utils.InferTool("addUser", "add a user", AddUser)
    getTool, _ := utils.InferTool("getUser", "get a user", GetUser)
    listTool, _ := utils.InferTool("listUsers", "list all users", ListUsers)
    deleteTool, _ := utils.InferTool("deleteUser", "delete a user", DeleteUser)

    tools := []tool.BaseTool{addTool, getTool, listTool, deleteTool}

    var toolInfos []*schema.ToolInfo
    for _, t := range tools {
        info, err := t.Info(ctx)
        if err != nil {
            log.Fatalf("get tool info failed: %v", err)
        }
        toolInfos = append(toolInfos, info)
    }

    return tools, toolInfos
}

var (
    userDb = map[string]UserInfo{
        "leon": {
            Username:   "leon",
            Age:        34,
            Department: "软件开发",
        },
    }
)

func AddUser(ctx context.Context, params *UserInfo) (string, error) {
    userDb[params.Username] = *params
    log.Printf("%s added, 年龄: %d, 部门: %s", params.Username, params.Age, params.Department)
    return "ok", nil
}

func GetUser(ctx context.Context, params *UserInfo) (string, error) {
    user, ok := userDb[params.Username]
    if !ok {
        return "", nil
    }
    log.Printf("查询到用户: %s, 年龄: %d, 部门: %s", user.Username, user.Age, user.Department)
    userString, _ := json.Marshal(user)
    return string(userString), nil
}

func ListUsers(ctx context.Context, _ *UserInfo) (string, error) {
    var users []UserInfo
    for _, user := range userDb {
        users = append(users, user)
    }
    log.Printf("查询到用户: %v", users)
    userString, _ := json.Marshal(users)
    return string(userString), nil
}

func DeleteUser(ctx context.Context, params *UserInfo) (string, error) {
    log.Printf("删除用户: %s", params.Username)
    delete(userDb, params.Username)
    log.Printf("删除用户: %s", params.Username)
    return "ok", nil
}
