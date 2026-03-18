package models

type Subscriber struct {
	Endpoint string
	Channel  chan struct{}
}
