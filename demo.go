package main

import (
	"flag"
	"io"
	"log"
	"net"

	"github.com/pkg/errors"
)

const (
	port = ":61000"
)

func main() {
	connect := flag.String("connect", "", "IP address of server. If empty, run in server mode")
	flag.Parse()

	if *connect != "" {
		err := client(*connect)
		if err != nil {
			log.Println("client:", errors.WithStack(err))
		}
		log.Println("client: done")
		return
	}

	err := server()
	if err != nil {
		log.Println("server: ", errors.WithStack(err))
	}
	log.Println("server: done")
}

func init() {
	log.SetFlags(log.Lshortfile)
}

func server() (e error) {

	listener, e := net.Listen("tcp", port)
	if e != nil {
		return errors.Wrap(e, "server: unable to listen on "+listener.Addr().String()+"\n")
	}
	log.Println("server: listen on", listener.Addr().String())

	for {
		log.Println("server: accept a connection request")
		conn, err := listener.Accept()
		if err != nil {
			log.Println("server: failed accepting a connection request:", err)
			continue
		}
		log.Println("server: handle incoming messages")
		go handleMessages(conn)
	}

	return
}

func client(ip string) (e error) {
	return
}

func handleMessages(c net.Conn) {
	io.Copy(c, c)
	defer c.Close()
}
