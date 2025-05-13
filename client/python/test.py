import sys
import os

sys.path.append(os.path.split(__file__)[0])
from lib import postman

def test():
    def on_connect():
        print("\n>>> connect postman...")

    def on_message(ch, msg, tag, ext):
        print("\n>>> message recieved: %s" % msg)

    def on_close():
        print("\n>>> close postman...")

    def on_error(err):
        print("\n>>> error: %s" % err)

    pstmn = postman.Postman("127.0.0.1:8800", on_connect=on_connect, on_message=on_message, on_close=on_close, on_error=on_error)
    pstmn.connect()

    while True:
        s = input("<<< postman test sub/pub/unsub/discon?: ")
        if s == "sub":
            pstmn.subscribe("TEST")
        elif s == "pub":
            pstmn.publish("TEST", "@@@@")
        elif s == "unsub":
            pstmn.unsubscribe("TEST")
        elif s == "discon":
            pstmn.disconnect()

if __name__ == "__main__":
    test()
