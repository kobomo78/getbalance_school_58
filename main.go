package main

import "flag"

import "encoding/json"
import "os"

import "time"

import "net/http"
import "log"
import "golang.org/x/net/idna"
import "io/ioutil"
import "strings"
import "container/list"

import "regexp"

import "fmt"

type configCardData struct {
	Base_Type  bool   `json:"base_type"`
	FieldName  string `json:"field"`
	CardNumber string `json:"card_number"`
}
type ConfigChannel struct {
	Api_key string           `json:"api_key"`
	Cards   []configCardData `json:"card_data"`
}
type Config struct {
	Data []ConfigChannel `json:"channel"`
}

type cardData struct {
	FieldName   string
	CardBalance string
}

type dataChannel struct {
	Api_key string `json:"api_key"`
	Cards   []cardData
}

type SendData struct {
	Data []dataChannel
}

func main() {

	configPath := flag.String("d", "", "a string")

	flag.Parse()

	config := LoadConfiguration(*configPath + "base_food.conf")

	var balanceList list.List

	for j := 0; j < len(config.Data); j++ {
		for i := 1; i <= len(config.Data[j].Cards); i++ {
			var card cardData
			card.FieldName = config.Data[j].Cards[i-1].FieldName
			card.CardBalance = GetBalance(config.Data[j].Cards[i-1].CardNumber, config.Data[j].Cards[i-1].Base_Type)
			balanceList.PushBack(card)

		}
		if config.Data[j].Api_key != "" {
			SetThingSpeakBalance(balanceList, config.Data[j].Api_key)
		}
		balanceList.Init()
		if j < len(config.Data)-1 {
			time.Sleep(20 * time.Second)
		}

	}

}
func LoadConfiguration(file string) Config {

	var config Config
	configFile, err := os.Open(file)
	if err != nil {
		fmt.Println(err.Error())
	}

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

func GetBalance(cardNumber string, base_type bool) string {

	var str_query strings.Builder
	var p *idna.Profile

	// Формирование адреса в ponycode
	p = idna.New()
	ponycode, err := p.ToASCII("школа58.рф")

	str_query.WriteString("http://")
	str_query.WriteString(ponycode)
	str_query.WriteString("/ajax?card=" + cardNumber + "&act=FreeCheckBalance")

	resp, err := http.Get(str_query.String())
	if err != nil {
		log.Println(err)
		return ""
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return ""
	}

	var regexpStr string

	if base_type {
		regexpStr = "Основное .*? руб."
	} else {
		regexpStr = "Дополнительное .*? руб."
	}
	re := regexp.MustCompile(regexpStr)

	balance := re.Find(body)

	re = regexp.MustCompile(`[0-9]+\.?[0-9]*`)

	balance = re.Find(balance)

	if base_type {
		log.Println("Карта=" + cardNumber + " Основной питание=" + string(balance))
	} else {
		log.Println("Карта=" + cardNumber + " Дополнительное питание=" + string(balance))
	}

	return string(balance)

}
func SetThingSpeakBalance(balanceList list.List, apikey string) {

	var data map[string]string
	data = make(map[string]string)

	data["api_key"] = apikey

	for i := balanceList.Front(); i != nil; i = i.Next() {
		data[i.Value.(cardData).FieldName] = i.Value.(cardData).CardBalance
	}

	query, _ := json.Marshal(data)
	log.Println("Query=" + string(query))
	http.Post("https://api.thingspeak.com/update.json", "application/json", strings.NewReader(string(query)))

}
