package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/tsenart/vegeta/lib"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	defaultFreq        = 1000
	defaultConnections = 10000
	defaultWorkers     = 10
	defaultTime        = 1
	letterBytes        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits      = 6                    // 6 bits to represent a letter index
	letterIdxMask      = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax       = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	defaultHost        = "https://localhost:8822"
	bOPrefix           = "/api/air-bo"
	testPrefix         = "/api/test"
	authPrefix         = "/api/user/login"
	defaultLocation    = 2
	contentJSONHeader  = "application/json"
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

func createArticleDependencies() []byte {
	jsonFile, err := os.Open("operations.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	data, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}
	return data
}

func createArticle(index int) []byte {
	jsonFile, err := os.Open("art.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	data, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}
	var abc map[string][]map[string]interface{}
	err = json.Unmarshal(data, &abc)
	if err != nil {
		panic(err)
	}
	abc["operations"][3]["data"].(map[string]interface{})["name"] = randStringBytesMaskImprSrc(10)
	if index&1 == 1 {
		abc["operations"][3]["data"].(map[string]interface{})["state"] = 0
	}
	res, err := json.Marshal(abc)
	if err != nil {
		panic(err)
	}
	return res
}

func writeSupplyToLocations(token, uRL string) {
	targets := []vegeta.Target{{
		Method: "POST",
		Body:   createArticleDependencies(),
		Header: map[string][]string{"Content-Type": {contentJSONHeader}},
		URL:    uRL}}
	if token != "" {
		targets[0].Header["Authorization"] = []string{"Bearer " + token}
	}
	targeter := vegeta.NewStaticTargeter(targets...)
	attacker := vegeta.NewAttacker(vegeta.Connections(defaultConnections), vegeta.Workers(uint64(defaultWorkers)))
	attacker.Attack(targeter, vegeta.Rate{Freq: 1, Per: time.Second}, time.Duration(1)*time.Second, "Supply")
}

func main() {
	var frequency = flag.Int("f", defaultFreq, "Posts in second")
	var login = flag.String("l", "", "Login in unTill Air")
	var password = flag.String("p", "", "Password")
	var location = flag.Int("loc", defaultLocation, "Location number")
	var host = flag.String("h", defaultHost, "Server host or few separated by comma")
	var minutes = flag.Uint("m", defaultTime, "Time in minutes")
	var numOfConnections = flag.Int("c", defaultConnections, "Connections num")
	var numOfWorkers = flag.Int("w", defaultWorkers, "Workers num")
	var test = flag.Bool("t", false, "Writes in /api/test queue")
	var noAuth = flag.Bool("na", false, "If true don't try auth on server")

	flag.Parse()

	log.Println("Host:", *host)

	rate := vegeta.Rate{Freq: *frequency, Per: time.Second}
	duration := time.Duration(50) * time.Second
	targetsNum := 10000
	targets := make([]vegeta.Target, targetsNum, targetsNum)
	log.Println(*frequency, "requests per second")
	log.Printf("For %d minutes", *minutes)

	token := ""

	if !*noAuth {
		token = authOnServer(*login, *password, *host)
	}

	var uRL string
	if *test {
		uRL = *host + testPrefix + "/" + strconv.Itoa(*location)
	} else {
		uRL = *host + bOPrefix + "/" + strconv.Itoa(*location)
	}

	writeSupplyToLocations(token, uRL)

	for i := 1; i < targetsNum; i++ {
		targets[i] = vegeta.Target{
			Method: "POST",
			Body:   createArticle(i),
			Header: map[string][]string{"Content-Type": {contentJSONHeader}},
			URL:    uRL,
		}
		if !*noAuth {
			targets[i].Header["Authorization"] = []string{"Bearer " + token}
		}
	}
	log.Println("Targets created")

	targeter := vegeta.NewStaticTargeter(targets...)
	attacker := vegeta.NewAttacker(vegeta.Connections(*numOfConnections), vegeta.Workers(uint64(*numOfWorkers)))

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
	log.Println(string(data))
}

func authOnServer(login, password, host string) string {
	creds := map[string]string{"login": login, "password": password}
	data, _ := json.Marshal(creds)
	resp, err := http.Post(host+authPrefix, contentJSONHeader, bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	respStr := struct {
		Status     string
		StatusCode int
		Data       interface{}
	}{}
	err = json.Unmarshal(bodyBytes, &respStr)
	if err != nil {
		panic(err)
	}
	if respStr.StatusCode != http.StatusOK {
		panic("can't login to server: " + respStr.Data.(string))
	}
	respData := respStr.Data.(map[string]interface{})
	return respData["token"].(string)
}
