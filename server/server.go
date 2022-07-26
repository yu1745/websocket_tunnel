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

var (
	addr   string
	https  bool
	cert   string
	key    string
	dst    string
	header string
)

func init() {
	flag.StringVar(&addr, "a", ":80", "http service address")
	flag.BoolVar(&https, "s", false, "enable https")
	flag.StringVar(&cert, "c", "", "cert file")
	flag.StringVar(&key, "k", "", "private key file")
	flag.StringVar(&dst, "d", "localhost:25565", "destination address")
	flag.StringVar(&header, "header", "X-Real-IP", "the http header key implying the client ip, generally it's X-Real-IP or True-Client-Ip or X-Forwarded-For")
	flag.Parse()
	if https {
		if cert == "" || key == "" {
			println("error: when enabling https, you should provide cert and private key by adding -c xxx and -k yyy in the commandline")
			flag.PrintDefaults()
			os.Exit(0)
		}
	} else {
		//println("It is recommended to enable https to avoid HUGE traffic bill")
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	printConfig()
}

func printConfig() {
	log.Printf("listen address: %s", addr)
	log.Printf("enable https: %t", https)
	if https {
		log.Printf("using cert file: %s", cert)
		log.Printf("using key file: %s", key)
	}
	log.Printf("proxy destination: %s", dst)
	log.Printf("client ip http header key: %s", header)
}

func main() {
	http.HandleFunc("/proxy", proxy)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Header.Get(header))
		_, _ = w.Write([]byte("hello"))
	})
	if https {
		log.Fatalln(http.ListenAndServeTLS(addr, cert, key, nil))
	} else {
		log.Fatalln(http.ListenAndServe(addr, nil))
	}
}

func proxy(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("[proxy] from %s", r.Header.Get(header))
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

func echo(w http.ResponseWriter, r *http.Request) {
	log.Printf("[echo] from %s", r.Header.Get(header))
	/*for k, v := range r.Header {
		println(k, ":", strings.Join(v, ";"))
	}*/
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "")
	conn := websocket.NetConn(context.TODO(), c, websocket.MessageBinary)
	defer conn.Close()
	go io.Copy(conn, conn)
	io.Copy(conn, conn)
}
