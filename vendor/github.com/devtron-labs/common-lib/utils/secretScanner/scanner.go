package secretScanner

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
)

const maxCapacity int = 256 * 1024 // 256KB
const DEVTRON = "DEVTRON"

// MaskSecretsOnString takes an input string and masks any secrets found based on the provided rules
func MaskSecretsOnString(input string) string {
	maskedInput := input

	for _, rule := range BuiltinRules {
		maskedInput = rule.Regex.ReplaceAllString(maskedInput, "******")

	}
	return maskedInput
}

// MaskCredentialKeyValues masks values assigned to credential-named keys, keeping the surrounding
// structure intact so the result stays parseable (e.g. valid JSON).
func MaskCredentialKeyValues(input string) string {
	return credentialAssignmentRegex.ReplaceAllString(input, "${pre}******")
}

// MaskSecretsOnStream processes an input stream, masking secrets according to built-in rules.
func MaskSecretsOnStream(input io.Reader) (io.Reader, error) {
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		scanner := bufio.NewScanner(input)
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		processLines(scanner, input, pw, buf)
	}()
	return pr, nil
}

func MaskSecretsOnStreamWithCtx(ctx context.Context, input io.Reader) (io.Reader, error) {
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		scanner := bufio.NewScanner(input)
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		processLinesWithCtx(ctx, scanner, input, pw, buf)
	}()
	return pr, nil
}

// processLines handles the main scanning and processing of lines from the input.
func processLines(scanner *bufio.Scanner, input io.Reader, pw *io.PipeWriter, buf []byte) {
	for scanner.Scan() {
		line := scanner.Text()
		writeMaskedLine(pw, line, true)
	}

	if err := scanner.Err(); err != nil {
		handleScanError(err, input, pw, buf)
	}
}

func processLinesWithCtx(ctx context.Context, scanner *bufio.Scanner, input io.Reader, pw *io.PipeWriter, buf []byte) {
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			line := scanner.Text()
			writeMaskedLine(pw, line, true)
		}
	}

	if err := scanner.Err(); err != nil {
		handleScanError(err, input, pw, buf)
	}
}

// writeMaskedLine writes the masked version of a line to the pipe writer.
func writeMaskedLine(pw *io.PipeWriter, line string, lineChange bool) {
	var err error
	if len(line) == 0 {
		_, err = pw.Write([]byte("\n"))
	} else {
		maskedString := MaskSecretsOnString(line)
		if lineChange {
			_, err = pw.Write([]byte(maskedString + "\n"))
		} else {
			_, err = pw.Write([]byte(maskedString))
		}
	}
	if err != nil {
		log.Println(DEVTRON, "Error writing to pipe: %v\n", err)
	}
}

// handleScanError handles errors encountered during scanning.
func handleScanError(err error, input io.Reader, pw *io.PipeWriter, buf []byte) {
	if errors.Is(err, bufio.ErrTooLong) {
		processLargeInput(input, pw, buf)
	} else {
		log.Println(DEVTRON, "Scanner error: %v\n", err)
	}
}

// processLargeInput handles processing of large inputs that exceed the scanner buffer size.
func processLargeInput(input io.Reader, pw *io.PipeWriter, buf []byte) {
	for {
		n, err := input.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println(DEVTRON, "Error reading input: %v\n", err)
			return
		}
		line := string(buf[:n])
		writeMaskedLine(pw, line, false)
	}
}
