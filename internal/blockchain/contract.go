package blockchain

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/meQlause/hara-core-blockchain-lib/utils"
)

type Contract struct {
	Address  common.Address
	ABI      abi.ABI
	Caller   bind.ContractCaller
	Transact bind.ContractTransactor
	Filter   bind.ContractFilterer
	Bound    *bind.BoundContract
}

func NewContract(cfg utils.ContractConfig) (*Contract, error) {
	parsedABI, err := abi.JSON(strings.NewReader(cfg.ABIJSON))
	if err != nil {
		return nil, err
	}

	address := common.HexToAddress(cfg.Address)

	bound := bind.NewBoundContract(
		address,
		parsedABI,
		cfg.CallBackend,
		cfg.TransactBackend,
		cfg.LogBackend,
	)

	return &Contract{
		Address:  address,
		ABI:      parsedABI,
		Caller:   cfg.CallBackend,
		Transact: cfg.TransactBackend,
		Filter:   cfg.LogBackend,
		Bound:    bound,
	}, nil
}
