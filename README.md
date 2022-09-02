# postman
easy pub/sub messaging server using websocket.

Application Options:
- `-p, --port`: listen port number (default: 8800)
- `-l, --log`: output log location
- `-c, --chlist`: safelist for channels
- `-i, --iplist`: connectable ip_address list
- `-f, --file`: enable file server api
- `-u, --plugin`: enable plugin api
- `-s, --secure`: enable secure mode
- `-g, --generate`: genarate token from environment variable [SECRET]

Help Options:
- `-h, --help`: Show this help message

### Websocket API

- `Ping`
  - <- "ping {}"
- `Status`
  - <- "status {}"
- `Subscribe`
  - <- "subscribe {"ch": "CHANNEL", ["ci": "CLIENT_INFO"]}"
- `Unsubscribe`
  - <- "unsubscribe {"ch": "CHANNEL"}"
- `Publish`
  - <- "publish {"ch": "CHANNEL", "msg": "MESSAGE", ["tag": "TAG", "ext": "OTHER"]}"

### Http API

http://XXX.XXX.XXX.XXX:8800/postman

- `Status`
  - (GET) [/status]()
  - (GET) [/status_pp]()
- `Publish`
  - (GET) [/publish?ch=CHANNEL&msg=MESSAGE[&tag=TAG&ext=OTHER&ci=CLIENT_INFO]]()
  - (POST) [/publish]() <- json={"ch": "CHANNEL", "msg": "MESSAGE", ["tag": "TAG", "ext": "OTHER", "ci": "CLIENT_INFO"]}
- `Store`
  - (GET) [/store?cmd=(GET|SET|HAS|DEL)&key=KEY[&val=VALUE]]()
  - (POST) [/store]() <- json={"cmd": "(GET|SET|HAS|DEL)", "key": "KEY", ["val": "VALUE"]}
- `File`
  - (GET) [/file?name=FILE_NAME]()
  - (POST) [/file]() <- file=FILE_BINARY
- `Plugin`
  - (GET) [/plugin?cmd=COMMAND]()
  - (POST) [/plugin]() <- json={"cmd": "COMMAND"}
  - see `./plugin/plugin.json`

### Client Library

- `Unity`
  - require
    - [websocket-sharp.dll](https://github.com/sta/websocket-sharp)
    - [Json.NET for Unity](https://assetstore.unity.com/packages/tools/input-management/json-net-for-unity-11347)
- `js`
- `python`

### Build Tags for Windows

`$ go build -tags windows ./...` others `$ go build ./...`

### Using on Heroku

change code in `main.go`.
```
TARGET_HEROKU = true
```

and `$ git push heroku master`
