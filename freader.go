package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
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
	reg, err := regexp.Compile(`(?sU:\{.*\}\n\};)`)
	if err != nil {
		log.Printf("regex pattern compile error %s", err)
	}
	names := reg.Find(c)
	names = bytes.TrimRight(names, ";")
	var tmp map[string]map[string]int
	err = json.Unmarshal(names, &tmp)
	if err != nil {
		log.Printf("pac file format error %s", err)
		return nil
	}
	result := make(map[string]bool, 0)
	for k, v := range tmp {
		for d1, _ := range v {
			domain := d1 + "." + k
			result[domain] = true
		}
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

func ShortenDomain(d string) string {
	parts := strings.Split(d, ".")
	l := len(parts)
	if l > 2 {
		return parts[l-2] + "." + parts[l-1]
	}
	return d
}

func IsFileExists(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}
