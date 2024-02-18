package llama

import (
	"context"
	"errors"
	"feishu-gpt-search/src/feishu"
	chatgpt "github.com/go-zoox/chatgpt-client"
	"github.com/sashabaranov/go-openai"
	"log"
	"os"
	"time"
)

type FeiShuAssistant struct {
	openAiClient  *openai.Client
	feishuClient  *feishu.FeishuClient
	assistantId   string
	ctx           context.Context
	chatThreadMap map[string]string
}

func (assistant *FeiShuAssistant) UploadFile(url string, args map[string]string) error {
	mdName, mdPath, err := assistant.feishuClient.GetDocumentByUrl(url)
	if err != nil {
		return err
	}
	request := openai.FileRequest{
		FileName: mdName,
		FilePath: mdPath,
		Purpose:  "assistants",
	}
	assistantFile, err := assistant.openAiClient.CreateFile(assistant.ctx, request)
	rsp, err := assistant.openAiClient.CreateAssistantFile(
		assistant.ctx,
		assistant.assistantId,
		openai.AssistantFileRequest{
			FileID: assistantFile.ID,
		},
	)
	log.Printf("rsp: %v", rsp)
	if err != nil {
		return err
	}
	return nil
}

/*
 * Annotation 转回文档
 */
func (assistant *FeiShuAssistant) ExtractMessage(listMessages openai.MessagesList) (string, map[string]string) {
	infoMsg := ""
	for _, message := range listMessages.Messages {
		messageContent := message.Content[0].Text
		infoMsg = infoMsg + "\r" + messageContent.Value
	}
	links := map[string]string{}

	return infoMsg, links
}

func (assistant *FeiShuAssistant) AskQuestion(msgId string, question string, args map[string]string) (string, map[string]string, error) {
	var threadId string
	metadata := map[string]any{}
	if threadId, ok := assistant.chatThreadMap[msgId]; ok {

	} else {
		thread, err := assistant.openAiClient.CreateThread(assistant.ctx, openai.ThreadRequest{
			Metadata: metadata,
		})
		threadId = thread.ID
		if err != nil {
			return "", nil, err
		}
		go func(threadId string, msgId string) {
			time.Sleep(1 * time.Hour)
			assistant.openAiClient.DeleteThread(assistant.ctx, threadId)
			delete(assistant.chatThreadMap, msgId)
		}(threadId, msgId)
		assistant.chatThreadMap[msgId] = threadId
	}
	messageRsp, err := assistant.openAiClient.CreateMessage(assistant.ctx, threadId, openai.MessageRequest{
		Content: question,
		//FileIds: 应该不需要,
		Role:     "user",
		Metadata: metadata,
	})
	messageId := messageRsp.ID

	runResponse, err := assistant.openAiClient.CreateRun(assistant.ctx, threadId, openai.RunRequest{
		AssistantID: assistant.assistantId,
	})
	if err != nil {
		return "", nil, err
	}
	responseId := runResponse.ID
	loopTime := 0
	for {
		loopTime = loopTime + 1
		if loopTime >= 10 {
			return "", nil, errors.New("超时")
		}
		time.Sleep(1 * time.Second)
		response, err := assistant.openAiClient.RetrieveRun(assistant.ctx, threadId, responseId)
		if err != nil {
			return "", nil, err
		}
		if response.Status == openai.RunStatusQueued || response.Status == openai.RunStatusInProgress {
			continue
		} else if response.Status == openai.RunStatusRequiresAction {
			log.Printf("required action unsupported")
		} else if response.Status == openai.RunStatusCompleted {
			order := "asc"
			messageLists, err := assistant.openAiClient.ListMessage(assistant.ctx, threadId, nil, &order, &messageId, nil)
			if err != nil {
				return "", nil, err
			}
			messageResponse, links := assistant.ExtractMessage(messageLists)
			return messageResponse, links, nil

		} else {
			return "", nil, errors.New(response.LastError.Message)
		}

	}
	//
	return "", nil, errors.New("不应该触发的情况")
}

func NewFeishuAssistant(config *chatgpt.Config, feishuClient *feishu.FeishuClient) AssistantClient {
	openAiConfig := openai.DefaultConfig(config.APIKey)
	openAiConfig.BaseURL = config.APIServer
	client := openai.NewClientWithConfig(openAiConfig)
	return &FeiShuAssistant{
		openAiClient: client,
		feishuClient: feishuClient,
		ctx:          feishuClient.Ctx,
		assistantId:  os.Getenv("CHATGPT_ASSISTANT_ID"),
	}
}
