package main

import (
	"io/ioutil"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

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

	key := &keystore.Key{
		Id:         uuid.NewRandom(),
		Address:    crypto.PubkeyToAddress(privateKey.PublicKey),
		PrivateKey: privateKey,
	}

	keyjson, err := keystore.EncryptKey(key, "", keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		log.Fatalf(err.Error())
	}

	keystorePath := strings.Join([]string{HomeDir(), config.Keystore, "keystore"}, "/")
	if _, err := os.Stat(keystorePath); os.IsNotExist(err) {
		if err := os.MkdirAll(keystorePath, 0700); err != nil {
			log.Fatalln("Could not create directory", err.Error())
		}
	}

	keystoreName := strings.Join([]string{address, "json"}, ".")
	keystorefile := strings.Join([]string{keystorePath, keystoreName}, "/")
	if err := ioutil.WriteFile(keystorefile, keyjson, 0600); err != nil {
		log.Fatalln("Failed to write keyfile to", err.Error())
	}

	log.WithFields(log.Fields{
		"Generate Ethereum account": address,
		"Time:":                     time.Now().Format("Mon Jan _2 15:04:05 2006"),
	}).Info("")
}
