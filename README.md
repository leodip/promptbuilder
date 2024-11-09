# PromptBuilder

A command-line tool that generates a formatted output of multiple source files. It's particularly useful for preparing code for AI tools.

## Features

- Include specific files or entire directories
- Exclude files by name, extension, or folder
- Support for both relative and absolute paths
- Binary file detection and skipping
- Customizable header text in output
- Markdown-formatted output

## Binaries

Build it yourself or get the binaries provided [here](https://github.com/leodip/promptbuilder/releases).

## Usage

```bash
promptbuilder -input input.txt -output output.txt
```

Options:
- `-input`: Input configuration file (default: "input.txt")
- `-output`: Output file path (default: "output.txt")

## Configuration File Format

The configuration file consists of two parts:
1. Header text (optional) - appears at the start of the output file
2. Configuration directives - separated from header by `---`

### Directives

- `basedir`: Base directory for file operations
- `include`: Files or directories to include
- `excludeFolder`: Folders to exclude
- `excludeExtension`: File extensions to exclude (without the dot)
- `excludeFile`: Specific files to exclude

### Example Configuration Files

#### Basic Example
```
Please review these source files.
---
basedir=.
include=src
excludeFolder=node_modules
excludeExtension=json
excludeExtension=md
```

#### Full Stack Project Example
```
Review my React + Express application.
---
basedir=/home/user/projects/myapp
include=client/src
include=server/src
excludeFolder=node_modules
excludeFolder=dist
excludeExtension=json
excludeExtension=map
excludeExtension=svg
excludeFile=.env
excludeFile=.gitignore
```

#### Multiple Directories Example
```
Review all our utility functions.
---
basedir=/home/user/project
include=src/utils
include=src/helpers
include=src/lib
excludeFolder=node_modules
excludeFolder=tests
excludeExtension=js
excludeFile=index.js
```

## Output Format

The tool generates a Markdown-formatted output file with:
1. Optional header text
2. File contents in Markdown code blocks
3. Full file paths as headers

## Tips

1. Use relative paths with `basedir=.` for portable configurations
2. Extensions in `excludeExtension` should be specified without the dot (e.g., `excludeExtension=json` not `excludeExtension=.json`)
3. Exclude unnecessary files to keep output focused
4. You can exclude specific files using their full path (e.g., `excludeFile=src/config/dev.js`)
5. Binary files are automatically detected and skipped
6. Use multiple include directives to select specific directories or files

## Common Extension Exclusions

Here are some commonly used extension exclusions:
```
excludeExtension=json
excludeExtension=md
excludeExtension=gitignore
excludeExtension=svg
excludeExtension=map
excludeExtension=lock
excludeExtension=sum
excludeExtension=mod
excludeExtension=log
```

## License

MIT License - feel free to use this in your own projects.