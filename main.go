// SPDX-FileCopyrightText: Contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/headzoo/surf"
)

var (
	jsonResponse []byte
	config       Config
)

// Config store the username and password for the scraper
type MQTTConfig struct {
	Host, Port, User, Pass, AutoDiscovery, Address string
}

type Config struct {
	Username, Password, Port, LoginUrl, ServePath string
	OpenBrowser                                   bool
	RefreshSeconds                                int
	MQTT                                          MQTTConfig
}

// SolarOSReading Data
type SolarOSReading struct {
	InstantaneousPower string `json:"instant_power"`
	LifeMeter          string `json:"life_meter"`
	// MoneySaved         string `json:"money_saved"`
	TreesSaved string `json:"trees_saved"`
	OilOffset  string `json:"oil_offset"`
	CO2Offset  string `json:"co2_offset"`
	// SystemSize string `json:"size"`
	// Inverter   string `json:"inverter"`
	LastUpdate int64 `json:"last_update"`
}

func dataURL(cookie string, page string, watch []string) string {
	// return url of javascript which contains data
	now := time.Now().Unix()
	url := fmt.Sprintf("https://solaros.datareadings.com/comet_request/15725446441/%v/%v?%v=%v&_=%v", cookie, page, watch[0], watch[1], now)
	return url
}

func toWatch(body string) []string {
	// return lift_toWatch as 2 strings
	page := strings.SplitN(body, "var lift_toWatch =", 2)[1]
	page = strings.SplitN(page, ";", 2)[0]
	page = strings.TrimSpace(page)
	page = strings.Trim(page, "{}")
	page = strings.Replace(page, "\"", "", -1)
	watchJSON := strings.SplitN(page, ":", 2)
	watchJSON[1] = strings.TrimSpace(watchJSON[1])
	return watchJSON
}

func liftPage(body string) string {
	// return lift_page as a string
	page := strings.SplitN(body, "var lift_page =", 2)[1]
	page = strings.SplitN(page, ";", 2)[0]
	page = strings.TrimSpace(page)
	page = strings.Trim(page, "\"")
	return page
}

func treesSaved(body string) string {
	// return treesSaved reading
	r, _ := regexp.Compile("benefitValue1(.*?)span")
	str := r.FindString(body)
	r, _ = regexp.Compile(">(.*?)<")
	str = r.FindString(str)
	str = strings.Trim(str, "<>")
	str = strings.Replace(str, ",", "", -1)
	return str
}
func oilOffset(body string) string {
	// return oilOffset reading
	r, _ := regexp.Compile("benefitValue2(.*?)span")
	str := r.FindString(body)
	r, _ = regexp.Compile(">(.*?)<")
	str = r.FindString(str)
	str = strings.Trim(str, "<>")
	str = strings.Replace(str, ",", "", -1)
	return str
}
func co2Offset(body string) string {
	// return co2Offset reading
	r, _ := regexp.Compile("benefitValue3(.*?)span")
	str := r.FindString(body)
	r, _ = regexp.Compile(">(.*?)<")
	str = r.FindString(str)
	str = strings.Trim(str, "<>")
	str = strings.Replace(str, ",", "", -1)
	return str
}
func instantMeter(body string) string {
	// return instantMeter reading
	r, _ := regexp.Compile("instantValue(.*?)span")
	str := r.FindString(body)
	r, _ = regexp.Compile(">(.*?)<")
	str = r.FindString(str)
	str = strings.Trim(str, "<>")
	str = strings.Replace(str, ",", "", -1)
	if str == "Offline" {
		str = "0"
	}
	return str
}
func dollarsSaved(body string) string {
	// return dollarsSaved reading
	r, _ := regexp.Compile("dollarsSaved(.*?)</span")
	str := r.FindString(body)
	r, _ = regexp.Compile(">(.*?)<")
	str = r.FindString(str)
	str = strings.Trim(str, "><")
	str = strings.Trim(str, "$")
	return str
}

func meter(body string) string {
	// return lifeMeter reading
	r, _ := regexp.Compile("lifetimeMeter(.*?)div")
	str := r.FindString(body)
	r, _ = regexp.Compile(">(.*?)<")
	str = r.FindString(str)
	str = strings.Trim(str, "<>")
	str = strings.Replace(str, ",", "", -1)
	return str
}

func size(body string) string {
	// return size reading
	r, _ := regexp.Compile("systemSize(.*?)span")
	str := r.FindString(body)
	r, _ = regexp.Compile(">(.*?)<")
	str = r.FindString(str)
	str = strings.Trim(str, "<>")
	return str
}
func inverter(body string) string {
	// return inverter reading
	r, _ := regexp.Compile("inverterType(.*?)span")
	str := r.FindString(body)
	r, _ = regexp.Compile(">(.*?)<")
	str = r.FindString(str)
	str = strings.Trim(str, "<>")
	return str
}
func getScript() (string, error) {
	// logs into SolarOS and scrapes the page, returning a usable URL
	bow := surf.NewBrowser()
	err := bow.Open(config.LoginUrl)
	if err != nil || bow.StatusCode() != 200 {
		return "", err
	}
	fm := bow.Forms()[1]
	username := fm.Dom().Find("input").Nodes[0].Attr[3].Val
	password := fm.Dom().Find("input").Nodes[1].Attr[3].Val
	fm.Input(username, config.Username)
	fm.Input(password, config.Password)
	err = fm.Submit()
	if err != nil || bow.StatusCode() != 200 {
		return "", nil
	}
	cookie := fmt.Sprintf("%v", bow.SiteCookies()[0])
	cookie = strings.SplitN(cookie, "=", 2)[1]
	body := bow.Body()
	page := liftPage(body)
	watch := toWatch(body)
	url := dataURL(cookie, page, watch)
	return url, nil
}

func scrape() (SolarOSReading, error) {
	bow := surf.NewBrowser()
	var reading SolarOSReading
	url, err := getScript()
	if err != nil {
		return SolarOSReading{}, err
	}
	err = bow.Open(url)
	if err != nil || bow.StatusCode() != 200 {
		return SolarOSReading{}, err
	}
	body := bow.Body()
	lifemeter := meter(body)
	if lifemeter != "" {
		reading.LifeMeter = lifemeter
	}
	// reading.MoneySaved = dollarsSaved(body)
	reading.InstantaneousPower = instantMeter(body)
	reading.TreesSaved = treesSaved(body)
	reading.OilOffset = oilOffset(body)
	reading.CO2Offset = co2Offset(body)
	// reading.SystemSize = size(body)
	// reading.Inverter = inverter(body)
	reading.LastUpdate = time.Now().Unix()
	return reading, nil
}

func query(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(jsonResponse)
}

func serve() {
	log.Println(fmt.Sprintf("Starting server on port %v", config.Port))
	http.HandleFunc(config.ServePath, query)
	http.ListenAndServe(fmt.Sprintf(":%v", config.Port), nil)
}
func makeJSON() {
	reading, err := scrape()
	if err != nil {
		log.Println(err)
		return
	}
	jsn, err := json.MarshalIndent(reading, "", "\t")
	if err != nil {
		log.Println(err)
		return
	}
	jsonResponse = jsn
	if config.MQTT.AutoDiscovery != "" {
		if client == nil || !client.IsConnected() {
			login()
		}
		publish()
	}
}

func init() {
	path := "config.toml"
	if _, err := toml.DecodeFile(path, &config); err != nil {
		log.Fatal(err)
	}
}
func main() {
	if config.Username == "" || config.Password == "" {
		log.Println("Username or Password is missing; please check config.toml")
		os.Exit(1)
	}
	makeJSON()
	if config.ServePath == "" || config.Port == "" {
		log.Println("ServePath or Port not configured, disabling server")
	} else {
		go serve()
		url := fmt.Sprintf("http://%v:%v%v", GetOutboundIP(), config.Port, config.ServePath)
		if config.OpenBrowser {
			log.Println(fmt.Sprintf("Opening %v", url))
			go openbrowser(url)
		} else {
			log.Println(fmt.Sprintf("Access %v to see json data", url))
		}
	}
	tickChan := time.NewTicker(time.Second * time.Duration(config.RefreshSeconds)).C
	for {
		select {
		case <-tickChan:
			makeJSON()
		}
	}

}
