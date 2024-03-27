package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"

	// "github.com/stevommm/p2pdav/transport"
	gologme "github.com/gologme/log"
	"github.com/yggdrasil-network/yggdrasil-go/src/config"
	"github.com/yggdrasil-network/yggdrasil-go/src/core"
)

import ()

type node struct {
	core   *core.Core
	config *config.NodeConfig
	log    *gologme.Logger
}

func main() {
	flag.Parse()

	var err error

	n := node{}
	n.log = gologme.New(os.Stdout, "", gologme.LstdFlags)
	n.config = config.GenerateConfig()

	n.core, err = core.New(n.config.Certificate, n.log)
	if err != nil {
		log.Fatal(err)
	}
	if peer, err := url.Parse("tcp://sin.yuetau.net:6642"); err == nil {
		n.core.AddPeer(peer, "")
	}

	n.log.Println("My public key is", n.core.PublicKey())
	n.log.Println("My address is", n.core.Address())

	listenaddr, _ := url.Parse("tcp://0.0.0.0:80")
	listener, err := n.core.Listen(listenaddr, "")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hi!"))
	})
	server := &http.Server{}
	server.Serve(listener)
}
