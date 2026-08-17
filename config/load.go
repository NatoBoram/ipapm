package config

import (
	"errors"
	"fmt"
	"io"
	"os"

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
	Name       string
}

func Load(env Env) (Config, error) {
	dir, err := ConfigDir(env)
	if err != nil {
		return Config{}, fmt.Errorf("getting config dir: %w", err)
	}

	name := ConfigFile(dir)

	config, err := read(name)
	if errors.Is(err, os.ErrNotExist) {
		err := create(env.Name, name)
		if err != nil {
			return Config{}, fmt.Errorf("creating config: %w", err)
		}

		return Config{}, fmt.Errorf("created new config at %q", name)
	}
	if err != nil {
		return Config{}, fmt.Errorf("reading config: %w", err)
	}

	return config, nil
}

func read(filename string) (Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Config{}, fmt.Errorf("opening config file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return Config{}, fmt.Errorf("reading config file: %w", err)
	}

	var c Config
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return c, fmt.Errorf("unmarshalling config: %w", err)
	}

	return c, err
}

func create(name, filename string) error {
	data, err := yaml.Marshal(Config{
		Kubo:    Kubo{MFS: fmt.Sprintf("/%s", name)},
		Port:    9090,
		Sources: []Source{},
	})
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return err
}
