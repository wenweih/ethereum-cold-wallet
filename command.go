package main

import (
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	number int
	node   string
)

// EtherScan 配置
type EtherScan struct {
	Key string
	URL string
}

type configure struct {
	ElasticURL   string
	ElasticSniff bool
	EthRPC       string
	MaxBalance   float64
	To           []string
	NetMode      string
	RawTx        string
	SignedTx     string
	DB           string
	GethRPC      string
	ParityRPC    string
	EtherscanRPC string
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ethereum-service",
	Short: "Generate Ethereum account and sign tx",
}

// apiCmd represents the chain command
var genAccountCmd = &cobra.Command{
	Use:   "genaccount",
	Short: "Generate ethereum account",
	Run: func(cmd *cobra.Command, args []string) {
		fixedPwd, err := promptPwd()
		if err != nil {
			return
		}

		timeFormat := time.Now().Format("2006-01-02 15:04:05")
		dir := strings.Join([]string{"version_1", timeFormat}, " ")
		accountDir, err := mkdirBySlice([]string{HomeDir(), "account", dir})
		if err != nil {
			log.Fatalln("Fail to create account directory")
		}
		addresses := []*csvAddress{}
		for index := 0; index < number; index++ {
			address, err := createAccount(*fixedPwd, *accountDir)
			if err != nil {
				log.Fatalln(err.Error())
			}
			addresses = append(addresses, &csvAddress{Address: *address})
		}
		exportCSV(addresses)
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync chain data to elasticsearch",
	Run: func(cmd *cobra.Command, args []string) {
		config.InitConfig()
		sync()
	},
}

var subscribeNewBlockCmd = &cobra.Command{
	Use:   "sub",
	Short: "subscribe new block event",
	Run: func(cmd *cobra.Command, args []string) {
		config.InitConfig()
		subNewBlockCmd()
	},
}

var constructCmd = &cobra.Command{
	Use:   "construct",
	Short: "construct transactio",
	Run: func(cmd *cobra.Command, args []string) {
		config.InitConfig()
		if !Contains([]string{"geth", "parity", "etherscan"}, node) {
			log.Errorln("Only support geth, parity, etherscan")
			return
		}
		constructTxCmd()
	},
}

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "sigin transactio",
	Run: func(cmd *cobra.Command, args []string) {
		config.InitConfig()
		signTxCmd()
	},
}

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "broadcast signex transaction to ethereum network",
	Run: func(cmd *cobra.Command, args []string) {
		config.InitConfig()
		nodeClient, err := ethclient.Dial(config.EthRPC)
		if err != nil {
			log.Fatalln(err.Error())
		}
		sendTxCmd(nodeClient)
	},
}

// Execute 命令行入口
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf(err.Error())
	}
}

func (conf *configure) InitConfig() {
	viper.SetConfigType("yaml")
	viper.AddConfigPath(HomeDir())
	viper.SetConfigName("ethereum-service")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err == nil {
		log.WithFields(log.Fields{
			"Using Configure file": viper.ConfigFileUsed(),
			"Time:":                time.Now().Format("Mon Jan _2 15:04:05 2006"),
		}).Info()
	} else {
		log.Fatal("Error: ethereum-service.yml not found in: ", HomeDir())
	}

	for key, value := range viper.AllSettings() {
		switch key {
		case "elastic_url":
			conf.ElasticURL = value.(string)
		case "elastic_sniff":
			conf.ElasticSniff = value.(bool)
		case "eth_rpc":
			conf.EthRPC = value.(string)
		case "max_balance":
			conf.MaxBalance = value.(float64)
		case "to":
			conf.To = viper.GetStringSlice(key)
		case "net_mode":
			conf.NetMode = value.(string)
		case "raw_tx_path":
			conf.RawTx = value.(string)
		case "signed_tx_path":
			conf.SignedTx = value.(string)
		case "db_mysql":
			conf.DB = value.(string)
		case "geth_rpc":
			conf.GethRPC = value.(string)
		case "parity_rpc":
			conf.ParityRPC = value.(string)
		case "etherscan_rpc":
			subv := viper.Sub("etherscan_rpc")
			for subKey, subValue := range subv.AllSettings() {
				switch subKey {
				case "key":
					etherscan.Key = subValue.(string)
				case "url":
					etherscan.URL = subValue.(string)
				}
			}
		}
	}
}

func init() {
	config = new(configure)
	etherscan = new(EtherScan)
	initLogger()
	rootCmd.AddCommand(genAccountCmd)
	rootCmd.AddCommand(subscribeNewBlockCmd)
	rootCmd.AddCommand(constructCmd)
	rootCmd.AddCommand(signCmd)
	rootCmd.AddCommand(sendCmd)
	// rootCmd.AddCommand(syncCmd)
	genAccountCmd.Flags().IntVarP(&number, "number", "n", 10, "Generate ethereum accounts")
	genAccountCmd.MarkFlagRequired("number")

	constructCmd.Flags().StringVarP(&node, "node", "n", "parity", "Ethereum node type, support geth, parity, etherscan")
	constructCmd.MarkFlagRequired("node")
}
