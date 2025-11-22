# HARA Core Blockchain Library

Library ini dibuat untuk mendukung pengembangan **HARA SDK**,
menyediakan abstraksi level tinggi untuk:

-   Interaksi RPC Ethereum (read & write)
-   Pembuatan & penandatanganan transaksi
-   Integrasi ENS/HNS (Registry, Resolver, ABI storage)
-   Call & execute smart contract
-   Manajemen wallet dan signing

## Fitur Utama

### Wallet

-   Derivasi private key
-   Signing EIP-191 & EIP-155
-   Public address helper

### Network

Wrapper RPC client untuk Ethereum: - eth_call - eth_sendRawTransaction -
gasPrice - blockNumber - chainId - pendingNonce

### Blockchain

-   Resolve smart contract melalui ENS/HNS
-   Cache ABI & Address
-   CallContract (read)
-   SendContractTx (write)
-   BuildTx struct-friendly

## Contoh Penggunaan

### Inisialisasi

``` go
net := pkg.NewNetwork("https://rpc.local", "2.0", 1)
bc := pkg.NewBlockchain("seed", net, 1)
```

### Resolve contract & ABI

``` go
abiRes, _ := bc.GetAddressABI(ctx, "contract.hara")
```

### Call contract

``` go
res, _ := bc.CallContract(ctx, contract, "balanceOf", []any{addr})
```

### Send transaksi

``` go
tx := bc.BuildTx(params)
hash, _ := bc.SendContractTx(ctx, tx)
```