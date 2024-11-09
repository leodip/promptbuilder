package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type Config struct {
	HeaderText string
	BaseDir    string
	Includes   []string
	Excludes   []string
}

func (c *Config) validate() error {
	if c.BaseDir == "" {
		return fmt.Errorf("basedir is required")
	}
	if !filepath.IsAbs(c.BaseDir) {
		return fmt.Errorf("basedir must be an absolute path")
	}
	if _, err := os.Stat(c.BaseDir); os.IsNotExist(err) {
		return fmt.Errorf("basedir does not exist: %s", c.BaseDir)
	}
	if len(c.Includes) == 0 {
		return fmt.Errorf("at least one include pattern is required")
	}
	return nil
}

func readInputFile(filepath string) (*Config, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	config := &Config{
		Includes: make([]string, 0),
		Excludes: make([]string, 0),
	}

	scanner := bufio.NewScanner(file)
	headerLines := []string{}
	isHeader := true

	for scanner.Scan() {
		line := scanner.Text()

		if line == "---" {
			isHeader = false
			config.HeaderText = strings.Join(headerLines, "\n")
			continue
		}

		if isHeader {
			headerLines = append(headerLines, line)
		} else {
			if line == "" {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.ToLower(strings.TrimSpace(parts[0]))
			value := strings.TrimSpace(parts[1])

			switch key {
			case "basedir":
				config.BaseDir = value
			case "include":
				config.Includes = append(config.Includes, value)
			case "exclude":
				config.Excludes = append(config.Excludes, value)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return config, nil
}

func isBinaryFile(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Read the first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}
	buf = buf[:n]

	// Check for null bytes (common in binary files)
	if bytes.IndexByte(buf, 0) != -1 {
		return true, nil
	}

	// Check if the content appears to be UTF-8 encoded text
	return !utf8IsValid(buf), nil
}

func utf8IsValid(buf []byte) bool {
	return utf8.Valid(buf)
}

func matchesAnyPattern(path string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}
	}
	return false
}

func findFiles(config *Config) ([]string, error) {
	var files []string

	err := filepath.Walk(config.BaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Convert to relative path for pattern matching
		relPath, err := filepath.Rel(config.BaseDir, path)
		if err != nil {
			return err
		}

		// Check if file should be included
		if matchesAnyPattern(path, config.Includes) {
			// Check if file should be excluded
			if matchesAnyPattern(path, config.Excludes) {
				return nil
			}

			// Check if file is binary
			isBinary, err := isBinaryFile(path)
			if err != nil {
				return fmt.Errorf("error checking if file is binary %s: %v", relPath, err)
			}
			if isBinary {
				fmt.Printf("Skipping binary file: %s\n", relPath)
				return nil
			}

			files = append(files, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	return files, nil
}

func generateOutput(config *Config, files []string, outputPath string) error {
	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer output.Close()

	// Write header
	if config.HeaderText != "" {
		fmt.Fprintln(output, config.HeaderText)
		fmt.Fprintln(output) // Empty line after header
	}

	// Process each file
	for _, relPath := range files {
		fullPath := filepath.Join(config.BaseDir, relPath)

		// Read file content
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("error reading file %s: %v", relPath, err)
		}

		// Write file name as markdown header
		fmt.Fprintf(output, "# %s\n", relPath)

		// Write file content as markdown code block
		fmt.Fprintln(output, "```")
		fmt.Fprintln(output, string(content))
		fmt.Fprintln(output, "```")
		fmt.Fprintln(output) // Empty line between files
	}

	return nil
}

func main() {
	// Define command line flags with default values
	inputFile := flag.String("input", "input.txt", "Input file path (default: input.txt)")
	outputFile := flag.String("output", "output.txt", "Output file path (default: output.txt)")

	// Parse the command line arguments
	flag.Parse()

	// Read and parse the input file
	config, err := readInputFile(*inputFile)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := config.validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Find matching files
	files, err := findFiles(config)
	if err != nil {
		fmt.Printf("Error finding files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("Warning: No files found matching the include/exclude patterns")
	}

	// Generate output file
	if err := generateOutput(config, files, *outputFile); err != nil {
		fmt.Printf("Error generating output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully processed %d files\n", len(files))
	fmt.Printf("Output written to: %s\n", *outputFile)
}
