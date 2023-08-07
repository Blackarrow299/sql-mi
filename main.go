package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	cfg, err := ParseArgs()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	// Check if the file exists
	if _, err := os.Stat(cfg.InputFilePath); os.IsNotExist(err) {
		fmt.Printf("File '%s' does not exist.\n", cfg.InputFilePath)
		os.Exit(1)
	}

	content, err := ioutil.ReadFile(cfg.InputFilePath)
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		os.Exit(1)
	}

	tokenizer := NewTokenizer(string(content))

	ast, err := Parse(tokenizer)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	sql, err := GenerateSQL(ast)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	file, err := os.Create(cfg.OutputFilePath)
	if err != nil {
		fmt.Printf("Error creating file '%s': %v\n", cfg.OutputFilePath, err)
		os.Exit(1)
	}
	defer file.Close()

	_, err = file.WriteString(sql)
	if err != nil {
		fmt.Printf("Error writing to file '%s': %v\n", cfg.OutputFilePath, err)
		os.Exit(1)
	}
}
