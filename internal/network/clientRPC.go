package network

import (
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

type RPCClient struct {
	URL        string
	Client     *ethclient.Client
	HttpClient *http.Client
}

func NewClientRPC(url string) (*RPCClient, error) {
	if client, err := ethclient.Dial(url); err != nil {
		return nil, err
	} else {
		return &RPCClient{
			URL:        url,
			Client:     client,
			HttpClient: &http.Client{Timeout: 10 * time.Second},
		}, nil
	}
}
