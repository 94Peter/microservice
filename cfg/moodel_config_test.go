package cfg

import (
	"context"
	"testing"

	"github.com/94peter/microservice/di"
)

type TestModel struct {
	Name string
}

func (m TestModel) Close() error {
	return nil
}

func (m TestModel) Init(uuid string, di di.DI) error {
	return nil
}

func (m TestModel) Copy() ModelCfg {
	return m
}

func TestGetFromCtx(t *testing.T) {
	origin := TestModel{Name: "test"}
	ctx := setToCtx(context.Background(), origin)
	// Testing the case when the value is present in the context and the correct type is returned
	result, _ := GetFromCtx[TestModel](ctx)

	if result.Name != "test" {
		t.Errorf("Expected result.Name to be 'test', got %s", result.Name)
	}
}

func TestHandler(t *testing.T) {

	// Test case 1
	t.Run("Valid gin.HandlerFunc", func(t *testing.T) {
		// Your test logic here
	})

	// Test case 2
	t.Run("Error handling when servDi is nil", func(t *testing.T) {
		// Your test logic here
	})

	// Test case 3
	t.Run("Error handling when servDi is not empty", func(t *testing.T) {
		// Your test logic here
	})

	// Test case 4
	t.Run("Error handling when data initialization fails", func(t *testing.T) {
		// Your test logic here
	})
}
