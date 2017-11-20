package main

import (
    "fmt"
    "net"
)

func main() {
    addr := net.UDPAddr{
        Port: 1119,
        IP:   net.ParseIP("0.0.0.0"),
    }
    conn, err := net.ListenUDP("udp", &addr)
    // code does not block here
    defer conn.Close()
    if err != nil {
        panic(err)
    }

    var buf [1024]byte
    for {
        rlen, remote, err := conn.ReadFromUDP(buf[:])
        // Do stuff with the read bytes
        fmt.Println(rlen)
        fmt.Println(remote)
        if err != nil {
            panic(err)
        }
    }

    var testPayload []byte = []byte("This is a test")

    conn.Write(testPayload)

}
