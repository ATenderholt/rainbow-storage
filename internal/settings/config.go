package settings

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultAccountNumber = "271828182845"
	DefaultRegion        = "us-west-2"

	DefaultBasePort = 9003
	DefaultDataPath = "data"

	DefaultNetworks = "rainbow"
)

type Config struct {
	AccountNumber string
	IsDebug       bool
	IsLocal       bool
	Region        string

	BasePort int
	dataPath string

	Networks []string
}

func (config *Config) DataPath() string {
	if config.dataPath[0] == '/' {
		return config.dataPath
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return filepath.Join(cwd, config.dataPath)
}

func DefaultConfig() *Config {
	logger.Debugf("Creating directory %s if necessary ...", DefaultDataPath)

	err := os.MkdirAll(DefaultDataPath, 0755)
	if err != nil {
		panic(err)
	}

	return &Config{
		AccountNumber: DefaultAccountNumber,
		IsDebug:       false,
		IsLocal:       true,
		Region:        DefaultRegion,
		BasePort:      DefaultBasePort,
		dataPath:      DefaultDataPath,
		Networks:      []string{DefaultNetworks},
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
	flags.BoolVar(&cfg.IsLocal, "local", true, "Application should use localhost when routing lambda")
	flags.StringVar(&cfg.Region, "region", DefaultRegion, "Region returned in ARNs")
	flags.IntVar(&cfg.BasePort, "port", DefaultBasePort, "Port used for HTTP and start of port range for individual lambdas")
	flags.StringVar(&cfg.dataPath, "data-path", DefaultDataPath, "Path to persist data and lambdas")
	flags.Var(&networks, "networks", "Comma-separated list of Networks for lambda containers")

	err := flags.Parse(args)
	if err != nil {
		return nil, buf.String(), err
	}

	cfg.Networks = networks.networks

	return &cfg, buf.String(), err
}
