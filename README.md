
# Aluminium Passport Backend

## Run Instructions

### 1. Setup

Copy `.env.example` to `.env` and fill in:

```
JWT_SECRET=supersecretkey
WEB3_RPC_URL=https://your.ethereum.node
PRIVATE_KEY=your_wallet_private_key
CONTRACT_ADDRESS=deployed_contract_address
```

### 2. Generate ABI bindings

Make sure Go is installed and `abigen` is installed (from `geth` tools).

Run this from project root:

```
abigen --sol contracts/AluminiumPassport.sol --pkg abi --out abi/aluminium_passport.go
```

This will generate Go bindings in `abi/`.

### 3. Run the Server

```
go mod tidy
go run main.go
```

### 4. Docker Support

```
docker-compose up --build
```

This runs the backend on port 8080.
