package main

import (
	"io"
	"os"
	"strings"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
)

// HomeDir 获取服务器当亲用户目录路径
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
