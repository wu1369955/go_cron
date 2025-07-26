package alerter

type noopAlerter struct{}

func NewNoopAlerter() Alerter {
	return &noopAlerter{}
}

func (n *noopAlerter) Alert(_ map[string]any) {}
