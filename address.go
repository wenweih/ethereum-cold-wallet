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

// RandomPwdJSON 随机密码
type RandomPwdJSON struct {
	Address   string `json:"address"`
	Randompwd string `json:"randompwd"`
}

// FixedPwdJSON 固定密码
type FixedPwdJSON struct {
	Address  string `json:"address"`
	FixedPwd string `json:"fixedpwd"`
}

// MnemonicJSON 助记词
type MnemonicJSON struct {
	Address  string `json:"address"`
	Mnemonic string `json:"mnemonic"`
	PATH     string `json:"path"`
}

func createAccount(fixedPwd string) {
	// Generate a mnemonic for memorization or user-friendly seeds
	mnemonic, err := mnemonicFun()
	if err != nil {
		log.Fatalln(err.Error())
	}

	privateKey, path, err := hdWallet(*mnemonic)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// pristr := hex.EncodeToString(privateKey.D.Bytes())

	// get the address
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	// generate rondom password
	randomPwd := RandStringBytesMaskImprSrc(50)

	// save mnemonic
	saveMnemonic(address, *mnemonic, *path)

	// save keystore to configure path
	saveKetstore(privateKey, fixedPwd, randomPwd)
	// save random pwd with address to configure path
	saveRandomPwd(address, randomPwd)
	// save fixed pwd with address to configure path
	saveFixedPwd(address, fixedPwd)

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

func mnemonicFun() (*string, error) {
	// Generate a mnemonic for memorization or user-friendly seeds
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return nil, err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}

	return &mnemonic, nil
}

func hdWallet(mnemonic string) (*ecdsa.PrivateKey, *string, error) {
	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed := bip39.NewSeed(mnemonic, "")

	// Generate a new master node using the seed.
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, nil, err
	}

	// This gives the path: m/44H
	acc44H, err := masterKey.Child(hdkeychain.HardenedKeyStart + 44)
	if err != nil {
		return nil, nil, err
	}

	// This gives the path: m/44H/60H
	acc44H60H, err := acc44H.Child(hdkeychain.HardenedKeyStart + 60)
	if err != nil {
		return nil, nil, err
	}

	// This gives the path: m/44H/60H/0H
	acc44H60H0H, err := acc44H60H.Child(hdkeychain.HardenedKeyStart + 0)
	if err != nil {
		return nil, nil, err
	}

	// This gives the path: m/44H/60H/0H/0
	acc44H60H0H0, err := acc44H60H0H.Child(0)
	if err != nil {
		return nil, nil, err
	}

	// This gives the path: m/44H/60H/0H/0/0
	acc44H60H0H00, err := acc44H60H0H0.Child(0)
	if err != nil {
		return nil, nil, err
	}

	btcecPrivKey, err := acc44H60H0H00.ECPrivKey()
	if err != nil {
		return nil, nil, err
	}

	privateKey := btcecPrivKey.ToECDSA()

	path := "m/44H/60H/0H/0/0"

	return privateKey, &path, nil
}

func saveFixedPwd(address, fixedPwd string) {
	fixedPwdJSON := FixedPwdJSON{
		address,
		fixedPwd,
	}
	hexFixedPwdJSON, err := json.Marshal(fixedPwdJSON)
	if err != nil {
		log.Fatalf(err.Error())
	}
	hexFixedPwdJSON = append(hexFixedPwdJSON, '\n')
	fixedPwdPath, err := mkdirBySlice([]string{HomeDir(), config.FixedPwd, "fixedpwd"})
	if err != nil {
		log.Fatalln("Could not create directory", err.Error())
	}
	fixedPwdFile := strings.Join([]string{*fixedPwdPath, "randompwd.json"}, "/")
	if err = appenFile(fixedPwdFile, hexFixedPwdJSON, 0600); err != nil {
		log.Fatalln("Failed to write keyfile to", err.Error())
	}
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

func saveMnemonic(address, mnemonic, path string) {
	m := &MnemonicJSON{
		address,
		mnemonic,
		path,
	}

	hexMnemonicJSON, err := json.Marshal(m)
	mnemonicPath, err := mkdirBySlice([]string{HomeDir(), config.Mnemonic, "mnemonic"})
	if err != nil {
		log.Fatalln("Could not create directory", err.Error())
	}

	mnemonicName := strings.Join([]string{address, "json"}, ".")
	mnemonicfile := strings.Join([]string{*mnemonicPath, mnemonicName}, "/")
	if err := ioutil.WriteFile(mnemonicfile, hexMnemonicJSON, 0600); err != nil {
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
