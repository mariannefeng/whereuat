package main

import (
	"github.com/k0kubun/pp"
	"log"
	"net"
	"os"
	"os/exec"
	"fmt"
	"bufio"
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
		//  fmt.Println("received:", pong)
		if pong.Alive {
			alives = append(alives, pong)
		}
	}
	doneChan <- alives
}

func main() {
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
	fmt.Fprintf(conn, "Hi UDP Server, How are you doing?")
	_, err = bufio.NewReader(conn).Read(p)
	if err == nil {
		fmt.Printf("%s\n", p)
	} else {
		fmt.Printf("Some error %v\n", err)
	}
	conn.Close()
}
