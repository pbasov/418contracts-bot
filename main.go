package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/antihax/goesi"
	"github.com/dustin/go-humanize"
	"github.com/gorilla/sessions"
	"github.com/gregjones/httpcache"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

var CLIENT_ID = os.Getenv("CLIENT_ID")
var SECRET_KEY = os.Getenv("SECRET_KEY")
var STORAGE_DIR = os.Getenv("STORAGE_DIR")

var corp_str, _ = strconv.ParseInt(os.Getenv("CORP_ID"), 10, 32)
var CORP_ID int32 = int32(corp_str)

type ESI struct {
	api     *goesi.APIClient
	sso     *goesi.SSOAuthenticator
	store   *sessions.CookieStore
	scopes  []string
	token   *oauth2.Token
	items   map[int32]ItemSDE
	discord *DiscordBot
}

type ItemSDE struct {
	Name        map[string]string
	Description map[string]string
}

func main() {
	transport := httpcache.NewTransport(httpcache.NewMemoryCache())
	transport.Transport = &http.Transport{}
	httpClient := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	scopes := []string{"esi-contracts.read_corporation_contracts.v1", "publicData"}
	apiClient := goesi.NewAPIClient(httpClient, "Outfit 418 contracts discord bot (mail@weystrom.dev)")
	ssoClient := goesi.NewSSOAuthenticator(&http.Client{}, CLIENT_ID, SECRET_KEY, "http://localhost:4180/callback", scopes)
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

	log.Println("Loading SDE typeIds into memory, may take some time..")
	itemsDb := make(map[int32]ItemSDE)
	sdeFile, err := ioutil.ReadFile(STORAGE_DIR + "/typeIDs.yaml")
	if err != nil {
		log.Println("[ERROR] Failed to read the SDE file")
	}
	err = yaml.Unmarshal(sdeFile, &itemsDb)
	if err != nil {
		log.Printf("[ERROR] Error parning YAML file: %s\n", err)
	}
	log.Println("[DEBUG] SDE file load complete")

	api := ESI{
		api:    apiClient,
		sso:    ssoClient,
		store:  store,
		scopes: scopes,
		items:  itemsDb,
	}

	api.discord = &DiscordBot{Token: DISCORD_TOKEN, Notified: make(map[int32]bool)}
	api.discord.Startup = make(chan string)

	go api.discord.NewDiscordBot()

	defer api.discord.Session.Close()
	http.HandleFunc("/login", api.handleEsiLogin)
	http.HandleFunc("/callback", api.handleEsiCallback)
	http.HandleFunc("/raw", api.handleContractsRequest)
	http.HandleFunc("/parsed", api.handleParsedContractsRequest)
	http.HandleFunc("/sendit", api.handleSendContracts)

	// Block until discord bot is fully running
	log.Println(<-api.discord.Startup)
	api.readNotifiedContracts()
	go api.loop()

	log.Println("[DEBUG] Starting API server on port 4180")
	http.ListenAndServe(":4180", nil)
}

func (esi *ESI) loop() {
	for {
		err := esi.CheckContracts()
		if err != nil {
			log.Printf("[ERROR] Error in loop: %s\n", err)
		}
		log.Println("[INFO] Sleeping for 60s")
		time.Sleep(time.Second * 60)
	}
}

func (esi *ESI) CheckContracts() error {
	// Get token
	tokenSource, err := esi.tokenSource()
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to get token: %s", err)
	}
	// Get raw contracts
	ctx := context.WithValue(context.Background(), goesi.ContextOAuth2, tokenSource)
	rawContracts, err := esi.GetRawContracts(ctx)
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to get raw contracts: %s", err)
	}

	for _, contract := range rawContracts {
		if contract.Status != "outstanding" {
			continue
		}
		// TODO: multithread this
		// Create a pool of workers
		// Create a pool of channels
		// Create a channel for each contract
		if _, ok := esi.discord.Notified[contract.ContractId]; !ok {
			parsed_contract, err := esi.ParseContract(contract, ctx)
			if err != nil {
				log.Printf("[ERROR] Failed to parse contract: %s\n", err)
				continue
			}
			log.Printf("[DEBUG] Posting contract to discord: %v", contract.ContractId)
			esi.discord.Session.ChannelMessageSend(DISCORD_CHANNEL, fmt.Sprintf(
				":warning: New contract!\n**%s** from **%s**\nPrice: %s ISK \nEvepraisal: %s ISK - Jita Buy\n%s",
				parsed_contract.Title, parsed_contract.Issuer,
				humanize.Commaf(parsed_contract.Price),
				humanize.Commaf(parsed_contract.EPValue),
				parsed_contract.EPLink))
			esi.discord.Notified[contract.ContractId] = true
		}
	}
	err = esi.storeNotifiedContracts()
	if err != nil {
		return err
	}
	return nil
}
