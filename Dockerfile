FROM  golang:1.20-alpine as builder
WORKDIR src
COPY . .
ENV GOPROXY https://goproxy.cn,direct
ENV GOPRIVATE git.laiye.com
RUN go mod download
RUN go build -o dist/feishu_sre_bot main.go

FROM alpine:3.36
WORKDIR /app/
COPY --from=builder  /src/dist/feishu_sre_bot /app/feishu_sre_bot
VOLUME /app/.feishu.env
VOLUME /app/.chatgpt.env
ENTRYPOINT ["/app/feishu_gpt_search"]
