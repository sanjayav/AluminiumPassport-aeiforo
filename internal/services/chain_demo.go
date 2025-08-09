package services

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Minimal ABI with only demo methods we call.
const aluminiumPassportDemoABI = `[
  {"inputs":[{"internalType":"string","name":"batchId","type":"string"},{"internalType":"string","name":"cid","type":"string"}],"name":"registerUpstreamBatch","outputs":[],"stateMutability":"nonpayable","type":"function"},
  {"inputs":[{"internalType":"string","name":"orgId","type":"string"},{"internalType":"string","name":"upstreamBatchId","type":"string"},{"internalType":"string","name":"metaCid","type":"string"}],"name":"createPassport","outputs":[{"internalType":"uint256","name":"passportId","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},
  {"inputs":[{"internalType":"uint256","name":"passportId","type":"uint256"},{"internalType":"string","name":"countryCode","type":"string"},{"internalType":"string","name":"dateISO","type":"string"},{"internalType":"string","name":"cid","type":"string"}],"name":"recordPlacedOnMarket","outputs":[],"stateMutability":"nonpayable","type":"function"},
  {"inputs":[{"internalType":"uint256","name":"passportId","type":"uint256"},{"internalType":"string","name":"cid","type":"string"}],"name":"addAttestation","outputs":[],"stateMutability":"nonpayable","type":"function"},
  {"inputs":[{"internalType":"uint256","name":"passportId","type":"uint256"},{"internalType":"uint8","name":"recoveryPercent","type":"uint8"},{"internalType":"string","name":"quality","type":"string"},{"internalType":"string","name":"cid","type":"string"}],"name":"recordRecovery","outputs":[],"stateMutability":"nonpayable","type":"function"},
  {"inputs":[{"internalType":"uint256","name":"parentId","type":"uint256"},{"internalType":"string","name":"metaCid","type":"string"}],"name":"spawnSecondaryPassport","outputs":[{"internalType":"uint256","name":"newPassportId","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},
  {"inputs":[{"internalType":"uint256","name":"passportId","type":"uint256"}],"name":"getPublicView","outputs":[{"internalType":"string","name":"orgId","type":"string"},{"internalType":"string","name":"upstreamBatchId","type":"string"},{"internalType":"string","name":"passportMetaCid","type":"string"},{"internalType":"bool","name":"placed","type":"bool"},{"internalType":"string","name":"countryCode","type":"string"},{"internalType":"string","name":"dateISO","type":"string"},{"internalType":"string","name":"placedCid","type":"string"},{"internalType":"bool","name":"hasAttestation","type":"bool"}],"stateMutability":"view","type":"function"}
]`

type ChainDemo struct {
	client   *ethclient.Client
	abi      abi.ABI
	contract *bind.BoundContract
	from     common.Address
	chainID  *big.Int
	privKey  *ecdsa.PrivateKey
}

func NewChainDemo(rpcURL, contractAddr, privateKey string) (*ChainDemo, error) {
	c, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}
	parsed, err := abi.JSON(strings.NewReader(aluminiumPassportDemoABI))
	if err != nil {
		return nil, err
	}
	addr := common.HexToAddress(contractAddr)
	bound := bind.NewBoundContract(addr, parsed, c, c, c)
	pk, err := crypto.HexToECDSA(strings.TrimPrefix(privateKey, "0x"))
	if err != nil {
		return nil, err
	}
	chainID, err := c.ChainID(context.Background())
	if err != nil {
		return nil, err
	}
	from := crypto.PubkeyToAddress(pk.PublicKey)
	return &ChainDemo{client: c, abi: parsed, contract: bound, from: from, chainID: chainID, privKey: pk}, nil
}

func (cd *ChainDemo) buildTxOpts(ctx context.Context) (*bind.TransactOpts, error) {
	return bind.NewKeyedTransactorWithChainID(cd.privKey, cd.chainID)
}

// Bridge methods. Note: caller must ensure the signer has the necessary roles in the contract.

func (cd *ChainDemo) RegisterUpstream(batchID, cid string) (common.Hash, error) {
	txOpts, err := cd.buildTxOpts(context.Background())
	if err != nil {
		return common.Hash{}, err
	}
	var tx *bind.TransactOpts = txOpts
	res, err := cd.contract.Transact(tx, "registerUpstreamBatch", batchID, cid)
	if err != nil {
		return common.Hash{}, err
	}
	return res.Hash(), nil
}

func (cd *ChainDemo) CreatePassport(orgID, upstreamBatchID, metaCID string) (*big.Int, common.Hash, error) {
	txOpts, err := cd.buildTxOpts(context.Background())
	if err != nil {
		return nil, common.Hash{}, err
	}
	var tx *bind.TransactOpts = txOpts
	res, err := cd.contract.Transact(tx, "createPassport", orgID, upstreamBatchID, metaCID)
	if err != nil {
		return nil, common.Hash{}, err
	}
	// We cannot get return value until mined unless we parse logs; for demo, just return nil and tx hash.
	return nil, res.Hash(), nil
}

func (cd *ChainDemo) RecordPlacedOnMarket(passportID *big.Int, country, dateISO, cid string) (common.Hash, error) {
	txOpts, err := cd.buildTxOpts(context.Background())
	if err != nil {
		return common.Hash{}, err
	}
	res, err := cd.contract.Transact(txOpts, "recordPlacedOnMarket", passportID, country, dateISO, cid)
	if err != nil {
		return common.Hash{}, err
	}
	return res.Hash(), nil
}

func (cd *ChainDemo) AddAttestation(passportID *big.Int, cid string) (common.Hash, error) {
	txOpts, err := cd.buildTxOpts(context.Background())
	if err != nil {
		return common.Hash{}, err
	}
	res, err := cd.contract.Transact(txOpts, "addAttestation", passportID, cid)
	if err != nil {
		return common.Hash{}, err
	}
	return res.Hash(), nil
}

func (cd *ChainDemo) RecordRecovery(passportID *big.Int, pct uint8, quality, cid string) (common.Hash, error) {
	txOpts, err := cd.buildTxOpts(context.Background())
	if err != nil {
		return common.Hash{}, err
	}
	res, err := cd.contract.Transact(txOpts, "recordRecovery", passportID, pct, quality, cid)
	if err != nil {
		return common.Hash{}, err
	}
	return res.Hash(), nil
}

func (cd *ChainDemo) SpawnSecondary(parentID *big.Int, metaCID string) (*big.Int, common.Hash, error) {
	txOpts, err := cd.buildTxOpts(context.Background())
	if err != nil {
		return nil, common.Hash{}, err
	}
	res, err := cd.contract.Transact(txOpts, "spawnSecondaryPassport", parentID, metaCID)
	if err != nil {
		return nil, common.Hash{}, err
	}
	return nil, res.Hash(), nil
}

func (cd *ChainDemo) GetPublicView(passportID *big.Int) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var out []interface{}
	err := cd.contract.Call(&bind.CallOpts{Context: ctx}, &out, "getPublicView", passportID)
	if err != nil {
		return nil, err
	}
	if len(out) != 8 {
		return nil, ErrUnexpectedResult
	}
	return map[string]interface{}{
		"orgId":           out[0].(string),
		"upstreamBatchId": out[1].(string),
		"passportMetaCid": out[2].(string),
		"placed":          out[3].(bool),
		"countryCode":     out[4].(string),
		"dateISO":         out[5].(string),
		"placedCid":       out[6].(string),
		"hasAttestation":  out[7].(bool),
	}, nil
}

var ErrUnexpectedResult = bind.ErrNoCode

// Helper to initialize from env
func NewChainDemoFromEnv() (*ChainDemo, error) {
	if strings.ToLower(os.Getenv("DEMO_CHAIN_ENABLED")) != "true" {
		return nil, nil
	}
	rpc := os.Getenv("DEMO_RPC_URL")
	addr := os.Getenv("DEMO_CONTRACT_ADDRESS")
	pk := os.Getenv("DEMO_PRIVATE_KEY")
	if rpc == "" || addr == "" || pk == "" {
		return nil, nil
	}
	return NewChainDemo(rpc, addr, pk)
}
