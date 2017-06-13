package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	argv struct {
		host string
		key  string
		rule string
		tag  string
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
	} else if len(argv.tag) == 0 {
		fmt.Fprintln(os.Stderr, "Hey! -tag is required")
		flag.Usage()
	} else if len(argv.rule) == 0 {
		fmt.Fprintln(os.Stderr, "Hey! -rule is required")
		flag.Usage()
	} else {
		needStop = false
	}

	return
}

func init() {
	flag.StringVar(&argv.host, `host`, ``, `streaming api host. REQUIRED`)
	flag.StringVar(&argv.key, `key`, ``, `client key. REQUIRED`)
	flag.StringVar(&argv.tag, `tag`, ``, `tag. REQUIRED`)
	flag.StringVar(&argv.rule, `rule`, ``, `rule. REQUIRED`)
	flag.BoolVar(&argv.help, `h`, false, `show this help`)

	flag.Parse()
}

func main() {
	if processArgs() {
		return
	}

	url := fmt.Sprintf("https://%s/rules/?key=%s", argv.host, argv.key)
	rule := `{"rule":{"value":"` + argv.rule + `","tag":"` + argv.tag + `"}}`

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(rule)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("http request error:", err)
	}
	defer resp.Body.Close()

	bodyBuf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("response body read error:", err)
	}
	log.Println(string(bodyBuf))
}
