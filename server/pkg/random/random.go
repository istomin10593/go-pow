package random

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"
)

const (
	// randLimit is the maximum number that can be generated.
	randLimit = 1000
)

// New generates a random string encoded in base-64 format.
func New() string {
	source := rand.NewSource(time.Now().UnixNano())
	generator := rand.New(source)

	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", generator.Intn(randLimit))))
}
