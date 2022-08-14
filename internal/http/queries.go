package http

import (
	"encoding/xml"
	"github.com/ATenderholt/rainbow-storage/internal/domain"
)

type SetDefaultFunc func([]byte) []byte

var supportedQueries map[string]SetDefaultFunc

func init() {
	supportedQueries = make(map[string]SetDefaultFunc)
	supportedQueries["accelerate"] = defaultAccelerationConfiguration
	supportedQueries["acl"] = bytesPassThrough
}

func bytesPassThrough(config []byte) []byte {
	return config
}

func defaultAccelerationConfiguration(config []byte) []byte {
	var accel domain.AccelerateConfiguration
	if len(config) == 0 {
		accel.Status = "Disabled"
	} else {
		err := xml.Unmarshal(config, &accel)
		if err != nil {
			logger.Panicf("unable to unmarshal %s: %v", string(config), err)
		}
	}

	if accel.Status == "" {
		accel.Status = "Disabled"
	}

	result, err := xml.Marshal(accel)
	if err != nil {
		logger.Panicf("unable to marshal %+v: %v", accel, err)
	}

	return result
}
