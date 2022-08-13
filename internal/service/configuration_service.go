package service

import (
	"encoding/xml"
	"errors"
	"github.com/ATenderholt/rainbow-storage/internal/domain"
	"io/fs"
	"os"
	"path/filepath"
)

const accelerationDir = "acceleration"

type ConfigurationService struct {
	cfg Config
}

func NewConfigurationService(config Config) *ConfigurationService {
	return &ConfigurationService{
		cfg: config,
	}
}

func (service ConfigurationService) SaveAccelerationConfiguration(bucket string, config []byte) (string, error) {
	basePath := filepath.Join(service.cfg.DataPath(), accelerationDir)
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

	path := filepath.Join(basePath, bucket+".xml")
	logger.Infof("Saving AccelerationConfiguration to %s", path)

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

	_, err = file.Write(config)
	if err != nil {
		err := SaveError{
			path:   path,
			bucket: bucket,
			base:   err,
		}
		logger.Error(err)
		return path, err
	}

	return path, nil
}

func (service ConfigurationService) LoadAccelerationConfiguration(bucket string) (domain.AccelerateConfiguration, error) {
	var config domain.AccelerateConfiguration
	path := filepath.Join(service.cfg.DataPath(), accelerationDir, bucket+".xml")
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return config, nil
		} else {
			err := LoadError{
				path: path,
				base: err,
			}
			logger.Error(err)
			return config, err
		}
	}

	err = xml.NewDecoder(file).Decode(&config)
	if err != nil {
		err := LoadError{
			path: path,
			base: err,
		}
		logger.Error(err)
		return config, err
	}

	return config, nil
}
