# Makefile for Aluminium Passport (Foundry-centric)

CONTRACT=contracts/AluminiumPassport.sol
GO_BINDINGS=abi/aluminium_passport.go

.PHONY: all build test abigen clean help

all: build abigen

build:
	@echo "[+] Compiling Solidity contract with Foundry (forge build)..."
	forge build

test:
	@echo "[+] Running Foundry tests..."
	forge test

abigen:
	@echo "[+] Generating Go bindings with abigen..."
	abigen --sol $(CONTRACT) --pkg abi --out $(GO_BINDINGS)

clean:
	@echo "[+] Cleaning build artifacts..."
	rm -rf out/ cache/ $(GO_BINDINGS)

help:
	@echo "Aluminium Passport Makefile (Foundry)"
	@echo "-------------------------------------"
	@echo "make           - Build contract and generate Go bindings"
	@echo "make build     - Compile Solidity contract with Foundry"
	@echo "make test      - Run Foundry tests"
	@echo "make abigen    - Generate Go bindings from contract"
	@echo "make clean     - Remove build artifacts and Go bindings"
	@echo "make help      - Show this help message"