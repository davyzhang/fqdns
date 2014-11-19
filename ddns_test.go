package main

import (
	"log"
	"strconv"
	"testing"

	"github.com/miekg/dns"
)

func DummyHandler(w dns.ResponseWriter, req *dns.Msg) {
	//log.Printf("req %s", req)
	qName := req.Question[0].Name
	if qName == "test_goes_to_local." || qName == "sss.xxoo.com." {
		m := new(dns.Msg)
		m.SetReply(req)
		m.Answer = make([]dns.RR, 1)
		m.Answer[0], _ = dns.NewRR("test_goes_to_local.	404	IN	A	182.140.167.44")
		w.WriteMsg(m)
		return
	}
	if qName == "test_goes_to_remote." || qName == "sss.xyz.com." || qName == "x.black.com." {
		m := new(dns.Msg)
		m.SetReply(req)
		m.Answer = make([]dns.RR, 1)
		m.Answer[0], _ = dns.NewRR("test_goes_to_remote.	404	IN	A	182.140.167.43")
		w.WriteMsg(m)
		return
	}
	m := new(dns.Msg)
	m.SetReply(req)
	m.Answer = make([]dns.RR, 1)
	m.Answer[0], _ = dns.NewRR("no_found.	404	IN	A	182.140.167.44")
	w.WriteMsg(m)
}

func SetUpUDP(port int) {
	var svr dns.Server
	dns.HandleFunc(".", DummyHandler)

	svr.Addr = "localhost:" + strconv.Itoa(port)
	svr.Net = "udp"
	err := svr.ListenAndServe()
	if err != nil {
		log.Printf("listen udp error %s", err)
	}
}

func newDNSReq(net, name, svr string) *dns.Msg {
	c := new(dns.Client)
	c.Net = net
	msg := new(dns.Msg)
	var q dns.Question
	q.Name = name
	msg.Question = []dns.Question{q}
	back, _, err := c.Exchange(msg, svr)
	if err != nil {
		log.Fatalf("dns request error %s", err)
		return nil
	}
	return back
}

func SetUPServers() {
	uconfig := new(FConfig)
	uconfig.Local = []string{"localhost:7777"}
	uconfig.Remote = []string{"localhost:8888"}
	uconfig.Port = 9999

	go func() { SetUpUDP(7777) }() //local dns
	go func() { SetUpUDP(8888) }() //remote dns
	ServeUDP(uconfig)              //local dispatcher
}

func TestDistribute(t *testing.T) {
	SetUPServers()

	dpac = make(map[string]bool)
	dpac["test_goes_to_local."] = true

	msg := newDNSReq("udp", "test_goes_to_local.", "localhost:9999")
	//log.Printf("msg back %v", msg)
	if msg.Answer[0].String() != "test_goes_to_local.	404	IN	A	182.140.167.44" {
		t.Fatalf("distribute wrong %s\n", msg.Answer[0].String())
	}

	msg = newDNSReq("udp", "test_goes_to_remote.", "localhost:9999")
	if msg.Answer[0].String() != "test_goes_to_remote.	404	IN	A	182.140.167.43" {
		t.Fatalf("distribute wrong %s\n", msg.Answer[0].String())
	}

	lblack = []string{"*.xyz.com"}
	dpac["sss.xyz.com"] = true
	dpac["sss.xxoo.com"] = true
	msg = newDNSReq("udp", "sss.xyz.com.", "localhost:9999")
	if msg.Answer[0].String() != "test_goes_to_remote.	404	IN	A	182.140.167.43" {
		t.Fatalf("black list wrong %s\n", msg.Answer[0].String())
	}

	lwhite = []string{"*.xxoo.com"}
	msg = newDNSReq("udp", "sss.xxoo.com.", "localhost:9999")
	if msg.Answer[0].String() != "test_goes_to_local.	404	IN	A	182.140.167.44" {
		t.Fatalf("white list wrong %s\n", msg.Answer[0].String())
	}

	lblack = []string{"x.black.com"}
	lwhite = []string{"x.black.com"}
	msg = newDNSReq("udp", "x.black.com.", "localhost:9999")
	if msg.Answer[0].String() != "test_goes_to_remote.	404	IN	A	182.140.167.43" {
		t.Fatalf("white list wrong %s\n", msg.Answer[0].String())
	}

}
