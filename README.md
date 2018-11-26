## ethereum-cold-wallet
Generate Ethereum HD wallet & offline sign && broadcast signed tx to network. Solution for eth cold wallet.
### Install
Environment Require
- Golang
- Ethereum Private chain
- dep
- cgo or [xgo](https://github.com/karalabe/xgo) (for go-ethereum dependency)
- MySQL (construct tx)

```bash
go get -u github.com/wenweih/ethereum-cold-wallet
cd $GOPATH/src/github.com/wenweih/ethereum-cold-wallet
dep ensure -v -update
```
because of the codebase import [go-ethereum](https://github.com/ethereum/go-ethereum), which is dependent on c, so cross compile need cgo. I hightly recommend a tool name [xgo](https://github.com/karalabe/xgo) for Go CGO cross compiler, which  is based on the concept of lightweight Linux containers.
```bash
go get github.com/karalabe/xgo
```
Next step is compile binary package for specify platform server, like this:
```bash
# for ubuner server
xgo --targets=linux/amd64 ./
```
if you are insterested in xgo usage, pls read the docutment: [xgo#usage](https://github.com/karalabe/xgo#usage)
### Usage
Firstly, modify configure and put it in ~/ethereum-cold-wallet.yml
```bash
./ethereum-cold-wallet -h
time="2018-08-12T23:47:55+08:00" level=warning Note="all operate is recorded" Time:="Sun Aug 12 23:47:55 2018"
Generate Ethereum account and sign tx

Usage:
  ethereum-service [command]

Available Commands:
  construct   construct transactio
  genaccount  Generate ethereum account
  help        Help about any command
  send        broadcast signex transaction to ethereum network
  sign        sigin transactio
  sub         subscribe new block event

Flags:
  -h, --help   help for ethereum-service

Use "ethereum-service [command] --help" for more information about a command.
```
#### Generate HD Wallet
```bash
./ethereum-cold-wallet genaccount -n 3
time="2018-08-13T15:40:10+08:00" level=warning Note="all operate is recorded" Time:="Mon Aug 13 15:40:10 2018"
time="2018-08-13T15:40:11+08:00" level=info Generate Ethereum account=0xe5379d64Cd7d2D963B03da01fB052218a9aCB0Ce Time:="Mon Aug 13 15:40:11 2018"
time="2018-08-13T15:40:12+08:00" level=info Generate Ethereum account=0x8Dc63ce8b979627C11f5EEf673990814D4815613 Time:="Mon Aug 13 15:40:12 2018"
time="2018-08-13T15:40:13+08:00" level=info Generate Ethereum account=0x48031a8E6150B6ED53F0342451D269f109934729 Time:="Mon Aug 13 15:40:13 2018"
time="2018-08-13T15:40:13+08:00" level=warning Time:="Mon Aug 13 15:40:13 2018" export address to file=/Users/hww/account/eth_address.csv
```
All elements about wallet are generated in **~/account** folder:
```bash
▶ tree account
account
├── eth_address.csv
├── keystore
│   └── version_1_2018-08-13_15-40-10
│       ├── 0x48031a8E6150B6ED53F0342451D269f109934729.json
│       ├── 0x8Dc63ce8b979627C11f5EEf673990814D4815613.json
│       └── 0xe5379d64Cd7d2D963B03da01fB052218a9aCB0Ce.json
├── mnemonic_qrcode
│   └── version_1_2018-08-13_15-40-10
│       ├── 0x48031a8E6150B6ED53F0342451D269f109934729
│       │   ├── 0x48031a8E6150B6ED53F0342451D269f109934729_aesdecrypt_key_marked.png
│       │   └── 0x48031a8E6150B6ED53F0342451D269f109934729_aesdecrypt_mnemonic_marked.png
│       ├── 0x8Dc63ce8b979627C11f5EEf673990814D4815613
│       │   ├── 0x8Dc63ce8b979627C11f5EEf673990814D4815613_aesdecrypt_key_marked.png
│       │   └── 0x8Dc63ce8b979627C11f5EEf673990814D4815613_aesdecrypt_mnemonic_marked.png
│       └── 0xe5379d64Cd7d2D963B03da01fB052218a9aCB0Ce
│           ├── 0xe5379d64Cd7d2D963B03da01fB052218a9aCB0Ce_aesdecrypt_key_marked.png
│           └── 0xe5379d64Cd7d2D963B03da01fB052218a9aCB0Ce_aesdecrypt_mnemonic_marked.png
├── random_pwd_first
│   └── version_1_2018-08-13_15-40-10
│       └── randompwd.json
└── random_pwd_second
    └── version_1_2018-08-13_15-40-10
        └── randompwd.json
```
#### construct transacion
we have generated some wallets, next step is send amount of eth to the address. By conveniently, we sent ETH using Private Ethereum in our laptop.
[Ethereum 私有链和 web3.js 使用](https://huangwenwei.com/blogs/ethereum-private-chain-and-web3js)

```bash
# deposit 30 ETH to test address
Welcome to the Geth JavaScript console!

instance: Geth/v1.8.10-unstable-7677ec1f/darwin-amd64/go1.9.2
coinbase: 0x37764d6eae4fad0c69cb7194896e0af7cf260885
at block: 570 (Wed, 08 Aug 2018 18:22:01 CST)
 datadir: /Users/hww/geth_private_data
 modules: admin:1.0 debug:1.0 eth:1.0 miner:1.0 net:1.0 personal:1.0 rpc:1.0 txpool:1.0 web3:1.0

>
> eth.sendTransaction({from: eth.coinbase, to: "0xe5379d64Cd7d2D963B03da01fB052218a9aCB0Ce",value:web3.toWei(30,"ether")})
"0xc8167e6ad819c6fa9dbbb8f45cf0da3c00213553af8bbfcf89bdb631e60d48da"
> web3.fromWei(web3.eth.getBalance("0xe5379d64Cd7d2D963B03da01fB052218a9aCB0Ce"),"ether")
30
```
soft link to **eth_address.csv** and construct transacion for address **0xe5379d64Cd7d2D963B03da01fB052218a9aCB0Ce**
```bash
ln /Users/hww/account/eth_address.csv ~/

▶ ethereum-cold-wallet construct -n geth
time="2018-08-13T15:45:46+08:00" level=warning Note="all operate is recorded" Time:="Mon Aug 13 15:45:46 2018"
time="2018-08-13T15:45:46+08:00" level=info Time:="Mon Aug 13 15:45:46 2018" Using Configure file=/Users/hww/ethereum-cold-wallet.yml
time="2018-08-13T15:45:46+08:00" level=info msg="csv2db done"
time="2018-08-13T15:45:46+08:00" level=info msg="Exported HexTx to /Users/hww/tx/unsign/unsign_from.0xe5379d64Cd7d2D963B03da01fB052218a9aCB0Ce.json"
time="2018-08-13T15:45:46+08:00" level=warning msg="Ignore: 0x8Dc63ce8b979627C11f5EEf673990814D4815613 balance not great than the configure amount"
time="2018-08-13T15:45:46+08:00" level=warning msg="Ignore: 0x48031a8E6150B6ED53F0342451D269f109934729 balance not great than the configure amount"
```
as you can see, the contructed transaction is export to ```/Users/hww/tx/unsign/``` folder, we can copy these unsign transaction to offline computer, which is holder our wallet keys, in this example, we handle it in my laptop too.
#### Sign raw transaction
```bash
▶ ethereum-cold-wallet sign
time="2018-08-13T15:59:03+08:00" level=warning Note="all operate is recorded" Time:="Mon Aug 13 15:59:03 2018"
time="2018-08-13T15:59:03+08:00" level=info Time:="Mon Aug 13 15:59:03 2018" Using Configure file=/Users/hww/ethereum-cold-wallet.yml
time="2018-08-13T15:59:03+08:00" level=info msg="签名交易： 0x3f00ff54245328604a6f43f4de279de100d4afc8d5e7536eeaee7b531c2d64d2  To: 0x8Dc63ce8b979627C11f5EEf673990814D4815613"
time="2018-08-13T15:59:04+08:00" level=info msg="Exported HexTx to /Users/hww/tx/signed/signed_from.0xe5379d64Cd7d2D963B03da01fB052218a9aCB0Ce.json"
```
The transaction we constructed is signed and export json file to ```/Users/hww/tx/signed/``` folder, copy the result to broadcast the signed sendTransaction.
#### broadcast signed transacion
```bash
▶ ethereum-cold-wallet send
time="2018-08-13T16:03:18+08:00" level=warning Note="all operate is recorded" Time:="Mon Aug 13 16:03:18 2018"
time="2018-08-13T16:03:18+08:00" level=info Time:="Mon Aug 13 16:03:18 2018" Using Configure file=/Users/hww/ethereum-cold-wallet.yml
time="2018-08-13T16:03:18+08:00" level=info msg="send tx:  0xbdfece2382b6e08c265928578b11b00292582670ad5b2c7a90243267b892d41b success"
```
log show we have send the transacion successfully, now we query the tx related addresses balance in web3 console:
- from 0xe5379d64Cd7d2D963B03da01fB052218a9aCB0Ce should be 0
- to 0x8Dc63ce8b979627C11f5EEf673990814D4815613 should be 30 - fee

```bash
> web3.fromWei(web3.eth.getBalance("0xe5379d64Cd7d2D963B03da01fB052218a9aCB0Ce"),"ether")
0
> web3.fromWei(web3.eth.getBalance("0x8Dc63ce8b979627C11f5EEf673990814D4815613"),"ether")
29.999999999999979
```
### Links
- [xgo](https://github.com/karalabe/xgo)
- [Cross compiling Ethereum](https://github.com/ethereum/go-ethereum/wiki/Cross-compiling-Ethereum)
