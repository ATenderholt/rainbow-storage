package service

import (
	"fmt"
	"github.com/ATenderholt/rainbow-storage/internal/domain"
	"github.com/reactivex/rxgo/v2"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

const notificationDir = "notifications"

type Config interface {
	DataPath() string
}

type NotificationService struct {
	cfg     Config
	invoker domain.CloudFunctionInvoker
	buckets map[string]chan rxgo.Item
}

func NewNotificationService(config Config, invoker domain.CloudFunctionInvoker) *NotificationService {
	return &NotificationService{
		cfg:     config,
		invoker: invoker,
		buckets: make(map[string]chan rxgo.Item),
	}
}

func (service NotificationService) Save(bucket string, config domain.NotificationConfiguration) (string, error) {
	basePath := filepath.Join(service.cfg.DataPath(), notificationDir)
	err := os.MkdirAll(basePath, 0755)
	if err != nil {
		err := SaveError{
			path:   basePath,
			bucket: bucket,
			base:   err,
		}
		logger.Error(err)
		return basePath, err
	}

	path := filepath.Join(basePath, bucket+".yaml")
	logger.Infof("Saving NotificationConfiguration to %s", path)

	file, err := os.Create(path)
	if err != nil {
		err := SaveError{
			path:   path,
			bucket: bucket,
			base:   err,
		}
		logger.Error(err)
		return path, err
	}
	defer file.Close()

	err = yaml.NewEncoder(file).Encode(config)
	if err != nil {
		err := EncodeError{
			config: config,
			base:   err,
		}
		logger.Error(err)
		return path, err
	}

	return path, nil
}

func (service NotificationService) Load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		err := LoadError{
			path: path,
			base: err,
		}
		logger.Error(err)
		return err
	}
	defer file.Close()

	var config domain.NotificationConfiguration
	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		err := DecodeError{
			path: path,
			base: err,
		}
		logger.Error(err)
		return err
	}

	ch, _ := config.Start(service.invoker)

	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	bucket := filename[0 : len(filename)-len(ext)]

	service.buckets[bucket] = ch

	return nil
}

func (service NotificationService) ProcessEvent(bucket string, event domain.NotificationEvent) error {
	ch, ok := service.buckets[bucket]
	if !ok {
		err := fmt.Errorf("no NotificationConfiguration for for bucket %s has been registered", bucket)
		logger.Error(err)
		return err
	}

	item := rxgo.Item{V: event}
	ch <- item

	return nil
}
