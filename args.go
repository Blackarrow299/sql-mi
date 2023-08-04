package main

import (
	"errors"
	"flag"
	"fmt"
)

type Config struct {
	InputFilePath  string
	OutputFilePath string
}

func ParseArgs() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.OutputFilePath, "o", "schema.sql", "Output file")
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		return cfg, errors.New(fmt.Sprintf(
			"Error: Please provide a filename as an argument.\nUsage: ./script <filename>\n",
		))
	}

	cfg.InputFilePath = args[0]

	if cfg.OutputFilePath == "schema.sql" && len(args) > 2 {
		return cfg, errors.New(fmt.Sprintf("Usage: ./script -o <outputfile> <sourcefile>"))
	}

	return cfg, nil
}
