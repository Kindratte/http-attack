package main

import (
	"flag"
	"fmt"
	"github.com/tsenart/vegeta/lib"
	"log"
	"math/rand"
	"time"
)

const (
	defaultFreq   = 10000
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

	flag.Parse()

	rate := vegeta.Rate{Freq: *frequency, Per: time.Second}
	duration := 5 * time.Minute
	targets := make([]vegeta.Target, *frequency, *frequency)
	ports := [3]string{"8080", "8081", "8082"}
	rand.Seed(time.Now().Unix())
	log.Println(*frequency)
	for i := 0; i < *frequency; i++ {
		port := ports[rand.Intn(len(ports))]
		protoName := RandStringBytesMaskImprSrc(6)
		locoName := RandStringBytesMaskImprSrc(6)
		targets[i] = vegeta.Target{
			Method: "POST",
			URL:    fmt.Sprintf("http://localhost:%s/%s/%s/log", port, protoName, locoName),
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

	fmt.Printf("99th percentile: %s\n", metrics.Latencies.P99)
}
