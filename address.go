package main

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	bip39 "github.com/tyler-smith/go-bip39"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pborman/uuid"
)

// RandomPwdJSON 随机密码 JSON
type RandomPwdJSON struct {
	Address   string `json:"address"`
	Randompwd string `json:"randompwd"`
}

func createAccount(fixedPwd string) {
	// Generate a mnemonic for memorization or user-friendly seeds
	mnemonic := mnemonicFun()
	privateKey := hdWallet(mnemonic)

	// pristr := hex.EncodeToString(privateKey.D.Bytes())

	// get the address
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	randomPwd := RandStringBytesMaskImprSrc(50)

	// save keystore to configure path
	saveKetstore(privateKey, fixedPwd, randomPwd)
	// save random pwd with address to configure path
	saveRandomPwd(address, randomPwd)

	log.WithFields(log.Fields{
		"Generate Ethereum account": address,
		"Time:":                     time.Now().Format("Mon Jan _2 15:04:05 2006"),
	}).Info("")
}

func accountAuth(fixPwd, randomPwd string) string {
	h := sha256.New()
	h.Write([]byte(fixPwd))
	h.Write([]byte(randomPwd))
	auth := hex.EncodeToString(h.Sum(nil))
	return auth
}

func mnemonicFun() string {
	// Generate a mnemonic for memorization or user-friendly seeds
	entropy, _ := bip39.NewEntropy(128)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	return mnemonic
}

func hdWallet(mnemonic string) *ecdsa.PrivateKey {
	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed := bip39.NewSeed(mnemonic, "")

	// Generate a new master node using the seed.
	masterKey, _ := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)

	// This gives the path: m/44H
	acc44H, _ := masterKey.Child(hdkeychain.HardenedKeyStart + 44)

	// This gives the path: m/44H/60H
	acc44H60H, _ := acc44H.Child(hdkeychain.HardenedKeyStart + 60)

	// This gives the path: m/44H/60H/0H
	acc44H60H0H, _ := acc44H60H.Child(hdkeychain.HardenedKeyStart + 0)

	// This gives the path: m/44H/60H/0H/0
	acc44H60H0H0, _ := acc44H60H0H.Child(0)

	// This gives the path: m/44H/60H/0H/0/0
	acc44H60H0H00, _ := acc44H60H0H0.Child(0)

	btcecPrivKey, _ := acc44H60H0H00.ECPrivKey()
	privateKey := btcecPrivKey.ToECDSA()

	return privateKey
}

func saveRandomPwd(address, randomPwd string) {
	randomPwdJSON := RandomPwdJSON{
		address,
		randomPwd,
	}
	hexRandomPwdJSON, err := json.Marshal(randomPwdJSON)
	if err != nil {
		log.Fatalf(err.Error())
	}
	hexRandomPwdJSON = append(hexRandomPwdJSON, '\n')
	randomPwdPath, err := mkdirBySlice([]string{HomeDir(), config.RandomPwd, "randompwd"})
	if err != nil {
		log.Fatalln("Could not create directory", err.Error())
	}
	randomPwdFile := strings.Join([]string{*randomPwdPath, "randompwd.json"}, "/")
	if err = appenFile(randomPwdFile, hexRandomPwdJSON, 0600); err != nil {
		log.Fatalln("Failed to write keyfile to", err.Error())
	}
}

func saveKetstore(key *ecdsa.PrivateKey, fixedPwd, randomPwd string) {
	ks := &keystore.Key{
		Id:         uuid.NewRandom(),
		Address:    crypto.PubkeyToAddress(key.PublicKey),
		PrivateKey: key,
	}
	auth := accountAuth(fixedPwd, randomPwd)
	keyjson, err := keystore.EncryptKey(ks, auth, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		log.Fatalf(err.Error())
	}

	keystorePath, err := mkdirBySlice([]string{HomeDir(), config.Keystore, "keystore"})
	if err != nil {
		log.Fatalln("Could not create directory", err.Error())
	}

	address := crypto.PubkeyToAddress(key.PublicKey).Hex()
	keystoreName := strings.Join([]string{address, "json"}, ".")
	keystorefile := strings.Join([]string{*keystorePath, keystoreName}, "/")
	if err := ioutil.WriteFile(keystorefile, keyjson, 0600); err != nil {
		log.Fatalln("Failed to write keyfile to", err.Error())
	}
}
