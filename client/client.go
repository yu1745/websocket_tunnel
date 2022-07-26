package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"nhooyr.io/websocket"
	"os"
	"strings"
)

var (
	addr   string
	https  string
	listen string
	fake   string
	real_  string
)

// CA from ubuntu 20.04
//go:embed RootCAs
var CA embed.FS
var pool *x509.CertPool

func init() {
	//flag.StringVar(&addr, "a", "", "server address+port(used to resolve the ip to be connected) or ip+port")
	//flag.StringVar(&https, "s", "false", "enable https")
	//flag.StringVar(&listen, "l", ":25565", "listen address")
	//flag.StringVar(&fake, "fake", "", "fake server name(used in sni)")
	//flag.StringVar(&real_, "real", "", "real server name(used in http host)")
	log.SetFlags(log.Lshortfile)
	flag.Parse()
	if addr == "" {
		log.Println("address should not be null, please use -a xxx to set the address")
		flag.PrintDefaults()
		os.Exit(0)
	}
	pool = x509.NewCertPool()
	b, err := CA.ReadFile("RootCAs")
	if err != nil {
		log.Fatalln(err)
	}
	pool.AppendCertsFromPEM(b)
}

func main() {
	/*generate := ReadSourceAndGenerate()
	ip := Tcping(generate)
	println(ip.String())*/

	listener, err := net.Listen("tcp", listen)
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
	if fake != "" {
		fake_ = fake
	} else {
		fake_ = strings.Split(addr, ":")[0]
	}
	if https == "true" {
		if real_ != "" {
			u = url.URL{Scheme: "wss", Host: real_, Path: "/proxy"}
		} else {
			u = url.URL{Scheme: "wss", Host: addr, Path: "/proxy"}
		}
	} else {
		if real_ != "" {
			u = url.URL{Scheme: "ws", Host: real_, Path: "/proxy"}
		} else {
			u = url.URL{Scheme: "ws", Host: addr, Path: "/proxy"}
		}
	}
	addr_ := addr
	log.Printf("connecting to %s", u.String())
	c, _, err := websocket.Dial(context.TODO(), u.String(), &websocket.DialOptions{HTTPClient: &http.Client{Transport: &http.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return tls.Dial(network, addr_, &tls.Config{ServerName: fake_, RootCAs: pool})
		},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial(network, addr_)
		},
	}}})
	return c, err
}
