package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
)

func main() {
	p, _ := filepath.Abs(os.Args[0])
	appPath := filepath.Dir(p)
	config := flag.String("config", path.Join(appPath, "config.json"), "config file for fqdns")
	mode := flag.String("mode", "", "local dispatcher or outside resolver disp/resolver")
	flag.Parse()

	if *config == "" || *mode == "" {
		flag.Usage()
		return
	}

	log.Printf("using config file %s  mode %v", *config, *mode)
	c, err := ioutil.ReadFile(*config)
	if err != nil {
		log.Fatalf("read config error %s", err)
	}
	var fconfig FConfig
	fconfig.Local = make([]string, 1)
	fconfig.Remote = make([]string, 1)
	err = json.Unmarshal(c, &fconfig)
	if err != nil {
		log.Fatalf("config file format invalid %s", err)
	}
	if *mode == "disp" {
		fconfig.White = ExpandHomePath(fconfig.White)
		fconfig.Black = ExpandHomePath(fconfig.Black)
		fconfig.Pac = ExpandHomePath(fconfig.Pac)

		guess := path.Join(appPath, fconfig.White)
		if !IsFileExists(fconfig.White) && IsFileExists(guess) {
			fconfig.White = guess
		}

		guess = path.Join(appPath, fconfig.Pac)
		if !IsFileExists(fconfig.Pac) && IsFileExists(guess) {
			fconfig.Pac = guess
		}

		guess = path.Join(appPath, fconfig.Black)
		if !IsFileExists(fconfig.Black) && IsFileExists(guess) {
			fconfig.Black = guess
		}

		c, err = ioutil.ReadFile(fconfig.Pac)
		if err != nil {
			log.Fatalf("read white list file error %s", err)
		}

		fpw, err := os.Open(fconfig.White)
		if err != nil {
			log.Fatalf("white list read err %s", err)
		}
		fpb, err := os.Open(fconfig.Black)
		if err != nil {
			log.Fatalf("black list read err %s", err)
		}
		initDomains(fpw, fpb, c)
	}
	if *mode == "disp" {
		ServeUDP(&fconfig)
	} else if *mode == "resolver" {
		ServeTCPProxy(&fconfig)
	}
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case s := <-sig:
			log.Fatalf("Signal (%v) received, stopping\n", s)
		}
	}
}
