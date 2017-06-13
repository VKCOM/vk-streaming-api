package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
)

var (
	argv struct {
		host string
		key  string
		help bool
	}
)

func processArgs() (needStop bool) {
	needStop = true

	if argv.help {
		flag.Usage()
	} else if len(argv.host) == 0 {
		fmt.Fprintln(os.Stderr, "Hey! -host is required")
		flag.Usage()
	} else if len(argv.key) == 0 {
		fmt.Fprintln(os.Stderr, "Hey! -key is required")
		flag.Usage()
	} else {
		needStop = false
	}

	return
}

func init() {
	flag.StringVar(&argv.host, `host`, ``, `streaming api host. REQUIRED`)
	flag.StringVar(&argv.key, `key`, ``, `client key. REQUIRED`)
	flag.BoolVar(&argv.help, `h`, false, `show this help`)

	flag.Parse()
}

func main() {
	if processArgs() {
		return
	}

	u := url.URL{Scheme: "wss", Host: argv.host, Path: "/stream/", RawQuery: "key=" + argv.key}
	log.Printf("connecting to %s\n", u.String())
	c, wsResp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		if err == websocket.ErrBadHandshake {
			log.Printf("handshake failed with status %d\n", wsResp.StatusCode)
			bodyBuf, _ := ioutil.ReadAll(wsResp.Body)
			log.Println("respBody:", string(bodyBuf))
		}
		log.Fatal("dial error:", err)
	}
	log.Println("connection established")
	defer c.Close()

	done := make(chan struct{})
	defer close(done)

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				done <- struct{}{}
				return
			}
			log.Printf("recv: %s", string(message))
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	select {
	case <-interrupt:
		log.Println("interrupt")
		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("write close error: ", err)
			return
		}
		select {
		case <-done:
		case <-time.After(time.Second):
		}
	case <-done:
	}
}
