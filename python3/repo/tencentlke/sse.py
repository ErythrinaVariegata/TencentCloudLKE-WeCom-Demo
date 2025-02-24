import requests
import json
import logging

class SSEClient:
    def __init__(self):
        self.url = 'https://wss.lke.cloud.tencent.com/v1/qbot/chat/sse'
        self.logger = logging.getLogger(__name__)
        self.__content__ = ''
        self.__content_previous_newlines_pos__ = 0
        self.__content_current_newlines_pos__ = 0
        self.__reasoning_content__ = ''
        self.__reasoning_content_snapshot__ = ''
        self.__reasoning_procedure_name__ = ''
        self.__reasoning_elapsed__ = 0.0
        self.__reasoning_elapsed_sent__ = False
        self.__last_procedure_name__ = ''
        self.__references__ = []

    def send_query(self, query):
        # 设置SSE的请求头
        headers = {
            'Accept': 'text/event-stream',
            'Content-Type': 'application/json',
            'Cache-Control': 'no-cache',
        }
        try:
            # 通过requests包发送POST请求，需要开启stream
            response = requests.post(
                self.url,
                data=json.dumps(query),
                headers=headers,
                stream=True
            )
            # 如果请求中出现异常需要报错
            response.raise_for_status()
        except requests.RequestException as e:
            self.logger.error(f'Connect to server failed, url: {self.url}, err: {str(e)}')
        # 从服务器响应中读取并解析事件
        buffer = '' # 事件缓存，用于debug
        event_type = '' # 事件类型，用于校验data和type是否一致
        for line in response.iter_lines():
            if line is None:
                continue
            line = line.decode('utf-8').strip()
            buffer += line + '\n'   # 累积事件缓存，用于debug
            # event 代表事件的类型
            if line.startswith("event:"):
                # 从数据中找到'event:'开头的行并移除，同时去掉多余的空格或换行符
                event_type = line.replace('event:', '').strip()
            # data 代表事件的具体数据
            if line.startswith("data:"):
                # 从数据中找到'data:'开头的行并移除，同时去掉多余的空格或换行符
                data = line.replace('data:', '').strip()
                try:
                    event_data = json.loads(data)
                    if event_data['type'] != event_type:
                        # 本次事件的数据中的事件类型与本次事件的类型不一致，告警但不中断处理
                        self.logger.error(f'Get event data not consist with type, expect: {event_type}, get: {event_data['type']}')
                    if event_data['type'] == 'error':
                        # 处理报错
                        err_code = event_data['error']['code']
                        err_msg = event_data['error']['message']
                        self.logger.error(f'Get error from server, code: {err_code}, message: {err_msg}')
                        yield f'调用大模型知识引擎出错：{err_msg}'
                    elif event_data['type'] == 'reference':
                        self.__references__ += event_data['payload']['references']
                    elif event_data['type'] == 'thought':
                        # 处理思考模型的思考过程
                        if len(event_data['payload']['procedures']) > 0:
                            self.__reasoning_content__ = event_data['payload']['procedures'][0]['debugging']['content']
                            self.__reasoning_elapsed__ = float(event_data['payload']['procedures'][0]['elapsed'])
                            self.__reasoning_procedure_name__ = event_data['payload']['procedures'][0]['title'].strip()
                            procedure_invoice = f'> {self.__reasoning_procedure_name__}...'
                            # 每遇到一段思考，输出一次，避免等待过久体验不佳以及回答过长企微强制截断。
                            # 先处理段落，再处理单句。单句一般就是思考的最后一段话。
                            should_yield = False
                            if self.__reasoning_content__.endswith('\n\n'):
                                should_yield = True
                            if should_yield:
                                # 先裁剪出新增的文字
                                reasoning_content_cut = self.__reasoning_content__.replace(self.__reasoning_content_snapshot__, '')
                                reasoning_content_cut = self.__format_references__(reasoning_content_cut.strip())
                                # 再保存快照
                                self.__reasoning_content_snapshot__ = self.__reasoning_content__
                                # 输出思考过程片段
                                if len(reasoning_content_cut) != 0:
                                    yield f'{procedure_invoice}\n> \n> {reasoning_content_cut}'
                    elif event_data['type'] == 'reply':
                        # 处理大模型回复
                        prefix = ''
                        # 思考过程的最后才是回复，需要打印思考耗时
                        if not self.__reasoning_elapsed_sent__ and self.__reasoning_elapsed__ > 0:
                            self.__reasoning_procedure_name__ = self.__reasoning_procedure_name__.replace('中', '')
                            self.__reasoning_procedure_name__ = self.__reasoning_procedure_name__.replace('已完成', '')
                            prefix = f'> <font color=\"comment\">{self.__reasoning_procedure_name__}共用时{self.__reasoning_elapsed__/1000}秒</font>{prefix}'
                            self.__reasoning_elapsed_sent__ = True
                        # 如果思考内容还有部分没输出，需要输出
                        if len(self.__reasoning_content__) > len(self.__reasoning_content_snapshot__):
                            reasoning_content_cut = self.__reasoning_content__.replace(self.__reasoning_content_snapshot__, '')
                            reasoning_content_cut = self.__format_references__(reasoning_content_cut.strip())
                            # 快照保存为最终的思考内容，避免重复发送
                            self.__reasoning_content_snapshot__ = self.__reasoning_content__
                            # 输出思考过程片段
                            if len(reasoning_content_cut) != 0:
                                yield f'> {self.__reasoning_procedure_name__}...\n> \n> {reasoning_content_cut}\n{prefix}'
                        if not event_data['payload']['is_from_self']:
                            if event_data['payload']['is_final']:
                                self.logger.info(f'Get final event, traceId: {event_data['payload']['trace_id']}, data: {event_data}')
                            self.__content__ = event_data['payload']['content']
                            # 每遇到一段回答，输出一次，避免等待过久体验不佳以及回答过长企微强制截断。
                            newlines_pos = self.__content__.find('\n\n')
                            if newlines_pos != -1 and newlines_pos != self.__content_current_newlines_pos__:
                                self.__content_current_newlines_pos__ = newlines_pos
                                # 裁剪出完整段落的文字
                                content_cut = self.__content__[self.__content_previous_newlines_pos__:self.__content_current_newlines_pos__]
                                self.__content_previous_newlines_pos__ = self.__content_current_newlines_pos__
                                # 输出回复片段
                                content_cut = self.__format_references__(content_cut.strip())
                                if len(prefix) > 0:
                                    content_cut = f'{prefix}\n\n{content_cut}'
                                if len(content_cut) != 0:
                                    yield content_cut
                        else:
                            self.logger.info(f'Get input event, traceId: {event_data['payload']['trace_id']}, data: {event_data}')
                        continue
                    elif event_data['type'] == 'token_stat':
                        # 处理token相关信息
                        status_summary = event_data['payload']['status_summary'].strip()
                        if status_summary == "processing":
                            procedure_count = len(event_data['payload']['procedures'])
                            if procedure_count > 0:
                                procedure_name = event_data['payload']['status_summary_title'].strip()
                                if procedure_name != self.__last_procedure_name__:
                                    self.__last_procedure_name__ = procedure_name
                                    procedure_invoice = f'> {self.__last_procedure_name__}，请稍等...'
                                    yield procedure_invoice
                        elif status_summary == "success":
                            self.logger.info(f'{event_data}')
                    else:
                        self.logger.error(f'Get unsupported event: {data}')
                except json.JSONDecodeError:
                    self.logger.error(f'Parse event into JSON failed, data: {data}')
            # 空行代表一个事件的结束
            if line == '':
                self.logger.debug(f'Get event: \n{buffer}') # 打印事件缓存，用于debug
                buffer = ''
        # 发送最后的一段回复
        content_cut = self.__format_references__(self.__content__[self.__content_current_newlines_pos__:].strip())
        if not self.__reasoning_elapsed_sent__ and self.__reasoning_elapsed__ > 0:
            self.__reasoning_procedure_name__ = self.__reasoning_procedure_name__.replace('中', '')
            self.__reasoning_procedure_name__ = self.__reasoning_procedure_name__.replace('已完成', '')
            content_cut = f'> <font color=\"comment\">{self.__reasoning_procedure_name__}共用时{self.__reasoning_elapsed__/1000}秒</font>{content_cut}'
            self.__reasoning_elapsed_sent__ = True
        yield content_cut
        response.close()

    def __format_references__(self, text):
        formatted_text = text
        if len(self.__references__) > 0:
            for ref in self.__references__:
                ref_holder = f'[{ref['id']}]'
                ref_link = f'[【资料{ref['id']}】]({ref['url']})'
                formatted_text = formatted_text.replace(ref_holder, ref_link)
        return formatted_text
