class Postman {
    constructor(serverIp, reconnectOnClose) {
        this.ws = new WebSocket("ws://" + serverIp + "/postman");
		this.ws.onopen = function(we) {
			//console.log(we);
		};
		this.ws.onmessage = function(we) {
            //console.log(we);
            if(we.data == "" || we.data.length < 8)
                return;
            
            let head = "";
            for(let i = 0; i < 8; i++)
                head += we.data[i]
            if(head != "message ")
                return;

            let msg = "";
            for(let i = 8; i < we.data.length; i++)
                msg += we.data[i]
            
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
            //console.log(we);
            let e = new Event("on_postman_close");
            document.dispatchEvent(e);
		};
		this.ws.onerror = function(we) {
            //console.log(we);
            let e = new Event("on_postman_error");
            document.dispatchEvent(e);
        };
    }

    on(eventType, func) {
        if(eventType == "pingpong")
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
}