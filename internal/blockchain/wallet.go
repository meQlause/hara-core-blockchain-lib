package blockchain

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/meQlause/hara-core-blockchain-lib/utils"
)

type Wallet struct {
	seed string
}

func NewWallet(seed string) *Wallet {
	return &Wallet{seed: seed}
}

func (w *Wallet) derivePrivateKey() (*ecdsa.PrivateKey, error) {
	hash := crypto.Keccak256Hash([]byte(w.seed))

	privKey, err := crypto.ToECDSA(hash.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to derive private key: %w", err)
	}

	return privKey, nil
}

func (w *Wallet) GetPrivateKey() (string, error) {
	privKey, err := w.derivePrivateKey()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("0x%x", crypto.FromECDSA(privKey)), nil
}

func (w *Wallet) GetAddress() (string, error) {
	privKey, err := w.derivePrivateKey()
	if err != nil {
		return "", err
	}

	address := crypto.PubkeyToAddress(privKey.PublicKey)
	return address.Hex(), nil
}

func (w *Wallet) SignEIP191(message string) (*utils.SignResult, error) {
	privKey, err := w.derivePrivateKey()
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(message))
	prefixedMessage := []byte(prefix + message)

	hash := crypto.Keccak256Hash(prefixedMessage)

	sig, err := crypto.Sign(hash.Bytes(), privKey)
	if err != nil {
		return nil, err
	}

	r := sig[:32]
	s := sig[32:64]
	v := sig[64] + 27

	return &utils.SignResult{
		Message:     message,
		MessageHash: hash.Hex(),
		R:           "0x" + common.Bytes2Hex(r),
		S:           "0x" + common.Bytes2Hex(s),
		V:           v,
		Signature:   "0x" + common.Bytes2Hex(sig),
	}, nil
}

func (w *Wallet) SignTransaction(tx *types.Transaction, chainID *big.Int) ([]byte, error) {
	privKey, err := w.derivePrivateKey()
	if err != nil {
		return nil, err
	}

	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, privKey)
	if err != nil {
		return nil, err
	}

	return signedTx.MarshalBinary()
}
