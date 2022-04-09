package service_test

import (
	"github.com/ATenderholt/rainbow-storage/internal/domain"
	"github.com/ATenderholt/rainbow-storage/internal/service"
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

var factory = func() domain.EventFunction {
	return func(string, interface{}) {}
}

func TestNotificationServiceReadAndWrite(t *testing.T) {
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
}
