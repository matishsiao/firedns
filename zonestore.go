package main

import (
	"github.com/matishsiao/dns"
	"log"
	"sync"
	"github.com/matishsiao/gossdb/ssdb"
	"code.google.com/p/go.net/idna"
	"encoding/json"
	"strconv"
	_"strings"
	"time"
)

type ZoneStore struct {
	store map[string]Zone
	seri  map[string]uint64
	zones map[string]string
	m     *sync.RWMutex
}

type Zone map[dns.RR_Header][]dns.RR

var db *ssdb.Client

func Connect(ip string,port int,auth string) (*ssdb.Client,error) {
    cli, err := ssdb.Connect(ip, port)
    if err != nil {
    	log.Println("ssdb connect error:%v\n",err)
    	return nil,err
    }
    if auth != "" {
    	cli.Auth(auth)
    }
    db = cli
    return cli,nil
 }

func GetZones() map[string][]Record {
	/*db,err := Connect(config.ip,config.port,config.auth)
	if err != nil {
		log.Println("GetZones Error get zones file: ", err)
		return nil
	}*/
	zones,err := db.HashGetAll("zones")
	if err != nil {
		log.Println("Error get zones file: ", err)
		return nil		
	}
	maps := make(map[string][]Record)
	for k,v := range zones {
		var zone []Record
		zonestr := v.(string)
		err := json.Unmarshal([]byte(zonestr), &zone)
		if err != nil {
			log.Println("Error parsing JSON zones file: ", err, v)
			continue
		} 
		
		maps[k] = zone
	}
	return maps
}

func (zs *ZoneStore) GetZone(zonename string) bool {
	log.Println("GetZone:",zonename)
	newzone,err := db.HashGet("zones",zonename)
	if err != nil || newzone == nil {
		//log.Println("Error get zones file from GetZone: ", err)
		return false
	}
	maps := make(map[string][]Record)
	var zone []Record
	zonestr := newzone.(string)
	err = json.Unmarshal([]byte(zonestr), &zone)
	if err != nil {
		log.Println("Error parsing JSON zones: ", err, newzone)
		return false
	}
	maps[zonename] = zone
	log.Println("zones records:",len(maps))
	zs.updateZones(maps)
	return true
}

func (zs *ZoneStore) GetZoneList() {
	log.Println("GetZoneList:")
	size,err := db.HashSize("zones")
	if err != nil {
		return 
	}
	keys,err := db.HashKeys("zones","","",size.(int64))
	if err != nil || keys == nil {
		//log.Println("Error get zones file from GetZone: ", err)
		return
	}
	log.Println(keys)
	for _,v := range keys.([]string) {
		zs.zones[v] = v
	}
	log.Println("zs.zones:",zs.zones)
}

func (zs *ZoneStore) CheckZone(zone string) bool {
	if _,ok := zs.zones[zone];ok{
		return true
	}
	return false
}

func (zs *ZoneStore) Lookup() {
	for {
		go zs.lookup()
		time.Sleep(30 * time.Second)
	}
}

func (zs *ZoneStore) lookup() {
	/*db,err := Connect(config.ip,config.port,config.auth)
	if err != nil {
		log.Println("lookup Error get zones file: ", err)
		return
	}*/
	zones,err := db.HashGetAll("zones_ser")
	if err != nil {
		log.Println("lookup Error get zones serial numbers: ", err)
		return
	}
	
	/*for k,v := range zones {
		log.Printf("ssdb zones ser number[%s]:%v\n",k,v)
	}*/
	
	var updatelist []string
	if len(zs.seri) == 0 {
		for k,v := range zones {
			i,err := strconv.ParseUint(v.(string),10,64)
			if err != nil {
				i = 0
			}
			zs.seri[k] = i
		}
	} else {
		for k,v := range zones {
			i,err := strconv.ParseUint(v.(string),10,64)
			if err != nil {
				i = 0
			}
			if zv,ok := zs.seri[k]; ok{
				//if serial number less current number then update it
				
				if zv < i {
					updatelist = append(updatelist,k)
				}
				
			} else {
				//new zone update it
				updatelist = append(updatelist,k)
			}
			zs.seri[k] = i
			//log.Printf("zones ser number[%s]:%v now:%d\n",k,v,zs.seri[k])
		}
	}
	
	if len(updatelist) > 0 {
		newzones,err := db.HashMultiGet("zones",updatelist)
		if err != nil {
			log.Println("Error get zones file from lookup: ", err)
			return
		}
		maps := make(map[string][]Record)
		for k,v := range newzones {
			var zone []Record
			zonestr := v.(string)
			err := json.Unmarshal([]byte(zonestr), &zone)
			if err != nil {
				log.Fatal("Error parsing JSON zones: ", err, v)
			} else if err != nil {
				log.Println("Error parsing JSON zones: ", err)
			}
			log.Println(zone)
			maps[k] = zone
		}
		log.Println("lookup find new zones:",maps)
		zs.updateZones(maps)
	}
}

func (zs *ZoneStore) updateZones(tmpmap map[string][]Record) {
	zs.m.Lock()
	for key, value := range tmpmap {
		key = dns.Fqdn(key)
		if cdn, e := idna.ToASCII(key); e == nil {
			key = cdn
		}
		//flush old records
		zs.store[key] = make(map[dns.RR_Header][]dns.RR)
		
		for _, r := range value {
			if cdn, e := idna.ToASCII(r.Name); e == nil {
				r.Name = cdn
			}
			rr, err := dns.NewRR(dns.Fqdn(r.Name) + " " + r.Class + " " + r.Type + " " + r.Data)
			log.Println(rr)
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
}


func (zs *ZoneStore) match(q string, t uint16) (*Zone, string) {
	/*if debug {
		log.Println("match question:",q,t)
	}
	zs.m.RLock()
	defer zs.m.RUnlock()
	*/
	var zone *Zone
	var name string
	b := make([]byte, len(q)) // worst case, one label of length q
	off := 0
	end := false
	for {
		l := len(q[off:])
		for i := 0; i < l; i++ {
			b[i] = q[off+i]
			if b[i] >= 'A' && b[i] <= 'Z' {
				b[i] |= ('a' - 'A')
			}
		}
		
		/*if debug {
			log.Println("match:",string(b[:l]))
		}*/
		label := string(b[:l])
		//log.Println("label:",label)
		if z, ok := zs.store[label]; ok { // 'causes garbage, might want to change the map key
			//log.Println("find zone:",label)
			if t != dns.TypeDS {
				return &z, label
			} else {
				// Continue for DS to see if we have a parent too, if so delegeate to the parent
				zone = &z
				name = label
			}
		} else {
			if zs.CheckZone(label) {
				if zs.GetZone(label) {
					if z, ok := zs.store[label]; ok { // 'causes garbage, might want to change the map key
						//log.Println("find zone:",label)
						if t != dns.TypeDS {
							return &z, label
						} else {
							// Continue for DS to see if we have a parent too, if so delegeate to the parent
							zone = &z
							name = label
						}
					}
				}	
			}
		}	 /*else if !GetNZoneCache(name) {
			if zs.GetZone(name) {
				if z, ok := zs.store[name]; ok { // 'causes garbage, might want to change the map key
					if t != dns.TypeDS {
						return &z, name
					} else {
						// Continue for DS to see if we have a parent too, if so delegeate to the parent
						zone = &z
						name = name
					}
				}
			} else {
				SetNZoneCache(name)
				return nil, name
			}
		}*/
		off, end = dns.NextLabel(q, off)
		if end {
			break
		}
	}
	return zone, name
}

func (zs *ZoneStore) IsNCache(q string) (bool) {
	return false
}
