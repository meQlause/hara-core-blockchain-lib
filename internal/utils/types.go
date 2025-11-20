package utils

import "encoding/json"

type RPCRequest[T any] struct {
	JsonRPC string `json:"jsonrpc"`
	ID      uint64 `json:"id"`
	Method  string `json:"method"`
	Params  T      `json:"params"`
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
