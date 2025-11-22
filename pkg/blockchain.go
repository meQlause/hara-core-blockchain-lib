package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/meQlause/hara-core-blockchain-lib/internal/blockchain"
	"github.com/meQlause/hara-core-blockchain-lib/utils"
)

type Blockchain struct {
	Wallet     *blockchain.Wallet
	Network    *Network
	ChainID    int64
	haraAbiENS map[string]*utils.ABIResult
}

func NewBlockchain(wallet *blockchain.Wallet, network *Network, chainId int64) *Blockchain {
	return &Blockchain{
		Wallet:     wallet,
		Network:    network,
		ChainID:    chainId,
		haraAbiENS: make(map[string]*utils.ABIResult),
	}
}

func (bc *Blockchain) GetAddressABI(ctx context.Context, uri string) (*utils.ABIResult, error) {
	if v, ok := bc.haraAbiENS[uri]; ok {
		return v, nil
	}

	node := utils.Namehash(uri)
	resolverAddress, err := bc.getResolver(node)
	if err != nil {
		return nil, fmt.Errorf("resolver contract init: %w", err)
	}

	resolver, address, err := bc.getAddressFromResolver(*resolverAddress, node)
	if err != nil {
		return nil, fmt.Errorf("resolver contract init: %w", err)
	}

	result, err := bc.getABi(node, *address, resolver)
	if err != nil {
		return nil, fmt.Errorf("resolver contract init: %w", err)
	}

	bc.haraAbiENS[uri] = result
	return result, nil
}

func (bc *Blockchain) BuildTx(p utils.TransactionParams) *types.Transaction {
	return types.NewTransaction(
		p.Nonce,
		p.To,
		p.Value,
		p.GasLimit,
		p.GasPrice,
		p.Data,
	)
}

func (bc *Blockchain) CallContract(
	ctx context.Context,
	c *blockchain.Contract,
	method string,
	args []any,
) ([]byte, error) {
	data, err := c.ABI.Pack(method, args...)
	if err != nil {
		return nil, fmt.Errorf("abi pack error: %w", err)
	}
	
	raw := "0x" + common.Bytes2Hex(data)
	resp, err := bc.Network.Call(ctx, c.Address, raw)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (bc *Blockchain) SendContractTx(
	ctx context.Context,
	tx *types.Transaction,
) (string, error) {

	signedBytes, err := bc.Wallet.SignTransaction(tx, big.NewInt(bc.ChainID))
	if err != nil {
		return "", fmt.Errorf("sign error: %w", err)
	}

	raw := "0x" + common.Bytes2Hex(signedBytes)

	resp, err := bc.Network.SendRawTx(ctx, raw)
	if err != nil {
		return "", err
	}

	return string(resp), nil
}

func (bc *Blockchain) getResolver(node common.Hash) (*common.Address, error) {
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

	return &resolverAddress, nil
}

func (bc *Blockchain) getAddressFromResolver(resolverAddress common.Address, node common.Hash) (*blockchain.Contract, *common.Address, error) {
	resolver, err := blockchain.NewContract(utils.ContractConfig{
		ABIJSON:         utils.HNSResolverABI,
		Address:         resolverAddress.Hex(),
		CallBackend:     bc.Network.client.Client,
		TransactBackend: bc.Network.client.Client,
		LogBackend:      bc.Network.client.Client,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("resolver contract init: %w", err)
	}

	var addrOut []any
	err = resolver.Bound.Call(
		nil,
		&addrOut,
		"addr",
		node,
	)

	if err != nil {
		return nil, nil, fmt.Errorf("addr() call: %w", err)
	}

	address := addrOut[0].(common.Address)

	return resolver, &address, nil
}

func (bc *Blockchain) getABi(node common.Hash, address common.Address, resolver *blockchain.Contract) (*utils.ABIResult, error) {
	var abiOut []any
	err := resolver.Bound.Call(
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

	return &utils.ABIResult{
		Address: address.Hex(),
		ABI:     abiJSON,
	}, nil
}
