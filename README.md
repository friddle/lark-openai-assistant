# 此工具目标
本工具作为SRE内部提供文档服务的工具。基于的是OpenAi-Assistant。内部打通文档上传功能。
目标很明确。就是替我回答文档问题。包括执行部分操作

# 目标
1. 直接添加文档
2. 直接调用某些SRE的固定服务
3. 打包的语意



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