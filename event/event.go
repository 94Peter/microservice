package event

import (
	"fmt"
	"sync"
)

type Handler[T any] interface {
	EventHandle(service, model string, msg T) error
}

type Event[T any] interface {
	UnRegister(service, model string, handler Handler[T])
	Register(service, model string, handler Handler[T]) error
	Emit(service, model string, msg T) error
}

type singleEvent[T any] struct {
	handlers map[string][]Handler[T]
	sync.Mutex
}

func NewSingleService[T any]() Event[T] {
	return &singleEvent[T]{
		handlers: make(map[string][]Handler[T]),
	}
}

func key(service, model string) string {
	return fmt.Sprintf("%s:%s", service, model)
}

func (e *singleEvent[T]) UnRegister(service, model string, handler Handler[T]) {
	e.Lock()
	defer e.Unlock()
	key := key(service, model)
	for i, h := range e.handlers[key] {
		if h == handler {
			e.handlers[key] = append(e.handlers[key][:i], e.handlers[key][i+1:]...)
		}
	}
}

func (e *singleEvent[T]) Register(service, model string, handler Handler[T]) error {
	e.Lock()
	defer e.Unlock()
	key := key(service, model)
	if _, ok := e.handlers[key]; !ok {
		e.handlers[key] = make([]Handler[T], 0)
	}
	e.handlers[key] = append(e.handlers[key], handler)
	return nil
}

func (e *singleEvent[T]) Emit(service, model string, msg T) error {
	e.Lock()
	defer e.Unlock()
	var wg sync.WaitGroup
	key := key(service, model)
	for _, handler := range e.handlers[key] {
		wg.Add(1)
		go func(s, model string, h Handler[T], m T) {
			defer wg.Done()
			h.EventHandle(s, model, m)
		}(service, model, handler, msg)
	}
	wg.Wait()
	return nil
}
