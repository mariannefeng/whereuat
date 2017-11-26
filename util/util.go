package util

import (
	"net"
	"os"
	"os/exec"
)

func FindOthers() []string {
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

	hosts, _ := hosts(host)
	concurrentMax := 100
	pingChan := make(chan string, concurrentMax)
	pongChan := make(chan Pong, len(hosts))
	doneChan := make(chan []string)

	for i := 0; i < concurrentMax; i++ {
		go ping(pingChan, pongChan)
	}

	go receivePong(len(hosts), pongChan, doneChan)

	for _, ip := range hosts {
		pingChan <- ip
	}

	alives := <-doneChan
	return alives
}

////private functions to help out/////

func hosts(cidr string) ([]string, error) {
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

func receivePong(pongNum int, pongChan <-chan Pong, doneChan chan<- []string) {
	var alives []string
	for i := 0; i < pongNum; i++ {
		pong := <-pongChan
		if pong.Alive {
			alives = append(alives, pong.Ip)
		}
	}
	doneChan <- alives
}