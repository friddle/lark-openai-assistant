package llama

type AssistantClient interface {
	UploadFile(url string, args map[string]string) error
	AskQuestion(msgId string, question string, args map[string]string) (string, map[string]string, error)
}
