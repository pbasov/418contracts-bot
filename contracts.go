package main

import (
	"context"
	"fmt"
	"log"

	"github.com/antihax/goesi/esi"
)

type Contract struct {
	Issuer  string
	Price   float64
	Title   string
	Items   []ContractItems
	EPValue float64
	EPLink  string
	Expiry  string
}

type ContractItems struct {
	Name     string `json:"name"`
	Quantity int32  `json:"quantity"`
}

// GetRawContracts returns a list of contracts with unresolved issuer and items
func (esi *ESI) GetRawContracts(ctx context.Context) ([]esi.GetCorporationsCorporationIdContracts200Ok, error) {
	log.Println("[DEBUG] Pulling contracts")
	contracts, _, err := esi.api.ESI.ContractsApi.GetCorporationsCorporationIdContracts(ctx, CORP_ID, nil)
	if err != nil {
		return nil, err
	}
	return contracts, nil
}

// ParseContracts resolves the contract issuer and items, gets evepraisal value and link
func (esi *ESI) ParseContract(contract esi.GetCorporationsCorporationIdContracts200Ok, ctx context.Context) (Contract, error) {
	log.Printf("[DEBUG] Parsing contract %v", contract.ContractId)
	issuer, _, err := esi.api.ESI.CharacterApi.GetCharactersCharacterId(ctx, contract.IssuerId, nil)
	if err != nil {
		return Contract{}, fmt.Errorf("failed to get issuer: %v", err)
	}
	log.Printf("[DEBUG] Gettings items from contract %v", contract.ContractId)
	detailed, _, err := esi.api.ESI.ContractsApi.GetCorporationsCorporationIdContractsContractIdItems(ctx, contract.ContractId, CORP_ID, nil)
	if err != nil {
		return Contract{}, fmt.Errorf("failed to get items for contract %v", contract.ContractId)
	}
	items := make([]ContractItems, 0)
	for _, item := range detailed {
		item_detail := ContractItems{
			Name:     esi.items[item.TypeId].Name["en"],
			Quantity: item.Quantity,
		}
		items = append(items, item_detail)
	}
	epValue, epLink, err := esi.getEvePraisal(items)
	iskValue := epValue.Appraisal.Totals.Buy
	if err != nil {
		return Contract{}, fmt.Errorf("failed to get evepraisal value: %v", err)
	}
	expiry := contract.DateExpired.Format("2006-01-02")
	return Contract{
		Issuer:  issuer.Name,
		Items:   items,
		Title:   contract.Title,
		Expiry:  expiry,
		Price:   contract.Price,
		EPValue: iskValue,
		EPLink:  epLink,
	}, nil
}
