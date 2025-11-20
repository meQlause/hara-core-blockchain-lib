package pkg

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// NonceProvider defines interface for nonce retrieval
type NonceProvider interface {
	GetNonce(ctx context.Context, address common.Address) (uint64, error)
}

// Config holds transaction parameters
type Config struct {
	ContractAddress common.Address
	ABI             abi.ABI
	MethodName      string
	MethodParams    []interface{}
	ValueInWei      *big.Int
	GasPriceGwei    *big.Int // nil = auto
	PrivateKey      *ecdsa.PrivateKey
	MaxRetries      int
	UseNonceService bool
}

// TxSender handles Ethereum transactions with retry logic
type TxSender struct {
	client        *ethclient.Client
	nonceProvider NonceProvider
	httpClient    *http.Client
}

// NewTxSender creates a new transaction sender
func NewTxSender(rpcURL string, nonceProvider NonceProvider) (*TxSender, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	return &TxSender{
		client:        client,
		nonceProvider: nonceProvider,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// NodeNonceProvider fetches nonce from Ethereum node
type NodeNonceProvider struct {
	client *ethclient.Client
}

func NewNodeNonceProvider(client *ethclient.Client) *NodeNonceProvider {
	return &NodeNonceProvider{client: client}
}

func (p *NodeNonceProvider) GetNonce(ctx context.Context, address common.Address) (uint64, error) {
	nonce, err := p.client.PendingNonceAt(ctx, address)
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce from node: %w", err)
	}
	return nonce, nil
}

// ServiceNonceProvider fetches nonce from external service
type ServiceNonceProvider struct {
	serviceURL string
	httpClient *http.Client
}

func NewServiceNonceProvider(serviceURL string) *ServiceNonceProvider {
	return &ServiceNonceProvider{
		serviceURL: serviceURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (p *ServiceNonceProvider) GetNonce(ctx context.Context, address common.Address) (uint64, error) {
	url := p.serviceURL + address.Hex()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch nonce from service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("nonce service returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var result struct {
		Nonce uint64 `json:"nonce"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("failed to parse nonce response: %w", err)
	}

	return result.Nonce, nil
}

// ErrorClassifier helps identify error types
type ErrorClassifier struct {
	nonceErrorRegex  *regexp.Regexp
	revertErrorRegex *regexp.Regexp
}

func NewErrorClassifier() *ErrorClassifier {
	return &ErrorClassifier{
		nonceErrorRegex:  regexp.MustCompile(`(?i)nonce|replacement|already used|underpriced|known transaction`),
		revertErrorRegex: regexp.MustCompile(`(?i)revert(?:ed)?(?: with reason string)?:?\s*(.*)`),
	}
}

func (ec *ErrorClassifier) IsNonceError(err error) bool {
	if err == nil {
		return false
	}
	return ec.nonceErrorRegex.MatchString(err.Error())
}

func (ec *ErrorClassifier) ExtractRevertReason(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	matches := ec.revertErrorRegex.FindStringSubmatch(err.Error())
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1]), true
	}
	if strings.Contains(strings.ToLower(err.Error()), "revert") {
		return err.Error(), true
	}
	return "", false
}

// SendTransaction sends a transaction with retry logic
func (s *TxSender) SendTransaction(ctx context.Context, cfg Config) (*types.Receipt, error) {
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	if cfg.ValueInWei == nil {
		cfg.ValueInWei = big.NewInt(0)
	}

	address := crypto.PubkeyToAddress(cfg.PrivateKey.PublicKey)
	chainID, err := s.client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Encode method call
	data, err := cfg.ABI.Pack(cfg.MethodName, cfg.MethodParams...)
	if err != nil {
		return nil, fmt.Errorf("failed to encode method call: %w", err)
	}

	classifier := NewErrorClassifier()
	var lastNonce *uint64

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Get nonce
		var nonce uint64
		if lastNonce != nil && attempt > 0 {
			nonce = *lastNonce
		} else {
			nonce, err = s.nonceProvider.GetNonce(ctx, address)
			if err != nil {
				return nil, fmt.Errorf("failed to get nonce: %w", err)
			}
			lastNonce = &nonce
		}

		// Estimate gas
		gasLimit, err := s.client.EstimateGas(ctx, ethereum.CallMsg{
			From:  address,
			To:    &cfg.ContractAddress,
			Value: cfg.ValueInWei,
			Data:  data,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas: %w", err)
		}
		gasLimit = uint64(float64(gasLimit) * 1.1) // 10% buffer

		// Get gas price
		gasPrice := cfg.GasPriceGwei
		if gasPrice == nil {
			gasPrice, err = s.client.SuggestGasPrice(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get gas price: %w", err)
			}
		} else {
			gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(1e9)) // Gwei to Wei
		}

		// Create and sign transaction
		tx := types.NewTransaction(
			nonce,
			cfg.ContractAddress,
			cfg.ValueInWei,
			gasLimit,
			gasPrice,
			data,
		)

		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), cfg.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to sign transaction: %w", err)
		}

		// Send transaction
		err = s.client.SendTransaction(ctx, signedTx)
		if err != nil {
			// Check for revert errors (fail immediately)
			if reason, isRevert := classifier.ExtractRevertReason(err); isRevert {
				return nil, fmt.Errorf("transaction reverted: %s", reason)
			}

			// Check for nonce errors (retry)
			if classifier.IsNonceError(err) {
				if attempt == cfg.MaxRetries {
					return nil, fmt.Errorf("max retries reached, last nonce error: %w", err)
				}

				// Exponential backoff
				delay := time.Duration(1<<uint(attempt)) * time.Second
				fmt.Printf("[RETRY] Attempt %d/%d failed: %v. Retrying in %v...\n",
					attempt+1, cfg.MaxRetries+1, err, delay)

				time.Sleep(delay)
				lastNonce = nil // Force nonce refresh
				continue
			}

			// Other errors
			return nil, fmt.Errorf("failed to send transaction: %w", err)
		}

		fmt.Printf("[INFO] Transaction sent. Hash: %s\n", signedTx.Hash().Hex())

		// Wait for receipt
		receipt, err := s.waitForReceipt(ctx, signedTx.Hash())
		if err != nil {
			return nil, err
		}

		fmt.Printf("[SUCCESS] Transaction mined in block %d\n", receipt.BlockNumber.Uint64())
		time.Sleep(1 * time.Second) // 1 second delay
		return receipt, nil
	}

	return nil, fmt.Errorf("unexpected failure after %d attempts", cfg.MaxRetries)
}

func (s *TxSender) waitForReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for transaction receipt")
		case <-ticker.C:
			receipt, err := s.client.TransactionReceipt(ctx, txHash)
			if err == nil {
				return receipt, nil
			}
			// Continue waiting if receipt not found yet
		}
	}
}

// Close closes the Ethereum client connection
func (s *TxSender) Close() {
	s.client.Close()
}
