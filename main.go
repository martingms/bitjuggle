package main

import (
  "fmt"
  "net"
  "bytes"
  "flag"
  "os"
//  "time"
)

const ICMP_ECHO_REQUEST = 8
const ICMP_ECHO_REPLY = 0

// returns a suitable 'ping request' packet, with id & seq and a
// payload length of pktlen
func makeIcmpRequest(id, seq, pktlen int, data []byte) []byte {
    p := make([]byte, pktlen)
    copy(p[8:], bytes.Repeat(data, (pktlen - 8) / len(data) + 1))

    p[0] = ICMP_ECHO_REQUEST // type
    p[1] = 0                 // code
    p[2] = 0                 // cksum
    p[3] = 0                 // cksum
    p[4] = uint8(id >> 8)    // id
    p[5] = uint8(id & 0xff)  // id
    p[6] = uint8(seq >> 8)   // sequence
    p[7] = uint8(seq & 0xff) // sequence

    // calculate icmp checksum
    cklen := len(p)
    s := uint32(0)
    for i := 0; i < (cklen - 1); i += 2 {
        s += uint32(p[i+1])<<8 | uint32(p[i])
    }
    if cklen&1 == 1 {
        s += uint32(p[cklen-1])
    }
    s = (s >> 16) + (s & 0xffff)
    s = s + (s >> 16)

    // place checksum back in header; using ^= avoids the
    // assumption the checksum bytes are zero
    p[2] ^= uint8(^s & 0xff)
    p[3] ^= uint8(^s >> 8)

    return p
}

func parseIcmpReply(p []byte) (id, seq int) {
    id = int(p[4]) << 8 | int(p[5])
    seq = int(p[6]) << 8 | int(p[7])
    return
}

var srchost = flag.String("srchost", "", "Source of the ICMP ECHO request.")
var dsthost = flag.String("dsthost", "localhost", "Destination for the ICMP ECHO request.")

func main() {
    if os.Getuid() != 0 {
    fmt.Println("Must be root.")
        return
    }
  
    flag.Parse()

    var laddr *net.IPAddr
    if *srchost != "" {
        laddr, err := net.ResolveIPAddr("ip", *srchost)
        if err != nil {
            fmt.Println(`net.ResolveIPAddr("ip", "%v") = %v, %v`, *srchost, laddr, err)
        }
    }

    raddr, err := net.ResolveIPAddr("ip", *dsthost)
    if err != nil {
        fmt.Println(`net.ResolveIPAddr("ip", "%v") = %v, %v`, *dsthost, raddr, err)
    }

    c, err := net.ListenIP("ip4:icmp", laddr)
    if err != nil {
        fmt.Println(`net.ListenIP("ip4:icmp", %v) = %v, %v`, *srchost, c, err)
    }

    sendid := os.Getpid()
    const sendseq = 61455
    const pingpktlen = 128
    sendpkt := makeIcmpRequest(sendid, sendseq, pingpktlen, []byte("Testerlitt"))

    n, err := c.WriteToIP(sendpkt, raddr)
    if err != nil || n != pingpktlen {
        fmt.Println(`WriteToIP(..., %v) = %v, %v`, raddr, n, err)
    }

    //c.SetReadDeadline(time.Now().Add(time.Duration(10) * time.Second))
    resp := make([]byte, 1024)
    for {
        n, from, err := c.ReadFrom(resp)
        if err != nil {
            fmt.Println(`ReadFrom(...) = %v, %v, %v`, n, from, err)
        }
        if resp[0] != ICMP_ECHO_REPLY {
            continue
        }
        rcvid, rcvseq := parseIcmpReply(resp)
        if rcvid != sendid || rcvseq != sendseq {
            fmt.Println(`Ping reply saw id,seq=%v,%v (expected %v, %v)`, rcvid, rcvseq, sendid, sendseq)
        }
        fmt.Println(resp)
        return
    }
    fmt.Println("saw no ping return")
}
