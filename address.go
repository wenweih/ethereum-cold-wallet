package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pborman/uuid"
)

func createKeystore() {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf(err.Error())
	}

	// get the address
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	id := uuid.NewRandom()
	key := &keystore.Key{
		Id:         id,
		Address:    crypto.PubkeyToAddress(privateKey.PublicKey),
		PrivateKey: privateKey,
	}

	keyjson, err := keystore.EncryptKey(key, "", keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		log.Fatalf(err.Error())
	}

	keystoreName := strings.Join([]string{address, "json"}, ".")

	if err := os.MkdirAll(filepath.Dir("/tmp/keystore/"), 0700); err != nil {
		log.Fatalln("Could not create directory", err.Error())
	}

	if err := ioutil.WriteFile(strings.Join([]string{"/tmp/keystore", keystoreName}, "/"), keyjson, 0600); err != nil {
		log.Fatalln("Failed to write keyfile to", err.Error())
	}

	fmt.Println("Generate ", address)
}
