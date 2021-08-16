package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

const EP_URL = "https://evepraisal.com/appraisal.json"

type EvePraisal struct {
	Appraisal struct {
		Totals struct {
			Buy    float64
			Sell   float64
			Volume float64
		}
	}
}

func (esi *ESI) getEvePraisal(items []ContractItems) (EvePraisal, string, error) {
	ep := EvePraisal{}
	url := paramEncode(items)

	client := http.Client{}
	request := http.Request{
		Method: "POST",
		URL:    url,
		Header: make(http.Header),
	}
	request.Header.Set("User-Agent", "Outfit 418 contracts appraiser, mail@weystrom.dev")

	resp, err := client.Do(&request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	epId := resp.Header.Get("x-appraisal-id")
	if epId == "" {
		return ep, "", fmt.Errorf("no ep id found in header")
	}
	epLink := fmt.Sprintf("https://evepraisal.com/a/%s", epId)
	json.NewDecoder(resp.Body).Decode(&ep)
	return ep, epLink, nil
}

func paramEncode(items []ContractItems) *url.URL {
	var stringItems string
	for _, item := range items {
		stringItems += fmt.Sprintf("%s %v\n", item.Name, item.Quantity)
	}

	baseUrl, err := url.Parse(EP_URL)
	if err != nil {
		log.Println("Malformed URL: ", err.Error())
		return nil
	}

	queryParams := url.Values{}
	queryParams.Add("raw_textarea", stringItems)
	queryParams.Add("market", "jita")
	queryParams.Add("persist", "yes")

	baseUrl.RawQuery = queryParams.Encode()
	return baseUrl
}
