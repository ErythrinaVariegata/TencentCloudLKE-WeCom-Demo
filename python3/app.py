from quart import Quart, Response, request, abort
from repo.wecom.callback.WXBizMsgCrypt3 import WXBizMsgCrypt
from repo.wecom.api.CorpApi import CorpApi, CORP_API_TYPE
from repo.tencentlke.sse import SSEClient
import xml.etree.cElementTree as ET
import argparse
import asyncio
import logging
import os
import uuid


WX_TOKEN = ''
WX_ENCODING_AES_KEY = ''
WX_CORP_ID = ''
WX_APP_SECRET = ''
TENCENT_CLOUD_LKE_APP_KEY = ''

app = Quart(__name__)

# 解析脚本入参
parser = argparse.ArgumentParser()
parser.add_argument('--log_dir', help='Directory for log files. If not specified, logs will be printed to stdout')
args = parser.parse_args()

# 日志输出
app.logger.handlers.clear()
app.logger.propagate = False
app.logger.setLevel(logging.INFO)
formatter = logging.Formatter('[%(asctime)s] [%(levelname)s] %(message)s')
if args.log_dir:
    # 确保路径真实存在
    os.makedirs(args.log_dir, exist_ok=True)
    log_file = os.path.join(args.log_dir, 'app.log')
    # 构建日志文件处理
    file_handler = logging.FileHandler(log_file)
    file_handler.setLevel(logging.INFO)
    file_handler.setFormatter(formatter)
    app.logger.addHandler(file_handler)
    app.logger.info(f'Logging to file: {log_file}')
else:
    # 构建标准输出流式处理
    stream_handler = logging.StreamHandler()
    stream_handler.setLevel(logging.INFO)
    stream_handler.setFormatter(formatter)
    app.logger.addHandler(stream_handler)
    app.logger.info('Logging to console')

def go(coro) -> asyncio.Task:
    """
    类似于 Go 中的 go 关键字：
    在事件循环中启动协程任务，并返回对应的 Task 对象。
    注意：只有在事件循环已运行的情况下才能调用 asyncio.create_task().
    """
    return asyncio.create_task(coro)

@app.route('/', methods=['GET', 'POST'])
async def home():
    """
    主函数，主要用于处理企微`验证URL`和`接收消息回调`两个功能
    """
    # 如果URL中不包含参数，返回 404 Not found
    if not request.args:
        app.logger.error(f'No query parameters found, url: {request.url}')
        abort(404)

    # 获取URL参数
    msg_signature = request.args.get('msg_signature', '')
    timestamp = request.args.get('timestamp', '')
    nonce = request.args.get('nonce', '')
    echo_str = request.args.get('echostr', '')

    # 检查必要参数是否设置，如果没有，返回 400 Bad Request
    if msg_signature == '' or timestamp == '' or nonce == '':
        app.logger.error(f'Missing required parameters, msg_signature: {msg_signature}, timestamp: {timestamp}, nonce: {nonce}')
        abort(400)

    # 根据请求方式走不同的处理逻辑
    method = request.method
    if method == 'GET':
        # 企业微信校验URL
        if echo_str == '':
            app.logger.error(f'Missing required parameters, echo_str: {echo_str}')
            abort(400)
        try:
            wx_crypt = WXBizMsgCrypt(WX_TOKEN, WX_ENCODING_AES_KEY, WX_CORP_ID)
            ret, decrypted_echo_str = wx_crypt.VerifyURL(msg_signature, timestamp, nonce, echo_str)
            if ret != 0:
                app.logger.error(f'VerifyURL failed, ret: {ret}')
                abort(400)
            return decrypted_echo_str
        except Exception as e:
            app.logger.error(f'VerifyURL failed, runtime err: {str(e)}')
            abort(500)

    elif method == 'POST':
        # 企业微信接收消息回调，由于耗时较长，需要用协程实现
        content = ''
        user_id = ''
        app_id = ''
        wx_crypt = WXBizMsgCrypt(WX_TOKEN, WX_ENCODING_AES_KEY, WX_CORP_ID)
        data = await request.get_data(as_text=True)
        try:
            ret, decrypted_msg = wx_crypt.DecryptMsg(data, msg_signature, timestamp, nonce)
            if ret != 0:
                app.logger.error(f'DecryptMsg failed, ret: {ret}')
                abort(400)
            app.logger.info(f'DecryptMsg done, msg: {decrypted_msg}')
            xml_tree = ET.fromstring(decrypted_msg)
            content = xml_tree.find("Content").text
            user_id = xml_tree.find("FromUserName").text
            app_id = xml_tree.find("AgentID").text
        except Exception as e:
            app.logger.error(f'DecryptMsg failed, runtime err: {str(e)}')
            abort(500)
        if len(content) == 0:
            app.logger.error(f'Get empty message from {user_id}, quick return')
        go(process_user_input(app_id, user_id, content))
        return ''

async def process_user_input(app_id, user_id, content) -> None:
    """
    把用户消息发送给大模型知识引擎里的智能应用。
    智能应用流式返回响应时，收集一段文字后通过企业微信的接口回复用户。
    """
    lke_sse_client = SSEClient()
    wx_corp_api = CorpApi(WX_CORP_ID, WX_APP_SECRET)
    lke_query = {
        "session_id":         str(uuid.uuid4()),
        "bot_app_key":        TENCENT_CLOUD_LKE_APP_KEY,
        "visitor_biz_id":     user_id,
        "content":            content,
        "streaming_throttle": 1,
    }
    app.logger.info(f'Send query to Tencent LKE, req: {lke_query}')
    try:
        for lke_response in lke_sse_client.send_query(lke_query):
            app.logger.info(f'Send query to Tencent LKE, streaming: {lke_response}')
            if len(lke_response) == 0:
                continue
            wx_query = {
                'touser': user_id,
                'agentid': app_id,
                'msgtype': 'markdown',
                'markdown': {
                    'content': lke_response,
                },
            }
            app.logger.info(f'Send back message to WeComm, req: {wx_query}')
            try:
                wx_response = wx_corp_api.httpCall(CORP_API_TYPE['MESSAGE_SEND'], wx_query)
                app.logger.info(f'Send back message to WeComm done, resp: {wx_response}')
            except Exception as e:
                app.logger.error(f'Send back message to WeComm app failed, err: {str(e)}')
    except Exception as e:
        app.logger.error(f'Send event to LKE app failed, err: {str(e)}')

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=80, debug=True)
