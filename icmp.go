package main

import (
    "os"
    "net"
)

const (
    ICMP_ECHO_REQUEST = 8
    ICMP_ECHO_REPLY = 0
    PACKET_LENGTH = 1032
    SEQ = 61455
)

func sendIcmpPacket(data []byte, dsthost string, c *net.IPConn) {
    // Creating the packet to be sent.
    p := createIcmpPacket(SEQ, PACKET_LENGTH, data)

    // Resolving remote host address.
    raddr, err := net.ResolveIPAddr("ip", dsthost)
    if err != nil {
        panic(err)
    }

    n, err := c.WriteToIP(p, raddr)
    if err != nil || n != PACKET_LENGTH {
        panic(err)
    }
}

// returns a suitable 'ping request' packet, with id & seq and a
// payload length of pktlen
func createIcmpPacket(seq, pktlen int, data []byte) (p []byte) {
    p = make([]byte, pktlen)
    copy(p[8:], data) // FIXME: overflows

    id := os.Getpid()

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

    return
}

func parseIcmpReply(p []byte) (id, seq int, data []byte) {
    id = int(p[4]) << 8 | int(p[5])
    seq = int(p[6]) << 8 | int(p[7])
    data = p[8:]
    return
}
