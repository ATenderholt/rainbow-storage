package service

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

type ConfigurationService struct {
	cfg Config
}

func NewConfigurationService(config Config) *ConfigurationService {
	return &ConfigurationService{
		cfg: config,
	}
}

func (service ConfigurationService) SaveConfiguration(bucket string, configType string, config []byte) (string, error) {
	basePath := filepath.Join(service.cfg.DataPath(), configType)
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
	logger.Infof("Saving %s configuration for bucket %s to %s", configType, bucket, path)

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

func (service ConfigurationService) LoadConfiguration(bucket string, configType string) ([]byte, error) {
	path := filepath.Join(service.cfg.DataPath(), configType, bucket+".xml")
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return []byte{}, nil
		} else {
			err := LoadError{
				path: path,
				base: err,
			}
			logger.Error(err)
			return []byte{}, err
		}
	}

	config, err := ioutil.ReadAll(file)
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

func (service ConfigurationService) CleanupAllConfiguration(bucket string) {
	path := filepath.Join(service.cfg.DataPath())
	glob := fmt.Sprintf("%s/*/%s.xml", path, bucket)
	logger.Infof("cleaning up config files using glob %s", glob)

	files, err := filepath.Glob(glob)
	if err != nil {
		logger.Errorf("bad glob for cleaning up: %v", err)
		return
	}

	for _, file := range files {
		logger.Infof("removing config file %s", file)
		e := os.Remove(file)
		if e != nil {
			logger.Warnf("unable to delete %s: %v", file, e)
		}
	}
}
