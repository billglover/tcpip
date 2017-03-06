package main

import (
	"flag"
	"log"

	"github.com/pkg/errors"
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
	return
}

func client(ip string) (e error) {
	return
}
