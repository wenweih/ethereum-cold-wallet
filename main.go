package main

import (
	accounts "github.com/ethereum/go-ethereum/accounts"
	keystore "github.com/ethereum/go-ethereum/accounts/keystore"
)

func main() {
	am := accounts.NewManager("/tmp/keystore", keystore.StandardScryptN, keystore.StandardScryptP)

}
