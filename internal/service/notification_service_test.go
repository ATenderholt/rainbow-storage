package service_test

import (
	"github.com/ATenderholt/rainbow-storage/internal/domain"
	"github.com/ATenderholt/rainbow-storage/internal/service"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

type Config struct{}

func (c Config) DataPath() string {
	dir, err := ioutil.TempDir("", "rainbow-test-*")
	if err != nil {
		panic(err)
	}

	return dir
}

func TestNotificationServiceReadAndWrite(t *testing.T) {
	// visibility into events
	ch := make(chan domain.NotificationEvent)
	factory := func() domain.EventFunction {
		return func(_ string, i interface{}) {
			ch <- i.(domain.NotificationEvent)
		}
	}

	cfg := Config{}
	s := service.NewNotificationService(cfg, factory)

	data := domain.NotificationConfiguration{
		CloudFunctionConfigurations: []domain.CloudFunctionConfiguration{
			{
				Events:        []string{domain.ObjectCreatedEvent},
				Filter:        domain.Filter{},
				ID:            "some-id",
				CloudFunction: domain.CloudFunction("something"),
			},
		},
	}

	path, err := s.Save("test", data)
	if err != nil {
		t.Fatalf("Problem saving configuration: %v", err)
	}

	t.Cleanup(func() {
		err := os.RemoveAll(path)
		if err != nil {
			t.Logf("Unable to delete %s", path)
		}
	})

	err = s.Load(path)
	if err != nil {
		t.Fatalf("Problem loading configuration: %v", err)
	}

	// send two events, first should be filtered out
	testData := []domain.NotificationEvent{
		{Event: domain.ObjectRemovedEvent, Key: "test.txt"},
		{Event: domain.ObjectCreatedEvent, Key: "test.bin"},
	}

	for _, event := range testData {
		err = s.ProcessEvent("test", event)
		if err != nil {
			t.Fatalf("Error when processing event: %s", err)
		}
	}

	value := <-ch

	assert.Equal(t, testData[1], value)
}
