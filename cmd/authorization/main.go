package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	argv struct {
		accessToken string
		help        bool
	}
)

func processArgs() (needStop bool) {
	needStop = true

	if argv.help {
		flag.Usage()
	} else if len(argv.accessToken) == 0 {
		fmt.Fprintln(os.Stderr, "Hey! -token is required")
		flag.Usage()
	} else {
		needStop = false
	}

	return
}

func init() {
	flag.StringVar(&argv.accessToken, `token`, ``, `access token. REQUIRED`)
	flag.BoolVar(&argv.help, `h`, false, `show this help`)

	flag.Parse()
}

type (
	vkApiResponse struct {
		Response struct {
			Endpoint string
			Key      string
		}
	}
)

func main() {
	if processArgs() {
		return
	}

	resp, err := http.Get("https://api.vk.com/method/streaming.getServerUrl?access_token=" + argv.accessToken + "&v=5.64")
	if err != nil {
		log.Fatal("firehose api authorization failed:", err)
	}
	defer resp.Body.Close()

	bodyBuf, err := ioutil.ReadAll(resp.Body)
	var v vkApiResponse
	if err := json.Unmarshal(bodyBuf, &v); err != nil {
		log.Fatal("unmarshal response json failed:", err)
	}

	log.Println("host:", v.Response.Endpoint, "key:", v.Response.Key)
}
