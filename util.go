package main

import (
	"log"

	homedir "github.com/mitchellh/go-homedir"
)

// HomeDir 获取服务器当亲用户目录路径
func HomeDir() string {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(err.Error())
	}
	return home
}
