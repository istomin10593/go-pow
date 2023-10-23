package client

import (
	"errors"
	"go-pow/client/pkg/config"
	"go-pow/pkg/hashcash"
	"go-pow/pkg/pow"

	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

const (
	defaultServerHost      = "localhost"
	defaultServerPort      = ":8081"
	defaultSourceFile      = "quotes_test.txt"
	defaultPowTimeout      = time.Second * 1
	defaultCacheExpitation = time.Millisecond * 50
)

// getConfig returns a config.Config with the given parameters.
func getConfig(maxIteration int) *config.Config {
	config := &config.Config{}

	config.Server.Host = defaultServerHost
	config.Server.Port = defaultServerPort
	config.Server.Timeout = defaultPowTimeout

	config.Pow.MaxIterations = maxIteration

	return config
}

func TestHandleConnection(t *testing.T) {
	type args struct {
		serverAddress string
		maxIteration  int
		challenge     string
	}

	tests := []struct {
		name       string
		args       args
		wantErr    error
		testServer func(*testing.T, args, chan struct{})
	}{
		{
			name: "Error - failed to read a challenge header",
			args: args{
				serverAddress: defaultServerHost + defaultServerPort,
			},
			wantErr: pow.ErrReadConn,
			testServer: func(t *testing.T, args args, sync chan struct{}) {
				l, err := net.Listen("tcp", args.serverAddress)
				if err != nil {
					t.Fatal(err)
				}
				defer l.Close()

				t.Log("test server  listening on port", args.serverAddress)

				sync <- struct{}{}

				conn, err := l.Accept()
				if err != nil {
					t.Log("test server failed to accept connection")
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
				serverAddress: defaultServerHost + defaultServerPort,
				challenge:     "0:3:231019170010:127.0.0.1:58312::NzUx:MA==",
			},
			wantErr: hashcash.ErrInvalidVersion,
			testServer: func(t *testing.T, args args, sync chan struct{}) {
				l, err := net.Listen("tcp", args.serverAddress)
				if err != nil {
					t.Fatal(err)
				}
				defer l.Close()

				t.Log("test server  listening on", args.serverAddress)

				sync <- struct{}{}

				conn, err := l.Accept()
				if err != nil {
					t.Log("test server failed to accept connection")
					t.Fatal(err)
				}

				t.Log("test server accepted connection")

				prot := pow.New(conn, defaultPowTimeout)

				if err := prot.Read(); err != nil {
					t.Log("test server failed to read request for a challenge")
					t.Fatal(err)
				}

				if prot.Phase() != pow.InitPhase {
					t.Fatal("phase is not init")
				}

				t.Log("test server received  request for a challenge")

				prot.SetPayload([]byte(args.challenge))

				if err := prot.Write(); err != nil {
					t.Log("test server failed to send a challenge")
					t.Fatal(err)
				}

				t.Log("test server send a challenge")

				conn.Close()

				t.Log("test server shutdown")
			},
		},
		{
			name: "Error - failed to solve the pow",
			args: args{
				maxIteration:  100000,
				serverAddress: defaultServerHost + defaultServerPort,
				challenge:     "1:50:231019170010:127.0.0.1:58312::NzUx:MA==",
			},
			wantErr: hashcash.ErrMaxIterationsExceeded,
			testServer: func(t *testing.T, args args, sync chan struct{}) {
				l, err := net.Listen("tcp", args.serverAddress)
				if err != nil {
					t.Fatal(err)
				}
				defer l.Close()

				t.Log("test server  listening on", args.serverAddress)

				sync <- struct{}{}

				conn, err := l.Accept()
				if err != nil {
					t.Log("test server failed to accept connection")
					t.Fatal(err)
				}

				t.Log("test server accepted connection")

				prot := pow.New(conn, defaultPowTimeout)

				if err := prot.Read(); err != nil {
					t.Log("test server failed to read request for a challenge")
					t.Fatal(err)
				}

				if prot.Phase() != pow.InitPhase {
					t.Fatal("phase is not init")
				}

				t.Log("test server received  request for a challenge")

				prot.SetPayload([]byte(args.challenge))

				if err := prot.Write(); err != nil {
					t.Log("test server failed to send a challenge")
					t.Fatal(err)
				}

				t.Log("test server send a challenge")

				conn.Close()

				t.Log("test server shutdown")
			},
		},
		{
			name: "Error - failed to read a wisdom quote",
			args: args{
				maxIteration:  100000,
				serverAddress: defaultServerHost + defaultServerPort,
				challenge:     "1:3:231019170010:127.0.0.1:58312::NzUx:MA==",
			},
			wantErr: pow.ErrReadConn,
			testServer: func(t *testing.T, args args, sync chan struct{}) {
				l, err := net.Listen("tcp", args.serverAddress)
				if err != nil {
					t.Fatal(err)
				}
				defer l.Close()

				t.Log("test server  listening on", args.serverAddress)

				sync <- struct{}{}

				conn, err := l.Accept()
				if err != nil {
					t.Log("test server failed to accept connection")
					t.Fatal(err)
				}

				t.Log("test server accepted connection")

				prot := pow.New(conn, defaultPowTimeout)

				if err := prot.Read(); err != nil {
					t.Log("test server failed to read request for a challenge")
					t.Fatal(err)
				}

				if prot.Phase() != pow.InitPhase {
					t.Fatal("phase is not init")
				}

				t.Log("test server received  request for a challenge")

				prot.SetPayload([]byte(args.challenge))

				if err := prot.Write(); err != nil {
					t.Log("test server failed to send a challenge")
					t.Fatal(err)
				}

				t.Log("test server send a challenge")

				conn, err = l.Accept()
				if err != nil {
					t.Log("test server failed to accept connection")
					t.Fatal(err)
				}

				conn.Close()

				t.Log("test server shutdown")
			},
		},
		{
			name: "Success - solve the pow",
			args: args{
				maxIteration:  100000,
				serverAddress: defaultServerHost + defaultServerPort,
				challenge:     "1:3:231019170010:127.0.0.1:58312::NzUx:MA==",
			},
			testServer: func(t *testing.T, args args, sync chan struct{}) {
				l, err := net.Listen("tcp", args.serverAddress)
				if err != nil {
					t.Fatal(err)
				}
				defer l.Close()

				t.Log("test server  listening on", args.serverAddress)

				sync <- struct{}{}

				conn, err := l.Accept()
				if err != nil {
					t.Log("test server failed to accept connection")
					t.Fatal(err)
				}

				t.Log("test server accepted connection")

				prot := pow.New(conn, defaultPowTimeout)

				if err := prot.Read(); err != nil {
					t.Log("test server failed to read request for a challenge")
					t.Fatal(err)
				}

				if prot.Phase() != pow.InitPhase {
					t.Fatal("phase is not init")
				}

				t.Log("test server received  request for a challenge")

				prot.SetPayload([]byte(args.challenge))

				if err := prot.Write(); err != nil {
					t.Log("test server failed to send a challenge")
					t.Fatal(err)
				}

				t.Log("test server send a challenge")

				conn, err = l.Accept()
				if err != nil {
					t.Log("test server failed to accept connection")
					t.Fatal(err)
				}

				prot = pow.New(conn, defaultPowTimeout)

				if err := prot.Read(); err != nil {
					t.Log("test server failed to read a solution header")
					t.Fatal(err)
				}

				if prot.Phase() != pow.ValidPhase {
					t.Fatal("phase is not valid")
				}

				t.Log("test server received a solution header")

				var hashcashResponse hashcash.Hashcash
				if err := hashcashResponse.Parse(prot.Payload()); err != nil {
					t.Log("test server failed to parse a response header")
					t.Fatal(err)
				}

				if err := hashcashResponse.Calculate(hashcashResponse.GetCounter()); err != nil {
					t.Log("pow validation failed")
					t.Fatal(err)
				}

				prot.SetPayload([]byte("good job"))
				if err := prot.Write(); err != nil {
					t.Log("test server failed to send a quote")
					t.Fatal(err)
				}

				t.Log("test server sent a quote")

				conn.Close()

				t.Log("test server shutdown")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sync := make(chan struct{})
			defer close(sync)

			// Run test server.
			go tt.testServer(t, tt.args, sync)

			<-sync

			// Handle connection.
			log := zaptest.NewLogger(t)
			defer log.Sync()
			config := getConfig(tt.args.maxIteration)

			client := New(log, config)
			done := make(chan struct{})
			defer close(done)

			go func() {
				err := client.handleConnection(1, done)
				if err != nil {
					assert.True(t, errors.Is(err, tt.wantErr))
				}
			}()

			assert.Equal(t, struct{}{}, <-done)
		})
	}
}
