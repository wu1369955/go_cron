package alerter

type Alerter interface {
	// Alert 发送告警
	Alert(data map[string]any)
}
