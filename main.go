package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

type XRPacket struct {
	addr *net.UDPAddr
	data []byte
}

const MAX_XR_PACKET_SIZE = 4096
const VERSION = "heplify-xrcollector 0.1"

var (
	hepServerAddress string
	collectorAddress string
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Use %s like: %s [option]\n", VERSION, os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(&hepServerAddress, "hs", "127.0.0.1:9060", "HEP UDP server address")
	flag.StringVar(&collectorAddress, "xs", ":9064", "XR collector UDP listen address")
	flag.Parse()
}

func main() {
	UDPAddr, err := net.ResolveUDPAddr("udp", hepServerAddress)
	if err != nil {
		log.Fatalln(err)
	}
	connXR, err := net.ListenUDP("udp", UDPAddr)
	if err != nil {
		log.Fatalln(err)
	}

	connHEP, err := net.Dial("udp", hepServerAddress)
	if err != nil {
		log.Fatalln(err)
	}

	outXRCh := make(chan XRPacket, 100)
	outHEPCh := make(chan []byte, 100)
	inXRCh := make(chan XRPacket, 100)

	go sendXR(connXR, outXRCh)
	go recvXR(connXR, inXRCh, outHEPCh)
	go sendHEP(connHEP, outHEPCh)

	for packet := range inXRCh {
		outXRCh <- packet
	}
}
