package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pborman/uuid"
)

// RandomPwdJSON 随机密码 JSON
type RandomPwdJSON struct {
	Address   string `json:"address"`
	Randompwd string `json:"randompwd"`
}

func createKeystore(pwd string) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf(err.Error())
	}

	// get the address
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	randomPwd := RandStringBytesMaskImprSrc(50)
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

	key := &keystore.Key{
		Id:         uuid.NewRandom(),
		Address:    crypto.PubkeyToAddress(privateKey.PublicKey),
		PrivateKey: privateKey,
	}
	auth := accountAuth(pwd, randomPwd)
	keyjson, err := keystore.EncryptKey(key, auth, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		log.Fatalf(err.Error())
	}

	keystorePath, err := mkdirBySlice([]string{HomeDir(), config.Keystore, "keystore"})
	if err != nil {
		log.Fatalln("Could not create directory", err.Error())
	}

	keystoreName := strings.Join([]string{address, "json"}, ".")
	keystorefile := strings.Join([]string{*keystorePath, keystoreName}, "/")
	if err := ioutil.WriteFile(keystorefile, keyjson, 0600); err != nil {
		log.Fatalln("Failed to write keyfile to", err.Error())
	}

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
