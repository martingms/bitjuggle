package main

import (
  "net"
  "fmt"
  "runtime"
)

func juggle_receive(c *net.IPConn, buffer *[]byte, buffer_avail *[10]int) {
    fmt.Println("--> juggle_receive-thread started.")
    resp := make([]byte, 1032)
    for {
        _, _, err := c.ReadFrom(resp)

        fmt.Println(*buffer_avail)

        if err!= nil {
            panic(err)
        }

        if resp[0] != ICMP_ECHO_REPLY {
            continue
        }

        _, _, rcvdata := parseIcmpReply(resp)

        var found = false
        for i := 0; i < len(*buffer_avail); i++ {
            if (*buffer_avail)[i] == 0 {
                copy((*buffer)[1024 * i:], rcvdata) // TODO: check numbers
                (*buffer_avail)[i] = 1
                fmt.Println("got data!", *buffer_avail)
                found = true
                break
            }
        }
        if !found { panic("Buffer full!") }
        runtime.Gosched()
    }
}

func juggle_send(c *net.IPConn, buffer *[]byte, buffer_avail *[10]int) {
    fmt.Println("--> juggle_send-thread started.")
    for {
        for i := 0; i < len(*buffer_avail); i++ {
            if (*buffer_avail)[i] == 1 {
                sendIcmpPacket((*buffer)[1024 * i:1024 * (i+1)], choose_rand_host(), c)
                (*buffer_avail)[i] = 0
                fmt.Println("sent data!", *buffer_avail)
            }
        }
        runtime.Gosched()
    }
}
