package main

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	number int
)

type configure struct {
	Keystore  string
	RandomPwd string
	FixedPwd  string
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
		fixedPwd, err := promptUtil()
		if err != nil {
			return
		}

		for index := 0; index < number; index++ {
			createAccount(*fixedPwd)
		}
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
		case "key_store_path":
			conf.Keystore = value.(string)
		case "random_pwd_path":
			conf.RandomPwd = value.(string)
		case "fixed_pwd_path":
			conf.FixedPwd = value.(string)
		}
	}
}

func init() {
	config = new(configure)
	config.InitConfig()
	initLogger()
	rootCmd.AddCommand(genAccountCmd)
	genAccountCmd.Flags().IntVarP(&number, "number", "n", 10, "Generate ethereum accounts")
}
