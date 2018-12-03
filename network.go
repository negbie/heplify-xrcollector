package main

import (
	"fmt"
	"log"
	"net"

	"github.com/negbie/sipparser"
)

const maxPktSize = 4096

func recvXR(conn *net.UDPConn, inXRCh chan XRPacket, outHEPCh chan []byte) {
	for {
		b := make([]byte, maxPktSize)
		n, addr, err := conn.ReadFromUDP(b)
		if err != nil {
			log.Println("Error on XR read: ", err)
			continue
		}
		if cfg.Debug {
			log.Printf("Received following RTCP-XR PUBLISH with %d bytes from %s:\n%s\n", n, addr, string(b[:n]))
		} else {
			log.Printf("Received RTCP-XR PUBLISH with %d bytes from %s\n", n, addr)
		}
		var msg []byte
		if msg, err = processPublish(b[:n]); err != nil {
			log.Println(err)
			continue
		}
		inXRCh <- XRPacket{addr, msg}
		outHEPCh <- b[:n]
	}
}

func sendXR(conn *net.UDPConn, outXRCh chan XRPacket) {
	for packet := range outXRCh {
		n, err := conn.WriteToUDP(packet.data, packet.addr)
		if err != nil {
			log.Println("Error on XR write: ", err)
			continue
		}
		if cfg.Debug {
			log.Printf("Sent following SIP/2.0 200 OK with %d bytes to %s:\n%s\n", n, packet.addr, string(packet.data))
		} else {
			log.Printf("Sent SIP/2.0 200 OK with %d bytes to %s\n", n, packet.addr)
		}
	}
}

func sendHEP(conn net.Conn, outHEPCh chan []byte) {
	for packet := range outHEPCh {
		_, err := conn.Write(encodeHEP(packet, 35))
		if err != nil {
			log.Println("Error on HEP write: ", err)
			continue
		}
	}
}

func processPublish(pkt []byte) ([]byte, error) {
	sip := sipparser.ParseMsg(string(pkt))
	if sip.Error != nil {
		return nil, sip.Error
	}
	if sip.CseqMethod != "PUBLISH" || sip.ContentType != "application/vq-rtcpxr" || len(sip.Body) < 32 {
		return nil, fmt.Errorf("No or malformed vq-rtcpxr inside SIP PUBLISH:\n%s", sip.Msg)
	}

	resp := fmt.Sprintf("SIP/2.0 200 OK\r\nVia: %s\r\nFrom: %s\r\nTo: %s;tag=Fg2Uy0r7geBQF\r\nContact: %s\r\n"+
		"Call-ID: %s\r\nCseq: %s\r\nUser-Agent: heplify-xrcollector\r\nContent-Length: 0\r\n\r\n",
		sip.ViaOne,
		sip.From.Val,
		sip.To.Val,
		sip.ContactVal,
		sip.CallID,
		sip.Cseq.Val)
	return []byte(resp), nil
}
