package main

import (
	"log"
	"net"
	"os"
	"fmt"
	"bufio"
	"bytes"
	"strings"
	"github.com/mariannefeng/whereuat/util"
)

//TODO: SET MAX NUMBER OF BYTES PER MESSAGE TO BE 1119?
//TODO: make it take in std.in (so we can write!)
//TODO: add cache
//TODO: ssl (this is a biggie)
//TODO: send protobuf so we can give ourselves names
//TODO: terminal coloring
//TODO: add support for different operating systems (not just mac)
//TODO: figure out why the fuck finding each other doesn't always work
//TODO: obviously there's a way to optimize (also once we add cache really won't be that bad)
//TODO: figure out connect protocol:
	//TODO: this is where you would tell other people what's YOUR listening port
	//TODO: this would also tell other people what to call you
//TODO: can we send protobuf over udp? if so, do it
//TODO: cleanup main, this shit's a mess. Move stuff to app. separate files if necessary

var port = "1119"
var others []string

var maxbytes = 1119
//TODO: this needs to be different for different operating systems
var whereuat = "/Users/mariannefeng/.whereuat"


func main() {
	quitCh := make(chan int)
	go letsListen(quitCh)

	checkForOthers()

	if len(others) == 0 {
		alives := util.FindOthers()
		for _, alive := range alives {
			checkUDP(alive)
		}
	} else {
		for _, other := range others {
			p :=  make([]byte, maxbytes)
			conn, err := net.Dial("udp", other)
			if err != nil {
				return
			}
			go letsTalk(conn, other, p)
		}
	}

	<-quitCh
}

func checkUDP(addr string) {
	//var status string
	otherAddr :=  addr + ":" + port
	p :=  make([]byte, maxbytes)

	conn, err := net.Dial("udp", otherAddr)

	if err != nil {
		//fmt.Printf("Some error %v", err)
		return
	}

	//opens connection
	go letsTalk(conn, otherAddr, p)

}

func letsListen(quitCh chan int) {
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

	quitCh <- 1
}
 
//client reads from connection
func letsTalk(conn net.Conn, addr string, p []byte) {
	fmt.Fprintf(conn, "haygurl")

	for {
		_, err := conn.Read(p)
		if err == nil {
			fmt.Printf("%s\n", p)
			writeToFile(addr)
		} else {
			fmt.Printf("Some error %v\n", err)
			break
		}
	}

	conn.Close()
}

func checkForOthers() {
	var f = &os.File{}
	if _, err := os.Stat(whereuat); !os.IsNotExist(err) {
		log.Println("file already exists")
		f, err = os.Open(whereuat)
		if err == os.ErrNotExist {
			fmt.Printf("%s not found\n", whereuat)
			return
		} else if err != nil {
			fmt.Printf("%s, %v\n", whereuat, err)
			return
		}
	} else {
		_, err := os.Create(whereuat)
		if err != nil {
			return
		}
	}

	defer f.Close()

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(f)
	if err != nil {
		return
	}

	ipaddrs := strings.Split(buf.String(), " ")
	for _, ip := range ipaddrs {
		if len(strings.TrimSpace(ip)) > 0 {
			others = append(others, ip)
			log.Println("appended " + ip + " to others")
		}
	}
}

func writeToFile(addr string) {
	log.Println("inside of write to file: " + addr)

	var f = &os.File{}
	var err = error(nil)
	//if file exists we append
	if _, err = os.Stat(whereuat); !os.IsNotExist(err) {
		log.Println("file already exists")
		f, err = os.OpenFile(whereuat, os.O_APPEND|os.O_RDWR, 0666)
		if err == os.ErrNotExist {
			fmt.Printf("%s not found\n", whereuat)
		} else if err != nil {
			fmt.Printf("%s, %v\n", whereuat, err)
		}
	}

	defer f.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(f)
	ipaddrs := strings.Split(buf.String(), " ")

	saved := false
	for _, ip := range ipaddrs {
		if ip == addr {
			saved = true
		}
	}

	if !saved {
		log.Println("gonna go ahead and append to file")
		w := bufio.NewWriter(f)
		fmt.Fprint(w, addr + " ")
		err = w.Flush() // Don't forget to flush!
		if err != nil {
			log.Fatal("fatal error: ", err)
		}
	}
}


func sendResponse(conn *net.UDPConn, addr *net.UDPAddr) {
	_,err := conn.WriteToUDP([]byte("From server: Hello I got your mesage "), addr)
	if err != nil {
		fmt.Printf("Couldn't send response %v", err)
	}
}
