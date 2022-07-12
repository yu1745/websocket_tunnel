/*package main

import (
	"flag"
	"html/template"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", ":8080", "")
var dst = flag.String("dst", "localhost:25565", "")

var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("type: %v", mt)
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", handle)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func handle(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	conn, err := net.Dial("tcp", *dst)
	if err != nil {
		log.Println(err)
		return
	}
	b := make([]byte, 1024)
	go func(c *websocket.Conn, conn net.Conn) {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println(err)
				closeWS(c)
				conn.Close()
				return
			}
			_, err = conn.Write(message)
			if err != nil {
				log.Println(err)
				closeWS(c)
				conn.Close()
				return
			}
			//time.Sleep(time.Millisecond)
		}
	}(c, conn)
	for {
		nR, err := conn.Read(b)
		if err != nil {
			log.Println(err)
			closeWS(c)
			conn.Close()
			return
		}
		println(nR)
		if nR > 0 {
			err = c.WriteMessage(websocket.BinaryMessage, b[:nR])
			if err != nil {
				log.Println(err)
				closeWS(c)
				conn.Close()
				return
			}
		}
		//time.Sleep(time.Millisecond)
	}
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>
window.addEventListener("load", function(evt) {

    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;

    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
        output.scroll(0, output.scrollHeight);
    };

    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };

    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };

    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };

});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server,
"Send" to send a message to the server and "Close" to close the connection.
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output" style="max-height: 70vh;overflow-y: scroll;"></div>
</td></tr></table>
</body>
</html>
`))

func closeWS(c *websocket.Conn) {
	err := c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
	if err != nil {
		log.Println("write close:", err)
		return
	}
}
*/
// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"nhooyr.io/websocket"
)

var addr = flag.String("a", "[::]:443", "http service address")
var https = flag.Bool("s", false, "enable https")
var cert = flag.String("c", "./cert.pem", "cert file")
var key = flag.String("k", "./private.pem", "private key file")
var dst = flag.String("d", "localhost:25565", "destination address")

func init() {
	flag.Parse()
	log.SetFlags(log.Lshortfile)
}

func main() {
	http.HandleFunc("/echo", echo)
	//http.HandleFunc("/", home)
	if *https {
		log.Fatal(http.ListenAndServeTLS(*addr, *cert, *key, nil))
	} else {
		log.Fatalln(http.ListenAndServe(*addr, nil))
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	log.Println("accept a connection")
	if err != nil {
		log.Println(err)
	}
	defer c.Close(websocket.StatusInternalError, "")
	conn := websocket.NetConn(context.TODO(), c, websocket.MessageBinary)
	defer conn.Close()
	dial, err := net.Dial("tcp", *dst)
	if err != nil {
		log.Println(err)
		return
	}
	defer dial.Close()
	go func() {
		_, err := io.Copy(dial, conn)
		if err != nil {
			log.Println(err)
		}
	}()
	_, err = io.Copy(conn, dial)
	if err != nil {
		log.Println(err)
	}
	conn.Close()
	dial.Close()
}
