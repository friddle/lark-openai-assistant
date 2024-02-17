# 此工具目标
本工具作为SRE内部提供文档服务的工具。基于的是OpenAi-Assistant。内部打通文档上传功能。
目标很明确。就是替我回答文档问题。

# 工具流程
0. 用户进行登录
1. 用户的回答看看GPT是否能直接回答
2. 用户的输入进入OpenAI进行翻译(搜索)
3. 生成相应的关键词。调用飞书的API进行搜索(function calls)
4. 把结果文件都丢给OpenAi..文件进行学习。(add files and knowledge)
5. 得出结果返回给用户..

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