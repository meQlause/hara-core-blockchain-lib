package network

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	utils "github.com/meQlause/hara-core-blockchain-lib/internal/utils"
)

type RPCBuilder[T any] struct {
	body          *utils.RPCRequest[T]
	requestToSend *http.Request
	client        *http.Client
	state         utils.BuilderState
	headers       http.Header
	url           string
}

func NewRPCBuilder[T any](url string, client *http.Client) *RPCBuilder[T] {
	return &RPCBuilder[T]{
		url:     url,
		client:  client,
		headers: make(http.Header),
		state:   utils.StateInitial,
	}
}

func (b *RPCBuilder[T]) BuildBody(version string, id uint64, method string, params T) *RPCBuilder[T] {
	if b.state != utils.StateInitial {
		panic(fmt.Sprintf("invalid state transition: cannot build body from state %d", b.state))
	}

	b.body = &utils.RPCRequest[T]{
		JsonRPC: version,
		ID:      id,
		Method:  method,
		Params:  params,
	}

	b.state = utils.StateBodyBuilt
	return b
}

func (b *RPCBuilder[T]) SetHeader(key, value string) *RPCBuilder[T] {
	if b.state != utils.StateBodyBuilt && b.state != utils.StateHeadersSet {
		panic(fmt.Sprintf("invalid state transition: cannot set headers from state %d", b.state))
	}

	b.headers.Set(key, value)
	b.state = utils.StateHeadersSet
	return b
}

func (b *RPCBuilder[T]) BuildRequest(ctx context.Context) *RPCBuilder[T] {
	if b.state != utils.StateHeadersSet {
		panic(fmt.Sprintf("invalid state: cannot build request from state %d, headers must be set first", b.state))
	}

	bodyBytes, err := json.Marshal(b.body)
	if err != nil {
		panic(fmt.Errorf("failed to marshal RPC body: %w", err))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		panic(fmt.Errorf("failed to create request: %w", err))
	}

	for k, vals := range b.headers {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}

	b.requestToSend = req
	b.state = utils.StateRequestBuilt
	return b
}

func (b *RPCBuilder[T]) Execute(ctx context.Context) (*utils.RPCResponse, error) {
	if b.state != utils.StateRequestBuilt {
		return nil, fmt.Errorf("invalid state: cannot execute from state %d, request must be built first", b.state)
	}

	resp, err := b.client.Do(b.requestToSend)
	if err != nil {
		return nil, fmt.Errorf("RPC call failed: %w", err)
	}
	defer resp.Body.Close()

	var result utils.RPCResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode RPC response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", result.Error.Code, result.Error.Message)
	}

	b.state = utils.StateExecuted
	return &result, nil
}

func (b *RPCBuilder[T]) GetState() utils.BuilderState {
	return b.state
}

func (b *RPCBuilder[T]) Reset() *RPCBuilder[T] {
	b.body = nil
	b.requestToSend = nil
	b.headers = make(http.Header)
	b.state = utils.StateInitial
	return b
}
