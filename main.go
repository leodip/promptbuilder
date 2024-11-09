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

var version = "0.2"

type Config struct {
	HeaderText        string
	BaseDir           string
	Includes          []string
	ExcludeFolders    []string
	ExcludeExtensions []string
	ExcludeFiles      []string // New: list of specific files to exclude
}

func (c *Config) validate() error {
	if c.BaseDir == "" {
		return fmt.Errorf("basedir is required")
	}

	// Convert to absolute path if relative
	if !filepath.IsAbs(c.BaseDir) {
		absPath, err := filepath.Abs(c.BaseDir)
		if err != nil {
			return fmt.Errorf("failed to convert basedir to absolute path: %v", err)
		}
		c.BaseDir = absPath
	}

	if _, err := os.Stat(c.BaseDir); os.IsNotExist(err) {
		return fmt.Errorf("basedir does not exist: %s", c.BaseDir)
	}

	if len(c.Includes) == 0 {
		return fmt.Errorf("at least one include path is required")
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
		Includes:          make([]string, 0),
		ExcludeFolders:    make([]string, 0),
		ExcludeExtensions: make([]string, 0),
		ExcludeFiles:      make([]string, 0), // Initialize ExcludeFiles
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
			case "excludefolder":
				config.ExcludeFolders = append(config.ExcludeFolders, value)
			case "excludeextension":
				ext := value
				if !strings.HasPrefix(ext, "*.") {
					ext = "*." + ext
				}
				config.ExcludeExtensions = append(config.ExcludeExtensions, ext)
			case "excludefile":
				config.ExcludeFiles = append(config.ExcludeFiles, value)
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

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}
	buf = buf[:n]

	if bytes.IndexByte(buf, 0) != -1 {
		return true, nil
	}

	return !utf8.Valid(buf), nil
}

func isExcludedFolder(path string, excludeFolders []string) bool {
	for _, folder := range excludeFolders {
		if filepath.Base(path) == folder {
			return true
		}
	}
	return false
}

func isExcludedExtension(path string, excludeExtensions []string) bool {
	ext := filepath.Ext(path)
	if ext == "" {
		return false
	}

	for _, pattern := range excludeExtensions {
		if pattern == "*"+ext {
			return true
		}
	}
	return false
}

func isExcludedFile(path string, baseDir string, excludeFiles []string) bool {
	// Get the relative path from baseDir
	relPath, err := filepath.Rel(baseDir, path)
	if err != nil {
		return false
	}

	// Convert to forward slashes for consistency
	relPath = filepath.ToSlash(relPath)

	for _, excludeFile := range excludeFiles {
		// Convert exclude pattern to forward slashes
		excludePattern := filepath.ToSlash(excludeFile)

		// Try both exact match and filename-only match
		if relPath == excludePattern || filepath.Base(path) == excludePattern {
			return true
		}
	}
	return false
}

func collectFiles(path string, config *Config) ([]string, error) {
	var files []string

	err := filepath.Walk(path, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded folders
		if info.IsDir() && isExcludedFolder(currentPath, config.ExcludeFolders) {
			return filepath.SkipDir
		}

		// Skip directories, excluded extensions, and excluded files
		if !info.IsDir() &&
			!isExcludedExtension(currentPath, config.ExcludeExtensions) &&
			!isExcludedFile(currentPath, config.BaseDir, config.ExcludeFiles) {
			relPath, err := filepath.Rel(path, currentPath)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func findFiles(config *Config) ([]string, error) {
	var allFiles []string

	for _, includePath := range config.Includes {
		fullPath := filepath.Join(config.BaseDir, includePath)
		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			fmt.Printf("Warning: Cannot access path %s: %v\n", includePath, err)
			continue
		}

		if fileInfo.IsDir() {
			// If it's a directory, collect all files recursively
			files, err := collectFiles(fullPath, config)
			if err != nil {
				return nil, fmt.Errorf("error collecting files from %s: %v", includePath, err)
			}

			// Add directory prefix to found files
			for _, f := range files {
				allFiles = append(allFiles, filepath.Join(includePath, f))
			}
		} else {
			// If it's a file and not excluded
			if !isExcludedExtension(fullPath, config.ExcludeExtensions) &&
				!isExcludedFile(fullPath, config.BaseDir, config.ExcludeFiles) {
				allFiles = append(allFiles, includePath)
			}
		}
	}

	return allFiles, nil
}

func generateOutput(config *Config, files []string, outputPath string) error {
	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer output.Close()

	if config.HeaderText != "" {
		fmt.Fprintln(output, config.HeaderText)
		fmt.Fprintln(output)
	}

	for _, relPath := range files {
		fullPath := filepath.Join(config.BaseDir, relPath)

		// Check if file is binary
		isBinary, err := isBinaryFile(fullPath)
		if err != nil {
			fmt.Printf("Warning: Error checking if file is binary %s: %v\n", relPath, err)
			continue
		}
		if isBinary {
			fmt.Printf("Skipping binary file: %s\n", relPath)
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("error reading file %s: %v", relPath, err)
		}

		fmt.Fprintf(output, "# %s\n", fullPath)
		fmt.Fprintln(output, "```")
		fmt.Fprintln(output, string(content))
		fmt.Fprintln(output, "```")
		fmt.Fprintln(output)
	}

	return nil
}

func main() {
	fmt.Println("promptbuilder v" + version)

	inputFile := flag.String("input", "input.txt", "Input file path (default: input.txt)")
	outputFile := flag.String("output", "output.txt", "Output file path (default: output.txt)")
	flag.Parse()

	config, err := readInputFile(*inputFile)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	if err := config.validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	files, err := findFiles(config)
	if err != nil {
		fmt.Printf("Error finding files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("Warning: No files found matching the include paths")
	} else {
		fmt.Printf("Found %d matching files\n", len(files))
	}

	if err := generateOutput(config, files, *outputFile); err != nil {
		fmt.Printf("Error generating output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully processed %d files\n", len(files))
	fmt.Printf("Output written to: %s\n", *outputFile)
}
