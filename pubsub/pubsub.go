package pubsub

//PubsubMsg message for pubsub event
type PubsubMsg struct {
	Data    []byte
	Subject string
}

//Subscription subscription
type Subscription interface {
	Unsub() error
	Subject() string
}

//Pubsub interface used in registry for pubsub operation
type Pubsub interface {
	Sub(topic string, f func(m *PubsubMsg)) (Subscription, error)
	Pub(topic string, data []byte) error
}
