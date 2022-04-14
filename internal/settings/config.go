package settings

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultAccountNumber  = "271828182845"
	DefaultRegion         = "us-west-2"
	DefaultLambdaEndpoint = "http://localhost:9050"

	DefaultBasePort = 9000
	DefaultDataPath = "data"
	DefaultImage    = "bitnami/minio:2022.2.16"

	DefaultNetworks = "rainbow"
)

type Config struct {
	AccountNumber  string
	IsDebug        bool
	IsLocal        bool
	Region         string
	LambdaEndpoint string

	BasePort int
	dataPath string
	Image    string

	Networks []string
}

func (config *Config) DataPath() string {
	if config.dataPath[0] == '/' {
		return filepath.Join(config.dataPath, "s3")
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return filepath.Join(cwd, config.dataPath, "s3")
}

func (config *Config) MinioUrl() string {
	if config.IsLocal {
		return fmt.Sprintf("http://localhost:%d", config.BasePort+1)
	} else {
		return "http://s3:9000"
	}
}

func DefaultConfig() *Config {
	logger.Debugf("Creating directory %s if necessary ...", DefaultDataPath)

	err := os.MkdirAll(DefaultDataPath, 0755)
	if err != nil {
		panic(err)
	}

	return &Config{
		AccountNumber:  DefaultAccountNumber,
		IsDebug:        false,
		IsLocal:        true,
		Region:         DefaultRegion,
		LambdaEndpoint: DefaultLambdaEndpoint,
		BasePort:       DefaultBasePort,
		dataPath:       DefaultDataPath,
		Image:          DefaultImage,
		Networks:       []string{DefaultNetworks},
	}
}

type NetworkValue struct {
	networks []string
}

func (v *NetworkValue) Set(s string) error {
	v.networks = strings.Split(s, ",")
	return nil
}

func (v *NetworkValue) String() string {
	if len(v.networks) > 0 {
		return strings.Join(v.networks, ",")
	}

	return ""
}

func FromFlags(name string, args []string) (*Config, string, error) {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)

	var buf bytes.Buffer
	flags.SetOutput(&buf)

	var cfg Config
	networks := NetworkValue{[]string{DefaultNetworks}}
	flags.StringVar(&cfg.AccountNumber, "account-number", DefaultAccountNumber, "Account number returned in ARNs")
	flags.BoolVar(&cfg.IsDebug, "debug", false, "Enable debug logging")
	flags.BoolVar(&cfg.IsLocal, "local", true, "Application should use localhost when routing to s3 service")
	flags.StringVar(&cfg.Region, "region", DefaultRegion, "Region returned in ARNs")
	flags.StringVar(&cfg.LambdaEndpoint, "lambda-endpoint", DefaultLambdaEndpoint, "Endpoint URL for lambda service")
	flags.IntVar(&cfg.BasePort, "port", DefaultBasePort, "Port used for HTTP and start of port range for s3 service")
	flags.StringVar(&cfg.Image, "image", DefaultImage, "Image to use for backing storage")
	flags.StringVar(&cfg.dataPath, "data-path", DefaultDataPath, "Path to persist data and s3 configuration")
	flags.Var(&networks, "networks", "Comma-separated list of Networks for containers")

	err := flags.Parse(args)
	if err != nil {
		return nil, buf.String(), err
	}

	cfg.Networks = networks.networks

	return &cfg, buf.String(), err
}
