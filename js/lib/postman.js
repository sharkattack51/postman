/*
    // Postman Client JS

    let postman = new Postman(host, true, true);

    postman.on("open", () => {
        postman.subscribe("channel");
    });

    postman.on("message", (e) => {
        let ch = e.data.channel;
        let msg = e.data.message;
        let tag = e.data.tag;
        let ext = e.data.extention;

        console.log(e.data);
    });

    postman.on("close", () => {
        console.log("close");
    });

    postman.on("error", (e) => {
        console.log(e.data);
    });
*/

class Postman {
    constructor(serverIp, ssl = false, connectOnClose = true) {
        this.url = serverIp + "/postman";
        if(ssl)
            this.url = "wss://" + this.url;
        else
            this.url = "ws://" + this.url;

        this.ws = new WebSocket(this.url);

		this.onopen = function(we) {
            let e = new Event("on_postman_open");
            document.dispatchEvent(e);
        };
        this.ws.onopen = this.onopen;

		this.onmessage = function(we) {
            if(we.data == "" || we.data.length < 8)
                return;

            let head = we.data.substring(0, 8);
            if(head != "message ")
                return;

            let msg = we.data.substring(8, we.data.length);
            if(msg == "\"pong\"") {
                let e = new Event("on_postman_pingpong");
                document.dispatchEvent(e);
            } else {
                let e = new Event("on_postman_message");
                e.data = JSON.parse(msg);
                document.dispatchEvent(e);
            }
        }
        this.ws.onmessage = this.onmessage;

		this.onclose = (we) => {
            let e = new Event("on_postman_close");
            document.dispatchEvent(e);

            if(connectOnClose) {
                (async () => {
                    while(this.ws == undefined || this.ws.readyState !== 1) {
                        this.ws = new WebSocket(this.url);
                        await new Promise(resolve => setTimeout(resolve, 1000));
                    }

                    this.ws.onopen = this.onopen;
                    this.ws.onmessage = this.onmessage;
                    this.ws.onclose = this.onclose;
                    this.ws.onerror = this.onerror;

                    let e = new Event("on_postman_open");
                    document.dispatchEvent(e);
                })();
            }
        };
        this.ws.onclose = this.onclose;

		this.onerror = function(we) {
            let e = new Event("on_postman_error");
            document.dispatchEvent(e);
        };
        this.ws.onerror = this.onerror;
    }

    on(eventType, func) {
        if(eventType == "open")
            document.addEventListener("on_postman_open", func);
        else if(eventType == "pingpong")
            document.addEventListener("on_postman_pingpong", func);
        else if(eventType == "message")
            document.addEventListener("on_postman_message", func);
        else if(eventType == "close")
            document.addEventListener("on_postman_close", func);
        else if(eventType == "error")
            document.addEventListener("on_postman_error", func);
    }

    ping() {
        if(this.ws.readyState === 1)
            this.ws.send("ping {}");
    }

    status() {
        if(this.ws.readyState === 1)
            this.ws.send("status {}");
    }

    subscribe(channel, client_info = "") {
        if(this.ws.readyState === 1) {
            let sub_msg = {
                channel: channel,
                client_info: client_info
            }

            this.ws.send("subscribe " + JSON.stringify(sub_msg));
        }
    }

    unsubscribe(channel) {
        if(this.ws.readyState === 1) {
            let unsub_msg = {
                channel: channel
            }

            this.ws.send("unsubscribe " + JSON.stringify(unsub_msg));
        }
    }

    publish(channel, message, tag, extention) {
        if(this.ws.readyState === 1) {
            let pub_msg = {
                channel: channel,
                message: message,
                tag: tag,
                extention: extention
            }

            this.ws.send("publish " + JSON.stringify(pub_msg));
        }
    }

    disconnect() {
        if(this.ws.readyState === 1) {
            this.ws.close();
        }
    }

    isConnect() {
        return this.ws.readyState === 1;
    }
}
