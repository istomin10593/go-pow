package book

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

const (
	delimiter = "~"
)

// Quote with wisdom words
type Quote struct {
	Text   string
	Author string
}

// Book with quotes.
type Book []*Quote

// New create new Book.
func New(filePath string) (*Book, error) {
	filename, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("can't get source  path: %w", err)
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("can't open the source  file: %w", err)
	}
	defer file.Close()

	var quotesList Book
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), delimiter)
		if len(parts) == 2 {
			quote := Quote{
				Text:   strings.TrimSpace(parts[0]),
				Author: strings.TrimSpace(parts[1]),
			}
			quotesList = append(quotesList, &quote)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read the source file: %w", err)
	}

	return &quotesList, nil
}

// GetRandQuote returns random quote.
func (b Book) GetRandQuote() []byte {
	i := rand.Intn(len(b))
	return []byte(b[i].Text + delimiter + b[i].Author)
}
