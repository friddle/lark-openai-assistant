FROM  golang:1.20-alpine as builder
WORKDIR src
COPY . .
ENV GOPROXY https://goproxy.cn,direct
ENV GOPRIVATE git.laiye.com
RUN go mod download
RUN go build -o /src/dist/lark_sre_bot main.go

FROM alpine:3.16
WORKDIR /app/
COPY --from=builder  /src/dist/lark_sre_bot /app/lark_sre_bot
VOLUME /app/.feishu.env
VOLUME /app/.chatgpt.env
ENTRYPOINT ["/app/lark_sre_bot"]
