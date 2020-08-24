package main

import (
	"crypto/tls"
	"fmt"
	"github.com/miekg/dns"
	"log"
	"net/http"
)

type HttpHandler struct {
}

func (handler *HttpHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	if request.Host == "conntest.nintendowifi.net" {
		// this is the connection test page.
		// I think this just has to be up. The actual page can be anything.
		fmt.Fprintln(writer, "Hello Nintendo DS. How are you?")
		log.Println("completed connection test for", request.RemoteAddr)
	} else if request.Host == "nas.nintendowifi.net" {
		// This is the Nintendo Authentication Server. We will need to in someway replicate this
		// process, even in just a basic level.
		// Now, Nintendo, trying to be smart and secure, uses SSL, requiring a valid SSL certificate
		// signed by Nintendo. However, as outlined here https://github.com/KaeruTeam/nds-constraint,
		// it literally just needs to be signed by Nintendo, and client certificates can sign other certs.
		// So if we obtain a client certificate from a Wii as they suggest, we can get a signed certificate
		// and then it's just a matter fo responding to the requests.
		// But I currently have no idea how to obtain the client certificate. They said "guide coming soon"

		fmt.Println("received request from Nintendo Authentication Server!")
	}
}

func main() {

	dns.HandleFunc("nintendowifi.net.", func(writer dns.ResponseWriter, msg *dns.Msg) {

		newMsg := new(dns.Msg)

		newMsg.SetReply(msg)

		newMsg.Compress = false

		if msg.Opcode == dns.OpcodeQuery {

			for _, q := range msg.Question {
				if q.Qtype == dns.TypeA {
					fmt.Printf("Requested Host: %s\n", q.Name)

					record, err := dns.NewRR(fmt.Sprintf("%s A 192.168.1.63", q.Name))

					if err == nil {
						fmt.Println("server does not equal null")
						newMsg.Answer = append(newMsg.Answer, record)
					}
				}
			}
		}

		log.Println(writer.WriteMsg(newMsg))
	})

	go func() {
		err := http.ListenAndServe("0.0.0.0:80", &HttpHandler{})

		if err != nil {
			log.Panicln(err)
		}
	}()

	go func() {

		srv := &http.Server{Addr: "0.0.0.0:443",
			TLSConfig: &tls.Config{MinVersion: tls.VersionSSL30}, // Nintendo DS uses a severely outdated SSL version not even supported by Golang
			Handler: &HttpHandler{},
		}



		err := srv.ListenAndServeTLS("server.crt", "server.key")

		if err != nil {
			log.Panicln(err)
		}

	}()

	log.Fatalln(dns.ListenAndServe("0.0.0.0:53", "udp", nil))
}
