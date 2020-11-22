package registry

type fake struct {
}

func newFakePubsub() Pubsub {
	return &fake{}
}
func (p *fake) Pub(topic string, data []byte) error {
	return nil
}

func (p *fake) Sub(topic string, f func(m *PubsubMsg)) (Subscription, error) {
	return nil, nil
}
func (p *fake) Stop() {

}
