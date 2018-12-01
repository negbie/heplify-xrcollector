package main

import (
	"log"
	"net"
	"testing"
	"time"
)

var (
	listnXR = "127.0.0.1:9064"
	addrHEP = "localhost:9060"
	publish = "PUBLISH sip:87.103.120.253:9070 SIP/2.0\r\n" +
		"Via: SIP/2.0/UDP 10.0.3.13:3072;branch=z9hG4bK-2atcagwblzv2;rport\r\n" +
		"From: <sip:5004@10.0.3.252>;tag=2ygtpy7bgk\r\n" +
		"To: <sip:87.103.120.253:9070>\r\n" +
		"Call-ID: 89596257635d-ip18q8n0lp1b\r\n" +
		"CSeq: 2 PUBLISH\r\n" +
		"Max-Forwards: 70\r\n" +
		"Contact: <sip:5004@10.0.3.13:3072;line=swv8im3f>;reg-id=1\r\n" +
		"User-Agent: snom821/873_19_20130321\r\n" +
		"Event: vq-rtcpxr\r\n" +
		"Accept: application/sdp, message/sipfrag\r\n" +
		"Content-Type: application/vq-rtcpxr\r\n" +
		"Content-Length: 804\r\n\r\n" +
		"VQSessionReport: CallTerm\r\n" +
		"CallID:825962570309-8ds5sl3mca99\r\n" +
		"LocalID:<sip:5004@10.0.3.252>\r\n" +
		"RemoteID:<sip:520@10.0.3.252;user=phone>\r\n" +
		"OrigID:<sip:5004@10.0.3.252>\r\n" +
		"LocalAddr:IP=10.0.3.13 PORT=57460 SSRC=0x014EA261\r\n" +
		"LocalMAC:0004135310DB\r\n" +
		"RemoteAddr:IP=10.0.3.252 PORT=10034 SSRC=0x1F634EA2\r\n" +
		"DialogID:825962570309-8ds5sl3mca99;to-tag=gqj87t0stF-M8g.kPREKLthaGl030mze;from-tag=2ygtpy7bgk\r\n" +
		"x-UserAgent:snom821/873_19_20130321\r\n" +
		"LocalMetrics:\r\n" +
		"Timestamps:START=2016-06-16T07:47:14Z STOP=2016-06-16T07:47:21Z\r\n" +
		"SessionDesc:PT=8 PD=PCMA SR=8000 PPS=50 SSUP=off\r\n" +
		"x-SIPmetrics:SVA=RG SRD=392 SFC=0\r\n" +
		"x-SIPterm:SDC=OK SDT=7 SDR=OR\r\n" +
		"JitterBuffer:JBA=3 JBR=2 JBN=20 JBM=20JBX=240\r\n" +
		"PacketLoss:NLR=0.0 JDR=0.0\r\n" +
		"BurstGapLoss:BLD=0.0 BD=0 GLD=0.0 GD=5930 GMIN=16\r\n" +
		"Delay:RTD=0 ESD=0 IAJ=0\r\n" +
		"QualityEst:MOSLQ=4.1 MOSCQ=4.1\r\n"
)

func TestMain(t *testing.T) {
	addrXR, err := net.ResolveUDPAddr("udp", listnXR)
	if err != nil {
		log.Fatalln(err)
	}

	connXR, err := net.ListenUDP("udp", addrXR)
	if err != nil {
		log.Fatalln(err)
	}

	connHEP, err := net.Dial("udp", addrHEP)
	if err != nil {
		log.Fatalln(err)
	}

	connXROut, err := net.Dial("udp", listnXR)
	if err != nil {
		log.Fatalln(err)
	}

	inXRCh := make(chan XRPacket, 10)
	outXRCh := make(chan XRPacket, 10)
	outHEPCh := make(chan []byte, 10)

	go recvXR(connXR, inXRCh, outHEPCh)
	go sendXR(connXR, outXRCh)
	go sendHEP(connHEP, outHEPCh)

	for i := 0; i < 3; i++ {
		_, err := connXROut.Write([]byte(publish))
		if err != nil {
			log.Fatalln(err)
		}
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		close(inXRCh)
	}()

	for packet := range inXRCh {
		outXRCh <- packet
	}
}
