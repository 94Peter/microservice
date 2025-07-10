package event_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/94peter/microservice/event"
	"github.com/stretchr/testify/assert"
)

type TestMsg struct {
	Name string
}

func TestSingleEvent(t *testing.T) {
	singleEvent := event.NewSingleService[*TestMsg]()
	handler := singleEvent.RegisterFunc("service1", "model1", func(msg *TestMsg) {
		assert.Equal(t, msg.Name, "test")
	})

	singleEvent.Emit("service1", "model1", &TestMsg{Name: "test"})
	singleEvent.UnRegister("service1", "model1", handler)
}

func TestGoroutineSingleEvent(t *testing.T) {
	var wg sync.WaitGroup
	singleEvent := event.NewSingleService[*TestMsg]()
	concurrent := 5
	var cancelSlice = make([]context.CancelFunc, concurrent)
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		ctx, cancel := context.WithCancel(context.Background())
		cancelSlice[i] = cancel
		go func(ctx context.Context) {
			fmt.Println("test goroutine")
			handler := singleEvent.RegisterFunc("service1", "model1", func(msg *TestMsg) {
				assert.Equal(t, msg.Name, "test")
				fmt.Println("equial test", msg.Name)
				wg.Done()
			})
			<-ctx.Done()
			singleEvent.UnRegister("service1", "model1", handler)
		}(ctx)
	}

	time.Sleep(time.Second)
	singleEvent.Emit("service1", "model1", &TestMsg{Name: "test"})
	fmt.Println("test emit")
	wg.Wait()
	for i := 0; i < concurrent; i++ {
		if cancelSlice[i] != nil {
			cancelSlice[i]()
		}
	}
}
