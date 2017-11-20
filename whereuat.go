package main

import (
	"github.com/k0kubun/pp"
	"log"
	"net"
	"os"
	"os/exec"
	"fmt"
)

var port = "1119"

func Hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}

//  http://play.golang.org/p/m8TNTtygK0
//TODO: obviously there's a way to optimize (also once we add cache really won't be that bad)
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

type Pong struct {
	Ip    string
	Alive bool
}

func ping(pingChan <-chan string, pongChan chan<- Pong) {
	for ip := range pingChan {
		_, err := exec.Command("ping", "-c1", "-t1", ip).Output()
		//if ip == "192.168.1.12" {
		//	fmt.Println(string(pingOutput))
		//}
		var alive bool
		if err != nil {
			alive = false
		} else {
			alive = true
		}
		pongChan <- Pong{Ip: ip, Alive: alive}
	}
}

func receivePong(pongNum int, pongChan <-chan Pong, doneChan chan<- []Pong) {
	var alives []Pong
	for i := 0; i < pongNum; i++ {
		pong := <-pongChan
		//fmt.Println("received:", pong)
		if pong.Alive {
			alives = append(alives, pong)
		}
	}
	doneChan <- alives
}

func main() {
	go letsListen()

	var host = ""
	addrs, err := net.Interfaces()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, a := range addrs {
		//TODO: change this to use different values for different operating systems
		if a.Name == "en0" {
			addrList, err := a.Addrs()
			if err != nil {
				os.Stderr.WriteString("Oops: " + err.Error() + "\n")
				os.Exit(1)
			}
			for _, addr := range addrList {
				if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
					host = addr.String()
				}

			}
		}
	}

	hosts, _ := Hosts(host)
	concurrentMax := 100
	pingChan := make(chan string, concurrentMax)
	pongChan := make(chan Pong, len(hosts))
	doneChan := make(chan []Pong)

	for i := 0; i < concurrentMax; i++ {
		go ping(pingChan, pongChan)
	}

	go receivePong(len(hosts), pongChan, doneChan)

	for _, ip := range hosts {
		pingChan <- ip
		//  fmt.Println("sent: " + ip)
	}

	alives := <-doneChan
	for _, alive := range alives {
		checkUDP(alive.Ip)
	}
	pp.Println(alives)
}

func checkUDP(addr string) {
	//var status string
	otherAddr :=  addr + ":" + port
	log.Println(otherAddr)
	p :=  make([]byte, 2048)

	conn, err := net.Dial("udp", otherAddr)

	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}

	go letsTalk(conn, p)

}

func letsListen() {
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

func letsTalk(conn net.Conn, p []byte) {
	fmt.Fprintf(conn, "haygurl")

	for {
		_, err := conn.Read(p)
		if err == nil {
			fmt.Printf("%s\n", p)
		} else {
			fmt.Printf("Some error %v\n", err)
			break
		}
	}

	conn.Close()
}


func sendResponse(conn *net.UDPConn, addr *net.UDPAddr) {
	_,err := conn.WriteToUDP([]byte("From server: Hello I got your mesage "), addr)
	if err != nil {
		fmt.Printf("Couldn't send response %v", err)
	}
}
