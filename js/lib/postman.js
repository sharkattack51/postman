class Postman {
    constructor(serverIp, ssl) {
        this.url = serverIp + "/postman";
        if(ssl)
            this.url = "wss://" + this.url;
        else
            this.url = "ws://" + this.url;
        
        this.ws = new WebSocket(this.url);

		this.ws.onopen = function(we) {
            let e = new Event("on_postman_open");
            document.dispatchEvent(e);
        };
        
		this.ws.onmessage = function(we) {
            if(we.data == "" || we.data.length < 8)
                return;
            
            let head = we.data.substring(0, 8);
            if(head != "message ")
                return;

            let msg = we.data.substring(8);
            if(msg == "\"pong\"") {
                let e = new Event("on_postman_pingpong");
                document.dispatchEvent(e);
            } else {
                let e = new Event("on_postman_message");
                e.data = JSON.parse(msg);
                document.dispatchEvent(e);
            }
        }
        
		this.ws.onclose = function(we) {
            let e = new Event("on_postman_close");
            document.dispatchEvent(e);
        };
        
		this.ws.onerror = function(we) {
            let e = new Event("on_postman_error");
            document.dispatchEvent(e);
        };
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

    subscribe(channel) {
        if(this.ws.readyState === 1) {
            let sub_msg = {
                channel: channel
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
}