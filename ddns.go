package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

var defaultNSClient *dns.Client
var dpac map[string]bool
var lwhite, lblack []string

type FConfig struct {
	Local             []string
	Remote            []string
	Port              int
	Pac, White, Black string
	TCPRemote         bool
}

var fconfig *FConfig

func initDomains(white, black *os.File, pac []byte) {
	if pac != nil {
		dpac = GetDomainsFromPac(pac)
	}
	if white != nil {
		lwhite = GetDomains(white)
	}
	if black != nil {
		lblack = GetDomains(black)
	}

}

func nsDispatch(w dns.ResponseWriter, req *dns.Msg) {
	log.Printf("get request, forwarding...")
	for _, q := range req.Question {
		nssvr := ""
		d := strings.TrimRight(q.Name, ".")
		if IsDomainIn(d, lwhite) { //white list domains goes to local
			log.Println("hit white list")
			nssvr = fconfig.Local[0]
		} else if IsDomainIn(d, lblack) { //black list domains goes to remote
			log.Println("hit black list")
			nssvr = fconfig.Remote[0]
		} else if _, ok := dpac[d]; ok {
			log.Println("hit pac")
			nssvr = fconfig.Local[0] //todo:rand to dispatch
		} else {
			log.Println("hit nothing")
			nssvr = fconfig.Remote[0]
		}

		log.Printf("query %s goes to %s", d, nssvr)
		back, _, e := defaultNSClient.Exchange(req, nssvr)
		if e != nil {
			log.Printf("dns server request err %s", e)
			continue
		}
		if back != nil {
			w.WriteMsg(back)
		}
	}
}

func ServeUDP(config *FConfig) {
	var svr dns.Server
	dns.HandleFunc(".", nsDispatch)
	svr.Addr = ":" + strconv.Itoa(config.Port)
	svr.Net = "udp"

	fconfig = config
	defaultNSClient = new(dns.Client)
	if fconfig.TCPRemote {
		defaultNSClient.Net = "tcp"
	}
	go func() {
		err := svr.ListenAndServe()
		if err != nil {
			log.Fatalf("can not start server %s", err)
		}
	}()
}

func nsProxy(w dns.ResponseWriter, req *dns.Msg) {
	nssvr := fconfig.Remote[0]
	back, _, err := defaultNSClient.Exchange(req, nssvr)
	if err != nil {
		log.Printf("dns server request err %s", err)
	}
	if back != nil {
		w.WriteMsg(back)
	}
}

func ServeTCPProxy(config *FConfig) {
	var svr dns.Server
	dns.HandleFunc(".", nsProxy)
	svr.Addr = ":" + strconv.Itoa(config.Port)
	svr.Net = "tcp"

	fconfig = config
	defaultNSClient = new(dns.Client)
	go func() {
		err := svr.ListenAndServe()
		if err != nil {
			log.Fatalf("can not start server %s", err)
		}
	}()
}
