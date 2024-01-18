package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

var ignorePattern *ignore.GitIgnore
var hasIgnoreFile bool = false

func main() {
	dockerfilePath := flag.String("path", ".", "Path to the directory containing the Dockerfile")
	flag.Parse()

	fullDockerfilePath := filepath.Join(*dockerfilePath, "Dockerfile")
	gitignorePath := filepath.Join(*dockerfilePath, ".gitignore")
	dockerignorePath := filepath.Join(*dockerfilePath, ".dockerignore")

	if pathfile, err := os.Stat(gitignorePath); err != nil || !pathfile.IsDir() {
		ignorePattern, _ = ignore.CompileIgnoreFile(gitignorePath)
		hasIgnoreFile = true
	}

	if pathfile, err := os.Stat(dockerignorePath); err != nil || !pathfile.IsDir() {
		ignorePattern, _ = ignore.CompileIgnoreFile(dockerignorePath)
		hasIgnoreFile = true
	}

	file, err := os.Open(fullDockerfilePath)
	if err != nil {
		fmt.Printf("Error opening Dockerfile in %s: %s\n", *dockerfilePath, err)
		return
	}
	defer file.Close()

	newFilePath := filepath.Join(*dockerfilePath, "Dockerfile.script")
	newFile, err := os.Create(newFilePath)
	if err != nil {
		fmt.Printf("Error creating Dockerfile.script in %s: %s\n", *dockerfilePath, err)
		return
	}
	defer newFile.Close()

	writeShebang(newFile)
	copyDockerfileContent(file, newFile, *dockerfilePath)

	fmt.Printf("Dockerfile.script created successfully in %s\n", *dockerfilePath)
}

func writeShebang(file *os.File) {
	fmt.Fprintln(file, `#!/usr/bin/env -S bash -c "docker run --network host -it --rm \$(docker build --progress plain -f \$0 . 2>&1 | tee /dev/stderr | grep 'writing image sha256:' | grep -oP 'sha256:[0-9a-f]+' | cut -d' ' -f1)"`)
}

func copyDockerfileContent(originalFile *os.File, newFile *os.File, directory string) error {
	scanner := bufio.NewScanner(originalFile)
	fromCopied := false
	copyBefore := ""

	copyCmd := []string{"ADD", "COPY"}
	content, _ := io.ReadAll(originalFile)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.ToUpper(strings.TrimSpace(line))
		for _, cmd := range copyCmd {
			if strings.HasPrefix(line, cmd) {
				copyBefore = cmd
				goto APPEND
			}
		}
	}

APPEND:
	for i, line := range lines {

		// Write the line to the new file
		if _, err := fmt.Fprintln(newFile, line); err != nil {
			return err
		}

		// Check if the line is a FROM directive
		if !fromCopied && copyBefore != "" {
			if len(lines) > i+1 && strings.Contains(lines[i+1], copyBefore) {
				fromCopied = true

				// After copying FROM line, embed project files
				embedProjectFiles(directory, newFile)
			}
		}
	}

	return scanner.Err()
}

func embedProjectFiles(directory string, newFile *os.File) {
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}

		if hasIgnoreFile {
			// Ignore files based on .dockerignore patterns
			isSkip := ignorePattern.MatchesPath(relativePath)
			if relativePath == "Dockerfile" {
				isSkip = true
			}
			if isSkip || info.IsDir() {
				return nil
			}
		}

		fileContent, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Compress the file content using gzip
		var gzipBuffer bytes.Buffer
		gzipWriter := gzip.NewWriter(&gzipBuffer)
		_, err = gzipWriter.Write(fileContent)
		if err != nil {
			gzipWriter.Close()
			return err
		}
		gzipWriter.Close()

		// Encode the compressed content to base64
		encodedContent := base64.StdEncoding.EncodeToString(gzipBuffer.Bytes())
		fmt.Fprintf(newFile, "RUN mkdir -p %s\n", getParentPath(relativePath))
		fmt.Fprintf(newFile, "RUN echo '%s' | base64 -d | gunzip > %s\n", encodedContent, relativePath)

		return nil
	})

	if err != nil {
		fmt.Printf("Error embedding project files: %s\n", err)
	}
}

func getParentPath(fullPath string) string {
	// Clean the path to fix any irregularities
	fullPath = filepath.Clean(fullPath)

	// Split the path into directory and base
	dir, _ := filepath.Split(fullPath)

	// Clean the directory to remove the trailing slash
	return filepath.Clean(dir)
}
