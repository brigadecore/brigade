package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/deis/acid/pkg/webhook"
)

const usage = "lsd [-f FILE] URL"

var (
	file   string
	url    string
	secret string
)

const (
	envURL    = "LSD_URL"
	envSecret = "LSD_SECRET"
)

const hubSignature = "X-Hub-Signature"

func init() {
	flag.StringVar(&file, "file", "acid.js", "the script to run")
	flag.StringVar(&url, "url", os.Getenv(envURL), "the URL of Acid's exec hook")
	flag.StringVar(&secret, "secret", os.Getenv(envSecret), "the shared secret for computing an HMAC auth")
}

func main() {
	flag.Parse()
	if url == "" {
		bail("a URL is required")
	}

	script, err := ioutil.ReadFile(file)
	if err != nil {
		bail(err.Error())
	}

	send(script, url, secret)
}

func bail(msg string) {
	fmt.Fprintln(os.Stderr, usage)
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func send(data []byte, url, secret string) error {
	sum := webhook.SHA1HMAC([]byte(secret), data)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		bail(err.Error())
	}
	req.Header.Set(hubSignature, sum)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		bail(err.Error())
	}

	out, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		bail(err.Error())
	}

	fmt.Fprintf(os.Stdout, "%d\n%s\n", res.StatusCode, string(out))
	return nil
}
