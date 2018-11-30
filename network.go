package main

import (
	"fmt"
	"log"
	"net"

	"github.com/negbie/sipparser"
)

func sendXR(conn *net.UDPConn, outXRCh chan XRPacket) {
	for packet := range outXRCh {
		n, err := conn.WriteToUDP(packet.data, packet.addr)
		if err != nil {
			log.Println("Error on XR write: ", err)
			continue
		}
		log.Printf("Sent following RTCP-XR PUBLISH packet with %d bytes:\n%s\n", n, string(packet.data))
	}
}

func recvXR(conn *net.UDPConn, inXRCh chan XRPacket, outHEPCh chan []byte) {
	for {
		b := make([]byte, MAX_XR_PACKET_SIZE)
		n, addr, err := conn.ReadFromUDP(b)
		if err != nil {
			log.Println("Error on XR read: ", err)
			continue
		}
		log.Printf("Received following RTCP-XR PUBLISH packet with %d bytes:\n%s\n", n, string(b[:n]))
		if msg, err := processPublish(b[:n]); err == nil {
			inXRCh <- XRPacket{addr, msg}
			outHEPCh <- b[:n]
		} else {
			log.Println(err)
			inXRCh <- XRPacket{addr, b[:n]}
		}
	}
}

func sendHEP(conn net.Conn, outHEPCh chan []byte) {
	for packet := range outHEPCh {
		_, err := conn.Write(encodeHEP(packet))
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
	if sip.ContentType != "application/vq-rtcpxr" || len(sip.Body) < 32 {
		return nil, fmt.Errorf("No vq-rtcpxr inside SIP PUBLISH:\n%s", sip.Msg)
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
