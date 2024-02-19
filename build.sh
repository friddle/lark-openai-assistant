cd $(dirname $0)

go mod tidy
go mod download
go build -o dist/lark-openai-assistant ./main.go
