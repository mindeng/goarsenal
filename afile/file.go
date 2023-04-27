package afile

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// ReadFromFileOrStdin reads data from the specified inputFile.
// if the inputFile is empty or "-", it reads from stdin.
func ReadFromFileOrStdin(inputFile string) ([]byte, error) {
	var reader *bufio.Reader
	if inputFile == "" || inputFile == "-" {
		// read from stdin
		reader = bufio.NewReader(os.Stdin)
	} else {
		// read from file
		file, err := os.Open(inputFile)
		if err != nil {
			return nil, fmt.Errorf("Failed to open input file: %v", err)
		}
		defer file.Close()
		reader = bufio.NewReader(file)
	}

	var data []byte

	// read bytes from reader until EOF
	buf := make([]byte, 512)
	for {
		b, err := reader.Read(buf)
		if b > 0 {
			data = append(data, buf[:b]...)
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("Failed to read input: %v", err)
		}
	}

	return data, nil
}

// WriteToFileOrStdout writes data to the specified outputFile.
// if the outputFile is empty or "-", it writes to stdout.
func WriteToFileOrStdout(outputFile string, data []byte) error {
	var writer *bufio.Writer
	if outputFile == "" || outputFile == "-" {
		// write to stdout
		writer = bufio.NewWriter(os.Stdout)
	} else {
		// write to file
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("Failed to create output file: %v", err)
		}
		defer file.Close()
		writer = bufio.NewWriter(file)
	}

	_, err := writer.Write(data)
	if err != nil {
		return fmt.Errorf("Failed to write output: %v", err)
	}

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("Failed to flush output: %v", err)
	}

	return nil
}
