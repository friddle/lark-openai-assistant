package llama

import (
	"context"
	"feishu-gpt-search/src/config"
	"feishu-gpt-search/src/feishu"
	"fmt"
	"testing"
)

func getClient() AssistantClient {
	ctx := context.Background()
	feishuConf := config.ReadFeishuConfig()
	feishuApiClient := feishu.NewFeishuClient(ctx, feishuConf)
	println(fmt.Sprintf("info:%v", feishuConf))
	assistant := NewFeishuAssistant(config.ReadChatGptClient(), feishuApiClient)
	return assistant
}

func TestUpload(t *testing.T) {
	client := getClient()
	err := client.UploadFile("https://laiye-tech.feishu.cn/wiki/HZs4wZMsniQLiokuM6hcdJ6anae?from=from_lark_index_search&theme=DARK&contentTheme=DARK", map[string]string{})
	if err != nil {
		t.Error(err)
	}
}

func TestAskQuestion(t *testing.T) {
	client := getClient()
	_, _, err := client.AskQuestion("1000", "IDP大概怎么部署?", map[string]string{})
	if err != nil {
		t.Error(err)
	}
}
