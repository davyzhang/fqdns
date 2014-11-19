package main

import (
	"bytes"
	"os/user"
	"testing"
)

var c = []byte(`
var wall_proxy = "SOCKS5 127.0.0.1:1080;";
var nowall_proxy = "DIRECT;";
var direct = "DIRECT;";

var domains = {
	"0-100s.com": 1,
	"001en.com": 1,
	"001job.com": 1,
	"001sj.net": 1
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

function FindProxyForURL(url, host) {
	if ( isPlainHostName(host) === true ) {
		return direct;
	}
	if ( check_ipv4(host) === true ) {
		return nowall_proxy;
	}
	var suffix;
	var pos1 = host.lastIndexOf('.');
	var pos = host.lastIndexOf('.', pos1 - 1);

	suffix = host.substring(pos1 + 1);
	if (suffix == "cn") {
		return nowall_proxy;
	}

	while(1) {
		if (pos == -1) {
			if (hasOwnProperty.call(domains, host)) {
				return nowall_proxy;
			} else {
				return wall_proxy;
			}
		}
		suffix = host.substring(pos + 1);
		if (hasOwnProperty.call(domains, suffix)) {
			return nowall_proxy;
		}
		pos = host.lastIndexOf('.', pos - 1);
	}
}


`)

func TestReadPacDomains(t *testing.T) {
	names := GetDomainsFromPac(c)
	if _, ok := names["0-100s.com"]; !ok {
		t.Fatalf("want %s get %s", "0-100s.com", ok)
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
