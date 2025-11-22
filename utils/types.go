package utils

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type ABIResult struct {
	Address string
	ABI     any
}

type RPCRequest[T any] struct {
	JsonRPC string `json:"jsonrpc"`
	ID      uint8  `json:"id"`
	Method  string `json:"method"`
	Params  T      `json:"params"`
}

type ContractConfig struct {
	ABIJSON         string
	Address         string
	CallBackend     bind.ContractCaller
	TransactBackend bind.ContractTransactor
	LogBackend      bind.ContractFilterer
}

type SignResult struct {
	Message     string `json:"message"`
	MessageHash string `json:"messageHash"`
	R           string `json:"r"`
	S           string `json:"s"`
	V           uint8  `json:"v"`
	Signature   string `json:"signature"`
}

type RPCPayload struct {
	To   string `json:"to"`
	Data string `json:"data"`
}

type EthCallParams struct {
	Transaction *RPCPayload `json:"transaction"`
	BlockTag    string      `json:"blockTag"`
}

type RPCError struct {
	Code    uint8  `json:"code"`
	Message string `json:"message"`
}

type RPCResponse struct {
	JsonRPC string          `json:"jsonrpc"`
	ID      uint64          `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *RPCError       `json:"error"`
}
