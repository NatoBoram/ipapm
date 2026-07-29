package config

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Sources []Source `yaml:"Sources"`
	Port    uint     `yaml:"Port"`
}

type Source struct {
	Types         []string `yaml:"Types"`
	URIs          []string `yaml:"URIs"`
	Suites        []string `yaml:"Suites"`
	Components    []string `yaml:"Components"`
	Architectures []string `yaml:"Architectures"`
	SignedBy      string   `yaml:"SignedBy"`
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
	log.Printf("Loading config at %s", name)

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
	data, err := yaml.Marshal(Config{Port: 9090})
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
