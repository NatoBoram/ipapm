package config

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/NatoBoram/ipapm/env"
	"go.yaml.in/yaml/v4"
)

type Config struct {
	Kubo    Kubo     `yaml:"Kubo"`
	Port    uint     `yaml:"Port"`
	Sources []Source `yaml:"Sources"`
}

type Kubo struct {
	MFS string `yaml:"MFS"`
}

type Source struct {
	URIs     []string `yaml:"URIs"`
	Suites   []string `yaml:"Suites"`
	SignedBy string   `yaml:"Signed-By"`
}

type Env struct {
	CONFIG_DIR string
}

func Load(env Env) (Config, error) {
	dir, err := ConfigDir(env)
	if err != nil {
		return Config{}, fmt.Errorf("couldn't get config dir: %w", err)
	}

	name := ConfigFile(dir)

	config, err := read(name)
	if errors.Is(err, os.ErrNotExist) {
		err := create(name)
		if err != nil {
			return Config{}, fmt.Errorf("couldn't create config: %w", err)
		}

		return Config{}, fmt.Errorf("created new config at %s: %w", name, err)
	}
	if err != nil {
		return Config{}, fmt.Errorf("couldn't read config: %w", err)
	}

	return config, nil
}

func read(filename string) (Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var c Config
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return c, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return c, err
}

func create(filename string) error {
	name, err := env.Name()
	if err != nil {
		return fmt.Errorf("failed to get app name: %w", err)
	}

	data, err := yaml.Marshal(Config{
		Kubo:    Kubo{MFS: fmt.Sprintf("/%s", name)},
		Port:    9090,
		Sources: []Source{},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return err
}
