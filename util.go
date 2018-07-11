package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/manifoldco/promptui"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

// HomeDir 获取服务器当前用户目录路径
func HomeDir() string {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err.Error())
	}
	return home
}

func initLogger() {
	path := strings.Join([]string{HomeDir(), ".ethereum_service"}, "/")
	if err := os.MkdirAll(path, 0700); err != nil {
		log.Fatalln(err.Error())
	}

	filepath := strings.Join([]string{path, "out.log"}, "/")
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	mw := io.MultiWriter(os.Stdout, file)
	if err == nil {
		log.SetOutput(mw)
		log.WithFields(log.Fields{
			"Note":  "all operate is recorded",
			"Time:": time.Now().Format("Mon Jan _2 15:04:05 2006"),
		}).Warn("")
	} else {
		log.Error(err.Error())
	}
}

func promptPwd() (*string, error) {
	promptOne := promptui.Prompt{
		Label: "Password",
		Mask:  '*',
	}

	resultOne, err := promptOne.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return nil, err
	}

	validate := func(input string) error {
		if resultOne != input {
			return errors.New("password not match")
		}
		return nil
	}

	promptTwo := promptui.Prompt{
		Label:    "Password",
		Validate: validate,
		Mask:     '*',
	}

	resultTwo, err := promptTwo.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return nil, err
	}
	return &resultTwo, nil
}

func promptSign(to string) {
	prompt := promptui.Prompt{
		Label:     strings.Join([]string{"To 地址不在配置文件中，请确认是否转入地址:", to}, " "),
		IsConfirm: true,
	}

	_, err := prompt.Run()
	if err != nil {
		log.Fatalln("退出...")
	}
}

// RandStringBytesMaskImprSrc 随机数
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
func RandStringBytesMaskImprSrc(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// PKCS7Padding PKCS7 填充 https://www.jianshu.com/p/b63095c59361
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS7UnPadding 还原 PKCS7 填充
func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// AesEncrypt 加密
func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS7Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// AesDecrypt 解密
func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS7UnPadding(origData)
	return origData, nil
}

func nodeClient(node string) (*ethclient.Client, error) {
	var nodeConfig string
	if node == "geth" {
		nodeConfig = config.GethRPC
	} else if node == "parity" {
		nodeConfig = config.ParityRPC
	}

	client, err := ethclient.Dial(nodeConfig)
	if err != nil {
		return nil, errors.New(strings.Join([]string{"node error", err.Error()}, " "))
	}
	return client, nil
}

func balanceIsLessThanConfig(address string, balance *big.Int) error {
	balanceDecimal, _ := decimal.NewFromString(balance.String())
	ethFac, _ := decimal.NewFromString("0.000000000000000001")
	amount := balanceDecimal.Mul(ethFac)
	settingBalance := decimal.NewFromFloat(config.MaxBalance)
	if amount.LessThan(settingBalance) {
		return errors.New(strings.Join([]string{"Ignore:", address, "balance not great than the configure amount"}, " "))
	}
	return nil
}

func appenFile(filename string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

func mkdirBySlice(slice []string) (*string, error) {
	path := strings.Join(slice, "/")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0700); err != nil {
			return nil, err
		}
	}
	return &path, nil
}

// Contains tells whether a contains x.
func Contains(a []string, x string) bool {
	for _, n := range a {
		if strings.Compare(strings.ToLower(x), strings.ToLower(n)) == 0 {
			return true
		}
	}
	return false
}

func randomPickFromSlice(slice []string) string {
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	return slice[r.Intn(len(slice))]
}

func accountDir(address string) (*string, error) {
	rootDir := strings.Join([]string{HomeDir(), "account"}, "/")
	folders, err := ioutil.ReadDir(rootDir)
	if err != nil {
		return nil, errors.New("Read Keystore error")
	}
	keystoreName := strings.Join([]string{address, "json"}, ".")
	for _, folder := range folders {
		dir := strings.Join([]string{rootDir, folder.Name(), "keystore"}, "/")
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			return nil, errors.New("read file error")
		}
		for _, f := range files {
			if strings.Compare(strings.ToLower(f.Name()), strings.ToLower(keystoreName)) == 0 {
				path := path.Dir(dir)
				return &path, err
			}
		}
	}
	return nil, errors.New("Account directory not found")
}
