package main

import (
	"encoding/gob"
	"encoding/json"
	"log"
	"os"

	"golang.org/x/oauth2"
)

func (esi *ESI) storeNotifiedContracts() error {
	f, err := os.Create(STORAGE_DIR + "/notified.json")
	if err != nil {
		log.Println("[ERROR] Failed to store notified contracts")
		return err
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(esi.discord.Notified)
	if err != nil {
		log.Println("[ERROR] Failed to store notified contracts")
		return err
	}
	return nil
}

func (esi *ESI) readNotifiedContracts() error {
	f, err := os.Open(STORAGE_DIR + "/notified.json")
	if err != nil {
		log.Println("[ERROR] Failed reading notified contracts")
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&esi.discord.Notified)
	if err != nil {
		log.Println("[ERROR] Failed parsing notified contracts")
		return err
	}
	return nil
}

func (esi *ESI) storeToken(token *oauth2.Token) error {
	f, err := os.Create(STORAGE_DIR + "/esitoken.bin")
	if err != nil {
		log.Println("[ERROR] Failed to open token file for writing")
		return err
	}
	defer f.Close()
	gob.NewEncoder(f).Encode(*token)
	return nil
}

func (esi *ESI) readToken() (*oauth2.Token, error) {
	var out oauth2.Token
	f, err := os.Open(STORAGE_DIR + "/esitoken.bin")
	if err != nil {
		log.Println("[ERROR] Failed to read token file")
		return &out, err
	}

	defer f.Close()
	enc := gob.NewDecoder(f)
	enc.Decode(&out)
	return &out, nil
}
