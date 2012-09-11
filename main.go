package main

import (
  "fmt"
  "flag"
  "os"
  "net"
  "math/rand"
  "time"
  "io"
)
                        //yahoo.jp       yahoo.jp          //yahoo.com.au    yahoo.cn
var hostlist = []string{"183.79.23.196", "124.83.219.204", "203.84.217.229", "202.165.102.205"}

func choose_rand_host() (host string) {
    rand.Seed(time.Now().UnixNano())
    host = hostlist[rand.Intn(len(hostlist))]
    return
}

var file = flag.String("file", "", "File to be juggled.")
var srchost = flag.String("srchost", "", "Source of the ICMP ECHO request.")

func main() {
    flag.Parse()

    if os.Getuid() != 0 {
    fmt.Println("Must be root.")
        return
    }

    // Open file.
    fi, err := os.Open(*file)
    if err != nil {
        panic(err)
    }

    // Resolve local address.
    var laddr *net.IPAddr
    if *srchost != "" {
        laddr, err := net.ResolveIPAddr("ip", *srchost)
        if err != nil {
            fmt.Println(`net.ResolveIPAddr("ip", "%v") = %v, %v`, *srchost, laddr, err)
        }
    }

    // Create IP-connection
    c, err := net.ListenIP("ip4:icmp", laddr)
    if err != nil {
        fmt.Println(`net.ListenIP("ip4:icmp", %v) = %v, %v`, *srchost, c, err)
    }

    // Initialize the main channel
    channel := make(chan []byte, 16)

    go juggle_receive(c, channel)

    // To initialize, send each 1024B block of the file.
    file_buffer := make([]byte, 1024)
    for {
        n, err := fi.Read(file_buffer)
        if err != nil && err != io.EOF { panic(err) }
        if n == 0 { break }

        sendIcmpPacket(file_buffer, choose_rand_host(), c)
    }

    //os.Remove(*file)
    fmt.Println("--> File now only exists in teh wires...")

    juggle_send(c, channel)
}
