package main

import (
    "fmt"
    "net"
)
//

func sendResponse(conn *net.UDPConn, addr *net.UDPAddr) {
	_,err := conn.WriteToUDP([]byte("From server: Hello I got your mesage "), addr)
	if err != nil {
		fmt.Printf("Couldn't send response %v", err)
	}
}

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

	buf := make([]byte, 1024)
    for {
        rlen, remote, err := conn.ReadFromUDP(buf)
        // Do stuff with the read bytes
        fmt.Println(rlen)
        fmt.Println(remote)
        fmt.Println(string(buf[:rlen]))
        if err != nil {
            panic(err)
        }
        go sendResponse(conn, remote)
    }

    var testPayload = []byte("This is a test")

    conn.Write(testPayload)

}
//
//func sendResponse(conn *net.UDPConn, addr *net.UDPAddr) {
//	_,err := conn.WriteToUDP([]byte("From server: Hello I got your mesage "), addr)
//	if err != nil {
//		fmt.Printf("Couldn't send response %v", err)
//	}
//}
//
//
//func main() {
//	p := make([]byte, 2048)
//	addr := net.UDPAddr{
//		Port: 1119,
//		IP: net.ParseIP("0.0.0.0"),
//	}
//	ser, err := net.ListenUDP("udp", &addr)
//	if err != nil {
//		fmt.Printf("Some error %v\n", err)
//		return
//	}
//	for {
//		_,remoteaddr,err := ser.ReadFromUDP(p)
//		fmt.Printf("Read a message from %v %s \n", remoteaddr, p)
//		if err !=  nil {
//			fmt.Printf("Some error  %v", err)
//			continue
//		}
//		go sendResponse(ser, remoteaddr)
//	}
//}