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

func (service NotificationService) GetConfigurationPath(bucket string) string {
	return filepath.Join(service.cfg.DataPath(), notificationDir, bucket+".yaml")
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

func (service NotificationService) Start(bucket string, config domain.NotificationConfiguration) {
	logger.Infof("Starting NotificationConfigurations for bucket %s", bucket)

	ch, _ := config.Start(service.invoker)
	service.buckets[bucket] = ch
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

	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	bucket := filename[0 : len(filename)-len(ext)]

	service.Start(bucket, config)

	return nil
}

func (service NotificationService) LoadAll() error {
	rootPath := filepath.Join(service.cfg.DataPath(), notificationDir)
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		e := DirError{path: rootPath, base: err}
		logger.Error(e)
		return e
	}

	for _, entry := range entries {
		filename := entry.Name()
		ext := filepath.Ext(filename)
		if ext != ".yaml" {
			logger.Infof("Skipping unexpected file: %s", filename)
			continue
		}

		err = service.Load(filepath.Join(rootPath, filename))
		if err != nil {
			logger.Errorf("Unable load config file, not processing any more: %v", err)
			return err
		}
	}

	return nil
}

func (service NotificationService) ProcessEvent(event domain.NotificationEvent) error {
	ch, ok := service.buckets[event.Bucket]
	if !ok {
		err := fmt.Errorf("no NotificationConfiguration for for bucket %s has been registered", event.Bucket)
		logger.Error(err)
		return err
	}

	item := rxgo.Item{V: event}
	ch <- item

	return nil
}
