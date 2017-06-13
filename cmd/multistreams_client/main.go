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
	"strconv"
	"sync"
	"time"
)

var (
	argv struct {
		host         string
		key          string
		streamsCount int
		help         bool
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
	flag.IntVar(&argv.streamsCount, `streams`, 1, `streams count. 1 by default`)
	flag.BoolVar(&argv.help, `h`, false, `show this help`)

	flag.Parse()
}

func main() {
	if processArgs() {
		return
	}

	var wg sync.WaitGroup
	interruptChannels := make([]chan struct{}, argv.streamsCount)

	for streamId := 0; streamId < argv.streamsCount; streamId++ {
		u := url.URL{Scheme: "wss", Host: argv.host, Path: "/stream/", RawQuery: "key=" + argv.key + "&stream_id=" + strconv.Itoa(streamId)}
		log.Printf("connecting to %s\n", u.String())
		c, wsResp, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			if err == websocket.ErrBadHandshake {
				log.Printf("handshake failed with status %d\n", wsResp.StatusCode)
				bodyBuf, _ := ioutil.ReadAll(wsResp.Body)
				log.Println("respBody:", string(bodyBuf))
			}
			log.Printf("stream id: %d, dial error: %s\n", streamId, err.Error())
			continue
		}
		defer c.Close()

		log.Printf("stream id: %d, connection established\n", streamId)

		wg.Add(1)
		go func(id int) {
			done := make(chan struct{})
			defer close(done)

			go func() {

				for {
					_, message, err := c.ReadMessage()
					if err != nil {
						log.Printf("stream_id: %d, read: %s\n", id, err.Error())
						done <- struct{}{}
						return
					}
					log.Printf("recv: %s", message)
				}
			}()

			interruptChan := make(chan struct{})
			interruptChannels[id] = interruptChan

			select {
			case <-interruptChan:
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Printf("stream id: %d, write close error: %s\n", id, err.Error())
					wg.Done()
					return
				}
				select {
				case <-done:
				case <-time.After(time.Second):
				}
				wg.Done()
			case <-done:
				wg.Done()
			}
		}(streamId)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	select {
	case <-interrupt:
		log.Println("interrupt")
		for _, v := range interruptChannels {
			if v != nil {
				v <- struct{}{}
			}
		}
		wg.Wait()
	case <-done:
	}
}
