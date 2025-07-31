
// Code generated - DO NOT EDIT.
// This is a simulated placeholder for the Go ABI binding.

package abi

import (
    "math/big"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
)

type AluminiumPassport struct{}

func NewAluminiumPassport(addr common.Address, backend bind.ContractBackend) (*AluminiumPassport, error) {
    return &AluminiumPassport{}, nil
}

func (ap *AluminiumPassport) CreatePassport(opts *bind.TransactOpts, passportId, origin, alloy, agency, signature string) (*big.Int, error) {
    return big.NewInt(1), nil
}
