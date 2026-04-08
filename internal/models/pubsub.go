package models

type Subscriber struct {
	Endpoint string
	Channel  chan struct{}
}

type PubSubChanels struct {
	SubscriberChan   chan Subscriber
	UnsubscriberChan chan Subscriber
	PublisherChan    chan string
}

func NewPubSubChannels() PubSubChanels {
	return PubSubChanels{
		SubscriberChan:   make(chan Subscriber),
		UnsubscriberChan: make(chan Subscriber),
		PublisherChan:    make(chan string),
	}
}
