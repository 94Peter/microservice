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
	topicPrefix = "events/%s/%s/%s"
)

type MutiEventMsg interface {
	GetId() string
}

type MutiEventServ interface {
	Emit(service, model string, msg MutiEventMsg) error
	Register(service, model, msgId string, handler trans.Trans) error
	UnRegister(service, model, msgId string, handler trans.Trans)
	Run(context.Context)
}

type multiEventServ struct {
	topicHandler map[string][]trans.Trans
	mqttServ     mqtt.MqttServer
}

func NewMultiService(cfg *config.Config) (MutiEventServ, error) {
	event := &multiEventServ{
		topicHandler: make(map[string][]trans.Trans),
	}
	cfg.AddTopics("events/#")
	mqttServ, err := mqtt.NewMqttServ(cfg, map[string]trans.Trans{
		"events/#": event,
	})
	if err != nil {
		return nil, err
	}

	event.mqttServ = mqttServ
	return event, nil
}

func (m *multiEventServ) Emit(service, model string, msg MutiEventMsg) error {
	topic := fmt.Sprintf(topicPrefix, service, model, msg.GetId())
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return m.mqttServ.Publish(topic, 0, data)
}

func (m *multiEventServ) Register(service, model, id string, handler trans.Trans) error {
	key := fmt.Sprintf(topicPrefix, service, model, id)
	if _, ok := m.topicHandler[key]; !ok {
		m.topicHandler[key] = make([]trans.Trans, 0)
	}
	m.topicHandler[key] = append(m.topicHandler[key], handler)
	return nil
}

func (m *multiEventServ) UnRegister(service, model, id string, handler trans.Trans) {
	key := fmt.Sprintf(topicPrefix, service, model, id)
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
