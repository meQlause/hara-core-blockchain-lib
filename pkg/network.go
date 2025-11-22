package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/meQlause/hara-core-blockchain-lib/internal/network"
	"github.com/meQlause/hara-core-blockchain-lib/utils"
)

type Network struct {
	URL     string
	version string
	id      uint8
	client  *network.RPCClient
}

func NewNetwork(url string, version string, id uint8) *Network {
	client, err := network.NewClientRPC(url)
	if err != nil {
		panic(err)
	}
	return &Network{
		URL:     url,
		version: version,
		id:      id,
		client:  client,
	}
}

func (n *Network) ChainID(ctx context.Context) (uint64, error) {
	id, err := n.client.Client.ChainID(ctx)
	if err != nil {
		return 0, err
	}
	return id.Uint64(), nil
}

func (n *Network) LatestBlock(ctx context.Context) (uint64, error) {
	block, err := n.client.Client.BlockNumber(ctx)
	if err != nil {
		return 0, err
	}
	return block, nil
}

func (n *Network) GasPrice(ctx context.Context) (*big.Int, error) {
	return n.client.Client.SuggestGasPrice(ctx)
}

func (n *Network) ClientVersion(ctx context.Context) (string, error) {
	var result string
	err := n.client.Client.Client().CallContext(ctx, &result, "web3_clientVersion")
	return result, err
}

func (n *Network) PendingNonce(ctx context.Context, address common.Address) (uint64, error) {
	return n.client.Client.PendingNonceAt(ctx, address)
}

func (n *Network) IsOnline(ctx context.Context) bool {
	_, err := n.LatestBlock(ctx)
	return err == nil
}

func (n *Network) Call(ctx context.Context, to common.Address, data string) (json.RawMessage, error) {
	if n.client == nil {
		return nil, fmt.Errorf("RPC client not initialized")
	}

	response, err := network.NewRPCBuilder[utils.EthCallParams](n.URL, n.client.HttpClient).
		BuildBody(n.version, n.id, "eth_call", utils.EthCallParams{
			Transaction: &utils.RPCPayload{
				To:   to.Hex(),
				Data: data,
			},
			BlockTag: "latest",
		}).
		SetHeader("Content-Type", "application/json").
		BuildRequest(ctx).
		Execute(ctx)

	if err != nil {
		return nil, fmt.Errorf("RPC call failed: %w", err)
	}

	return response.Result, nil
}

func (n *Network) SendRawTx(ctx context.Context, rawHex string) (json.RawMessage, error) {
	response, err := network.NewRPCBuilder[string](n.URL, n.client.HttpClient).
		BuildBody(n.version, n.id, "eth_sendRawTransaction", rawHex).
		SetHeader("Content-Type", "application/json").
		BuildRequest(ctx).
		Execute(ctx)

	if err != nil {
		return nil, fmt.Errorf("eth_sendRawTransaction failed: %w", err)
	}

	return response.Result, nil
}
