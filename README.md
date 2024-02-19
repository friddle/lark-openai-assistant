# 此工具目标
本工具作为对接OpenAi的Assistant的飞书工具
目标很明确。就是飞书可以快速上传文档.并提供高质量的回答

# 限制条件
1. 只支持openai(个人开发者请谅解)
2. 文件转换过程中会明确丢失文件
3. 测试效果就gpt4好用.其他不好用

# 配置流程
1. 在openai中获得sk
2. 在openai中创建assistant的服务并获得sdk
2. 自己搭建proxy或者在cloudflare中申请openai-gateway获得基础地址(国内.可选)
3. 拷贝 .chatgpt.env.sample 为.chatgpt.env 和.feishu.env 为内部配置
4. 启动服务并配置好飞书

# 使用方式为
/auth 授权登陆.因为很多文档必须要权限
/upload {{文档链接}}
/clean 清除当前会话
直接提问


# 快速部署
## 直接运行
```
略
```

```docker
docker run -d -p 8080:8080 --env-file=.feishu.env --env-file=.chatgpt.env friddlecopper/lark-openai-assistant
```

```docker-compose
version: '3'
services:
  lark-openai-assistant:
    image: friddlecopper/lark-openai-assistant
    ports:
      - "8080:8080"
    env_file:
      - .feishu.env
      - .chatgpt.env
    restart: always

```

# 配置
机器人调用地址为 http://ip:port/event  
授权地址为 http://ip:port/auth

## 权限
im:*  建议消息所有权限  
docs:*readonly,sheets:*readonly,drive:*readonly,wiki:*readonly  文档所有只读权限

# 配置
参考相应的  
.feishu.env  
.chatgpt.env 