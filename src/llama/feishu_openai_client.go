package llama

import (
	"context"
	"errors"
	"feishu-gpt-search/src/feishu"
	"feishu-gpt-search/src/utils"
	chatgpt "github.com/go-zoox/chatgpt-client"
	"github.com/sashabaranov/go-openai"
	"log"
	"time"
)

type FeiShuAssistant struct {
	openAiClient         *openai.Client
	feishuClient         *feishu.FeishuClient
	assistantId          string
	assistantInstruction string
	assistantModel       string
	ctx                  context.Context
	chatThreadMap        map[string]string
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
	if err != nil {
		return err
	}
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

func (assistant *FeiShuAssistant) CleanMsgId(msgId string) error {
	if threadId, ok := assistant.chatThreadMap[msgId]; ok {
		assistant.openAiClient.DeleteThread(assistant.ctx, threadId)
		delete(assistant.chatThreadMap, msgId)
	} else {

	}
	return nil
}

func (assistant *FeiShuAssistant) CleanAll() error {

	return nil
}

func (assistant *FeiShuAssistant) AskQuestion(msgId string, question string, args map[string]string) (string, map[string]string, error) {
	var threadId string
	metadata := map[string]any{}
	if threadIdd, ok := assistant.chatThreadMap[msgId]; ok {
		threadId = threadIdd
	} else {
		thread, err := assistant.openAiClient.CreateThread(assistant.ctx, openai.ThreadRequest{
			Metadata: metadata,
		})
		threadId = thread.ID
		if err != nil {
			return "", nil, err
		}
		go func(threadId string, msgId string) {
			time.Sleep(5 * 24 * time.Hour)
			assistant.openAiClient.DeleteThread(assistant.ctx, threadId)
			delete(assistant.chatThreadMap, msgId)
		}(threadId, msgId)
		assistant.chatThreadMap[msgId] = threadId
	}
	//启动message
	messageRsp, err := assistant.openAiClient.CreateMessage(assistant.ctx, threadId, openai.MessageRequest{
		Content: question,
		Role:    "user",
	})
	if err != nil {
		return "", nil, err
	}
	messageId := messageRsp.ID

	//启动run
	runResponse, err := assistant.openAiClient.CreateRun(assistant.ctx, threadId, openai.RunRequest{
		AssistantID: assistant.assistantId,
		Model:       "gpt-4-turbo-preview",
		Instructions: "You are a programmer. " +
			"Your goal is to accurately guide customers step by step to install various laiye services and help them solve problems encountered during private deployment." +
			"All answers to questions are found in the documentation. If not found in the documentation, reply directly with \"我不知道\"." +
			"last use chinese",
	})
	if err != nil {
		return "", nil, err
	}
	responseId := runResponse.ID

	loopTime := 0
	for {
		time.Sleep(5 * time.Second)
		loopTime = loopTime + 1
		if loopTime >= 30 {
			return "", nil, errors.New("超时")
		}
		response, err := assistant.openAiClient.RetrieveRun(assistant.ctx, threadId, responseId)
		if err != nil {
			return "", nil, err
		}
		if response.Status == openai.RunStatusQueued || response.Status == openai.RunStatusInProgress {
			continue
		} else if response.Status == openai.RunStatusRequiresAction {
			log.Printf("required action unsupported")
			return "", nil, err
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
	return "", nil, errors.New("不应该触发的情况")
}

func NewFeishuAssistant(config *chatgpt.Config, feishuClient *feishu.FeishuClient) (AssistantClient, error) {
	openAiConfig := openai.DefaultConfig(config.APIKey)
	openAiConfig.BaseURL = config.APIServer
	client := openai.NewClientWithConfig(openAiConfig)

	assistantId := utils.GetOsEnv("CHATGPT_ASSISTANT_ID", "")
	if assistantId == "" {
		return nil, errors.New("assistant id 不能为空")
	}

	return &FeiShuAssistant{
		openAiClient:         client,
		feishuClient:         feishuClient,
		ctx:                  feishuClient.Ctx,
		assistantId:          utils.GetOsEnv("CHATGPT_ASSISTANT_ID", ""),
		assistantInstruction: utils.GetOsEnv("CHATGPT_ASSISTANT_INSTRUCTION", "你是一个智能助手.你需要通过文档的回答客户的各种问题."),
		assistantModel:       utils.GetOsEnv("CHATGPT_ASSISTANT_MODEL", "gpt-4-turbo-preview"),
		chatThreadMap:        make(map[string]string),
	}, nil
}
