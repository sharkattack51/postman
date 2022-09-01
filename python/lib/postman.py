import websocket
import json
import threading
import time

class Postman:
    ws = None
    url = ""

    on_connect = None
    on_message = None
    on_close = None
    on_error = None

    def __init__(self, serverIp="127.0.0.1:8800", ssl=False, on_connect=None, on_message=None, on_close=None, on_error=None):
        self.url = serverIp + "/postman"
        if ssl:
            self.url = "wss://" + self.url
        else:
            self.url = "ws://" + self.url

        self.on_connect = on_connect
        self.on_message = on_message
        self.on_close = on_close
        self.on_error = on_error

    def on_internal_open(self, ws):
        if self.on_connect != None:
            self.on_connect()

    def on_internal_message(self, ws, message):
        def thread_run():
            try:
                if message == "" or len(message) < 8:
                    return

                head = message[0:8]
                if head != "message ":
                    return

                msg = message[8:len(message)]
                if msg == '\"pong\"':
                    if self.on_pingpong != None:
                        self.on_pingpong()
                else:
                    if self.on_message != None:
                        j = json.loads(msg)
                        self.on_message(j["channel"], j["message"], j["tag"], j["extention"])
            except:
                if self.on_error != None:
                    self.on_error()

        threading.Thread(target=thread_run).start()

    def on_internal_close(self, ws, close_status_code, close_msg):
        if self.on_close != None:
            self.on_close()

    """
    Postman Public Function
    """
    def connect(self):
        if self.ws == None:
            try:
                self.ws = websocket.WebSocketApp(self.url, on_open=self.on_internal_open, on_message=self.on_internal_message, on_close=self.on_internal_close)
                def thread_run():
                    self.ws.run_forever()

                threading.Thread(target=thread_run).start()
                time.sleep(0.5)
            except:
                if self.on_error != None:
                    self.on_error()

    def connect_and_wait(self):
        self.connect()
        while True: pass # wait forever

    def subscribe(self, channel, client_info=""):
        if self.ws != None:
            try:
                sub_msg = {
                    "channel": channel,
                    "client_info": client_info
                }
                self.ws.send("subscribe " + json.dumps(sub_msg))
            except:
                if self.on_error != None:
                    self.on_error()

    def unsubscribe(self, channel):
        if self.ws != None:
            try:
                unsub_msg = {
                    "channel": channel
                }
                self.ws.send("unsubscribe " + json.dumps(unsub_msg))
            except:
                if self.on_error != None:
                    self.on_error()
    
    def publish(self, channel, message, tag="", extention=""):
        if self.ws != None:
            try:
                pub_msg = {
                    "channel": channel,
                    "message": message,
                    "tag": tag,
                    "extention": extention
                }
                self.ws.send("publish " + json.dumps(pub_msg))
            except:
                if self.on_error != None:
                    self.on_error()

    def disconnect(self):
        if self.ws != None:
            try:
                self.ws.keep_running = False # finish ws.run_forever()
                self.ws.close()
            except:
                if self.on_error != None:
                    self.on_error()
        self.ws = None
