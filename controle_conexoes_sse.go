package main

import (
	"fmt"
	"slices"
)

type Subscriber struct {
	Endpoint string
	Channel  chan struct{}
}

var subscriberChan chan Subscriber
var unsubscriberChan chan Subscriber
var publisherChan chan string

func controleConexoesSSE() {
	// Subscribers Channels
	subscriberChan = make(chan Subscriber)
	unsubscriberChan = make(chan Subscriber)

	// Publisher Channel
	publisherChan = make(chan string)

	// Mapa de controle
	mapaControle := make(map[string][]chan struct{})

	// Controla os eventos
	for {
		select {
		// quando abre uma conexão SSE é feito um subscribe
		case subs := <-subscriberChan:
			mapaControle[subs.Endpoint] = append(mapaControle[subs.Endpoint], subs.Channel)
			fmt.Println("Subscribed: ", subs)

		// quando fecha uma conexão SSE é feito um unsubscribe
		case unsubs := <-unsubscriberChan:
			if _, ok := mapaControle[unsubs.Endpoint]; ok {
				for i, ch := range mapaControle[unsubs.Endpoint] {
					if ch == unsubs.Channel {
						// mapaControle[unsubs.Endpoint] = append(mapaControle[unsubs.Endpoint][:i], mapaControle[unsubs.Endpoint][i+1:]...)
						mapaControle[unsubs.Endpoint] = slices.Delete(mapaControle[unsubs.Endpoint], i, i+1)
						break
					}
				}
			}
			fmt.Println("Unsubscribed: ", unsubs)

		// quando há uma mudança no banco de dados é feito um publish e o endpoint fica encarregado de reenviar a informação que lhe compete
		case endpoint := <-publisherChan:
			fmt.Println("Recebido evento para publicação: ", endpoint)
			c := 0
			if _, ok := mapaControle[endpoint]; ok {
				for _, ch := range mapaControle[endpoint] {
					ch <- struct{}{}
					c++
				}
			}
			if c > 0 {
				fmt.Println("Despachados para ", c, " subscribers")
			}
		}
	}
}
