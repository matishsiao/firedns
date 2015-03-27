package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	zones = &ZoneStore{
		store: make(map[string]Zone),
		m:     new(sync.RWMutex),
		seri:  make(map[string]uint64),
	}
	config     = &Configs{
		ip:		"10.5.4.59",
		port:	8888,
		auth:	"",
	}
	ssdbip	string
	ssdbport	int
	ssdbauth	string
	listenOn    string
	recurseTo   string
	apiKey      string
	buildtime   string
	buildcommit string
	debug		bool
	
)

type Configs struct {
	ip			string
	port		int
	auth		string
}

func main() {
	flag.StringVar(&ssdbip, "c", "10.5.4.59", "The ssdb ip")
	flag.IntVar(&ssdbport, "p", 8888, "The ssdb port")
	flag.BoolVar(&debug,"d",false,"debug mode")
	flag.StringVar(&ssdbauth, "a", "", "The ssdb auth password")
	flag.StringVar(&listenOn, "l", "", "The IP to listen on (default = blank = ALL)")
	flag.StringVar(&recurseTo, "r", "", "Pass-through requests that we can't answer to other DNS server (address:port or empty=disabled)")
	flag.StringVar(&apiKey, "k", "", "API key for http notifications")
	flag.Parse()
	
	log.Println("firedns (2015) by Matis Hsiao is starting...")
	
	config.ip = ssdbip
	config.port = ssdbport
	config.auth	= ssdbauth
	log.Printf("ssdb ip:%s:%d use auth:%s", config.ip,config.port,config.auth)
	Connect(config.ip,config.port,config.auth)
	prefetch(zones, true)

	server := &Server{
		host:     listenOn,
		port:     53,
		rTimeout: 5 * time.Second,
		wTimeout: 5 * time.Second,
		zones:    zones,
	}

	server.Run()

	go StartHTTP()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case s := <-sig:
			log.Fatalf("Signal (%d) received, stopping\n", s)
		}
	}
}
