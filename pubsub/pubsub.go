package pubsub

import "errors"

// PubsubMsg message for pubsub event
type PubsubMsg struct {
	Data    []byte
	Subject string
}

// Subscription subscription
type Subscription interface {
	Unsub() error
	Subject() string
}

// Pubsub interface used in registry for pubsub operation
type Pubsub interface {
	Sub(topic string, f func(m *PubsubMsg)) (Subscription, error)
	Pub(topic string, data []byte) error
	Stop()
}

//PubsubMock helper struct
type PubsubMock struct {
}

//ErrNotImplemented not implemented error
var ErrNotImplemented = errors.New("not implemented")

func (p *PubsubMock) Sub(topic string, f func(m *PubsubMsg)) (Subscription, error) {
	return nil, ErrNotImplemented
}
func (p *PubsubMock) Pub(topic string, data []byte) error {
	return ErrNotImplemented
}
func (p *PubsubMock) Stop() {

}
