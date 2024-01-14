package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	ignore "github.com/sabhiram/go-gitignore"
)

func main() {
	dockerfilePath := flag.String("path", ".", "Path to the directory containing the Dockerfile")
	flag.Parse()

	fullDockerfilePath := filepath.Join(*dockerfilePath, "Dockerfile")
	dockerignorePath := filepath.Join(*dockerfilePath, ".dockerignore")

	ignoredPatterns, _ := ignore.CompileIgnoreFile(dockerignorePath)

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
	copyDockerfileContent(file, newFile)
	embedProjectFiles(*dockerfilePath, ignoredPatterns, newFile)

	fmt.Printf("Dockerfile.script created successfully in %s\n", *dockerfilePath)
}

func writeShebang(file *os.File) {
	fmt.Fprintln(file, `#!/usr/bin/env -S bash -c "docker run -p 8080:8080 -it --rm \$(docker build --progress plain -f \$0 . 2>&1 | tee /dev/stderr | grep -oP 'sha256:[0-9a-f]')"`)
}

func copyDockerfileContent(originalFile *os.File, newFile *os.File) {
	scanner := bufio.NewScanner(originalFile)
	for scanner.Scan() {
		fmt.Fprintln(newFile, scanner.Text())
	}
}

func embedProjectFiles(directory string, ignorePattern *ignore.GitIgnore, newFile *os.File) {
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}

		// Ignore files based on .dockerignore patterns
		isSkip := ignorePattern.MatchesPath(relativePath)
		if relativePath == "Dockerfile" {
			isSkip = true
		}
		if isSkip || info.IsDir() {
			return nil
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
		fmt.Fprintf(newFile, "RUN echo '%s' | base64 -d | gunzip > %s\n", encodedContent, relativePath)

		return nil
	})

	if err != nil {
		fmt.Printf("Error embedding project files: %s\n", err)
	}
}
