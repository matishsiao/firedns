package main

import (
	"code.google.com/p/go.net/idna"
	"encoding/json"
	"github.com/matishsiao/dns"
	"io/ioutil"
	"log"
	"fmt"
	"os"
	_ "net/http"
)

// only used in JSON
type Record struct {
	Name  string
	Type  string
	Class string
	Ttl   uint32
	Data  string
}

func (r *Record) Dump() {
	log.Printf("Name:%s,Type:%s,Class:%s,Ttl:%d,Data:%s\n",r.Name,r.Type,r.Class,r.Ttl,r.Data)
}

func prefetch(zs *ZoneStore, critical bool) {
	tmpmap := GetZones()
	log.Println(tmpmap)

	zs.m.Lock()
	zs.store = make(map[string]Zone)
	for key, value := range tmpmap {
		key = dns.Fqdn(key)
		if cdn, e := idna.ToASCII(key); e == nil {
			key = cdn
		}
		if zs.store[key] == nil {
			zs.store[key] = make(map[dns.RR_Header][]dns.RR)
		}
		for _, r := range value {
			if cdn, e := idna.ToASCII(r.Name); e == nil {
				r.Name = cdn
			}
			rr, err := dns.NewRR(dns.Fqdn(r.Name) + " " + r.Class + " " + r.Type + " " + r.Data)
			fmt.Println(rr)
			if err == nil {
				rr.Header().Ttl = r.Ttl
				key2 := dns.RR_Header{Name: dns.Fqdn(rr.Header().Name), Rrtype: rr.Header().Rrtype, Class: rr.Header().Class}
				zs.store[key][key2] = append(zs.store[key][key2], rr)
			} else {
				log.Printf("Skipping problematic record: %+v\nError: %+v\n", r, err)
			}
		}
	}
	zs.m.Unlock()
	log.Printf("Loaded %d zones in memory", len(zs.store))
	go zs.Lookup()
}


func localPrefetch(zs *ZoneStore, critical bool) {
	fileName := "zone.json"
	body, e := ioutil.ReadFile(fileName)
	if e != nil {
		fmt.Printf("loadDDESymbolConfigs file:%s error:%v\n",fileName, e)
		os.Exit(1)	
	}

	tmpmap := make(map[string][]Record)
	err := json.Unmarshal(body, &tmpmap)
	if err != nil && critical {
		log.Fatal("Error parsing JSON zones file: ", err, string(body))
	} else if err != nil {
		log.Println("Error parsing JSON zones file: ", err)
	}

	zs.m.Lock()
	zs.store = make(map[string]Zone)
	for key, value := range tmpmap {
		key = dns.Fqdn(key)
		if cdn, e := idna.ToASCII(key); e == nil {
			key = cdn
		}
		if zs.store[key] == nil {
			zs.store[key] = make(map[dns.RR_Header][]dns.RR)
		}
		for _, r := range value {
			if cdn, e := idna.ToASCII(r.Name); e == nil {
				r.Name = cdn
			}
			rr, err := dns.NewRR(dns.Fqdn(r.Name) + " " + r.Class + " " + r.Type + " " + r.Data)
			fmt.Println(rr)
			if err == nil {
				rr.Header().Ttl = r.Ttl
				key2 := dns.RR_Header{Name: dns.Fqdn(rr.Header().Name), Rrtype: rr.Header().Rrtype, Class: rr.Header().Class}
				zs.store[key][key2] = append(zs.store[key][key2], rr)
			} else {
				log.Printf("Skipping problematic record: %+v\nError: %+v\n", r, err)
			}
		}
	}
	zs.m.Unlock()
	log.Printf("Loaded %d zones in memory", len(zs.store))
}
