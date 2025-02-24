# LKE WeCom Demo

这是一个将腾讯云 LLM Knowledge Engine (LKE) 与企业微信集成的演示项目。该项目可以帮助用户通过企业微信与 LKE 上的智能应用进行对话

## 功能特性

- 企业微信消息回调处理
- 自动刷新企业微信 access token
- 支持多平台构建（Linux、Windows、macOS）
- 支持多种CPU架构（amd64、arm64）

## 配置参数

项目支持通过命令行参数或环境变量进行配置：

### 环境变量

```bash
WX_TOKEN # 企业微信自建应用接收消息配置【Token】
WX_ENCODING_AES_KEY # 企业微信自建应用接收消息配置【EncodingAESKey】
WX_CORP_ID # 企业微信企业信息【企业ID】
WX_APP_SECRET # 企业微信自建应用信息【Secret】
TENCENT_CLOUD_LKE_APP_KEY # 腾讯云大模型知识引擎智能应用发布管理配置【AppKey】
```

### 命令行参数

```bash
-wx_token string 企业微信自建应用接收消息配置【Token】
-wx_encodingaeskey string 企业微信自建应用接收消息配置【EncodingAESKey】
-wx_corpid string 企业微信企业信息【企业ID】
-wx_appsecret string 企业微信自建应用信息【Secret】
-lke_appkey string 腾讯云大模型知识引擎智能应用发布管理配置【AppKey】
```

## 构建说明

项目使用 Makefile 管理构建流程，支持以下命令：

```bash
#构建当前平台的二进制文件
make build
#构建所有平台的二进制文件
make build-all
#构建特定平台的二进制文件
make build-custom OS=linux ARCH=amd64
#运行测试
make test
#清理构建产物
make clean
```

构建后的二进制文件将保存在 `build` 目录下。

## 运行

1. 构建项目：

```bash
# make build-custom OS=linux ARCH=amd64
make build
```

2. 运行服务：
以 `lke-wecom-demo-linux-amd64` 为例，运行在x86架构下的Linux系统上。

```bash
./build/lke-wecom-demo-linux-amd64 -wx_token=xxx -wx_encodingaeskey=xxx -wx_corpid=xxx -wx_appsecret=xxx -lke_appkey=xxx
```
或使用环境变量
```bash
export WX_TOKEN=xxx
export WX_ENCODING_AES_KEY=xxx
export WX_CORP_ID=xxx
export WX_APP_SECRET=xxx
export TENCENT_CLOUD_LKE_APP_KEY=xxx
./build/lke-wecom-demo-linux-amd64
```

服务将在 80 端口启动，并开始监听企业微信的回调请求。

## 注意事项

- 确保服务器的 80 端口可访问
- 妥善保管各项密钥信息
- 建议在生产环境使用 HTTPS
