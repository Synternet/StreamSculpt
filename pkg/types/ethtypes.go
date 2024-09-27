package types

import (
	"encoding/json"
	"math/big"
)

type EthLogEvent struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	BlockNumber      string   `json:"blockNumber"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	BlockHash        string   `json:"blockHash"`
	LogIndex         string   `json:"logIndex"`
	Removed          bool     `json:"removed"`
}

type DecodedEthLogEvent struct {
	Address          string                 `json:"address"`
	Sig              string                 `json:"sig"`
	Topics           []string               `json:"topics"`
	Data             map[string]interface{} `json:"data"`
	BlockNumber      string                 `json:"blockNumber"`
	TransactionHash  string                 `json:"transactionHash"`
	TransactionIndex string                 `json:"transactionIndex"`
	BlockHash        string                 `json:"blockHash"`
	LogIndex         string                 `json:"logIndex"`
	Removed          bool                   `json:"removed"`
}

func (t *DecodedEthLogEvent) AsJSON() []byte {
	json, err := json.Marshal(t)
	if err != nil {
		return []byte{}
	}
	return json
}

type Swap struct {
	Amount0In  *big.Int
	Amount1In  *big.Int
	Amount0Out *big.Int
	Amount1Out *big.Int
}
