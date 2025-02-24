# WeCom LKE Python Demo

这是一个将腾讯云大模型知识引擎 (LKE) 与企业微信集成的演示项目。该项目可以帮助用户通过企业微信与 LKE 上的智能应用进行对话


## 功能特性

- 企业微信消息回调处理
- 通过 SSE 对接腾讯云大模型知识引擎
- 基于 Quart 框架异步处理

## 前置要求

- Python 3.7+ (本demo实际由python3.11版本编写及测试)

## 部署

1. 安装依赖:

```bash
pip install -r requirements.txt
```

2. 在 `app.py` 配置密钥:


```python
WX_TOKEN = '企业微信自建应用接收消息配置【Token】'
WX_ENCODING_AES_KEY = '企业微信自建应用接收消息配置【EncodingAESKey】'
WX_CORP_ID = '企业微信企业信息【企业ID】'
WX_APP_SECRET = '企业微信自建应用信息【Secret】'
TENCENT_CLOUD_LKE_APP_KEY = '腾讯云大模型知识引擎智能应用发布管理配置【AppKey】'
```

## 运行

1. 启动服务:

```bash
# 打印日志到控制台
python app.py
# 打印日志到文件
# python app.py --log_dir /path/to/logs
```

2. 在企业微信自建应用配置页面上配置回调 URL 指向你的服务器:

```
http://your-server-address/
```

3. 在企业微信自建应用中发送消息与你的 LKE 智能应用对话
