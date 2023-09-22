package client

import (
	"go-pow/client/pkg/config"
	"go-pow/pkg/hashcash"

	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

// getConfig returns a config.Config with the given parameters.
func getConfig(serverPort string, maxIteration int) *config.Config {
	config := &config.Config{}

	config.Server.Host = "localhost"
	config.Server.Port = serverPort
	config.Server.Timeout = time.Second * 1

	config.Pow.MaxIterations = maxIteration

	return config
}

func TestHandleConnection(t *testing.T) {
	defaultserverPort := ":8081"

	type args struct {
		serverPort   string
		maxIteration int
		challenge    string
	}

	tests := []struct {
		name       string
		args       args
		wantErr    error
		testServer func(*testing.T, args, chan struct{})
	}{
		{
			name: "Error - failed to read a challenge response",
			args: args{
				serverPort: defaultserverPort,
			},
			wantErr: io.EOF,
			testServer: func(t *testing.T, args args, syncChan chan struct{}) {
				l, err := net.Listen("tcp", args.serverPort)
				if err != nil {
					t.Fatal(err)
				}
				defer l.Close()

				t.Log("test server  listening on port", args.serverPort)

				syncChan <- struct{}{}

				conn, err := l.Accept()
				if err != nil {
					t.Fatal(err)
				}

				t.Log("test server accepted connection")

				conn.Close()

				t.Log("test server shutdown")
			},
		},
		{
			name: "Error - failed to parse the challenge header",
			args: args{
				serverPort: defaultserverPort,
				challenge:  "0:4:230919221643:127.0.0.1:44736::OTg3:MA==",
			},
			wantErr: hashcash.ErrInvalidVersion,
			testServer: func(t *testing.T, args args, syncChan chan struct{}) {
				l, err := net.Listen("tcp", args.serverPort)
				if err != nil {
					t.Fatal(err)
				}
				defer l.Close()

				t.Log("test server  listening on port", args.serverPort)

				syncChan <- struct{}{}

				conn, err := l.Accept()
				if err != nil {
					t.Fatal(err)
				}

				t.Log("test server accepted connection")

				if _, err := conn.Write([]byte(args.challenge)); err != nil {
					t.Fatal(err)
				}

				t.Log("test server send challenge")

				conn.Close()

				t.Log("test server shutdown")
			},
		},
		{
			name: "Error - failed to solve the pow",
			args: args{
				maxIteration: 100000,
				serverPort:   defaultserverPort,
				challenge:    "1:50:230919221643:127.0.0.1:44736::OTg3:MA==",
			},
			wantErr: hashcash.ErrExceedHashLength,
			testServer: func(t *testing.T, args args, syncChan chan struct{}) {
				l, err := net.Listen("tcp", args.serverPort)
				if err != nil {
					t.Fatal(err)
				}
				defer l.Close()

				t.Log("test server  listening on port", args.serverPort)

				syncChan <- struct{}{}

				conn, err := l.Accept()
				if err != nil {
					t.Fatal(err)
				}

				t.Log("test server accepted connection")

				if _, err := conn.Write([]byte(args.challenge)); err != nil {
					t.Fatal(err)
				}

				t.Log("test server send challenge")

				conn.Close()

				t.Log("test server shutdown")
			},
		},
		{
			name: "Error - failed to send the solution header",
			args: args{
				maxIteration: 100000,
				serverPort:   defaultserverPort,
				challenge:    "1:4:230919221643:127.0.0.1:44736::OTg3:MA==",
			},
			wantErr: io.EOF,
			testServer: func(t *testing.T, args args, syncChan chan struct{}) {
				l, err := net.Listen("tcp", args.serverPort)
				if err != nil {
					t.Fatal(err)
				}
				defer l.Close()

				t.Log("test server  listening on port", args.serverPort)

				syncChan <- struct{}{}

				conn, err := l.Accept()
				if err != nil {
					t.Fatal(err)
				}

				t.Log("test server accepted connection")

				if _, err := conn.Write([]byte(args.challenge)); err != nil {
					t.Fatal(err)
				}

				t.Log("test server send challenge")

				conn.Close()

				t.Log("test server shutdown")
			},
		},
		{
			name: "Success - solve the pow",
			args: args{
				maxIteration: 100000,
				serverPort:   defaultserverPort,
				challenge:    "1:4:230919221643:127.0.0.1:44736::OTg3:MA==",
			},
			testServer: func(t *testing.T, args args, syncChan chan struct{}) {
				l, err := net.Listen("tcp", args.serverPort)
				if err != nil {
					t.Fatal(err)
				}
				defer l.Close()

				t.Log("test server  listening on port", args.serverPort)

				syncChan <- struct{}{}

				conn, err := l.Accept()
				if err != nil {
					t.Fatal(err)
				}

				t.Log("test server accepted connection")

				if _, err := conn.Write([]byte(args.challenge)); err != nil {
					t.Fatal(err)
				}

				t.Log("test server send a challenge")

				buf := make([]byte, 1024)
				n, err := conn.Read(buf)
				if err != nil {
					t.Fatal(err)
				}

				headerSolution := string(buf[:n])

				t.Log("test server received a solution header")

				var hashcashResponse hashcash.Hashcash
				if err := hashcashResponse.Parse(headerSolution); err != nil {
					t.Log("test server failed to parse a response header")
					t.Fatal(err)
				}

				if err := hashcashResponse.Calculate(hashcashResponse.GetCounter()); err != nil {
					t.Log("pow validation failed")
					t.Fatal(err)
				}

				if _, err := conn.Write([]byte("good job")); err != nil {
					t.Log("est server failed to send a quote")
					t.Fatal(err)
				}

				conn.Close()

				t.Log("test server shutdown")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncChan := make(chan struct{})
			defer close(syncChan)

			// Run test server.
			go tt.testServer(t, tt.args, syncChan)

			// Handle connection.
			log := zaptest.NewLogger(t)
			defer log.Sync()
			config := getConfig(tt.args.serverPort, tt.args.maxIteration)

			done := make(chan struct{})
			defer close(done)

			go func() {
				<-syncChan
				err := handleConnection(log, config, done, 1)
				if err != nil {
					assert.Equal(t, tt.wantErr, err)
				}
			}()

			assert.Equal(t, struct{}{}, <-done)
		})
	}
}
