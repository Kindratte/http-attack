package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/tsenart/vegeta/lib"
	"io/ioutil"
	"log"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultFreq   = 10000
	defaultHost   = "localhost"
	defaultTime   = 1
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func randStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func processUrlString(url string) []string {
	urls := strings.Split(url, ",")
	for i, s := range urls {
		urls[i] = strings.TrimSpace(s)
	}
	return urls
}

func pickRandomElem(arr []string) string {
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	return arr[r.Intn(len(arr))]
}

func main() {
	var frequency = flag.Int("f", defaultFreq, "Posts in second")
	var host = flag.String("h", defaultHost, "Server host or few separated by comma")
	var minutes = flag.Uint("m", defaultTime, "Time in minutes")

	flag.Parse()

	hosts := processUrlString(*host)

	rate := vegeta.Rate{Freq: *frequency, Per: time.Second}
	duration := time.Duration(*minutes) * time.Minute
	targets := make([]vegeta.Target, *frequency, *frequency)
	log.Println(*frequency, "requests per second")
	log.Printf("For %d minutes", *minutes)
	for i := 0; i < *frequency; i++ {
		protoName := randStringBytesMaskImprSrc(6)
		locoName := randStringBytesMaskImprSrc(6)
		targets[i] = vegeta.Target{
			Method: "POST",
			URL:    fmt.Sprintf("http://%s:8080/%s/%s/log", pickRandomElem(hosts), protoName, locoName),
		}
	}
	log.Println("Targets created")

	targeter := vegeta.NewStaticTargeter(targets...)
	attacker := vegeta.NewAttacker()

	var metrics vegeta.Metrics
	log.Println("Start attack")
	for res := range attacker.Attack(targeter, rate, duration, "Big Bang!") {
		metrics.Add(res)
	}
	metrics.Close()

	data, err := json.MarshalIndent(metrics, "", "	")
	if err != nil {
		panic(err)
	}
	home, _ := homedir.Dir()
	err = ioutil.WriteFile(filepath.Join(home, strconv.Itoa(*frequency)+".json"), data, 0644)
	if err != nil {
		panic(err)
	}
	log.Println("Attack ends")
}
