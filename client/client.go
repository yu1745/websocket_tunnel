/*package main

import (
	"flag"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("r", "wangyu175.cf:8080", "cdn")
var l = flag.String("l", ":25565", "local port")
var u = url.URL{Scheme: "ws", Host: *addr, Path: "/"}

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)

	conns, err := accept()
	if err != nil {
		log.Fatalln(err)
	}
	for conn := range conns {
		connect(conn)
	}
}

func accept() (<-chan net.Conn, error) {
	listener, err := net.Listen("tcp", *l)
	if err != nil {
		return nil, err
	}
	ch := make(chan net.Conn)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println(err)
				continue
			}
			ch <- conn
		}
	}()
	go func() {
		ch := make(chan os.Signal)
		signal.Notify(ch, os.Interrupt)
		<-ch
		close(ch)
		time.Sleep(time.Second * 2)
		os.Exit(0)
	}()
	return ch, nil
}

func connect(conn net.Conn) {
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	b := make([]byte, 1024)
	go func(c *websocket.Conn) {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println(err)
				closeWS(c)
				return
			}
			_, err = conn.Write(message)
			if err != nil {
				log.Println(err)
				closeWS(c)
				return
			}
			//time.Sleep(time.Millisecond)
		}
	}(c)
	for {
		nR, err := conn.Read(b)
		if err != nil {
			log.Println(err)
			closeWS(c)
			return
		}
		//println(nR)
		if nR > 0 {
			err = c.WriteMessage(websocket.BinaryMessage, b[:nR])
			if err != nil {
				log.Println(err)
				closeWS(c)
				return
			}
		}
		//time.Sleep(time.Millisecond)
	}
}

func closeWS(c *websocket.Conn) {
	err := c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
	if err != nil {
		log.Println(err)
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
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"nhooyr.io/websocket"
	"strings"
)

var addr = flag.String("a", "opencdn.jomodns.com:443", "server address")
var https = flag.Bool("s", false, "enable https")
var listen = flag.String("l", ":25565", "listen address")
var fake = flag.String("fake", "", "fake server name")

func init() {
	log.SetFlags(log.Lshortfile)
	flag.Parse()
}

func main() {
	/*generate := ReadSourceAndGenerate()
	ip := Tcping(generate)
	println(ip.String())*/

	listener, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalln(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		}
		//_ = conn.SetReadDeadline(time.Unix(0, 0))
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	c, err := NewWSConnection()
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close(websocket.StatusAbnormalClosure, "")
	netConn := websocket.NetConn(context.TODO(), c, websocket.MessageBinary)
	go func() {
		_, err := io.Copy(netConn, conn)
		if err != nil {
			log.Println(err)
		}
	}()
	_, err = io.Copy(conn, netConn)
	if err != nil {
		log.Println(err)
	}
	_ = c.Close(websocket.StatusNormalClosure, "")
	conn.Close()
	/*go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		for {
			<-ticker.C
			_, received, err := c.Read(context.Background())
			if err != nil {
				log.Println(err)
				return
			}
			_, err = conn.Write(received)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}()
	buf := make([]byte, 1024*1024)
	ticker := time.NewTicker(10 * time.Millisecond)
	for {
		<-ticker.C
		nR, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			return
		}
		err = c.Write(context.Background(), websocket.MessageBinary, buf[:nR])
		if err != nil {
			log.Println(err)
			return
		}

	}*/
}

func NewWSConnection() (*websocket.Conn, error) {
	var u url.URL
	var fake_ string
	if *fake != "" {
		fake_ = *fake
	} else {
		fake_ = strings.Split(*addr, ":")[0]
	}
	if *https {
		u = url.URL{Scheme: "wss", Host: *addr, Path: "/echo"}
	} else {
		u = url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	}
	log.Printf("connecting to %s", u.String())
	c, _, err := websocket.Dial(context.TODO(), u.String(), &websocket.DialOptions{HTTPClient: &http.Client{Transport: &http.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return tls.Dial(network, addr, &tls.Config{ServerName: fake_})
		},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial(network, addr)
		},
	}}})
	return c, err
}
