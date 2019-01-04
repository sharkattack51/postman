# postman
easy pub/sub messaging server using websocket.

### Websocket API

- `Ping`
  - <- "ping {}"
- `Status`
  - <- "status {}"
- `Subscribe`
  - <- "subscribe {"channel": "CHANNEL"}"
- `Unsubscribe`
  - <- "unsubscribe {"channel": "CHANNEL"}"
- `Publish`
  - <- "publish {"channel": "CHANNEL", "message": "MESSAGE"}"
  - <- "publish {"channel": "CHANNEL", "message": "MESSAGE", "tag": "TAG", "extention": "OTHER"}"

### Http API

- `Status`
  - GET http://XXX.XXX.XXX.XXX:8800/postman/status
- `Publish`
  - GET http://XXX.XXX.XXX.XXX:8800/postman/publish?channel=CHANNEL&message=MESSAGE
  - GET http://XXX.XXX.XXX.XXX:8800/postman/publish?channel=CHANNEL&message=MESSAGE&tag=TAG&extention=OTHER
  - POST http://XXX.XXX.XXX.XXX:8800/postman/publish <- "json={"channel": "CHANNEL", "message": "MESSAGE"}"
  - POST http://XXX.XXX.XXX.XXX:8800/postman/publish <- "json={"channel": "CHANNEL", "message": "MESSAGE", "tag": "TAG", "extention": "OTHER"}"

### Client Library

- `Unity`
  - require
    - [websocket-sharp.dll](https://github.com/sta/websocket-sharp)
    - [Json.NET for Unity](https://assetstore.unity.com/packages/tools/input-management/json-net-for-unity-11347)
- `js`

### Using on Heroku

change code in `main.go`.
```
TARGET_HEROKU = true
```

and `go build ./...`
