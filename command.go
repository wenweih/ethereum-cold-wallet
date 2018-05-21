package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type configure struct {
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ethereum-service",
	Short: "Generate wallet and sign tx",
}

// Execute 命令行入口
func Execute() {
	config.InitConfig()
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
		fmt.Println("Using Configure file:", viper.ConfigFileUsed())
	} else {
		log.Fatal("Error: ethereum-service.yml not found in: ", HomeDir())
	}

	// for key, value := range viper.AllSettings() {
	// 	switch key {
	// 	}
	// }
}

func init() {
}
