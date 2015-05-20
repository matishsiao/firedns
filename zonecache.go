package main

import (
	"github.com/matishsiao/dns"
	"sync"
	"crypto/sha1"
	_"strings"
	_"log"
)

var zonecachestore *ZoneCacheStore

type ZoneCacheStore struct {
	cache *sync.RWMutex
	ncache map[string]int
	rcache map[string]*Cache
	size int
}

/*type ZoneCache struct {
	rcache map[string]Cache
}*/

type Cache struct {
	Key	   string
	Answer *dns.Msg
	Ttl		int64
}

func GetZoneCache(question string) *dns.Msg {
	if zonecachestore == nil {
		//zonecachestore = &ZoneCacheStore{cache:&sync.RWMutex{},rcache:make([]Cache,1000)}
		zonecachestore = &ZoneCacheStore{cache:&sync.RWMutex{},rcache:make(map[string]*Cache),ncache:make(map[string]int)}
		//zonecachestore = &ZoneCacheStore{cache:&sync.RWMutex{}}
	}
	if v,ok := zonecachestore.rcache[question];ok{
		return v.Answer.Copy()
	}
	return nil
}

func SetZoneCache(question string,msg *dns.Msg) {
	cache := &Cache{Answer:msg,Ttl:300}
	zonecachestore.rcache[question] = cache
	zonecachestore.size++
}

func GetNZoneCache(zone string) bool {
	if zonecachestore == nil {
		//zonecachestore = &ZoneCacheStore{cache:&sync.RWMutex{},rcache:make([]Cache,1000)}
		zonecachestore = &ZoneCacheStore{cache:&sync.RWMutex{},rcache:make(map[string]*Cache),ncache:make(map[string]int)}
		//zonecachestore = &ZoneCacheStore{cache:&sync.RWMutex{}}
	}
	if _,ok := zonecachestore.ncache[zone];ok{
		zonecachestore.ncache[zone]++
		return true
	}
	return false
}

func SetNZoneCache(zone string) {
	if zonecachestore == nil {
		//zonecachestore = &ZoneCacheStore{cache:&sync.RWMutex{},rcache:make([]Cache,1000)}
		zonecachestore = &ZoneCacheStore{cache:&sync.RWMutex{},rcache:make(map[string]*Cache),ncache:make(map[string]int)}
		//zonecachestore = &ZoneCacheStore{cache:&sync.RWMutex{}}
	}
	zonecachestore.ncache[zone] = 1
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
	for k,v := range zcs.rcache {
		//for zk,_ := range zone.rcache {
			//zcs.zones[k].rcache[zk].Ttl = 0
			if v.Ttl < 0 {
				delete(zcs.rcache,k)
				//zcs.rcache = zcs.rcache[k:]
			}
		//}
	}
}

func packUint16(i uint16) []byte { return []byte{byte(i >> 8), byte(i)} }
func packUint32(i uint32) []byte { return []byte{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)} }