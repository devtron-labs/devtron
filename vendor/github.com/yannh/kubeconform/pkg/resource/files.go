package resource

import (
	"bufio"
	"context"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func isYAMLFile(info os.FileInfo) bool {
	return !info.IsDir() && (strings.HasSuffix(strings.ToLower(info.Name()), ".yaml") || strings.HasSuffix(strings.ToLower(info.Name()), ".yml"))
}

func isJSONFile(info os.FileInfo) bool {
	return !info.IsDir() && (strings.HasSuffix(strings.ToLower(info.Name()), ".json"))
}

type DiscoveryError struct {
	Path string
	Err  error
}

func (de DiscoveryError) Error() string {
	return de.Err.Error()
}

func isIgnored(path string, ignoreFilePatterns []string) (bool, error) {
	for _, p := range ignoreFilePatterns {
		m, err := regexp.MatchString(p, path)
		if err != nil {
			return false, err
		}
		if m {
			return true, nil
		}
	}
	return false, nil
}

func findFilesInFolders(ctx context.Context, paths []string, ignoreFilePatterns []string) (chan string, chan error) {
	files := make(chan string)
	errors := make(chan error)

	go func() {
		for _, path := range paths {
			// we handle errors in the walk function directly
			// so it should be safe to discard the outer error
			err := filepath.Walk(path, func(p string, i os.FileInfo, err error) error {
				select {
				case <-ctx.Done():
					return io.EOF
				default:
				}

				if err != nil {
					return err
				}

				if !isYAMLFile(i) && !isJSONFile(i) {
					return nil
				}

				ignored, err := isIgnored(p, ignoreFilePatterns)
				if err != nil {
					return err
				}
				if ignored {
					return nil
				}

				files <- p

				return nil
			})

			if err != nil && err != io.EOF {
				errors <- DiscoveryError{path, err}
			}
		}

		close(files)
	}()

	return files, errors
}

func findResourcesInReader(p string, f io.Reader, resources chan<- Resource, errors chan<- error, buf []byte) {
	maxBufSize := 256 * 1024 * 1024
	scanner := bufio.NewScanner(f)
	// We start with a buf that is 4MB, scanner will resize it up to 256MB if needed
	// https://github.com/golang/go/blob/aeea5bacbf79fb945edbeac6cd7630dd70c4d9ce/src/bufio/scan.go#L191
	scanner.Buffer(buf, maxBufSize)
	scanner.Split(SplitYAMLDocument)
	nRes := 0
	for scanner.Scan() {
		if len(scanner.Text()) > 0 {
			res := Resource{Path: p, Bytes: []byte(scanner.Text())}
			for _, subres := range res.Resources() {
				resources <- subres
				nRes++
			}
		}
	}
	if err := scanner.Err(); err != nil {
		errors <- DiscoveryError{p, err}
	}
	if nRes == 0 {
		resources <- Resource{Path: p, Bytes: []byte{}}
	}
}

func findResourcesInFile(p string, resources chan<- Resource, errors chan<- error, buf []byte) {
	f, err := os.Open(p)
	defer f.Close()

	if err != nil {
		errors <- DiscoveryError{p, err}
		return
	}

	findResourcesInReader(p, f, resources, errors, buf)
}

func FromFiles(ctx context.Context, paths []string, ignoreFilePatterns []string) (<-chan Resource, <-chan error) {
	resources := make(chan Resource)

	files, errors := findFilesInFolders(ctx, paths, ignoreFilePatterns)

	go func() {
		initialBufSize := 4 * 1024 * 1024   // This is the initial size - scanner will resize if needed
		buf := make([]byte, initialBufSize) // We reuse the same buffer to avoid multiple large memory allocations

		for p := range files {
			findResourcesInFile(p, resources, errors, buf)
		}

		close(errors)
		close(resources)
	}()

	return resources, errors
}
