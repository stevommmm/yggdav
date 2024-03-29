package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"

	gologme "github.com/gologme/log"
	"github.com/yggdrasil-network/yggdrasil-go/src/config"
	"github.com/yggdrasil-network/yggdrasil-go/src/core"
	"github.com/yggdrasil-network/yggstack/src/netstack"
	"golang.org/x/net/webdav"
)

var (
	dataDirectory string = "."
	localListener string = "127.0.0.1:8080"
)

type node struct {
	core   *core.Core
	config *config.NodeConfig
	log    *gologme.Logger
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.StringVar(&dataDirectory, "data", ".", "Directory contents to server via WebDAV")
	flag.StringVar(&localListener, "listen", localListener, "Local network binding address")
	flag.Parse()

	var err error
	var wg sync.WaitGroup

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

	listener, err := net.Listen("tcp", localListener)
	if err != nil {
		log.Fatal(err)
	}

	n.log.Printf("My yggdrasil address is dav://[%s]/\n", n.core.Address())
	n.log.Printf("My local address is dav://%s/\n", listener.Addr())
	s, err := netstack.CreateYggdrasilNetstack(n.core)
	if err != nil {
		panic(err)
	}
	ygglistener, err := s.ListenTCP(&net.TCPAddr{Port: 80})
	if err != nil {
		log.Fatal(err)
	}

	dav := &webdav.Handler{
		FileSystem: webdav.Dir(dataDirectory),
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
			if err != nil {
				log.Printf(" error:%s", err)
			}
			log.Println("")
		},
	}

	http.HandleFunc("/", dav.ServeHTTP)
	server := &http.Server{}

	wg.Add(2)
	go func() {
		log.Fatal(server.Serve(ygglistener))
		wg.Done()
	}()
	go func() {
		log.Fatal(server.Serve(listener))
		wg.Done()
	}()

	wg.Wait()
}
