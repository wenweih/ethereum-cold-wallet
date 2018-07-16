package main

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
	qrcode "github.com/skip2/go-qrcode"
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

type csvAddress struct {
	Address string `csv:"address"`
}

func createAccount(accoutDir string) (*string, error) {
	// Generate a mnemonic for memorization or user-friendly seeds
	mnemonic, err := mnemonicFun()
	if err != nil {
		return nil, err
	}

	privateKey, path, err := hdWallet(*mnemonic)
	if err != nil {
		return nil, err
	}

	// pristr := hex.EncodeToString(privateKey.D.Bytes())

	// get the address
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	// generate first rondom password
	randomPwdFirst := RandStringBytesMaskImprSrc(50)

	// generate second rondom password
	randomPwdSecond := RandStringBytesMaskImprSrc(60)

	// save mnemonic qrcode
	saveAESEncryptMnemonicQrcode(address, *mnemonic, *path, accoutDir)

	// save keystore to configure path
	saveKeystore(privateKey, randomPwdFirst, randomPwdSecond, accoutDir)
	// save random pwd with address to configure path
	saveRandomPwd(address, randomPwdFirst, accoutDir, "random_pwd_first")
	saveRandomPwd(address, randomPwdSecond, accoutDir, "random_pwd_second")

	log.WithFields(log.Fields{
		"Generate Ethereum account": address,
		"Time:":                     time.Now().Format("Mon Jan _2 15:04:05 2006"),
	}).Info("")

	return &address, nil
}

func accountAuth(randomPwdFirst, randomPwdSecond string) string {
	h := sha256.New()
	h.Write([]byte(randomPwdFirst))
	h.Write([]byte(randomPwdSecond))
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

func saveFixedPwd(address, fixedPwd, dir string) {
	fixedPwdJSON := FixedPwdJSON{
		address,
		fixedPwd,
	}
	hexFixedPwdJSON, err := json.Marshal(fixedPwdJSON)
	if err != nil {
		log.Fatalf(err.Error())
	}
	hexFixedPwdJSON = append(hexFixedPwdJSON, '\n')
	fixedPwdPath, err := mkdirBySlice([]string{dir, "fixed_pwd"})
	if err != nil {
		log.Fatalln("Could not create directory", err.Error())
	}
	fixedPwdFile := strings.Join([]string{*fixedPwdPath, "fixedpwd.json"}, "/")
	if err = appenFile(fixedPwdFile, hexFixedPwdJSON, 0600); err != nil {
		log.Fatalln("Failed to write keyfile to", err.Error())
	}
}

func saveRandomPwd(address, randomPwd, dir, rdname string) {
	randomPwdJSON := RandomPwdJSON{
		address,
		randomPwd,
	}
	hexRandomPwdJSON, err := json.Marshal(randomPwdJSON)
	if err != nil {
		log.Fatalf(err.Error())
	}
	hexRandomPwdJSON = append(hexRandomPwdJSON, '\n')
	randomPwdPath, err := mkdirBySlice([]string{dir, rdname})
	if err != nil {
		log.Fatalln("Could not create directory", err.Error())
	}
	randomPwdFile := strings.Join([]string{*randomPwdPath, "randompwd.json"}, "/")
	if err = appenFile(randomPwdFile, hexRandomPwdJSON, 0600); err != nil {
		log.Fatalln("Failed to write keyfile to", err.Error())
	}
}

func readPwd(address, pwdType, path string) (*string, error) {
	var (
		PwdFile string
	)
	switch pwdType {
	case "random_pwd_first":
		dir := strings.Join([]string{path, "random_pwd_first"}, "/")
		PwdFile = strings.Join([]string{dir, "randompwd.json"}, "/")
	case "random_pwd_second":
		dir := strings.Join([]string{path, "random_pwd_second"}, "/")
		PwdFile = strings.Join([]string{dir, "randompwd.json"}, "/")
	default:
		return nil, errors.New("pwdType error")
	}

	jsonFile, err := os.Open(PwdFile)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	reader := bufio.NewReader(jsonFile)

	var pwd = new(string)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		var randompwd RandomPwdJSON
		json.Unmarshal(line, &randompwd)
		if randompwd.Address == address {
			pwd = &(randompwd.Randompwd)
			return pwd, nil
		}
	}
	return pwd, nil
}

func saveAESEncryptMnemonicQrcode(address, mnemonic, path, dir string) {
	// AES encrypt key should be 16 bytes (AES-128) or 32 (AES-256).
	randomPwd := RandStringBytesMaskImprSrc(32)
	m := &MnemonicJSON{
		address,
		mnemonic,
		path,
	}
	bMnemonicJSON, _ := json.Marshal(m)

	mNemonicCrypted, err := AesEncrypt(bMnemonicJSON, []byte(randomPwd))
	if err != nil {
		log.Fatalln("crypted mnemonic error", err.Error())
	}

	// save ASE 256 encode mnemonic and randomPwd(AesDecrypt key) qrcode
	saveAES256EncodeMnemonicQrcode(mNemonicCrypted, randomPwd, address, dir)

}

func saveAES256EncodeMnemonicQrcode(mNemonicCrypted []byte, key, address, dir string) {
	h := sha256.New()
	h.Write(mNemonicCrypted)
	mnemonicSha := base64.URLEncoding.EncodeToString(h.Sum(nil))
	mnemonicScryptedStr := base64.StdEncoding.EncodeToString(mNemonicCrypted)
	mnemonicSha256AndAESResult := strings.Join([]string{mnemonicScryptedStr, mnemonicSha}, "")

	mnemonicPNGPath, err := mkdirBySlice([]string{dir, "mnemonic_qrcode", address})
	if err != nil {
		log.Fatalln("Could not create directory", err.Error())
	}

	mnemonicAesDecryptPNGName := strings.Join([]string{address, "aesdecrypt_mnemonic.png"}, "_")
	mnemonicAesDecryptPNGFile := strings.Join([]string{*mnemonicPNGPath, mnemonicAesDecryptPNGName}, "/")
	if err := qrcode.WriteFile(mnemonicSha256AndAESResult, qrcode.Highest, 256, mnemonicAesDecryptPNGFile); err != nil {
		log.Fatalln("encode encrypt qrcode error", err.Error())
	}

	wm(mnemonicAesDecryptPNGFile, address)

	AesDecryptKeyPNGName := strings.Join([]string{address, "aesdecrypt_key.png"}, "_")
	AesDecryptKeyPNGFile := strings.Join([]string{*mnemonicPNGPath, AesDecryptKeyPNGName}, "/")
	if err := qrcode.WriteFile(key, qrcode.Medium, 256, AesDecryptKeyPNGFile); err != nil {
		log.Fatalln("encode key qrcode error", err.Error())
	}

	wm(AesDecryptKeyPNGFile, address)

	os.Remove(mnemonicAesDecryptPNGFile)
	os.Remove(AesDecryptKeyPNGFile)
}

func saveMnemonic(address, mnemonic, path, dir string) {
	m := &MnemonicJSON{
		address,
		mnemonic,
		path,
	}

	hexMnemonicJSON, _ := json.Marshal(m)
	mnemonicPath, err := mkdirBySlice([]string{dir, "mnemonic"})
	if err != nil {
		log.Fatalln("Could not create directory", err.Error())
	}

	mnemonicName := strings.Join([]string{address, "json"}, ".")
	mnemonicfile := strings.Join([]string{*mnemonicPath, mnemonicName}, "/")
	if err := ioutil.WriteFile(mnemonicfile, hexMnemonicJSON, 0600); err != nil {
		log.Fatalln("Failed to write keyfile to", err.Error())
	}
}

func saveKeystore(key *ecdsa.PrivateKey, randomPwdFirst, randomPwdSecond, dir string) {
	ks := &keystore.Key{
		Id:         uuid.NewRandom(),
		Address:    crypto.PubkeyToAddress(key.PublicKey),
		PrivateKey: key,
	}
	auth := accountAuth(randomPwdFirst, randomPwdSecond)
	keyjson, err := keystore.EncryptKey(ks, auth, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		log.Fatalf(err.Error())
	}

	keystorePath, err := mkdirBySlice([]string{dir, "keystore"})
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

func readKeyStore(address, path string) ([]byte, error) {
	keystoreName := strings.Join([]string{address, "json"}, ".")
	keystorefile := strings.Join([]string{path, "keystore", keystoreName}, "/")
	return ioutil.ReadFile(keystorefile)
}

func decodeKS2Key(addressHex string) (*keystore.Key, error) {
	path, err := accountDir(addressHex)
	if err != nil {
		return nil, err
	}

	keyjson, err := readKeyStore(addressHex, *path)
	if err != nil {
		return nil, errors.New(strings.Join([]string{"read keystore error", err.Error()}, " "))
	}

	randomPwdFirst, err := readPwd(addressHex, "random_pwd_first", *path)
	if err != nil {
		return nil, errors.New(strings.Join([]string{"read random_pwd_first error", err.Error()}, " "))
	}

	randomPwdSecond, err := readPwd(addressHex, "random_pwd_second", *path)
	if err != nil {
		return nil, errors.New(strings.Join([]string{"read random_pwd_second error", err.Error()}, " "))
	}

	auth := accountAuth(*randomPwdFirst, *randomPwdSecond)
	key, err := keystore.DecryptKey(keyjson, auth)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func export2CSV(addresses []*csvAddress, path string) {
	addressPath := strings.Join([]string{path, "eth_address.csv"}, "/")
	addressFile, err := os.OpenFile(addressPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer addressFile.Close()
	originAddress := []*csvAddress{}
	if err := gocsv.UnmarshalFile(addressFile, &originAddress); err != nil {
		if err := gocsv.MarshalFile(&addresses, addressFile); err != nil {
			log.Fatalln(err.Error())
		}
	} else {
		gocsv.MarshalWithoutHeaders(&addresses, addressFile)
	}

	log.WithFields(log.Fields{
		"export address to file": addressFile.Name(),
		"Time:":                  time.Now().Format("Mon Jan _2 15:04:05 2006"),
	}).Warn()
}
