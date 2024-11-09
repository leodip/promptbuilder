# promptbuilder

A Go program that scans directories for text files based on patterns and combines their contents into a single Markdown file.

## Installation

```bash
go build -o fileprocessor
```

## Usage

```bash
./fileprocessor [-input input.txt] [-output output.txt]
```

## Configuration File Format

```
Optional header text here
Multiple lines allowed
---
basedir=/path/to/directory
include=*.go
include=*.js
exclude=*.test.js
```

### Parameters

- `basedir`: Starting directory (absolute path required)
- `include`: File patterns to include
- `exclude`: File patterns to exclude

Parameters are case-insensitive.

## Output

Generates a Markdown file with:
- Header text (if provided)
- Each file's content in a code block with its path as header

## Features

- Recursive directory scanning
- Pattern-based file inclusion/exclusion
- Binary file detection and skipping
- Case-insensitive configuration
- Markdown output format

