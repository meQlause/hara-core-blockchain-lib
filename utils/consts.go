package utils

const ENSAddress = "0xD712CF6D5ADd6ddee0F3CEfF9FEde7c7CB4e8412"

type BuilderState int

const (
	StateInitial BuilderState = iota
	StateBodyBuilt
	StateHeadersSet
	StateRequestBuilt
	StateExecuted
)

const HNSRegistryABI = `[
  {
    "inputs": [
      {
        "internalType": "bytes32",
        "name": "_node",
        "type": "bytes32"
      }
    ],
    "name": "resolver",
    "outputs": [
      {
        "internalType": "address",
        "name": "",
        "type": "address"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  }
]`

const HNSResolverABI = `[
  {
    "inputs": [
      {
        "internalType": "bytes32",
        "name": "_node",
        "type": "bytes32"
      },
      {
        "internalType": "uint256",
        "name": "_contentTypes",
        "type": "uint256"
      }
    ],
    "name": "ABI",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "contentType",
        "type": "uint256"
      },
      {
        "internalType": "bytes",
        "name": "data",
        "type": "bytes"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  },
  {
    "inputs": [
      {
        "internalType": "bytes32",
        "name": "_node",
        "type": "bytes32"
      }
    ],
    "name": "addr",
    "outputs": [
      {
        "internalType": "address",
        "name": "ret",
        "type": "address"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  }
]`
