package book

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
)

// delimiter for quote
const (
	delimiter byte = '~'
)

// Quote with wisdom words
type Quote struct {
	Text   []byte
	Author []byte
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
		parts := bytes.Split(scanner.Bytes(), []byte{delimiter})
		if len(parts) == 2 {
			quote := Quote{
				Text:   bytes.TrimSpace(parts[0]),
				Author: bytes.TrimSpace(parts[1]),
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

	buf := &bytes.Buffer{}
	buf.Write(b[i].Text)
	buf.WriteByte(delimiter)
	buf.Write(b[i].Author)

	return buf.Bytes()
}
