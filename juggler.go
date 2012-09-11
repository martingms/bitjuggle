package main

import (
  "net"
  "fmt"
)

func juggle_receive(c *net.IPConn, channel chan<- []byte) {
    fmt.Println("--> juggle_receive-thread started.")
    resp := make([]byte, 1032)
    for {
        _, _, err := c.ReadFrom(resp)
        if err!= nil {
            panic(err)
        }

        if resp[0] != ICMP_ECHO_REPLY {
            continue
        }

        _, _, rcvdata := parseIcmpReply(resp)

        fmt.Println("--> ", len(rcvdata), "bytes received")

        channel <- rcvdata
    }
}

func juggle_send(c *net.IPConn, channel <-chan []byte) {
    fmt.Println("--> juggle_send-thread started.")
    for {
        sendIcmpPacket(<- channel, choose_rand_host(), c)
        fmt.Println("--> Some bytes sent")
    }
}
