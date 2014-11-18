package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os/user"
	"path"
	"regexp"
	"strings"
)

func GetDomainsFromPac(c []byte) map[string]bool {
	//c, err := ioutil.ReadFile(fname)
	//if err != nil {
	//	log.Printf("read file error %s", err)
	//}
	reg, err := regexp.Compile(`\".*\":`)
	if err != nil {
		log.Printf("regex pattern compile error %s", err)
	}
	names := reg.FindAll(c, -1)
	result := make(map[string]bool, 0)
	for _, n := range names {
		result[string(bytes.TrimLeft(bytes.TrimRight(n, "\":"), "\""))] = true
	}
	return result
}

func GetDomains(fp io.Reader) []string {
	scanner := bufio.NewScanner(fp)
	result := make([]string, 0)

	for scanner.Scan() {
		l := scanner.Text()
		if strings.HasPrefix(l, "#") {
			continue
		}
		if strings.TrimSpace(l) != "" {
			//log.Printf("add line %s", l)
			result = append(result, l)
		}
	}
	return result
}

//perform wildcard path match
func MatchDomain(token, pattern string) bool {
	m, e := path.Match(pattern, token)
	if e != nil {
		log.Printf("match err %s", e)
		return false
	}
	return m
}

func IsDomainIn(domain string, list []string) bool {
	for _, p := range list {
		if MatchDomain(domain, p) {
			return true
		}
	}
	return false
}

func ExpandHomePath(p string) string {
	if strings.HasPrefix(p, "~") {
		u, err := user.Current()
		if err != nil {
			log.Printf("get user err %s", err)
			return p
		}
		part := strings.TrimLeft(p, "~")
		return path.Join(u.HomeDir, part)
	}
	return p
}
