package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/meQlause/hara-core-blockchain-lib/internal/blockchain"
	"github.com/meQlause/hara-core-blockchain-lib/utils"
)

type Blockchain struct {
	Wallet  *blockchain.Wallet
	Network *Network
	abi     map[string]*utils.ABIResult
}

func NewBlockchain(wallet *blockchain.Wallet, network *Network) *Blockchain {
	return &Blockchain{
		Wallet:  wallet,
		Network: network,
		abi:     make(map[string]*utils.ABIResult),
	}
}

func (bc *Blockchain) GetAddressABI(ctx context.Context, uri string) (*utils.ABIResult, error) {
	if v, ok := bc.abi[uri]; ok {
		return v, nil
	}

	node := utils.Namehash(uri)

	registry, err := blockchain.NewContract(utils.ContractConfig{
		ABIJSON:         utils.HNSRegistryABI,
		Address:         utils.ENSAddress,
		CallBackend:     bc.Network.client.Client,
		TransactBackend: bc.Network.client.Client,
		LogBackend:      bc.Network.client.Client,
	})
	if err != nil {
		return nil, fmt.Errorf("registry contract init: %w", err)
	}

	var resolverOut []any
	err = registry.Bound.Call(
		nil,
		&resolverOut,
		"resolver",
		node,
	)
	if err != nil {
		return nil, fmt.Errorf("resolver() call: %w", err)
	}

	resolverAddress := resolverOut[0].(common.Address)
	resolver, err := blockchain.NewContract(utils.ContractConfig{
		ABIJSON:         utils.HNSResolverABI,
		Address:         resolverAddress.Hex(),
		CallBackend:     bc.Network.client.Client,
		TransactBackend: bc.Network.client.Client,
		LogBackend:      bc.Network.client.Client,
	})
	if err != nil {
		return nil, fmt.Errorf("resolver contract init: %w", err)
	}

	var addrOut []any
	err = resolver.Bound.Call(
		nil,
		&addrOut,
		"addr",
		node,
	)
	if err != nil {
		return nil, fmt.Errorf("addr() call: %w", err)
	}

	address := addrOut[0].(common.Address)
	var abiOut []any
	err = resolver.Bound.Call(
		nil,
		&abiOut,
		"ABI",
		node,
		big.NewInt(1),
	)
	if err != nil {
		return nil, fmt.Errorf("ABI() call: %w", err)
	}

	contentType := abiOut[0].(*big.Int).Uint64()
	abiBytes := abiOut[1].([]byte)

	var abiJSON any
	if contentType != 0 && len(abiBytes) > 0 {
		if err := json.Unmarshal(abiBytes, &abiJSON); err != nil {
			return nil, fmt.Errorf("invalid ABI JSON: %w", err)
		}
	}

	result := &utils.ABIResult{
		Address: address.Hex(),
		ABI:     abiJSON,
	}

	bc.abi[uri] = result
	return result, nil
}
