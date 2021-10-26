package proxy

import (
	"log"

	"github.com/platform-edn/courier/message"
)

type Proxyer interface {
	Subscribe(subject string) chan message.Message
	MessageChannel() chan message.Message
	Subscriptions(string) ([]chan (message.Message), error)
}

func ForwardMessages(proxy Proxyer) {
	for m := range proxy.MessageChannel() {
		subs, err := proxy.Subscriptions(m.Subject)
		if err != nil {
			log.Print(err.Error())
		}

		for _, subscriber := range subs {
			subscriber <- m
		}
	}
}
