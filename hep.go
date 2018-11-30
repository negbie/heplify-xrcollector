package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"time"
)

var (
	hepBuf   bytes.Buffer
	hepVer   = []byte{0x48, 0x45, 0x50, 0x33} // "HEP3"
	hepLen   = []byte{0x00, 0x00}
	hepLen7  = []byte{0x00, 0x07}
	hepLen8  = []byte{0x00, 0x08}
	hepLen10 = []byte{0x00, 0x0a}
	chunck16 = []byte{0x00, 0x00}
	chunck32 = []byte{0x00, 0x00, 0x00, 0x00}
)

func encodeHEP(payload []byte) []byte {
	hepMsg := append([]byte{}, makeHEPChuncks(payload)...)
	binary.BigEndian.PutUint16(hepMsg[4:6], uint16(len(hepMsg)))
	return hepMsg
}

// makeHEPChuncks will construct the respective HEP chunck
func makeHEPChuncks(payload []byte) []byte {
	hepBuf.Reset()
	hepBuf.Write(hepVer)
	// hepMsg length placeholder. Will be written later
	hepBuf.Write(hepLen)

	// Chunk IP protocol family (0x02=IPv4, 0x0a=IPv6)
	hepBuf.Write([]byte{0x00, 0x00, 0x00, 0x01})
	hepBuf.Write(hepLen7)
	hepBuf.WriteByte(0x02)

	// Chunk IP protocol ID (0x06=TCP, 0x11=UDP)
	hepBuf.Write([]byte{0x00, 0x00, 0x00, 0x02})
	hepBuf.Write(hepLen7)
	hepBuf.WriteByte(0x11)

	// Chunk IPv4 source address
	srcIP := net.IPv4(1, 1, 1, 1).To4()
	hepBuf.Write([]byte{0x00, 0x00, 0x00, 0x03})
	binary.BigEndian.PutUint16(hepLen, 6+uint16(len(srcIP)))
	hepBuf.Write(hepLen)
	hepBuf.Write(srcIP)

	// Chunk IPv4 destination address
	dstIP := net.IPv4(2, 2, 2, 2).To4()
	hepBuf.Write([]byte{0x00, 0x00, 0x00, 0x04})
	binary.BigEndian.PutUint16(hepLen, 6+uint16(len(dstIP)))
	hepBuf.Write(hepLen)
	hepBuf.Write([]byte(dstIP))

	// Chunk protocol source port
	hepBuf.Write([]byte{0x00, 0x00, 0x00, 0x07})
	hepBuf.Write(hepLen8)
	binary.BigEndian.PutUint16(chunck16, 1111)
	hepBuf.Write(chunck16)

	// Chunk protocol destination port
	hepBuf.Write([]byte{0x00, 0x00, 0x00, 0x08})
	hepBuf.Write(hepLen8)
	binary.BigEndian.PutUint16(chunck16, 2222)
	hepBuf.Write(chunck16)

	// Chunk unix timestamp, seconds
	hepBuf.Write([]byte{0x00, 0x00, 0x00, 0x09})
	hepBuf.Write(hepLen10)
	binary.BigEndian.PutUint32(chunck32, uint32(time.Now().Unix()))
	hepBuf.Write(chunck32)

	// Chunk unix timestamp, microseconds offset
	hepBuf.Write([]byte{0x00, 0x00, 0x00, 0x0a})
	hepBuf.Write(hepLen10)
	binary.BigEndian.PutUint32(chunck32, uint32(time.Now().Nanosecond()/1000))
	hepBuf.Write(chunck32)

	// Chunk protocol type (DNS, LOG, RTCP, SIP)
	hepBuf.Write([]byte{0x00, 0x00, 0x00, 0x0b})
	hepBuf.Write(hepLen7)
	hepBuf.WriteByte(35)

	// Chunk capture agent ID
	hepBuf.Write([]byte{0x00, 0x00, 0x00, 0x0c})
	hepBuf.Write(hepLen10)
	binary.BigEndian.PutUint32(chunck32, 2222)
	hepBuf.Write(chunck32)

	// Chunk captured packet payload
	hepBuf.Write([]byte{0x00, 0x00, 0x00, 0x0f})
	binary.BigEndian.PutUint16(hepLen, 6+uint16(len(payload)))
	hepBuf.Write(hepLen)
	hepBuf.Write(payload)

	var cid []byte
	if posCallID := bytes.Index(payload, []byte("CallID:")); posCallID > 0 {
		restCallID := payload[posCallID:]
		// Minimum length of "CallID:x" = 8
		if posRestCallID := bytes.Index(restCallID, []byte("\r\n")); posRestCallID >= 8 {
			cid = restCallID[len("CallID:"):posRestCallID]
			i := 0
			for i < len(cid) && (cid[i] == ' ' || cid[i] == '\t') {
				i++
			}
			cid = cid[i:]
		} else {
			log.Printf("no end or fishy CallID in '%s'\n", payload)
		}
	}

	if cid != nil {
		// Chunk internal correlation id
		hepBuf.Write([]byte{0x00, 0x00, 0x00, 0x11})
		binary.BigEndian.PutUint16(hepLen, 6+uint16(len(cid)))
		hepBuf.Write(hepLen)
		hepBuf.Write(cid)
	}

	return hepBuf.Bytes()
}
