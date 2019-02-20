package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tsenart/vegeta/lib"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"time"
)

const (
	defaultFreq   = 10000
	defaultHost   = "localhost"
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func RandStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
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

func main() {
	var frequency = flag.Int("f", defaultFreq, "Posts in second")
	var host = flag.String("h", defaultHost, "Server host")

	flag.Parse()

	rate := vegeta.Rate{Freq: *frequency, Per: time.Second}
	duration := 1 * time.Minute
	targets := make([]vegeta.Target, *frequency, *frequency)
	log.Println(*frequency, "requests per second")
	for i := 0; i < *frequency; i++ {
		protoName := RandStringBytesMaskImprSrc(6)
		locoName := RandStringBytesMaskImprSrc(6)
		targets[i] = vegeta.Target{
			Method: "POST",
			URL:    fmt.Sprintf("http://%s:8080/%s/%s/log", *host, protoName, locoName),
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
	err = ioutil.WriteFile("c:/tools/vegeta"+strconv.Itoa(*frequency)+".json", data, 0644)
	if err != nil {
		panic(err)
	}
}
