package main

import (
	"encoding/binary"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
	"net/http"
	"net/netip"
	"time"
)

func accelerateTo88mph(pkt gopacket.Packet) []byte {
	icmpv4 := pkt.Layer(layers.LayerTypeICMPv4).(*layers.ICMPv4)
	payload := icmpv4.LayerPayload()

	// extract timestamp from icmp echo reply
	lil := binary.LittleEndian
	secs := lil.Uint64(payload)
	usecs := lil.Uint32(payload[8:])
	ts := time.Unix(int64(secs), int64(usecs*1_000))

	// aws vpc is pretty quick, we probably only need
	// to go back 5ms in the past for our shenanigans
	ts = ts.Add(5 * time.Millisecond)
	lil.PutUint64(payload, uint64(ts.Unix()))
	lil.PutUint32(payload[8:], uint32(ts.Nanosecond()/1_000))

	// lets create a new packet sandwich. we'll let gopacket
	// handle re-calculating checksums in icmp and ipv4
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{ComputeChecksums: true}
	err := gopacket.SerializeLayers(buf, opts,
		pkt.Layer(layers.LayerTypeIPv4).(*layers.IPv4),
		&layers.ICMPv4{TypeCode: icmpv4.TypeCode, Id: icmpv4.Id, Seq: icmpv4.Seq},
		gopacket.Payload(payload))
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}

	return buf.Bytes()
}

func main() {
	// it's important that we have healthchecks for this nonsense
	go healthcheck()

	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 6081})
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}

	buf := make([]byte, 9001)
	responders := map[uint16]*net.UDPConn{}

	for {
		n, addr, err := conn.ReadFromUDPAddrPort(buf)
		if err != nil {
			panic(fmt.Sprintf("%+v", err))
		}

		// we'll only rewrite icmpv4 echo reply packets
		pkt := gopacket.NewPacket(buf[:n], layers.LayerTypeGeneve, gopacket.Default)
		icmpv4, ok := pkt.Layer(layers.LayerTypeICMPv4).(*layers.ICMPv4)
		if ok && icmpv4.TypeCode == layers.ICMPv4TypeEchoReply {
			modifiedPayload := accelerateTo88mph(pkt)
			// we don't need to modify anything at the geneve layer
			genevelen := len(pkt.Layer(layers.LayerTypeGeneve).LayerContents())
			copy(buf[genevelen:], modifiedPayload)
		}

		// gwlb expects us to send responses to dstport 6081
		// and use its srcport to avoid packet reordering
		srcport := addr.Port()
		responder := responders[srcport]
		if responder == nil {
			responder, err = net.ListenUDP("udp", &net.UDPAddr{Port: int(srcport)})
			if err != nil {
				panic(fmt.Sprintf("%+v", err))
			}
			responders[srcport] = responder
		}

		dst := netip.AddrPortFrom(addr.Addr(), 6081)
		_, err = responder.WriteToUDPAddrPort(buf[:n], dst)
		if err != nil {
			panic(fmt.Sprintf("%+v", err))
		}
	}
}

func healthcheck() {
	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
}
