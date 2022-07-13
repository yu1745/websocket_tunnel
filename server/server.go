package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"nhooyr.io/websocket"
	"os"
)

var addr string
var https bool
var cert string
var key string
var dst string

func init() {
	flag.StringVar(&addr, "a", ":80", "http service address")
	flag.BoolVar(&https, "s", false, "enable https")
	flag.StringVar(&cert, "c", "", "cert file")
	flag.StringVar(&key, "k", "", "private key file")
	flag.StringVar(&dst, "d", "localhost:25565", "destination address")
	flag.Parse()
	if https {
		println("It is recommended to enable https to avoid HUGE traffic bill")
		if cert == "" || key == "" {
			println("error: when enabling https, you should provide cert and private key by adding -c xxx.pem and -k xxx.pem/xxx.key in the commandline")
			flag.PrintDefaults()
			os.Exit(0)
		}
	}
	log.SetFlags(log.Lshortfile)
}

func main() {
	http.HandleFunc("/proxy", proxy)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	})
	if https {
		log.Fatal(http.ListenAndServeTLS(addr, cert, key, nil))
	} else {
		log.Fatalln(http.ListenAndServe(addr, nil))
	}
}

func proxy(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	//log.Printf("accept a connection  %s", r.Header)
	if err != nil {
		log.Println(err)
	}
	defer c.Close(websocket.StatusInternalError, "")
	conn := websocket.NetConn(context.TODO(), c, websocket.MessageBinary)
	defer conn.Close()
	dial, err := net.Dial("tcp", dst)
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
