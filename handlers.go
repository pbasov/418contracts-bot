package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/antihax/goesi"
)

// /raw
func (esi *ESI) handleContractsRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tokenSource, err := esi.tokenSource()
	if err != nil {
		log.Println("No token found, redirecting to login page")
		http.Redirect(w, r, "http://localhost:4180/login", http.StatusFound)
		return
	}
	ctx := context.WithValue(context.Background(), goesi.ContextOAuth2, tokenSource)
	contracts, err := esi.GetRawContracts(ctx)
	if err != nil {
		http.Error(w, "Token Failure", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(contracts)
}

// /parsed
func (esi *ESI) handleParsedContractsRequest(w http.ResponseWriter, r *http.Request) {
	// Get token
	tokenSource, err := esi.tokenSource()
	if err != nil {
		log.Println("No token found, redirecting to login page")
		http.Redirect(w, r, "http://localhost:4180/login", http.StatusFound)
		return
	}

	// Get raw contracts
	ctx := context.WithValue(context.Background(), goesi.ContextOAuth2, tokenSource)
	contracts, err := esi.GetRawContracts(ctx)
	if err != nil {
		http.Error(w, "Token Failure", http.StatusInternalServerError)
		return
	}

	// Parse additional data
	parsed := make([]Contract, 0)
	for _, contract := range contracts {
		if contract.Status != "outstanding" {
			continue
		}
		parsed_contract, err := esi.ParseContract(contract, ctx)
		if err != nil {
			continue
		}
		parsed = append(parsed, parsed_contract)
	}

	// Return json
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parsed)
}

// /sendit
func (esi *ESI) handleSendContracts(w http.ResponseWriter, r *http.Request) {
	esi.CheckContracts()
	fmt.Fprintf(w, "Sent")
}
