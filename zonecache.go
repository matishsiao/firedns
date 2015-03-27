package main

import (
	"github.com/matishsiao/dns"
	"sync"
	"crypto/sha1"
)

var zonecachestore *ZoneCacheStore

type ZoneCacheStore struct {
	cache *sync.Mutex
	zones map[string]ZoneCache
}

type ZoneCache struct {
	rcache map[string]Cache
}

type Cache struct {
	Answer []dns.RR
	Ns	   []dns.RR
	Extra  []dns.RR
	Ttl		int
}

func GetZoneCache(zone string,question string) ([][]dns.RR) {
	if zonecachestore == nil {
		zonecachestore = &ZoneCacheStore{cache:&sync.Mutex{},zones:make(map[string]ZoneCache)}
	}
	if zonecache,ok := zonecachestore.zones[zone]; ok{
		if c,ok := zonecache.rcache[question]; ok{
			return [][]dns.RR{c.Answer,c.Ns,c.Extra}
		}
	} 
	return nil
}

func SetZoneCache(zone string,question string, ans []dns.RR,ns []dns.RR,extra []dns.RR) {
	var cache Cache
	cache.Answer = ans
	cache.Ns = ns
	cache.Extra = extra
	cache.Ttl = 300
	if _,ok := zonecachestore.zones[zone]; !ok{
		zonecachestore.zones[zone] = ZoneCache{rcache:make(map[string]Cache)}
	}
	zonecachestore.zones[zone].rcache[question] = cache
}

func QuestionKey(q dns.Question, dnssec bool) string {
	h := sha1.New()
	i := append([]byte(q.Name), packUint16(q.Qtype)...)
	if dnssec {
		i = append(i, byte(255))
	}
	return string(h.Sum(i))
}

func (zcs *ZoneCacheStore) CalcTtl() {
	zcs.cache.Lock()
	defer zcs.cache.Unlock()
	for k,zone := range zcs.zones {
		for zk,_ := range zone.rcache {
			//zcs.zones[k].rcache[zk].Ttl = 0
			if zcs.zones[k].rcache[zk].Ttl < 0 {
				delete(zcs.zones[k].rcache,zk)
			}
		}
	}
}

func packUint16(i uint16) []byte { return []byte{byte(i >> 8), byte(i)} }
func packUint32(i uint32) []byte { return []byte{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)} }