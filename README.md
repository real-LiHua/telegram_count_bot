# 一个统计 Telegram 群组一天内消息数量的机器人
暂不支持多群组

## 编译时依赖
[go](https://go.dev/dl/)

## 编译
进入到项目目录  

### 默认定时消息格式
```
go build
```

### 自定义定时消息格式
%d 有且仅有一个  
例：
```shell
go build -ldflags="-X 'main.Format=昨日消息数量：%d'"
```

## 使用环境变量进行配置

`TOKEN`：机器人令牌，从 @BotFather 获取  
`CHAT_ID`：群组 ID  
`HTTPS_PROXY`：网络代理配置（可选）  
`TZ`：本地时区配置（可选）  

## 启动
```
./telegram_count_bot
```

## 可用命令
`/last [MessageId]`，该命令逻辑如下  
当带有参数时，`LastID` 被设置为 MessageId  
当不带有参数时，如果是作为消息回复，`LastID` 被设置为被回复消息的消息 ID  
当不带有参数且非作为消息回复时， `LastID` 被设置为当前消息的消息ID
