package utils

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func Namehash(name string) common.Hash {
	var node common.Hash
	if name == "" {
		return node
	}

	labels := strings.Split(name, ".")

	for i := len(labels) - 1; i >= 0; i-- {
		label := strings.ToLower(labels[i])
		labelHash := crypto.Keccak256Hash([]byte(label))
		concat := append(node.Bytes(), labelHash.Bytes()...)
		node = crypto.Keccak256Hash(concat)
	}

	return node
}
