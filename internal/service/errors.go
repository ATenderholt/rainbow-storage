package service

import (
	"fmt"
	"github.com/ATenderholt/rainbow-storage/internal/domain"
)

type DirError struct {
	path string
	base error
}

func (e DirError) Error() string {
	return "Unable to list directory " + e.path + ": " + e.base.Error()
}

type LoadError struct {
	path string
	base error
}

func (e LoadError) Error() string {
	return fmt.Sprintf("Unable to load NotificationConfiguration from %s: %v", e.path, e.base)
}

type SaveError struct {
	path   string
	bucket string
	base   error
}

func (e SaveError) Error() string {
	return fmt.Sprintf("Unable to save NotificationConfiguration for bucket %s to %s: %v", e.bucket, e.path, e.base)
}

type DecodeError struct {
	path string
	base error
}

func (e DecodeError) Error() string {
	return fmt.Sprintf("Unable to decode file at %s from yaml: %v", e.path, e.base)
}

type EncodeError struct {
	config domain.NotificationConfiguration
	base   error
}

func (e EncodeError) Error() string {
	return fmt.Sprintf("Unable to encode %+v to yaml: %v", e.config, e.base)
}
