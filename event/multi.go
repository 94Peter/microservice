package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/94peter/mqtt"
	"github.com/94peter/mqtt/config"
	"github.com/94peter/mqtt/trans"
)

const (
	topicPrefix = "events/%s/%s"
)

type MutiEventServ interface {
	Emit(service, model string, msg any) error
	Register(service, model string, handler trans.Trans) error
	UnRegister(service, model string, handler trans.Trans)
	Run(context.Context)
}

type multiEventServ struct {
	topicHandler map[string][]trans.Trans
	mqttServ     mqtt.MqttServer
}

func NewMultiService(cfg *config.Config) MutiEventServ {
	event := &multiEventServ{
		topicHandler: make(map[string][]trans.Trans),
	}
	mqttServ := mqtt.NewMqttServ(cfg, map[string]trans.Trans{
		"event/#": event,
	})

	event.mqttServ = mqttServ
	return event
}

func (m *multiEventServ) Emit(service, model string, msg any) error {
	topic := fmt.Sprintf(topicPrefix, service, model)
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return m.mqttServ.Publish(topic, 0, data)
}

func (m *multiEventServ) Register(service, model string, handler trans.Trans) error {
	key := fmt.Sprintf(topicPrefix, service, model)
	if _, ok := m.topicHandler[key]; !ok {
		m.topicHandler[key] = make([]trans.Trans, 0)
	}
	m.topicHandler[key] = append(m.topicHandler[key], handler)
	return nil
}

func (m *multiEventServ) UnRegister(service, model string, handler trans.Trans) {
	key := fmt.Sprintf(topicPrefix, service, model)
	for i, h := range m.topicHandler[key] {
		if h == handler {
			m.topicHandler[key] = append(m.topicHandler[key][:i], m.topicHandler[key][i+1:]...)
		}
	}
}

func (m *multiEventServ) Send(topic string, msg []byte) error {
	for _, handler := range m.topicHandler[topic] {
		if err := handler.Send(topic, msg); err != nil {
			return err
		}
	}
	return nil
}

func (m *multiEventServ) Close() {
	m.mqttServ.Close()
}

func (m *multiEventServ) Run(ctx context.Context) {
	m.mqttServ.Run(ctx)
}
