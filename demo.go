package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	port = ":61000"
)

type P struct {
	X, Y, Z int
	Name    string
}

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

	// first that string
	addr := ip + port
	log.Println("client:", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return errors.Wrap(err, "client: dialing "+addr+" failed")
	}
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	log.Println("client: send the string request")
	n, err := rw.WriteString("STRING\n")
	if err != nil {
		return errors.Wrap(err, "client: could not send the STRING request ("+strconv.Itoa(n)+" bytes written)")
	}
	n, err = rw.WriteString("Additional data.\n")
	if err != nil {
		return errors.Wrap(err, "client: could not send additional STRING data ("+strconv.Itoa(n)+" bytes written)")
	}
	log.Println("client: flush the buffer")
	err = rw.Flush()
	if err != nil {
		return errors.Wrap(err, "client: flush failed")
	}
	log.Println("client: read the reply")
	response, err := rw.ReadString('\n')
	if err != nil {
		return errors.Wrap(err, "client: Failed to read the reply: '"+response+"'")
	}

	log.Println("client: STRING request - got a response:", response)

	// now for that gob
	log.Println("client: send the gob request")
	n, err = rw.WriteString("GOB\n")
	if err != nil {
		return errors.Wrap(err, "client: could not send the GOB request ("+strconv.Itoa(n)+" bytes written)")
	}

	enc := gob.NewEncoder(rw)
	err = enc.Encode(P{3, 4, 5, "Pythagoras"})
	if err != nil {
		log.Fatal("client: encode error:", err)
	}

	log.Println("client: flush the buffer")
	err = rw.Flush()
	if err != nil {
		return errors.Wrap(err, "client: flush failed")
	}

	return
}

func handleMessages(c net.Conn) {
	rw := bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))
	defer c.Close()

	for {
		log.Println("server: waiting for command")
		cmd, err := rw.ReadString('\n')
		switch {
		case err == io.EOF:
			log.Println("server: reached EOF - close this connection")
			return
		case err != nil:
			log.Println("server: error reading command - got: '"+cmd+"'", err)
			return
		}

		cmd = strings.Trim(cmd, "\n")
		log.Println("server: received command '" + cmd + "'")

		// TODO: it would be better to register a list of handlers here
		switch cmd {
		case "STRING":
			handleStrings(rw)
		case "GOB":
			handleGob(rw)
		default:
			log.Printf("server: unknown command '%s' - close this connection\n", cmd)
			return
		}

	}
}

func handleStrings(rw *bufio.ReadWriter) {
	log.Println("server: receive STRING message")

	s, err := rw.ReadString('\n')
	if err != nil {
		log.Println("server: cannot read from connection", err)
	}

	s = strings.Trim(s, "\n")
	log.Println("server:", s)

	_, err = rw.WriteString("Thank you for your STRING.\n")
	if err != nil {
		log.Println("server: cannot write to connection", err)
	}

	err = rw.Flush()
	if err != nil {
		log.Println("server: cannot flush connection", err)
	}

}

func handleGob(rw *bufio.ReadWriter) {
	log.Println("server: receive GOB message")

	var data P
	dec := gob.NewDecoder(rw)
	err := dec.Decode(&data)
	if err != nil {
		log.Println("server: error decoding GOB data", err)
		return
	}

	log.Printf("server: GOB data received: %#v\n", data)

}
