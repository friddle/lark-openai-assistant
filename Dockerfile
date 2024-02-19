FROM  golang:1.20-alpine as builder
WORKDIR src
COPY . .
ENV GOPROXY https://goproxy.cn,direct
RUN go mod download
RUN go build -o /src/dist/lark_openai_assistant main.go

FROM alpine:3.16
WORKDIR /app/
COPY --from=builder  /src/dist/lark_openai_assistant /app/lark_openai_assistant
VOLUME /app/.feishu.env
VOLUME /app/.chatgpt.env
ENTRYPOINT ["/app/lark_openai_assistant"]
