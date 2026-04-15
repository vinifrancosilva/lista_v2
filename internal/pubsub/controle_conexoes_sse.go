package pubsub

import (
	"fmt"
	"slices"

	"github.com/vinifrancosilva/lista_v2/internal/models"
)

func ControleConexoesSSE(pubsub *models.PubSubChanels) {
	// Mapa de controle
	mapaControle := make(map[string][]chan struct{})
	c := 0

	// Controla os eventos
	for {
		select {
		// quando abre uma conexão SSE é feito um subscribe
		case subs := <-pubsub.SubscriberChan:
			mapaControle[subs.Endpoint] = append(mapaControle[subs.Endpoint], subs.Channel)
			fmt.Println("Subscribed: ", subs)
			fmt.Println("Enviando evento para carregar:", subs.Endpoint)
			subs.Channel <- struct{}{}

		// quando fecha uma conexão SSE é feito um unsubscribe
		case unsubs := <-pubsub.UnsubscriberChan:
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
		case endpoint := <-pubsub.PublisherChan:
			fmt.Println("Recebido evento para publicação: ", endpoint)
			c = 0
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
