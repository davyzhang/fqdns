package main

import (
	"bytes"
	"fmt"
	"os/user"
	"testing"
)

var c = []byte(`
var wall_proxy = "SOCKS5 127.0.0.1:1080;";
var nowall_proxy = "DIRECT;";
var direct = "DIRECT;";

/*
 * Copyright (C) 2014 breakwa11
 * https://github.com/breakwa11/gfw_whitelist
 */

var white_domains = {"am":{
"126":1,
"51":1
},"biz":{
"7daysinn":1,
"changan":1,
"chinafastener":1,
"cnbearing":1,
"menchuang":1,
"qianyan":1
},"tw":{
"taiwandao":1
},"vg":{
"loli":1
},"xn--fiqs8s":{
"":1
},"za":{
"org":1
}
};

var hasOwnProperty = Object.hasOwnProperty;

function check_ipv4(host) {
	// check if the ipv4 format (TODO: ipv6)
	//   http://home.deds.nl/~aeron/regex/
	var re_ipv4 = /^\d+\.\d+\.\d+\.\d+$/g;
	if (re_ipv4.test(host)) {
		// in theory, we can add chnroutes test here.
		// but that is probably too much an overkill.
		return true;
	}
}
function isInDomains(domain_dict, host) {
	var suffix;
	var pos1 = host.lastIndexOf('.');

	suffix = host.substring(pos1 + 1);
	if (suffix == "cn") {
		return true;
	}

	var domains = domain_dict[suffix];
	if ( domains === undefined ) {
		return true;
	}
	host = host.substring(0, pos1);
	var pos = host.lastIndexOf('.');

	while(1) {
		if (pos <= 0) {
			if (hasOwnProperty.call(domains, host)) {
				return true;
			} else {
				return false;
			}
		}
		suffix = host.substring(pos + 1);
		if (hasOwnProperty.call(domains, suffix)) {
			return true;
		}
		pos = host.lastIndexOf('.', pos - 1);
	}
}
function FindProxyForURL(url, host) {
	if ( isPlainHostName(host) === true ) {
		return direct;
	}
	if ( check_ipv4(host) === true ) {
		return nowall_proxy;
	}
	if ( isInDomains(white_domains, host) === true ) {
		return nowall_proxy;
	}
	return wall_proxy;
}
`)

func TestReadPacDomains(t *testing.T) {
	names := GetDomainsFromPac(c)
	if v, ok := names["7daysinn.biz"]; !ok {

		for k, v := range names {
			fmt.Println(k, " > ", v)
		}
		t.Fatalf("want %s get %v", "7daysinn.biz", v)
	}
}

func TestReadDomains(t *testing.T) {
	txt := `
#comment1
www.somedomains.com
www.gfwshit.com
`
	r := bytes.NewBufferString(txt)
	names := GetDomains(r)
	if names[0] != "www.somedomains.com" {
		t.Fatalf("parse error 1 %s %s", names[0], names[1])
	}
	if names[1] != "www.gfwshit.com" {
		t.Fatalf("parse error 2")
	}

}

func TestMatchDomain(t *testing.T) {
	if !MatchDomain("xxx.cloudfront.com", "*.cloudfront.*") {
		t.Errorf("match failed 1")
	}
	if !MatchDomain("xxx.cloudfront.org", "*.cloudfront.*") {
		t.Errorf("match failed 2")
	}
	if MatchDomain("xxx.cloudfront", "*.cloudfront.*") {
		t.Errorf("match failed 3")
	}
}

func TestExpandHome(t *testing.T) {
	u, _ := user.Current()

	tstr := ExpandHomePath("~/fqdns/xxx")
	if tstr != u.HomeDir+"/fqdns/xxx" {
		t.Errorf("home expand not right %s", tstr)
	}
}

func TestShortenDomain(t *testing.T) {
	r := ShortenDomain("www.qq.com")
	if r != "qq.com" {
		t.Errorf("www.qq.com should be qq.com get %s", r)
	}
	r = ShortenDomain("p.p.qq.com")
	if r != "qq.com" {
		t.Errorf("www.qq.com should be qq.com get %s", r)
	}
	r = ShortenDomain("qq.com")
	if r != "qq.com" {
		t.Errorf("www.qq.com should be qq.com get %s", r)
	}
}
