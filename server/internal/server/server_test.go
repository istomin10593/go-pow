package server

import (
	hash "go-pow/pkg/hashcash"
	"go-pow/server/pkg/book"
	"go-pow/server/pkg/config"

	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

// getConfig returns a config.Config with the given parameters.
func getConfig(serverHost, serverPort string, zeroBits int) *config.Config {
	config := &config.Config{}

	config.Server.Host = serverHost
	config.Server.Port = serverPort
	config.Server.Timeout = time.Second * 1

	config.Pow.ZeroBits = zeroBits

	return config
}

func TestHandleConnection(t *testing.T) {
	defaultserverHost := "localhost"
	defaultserverPort := ":8081"
	defaultSourceFile := "quotes_test.txt"

	type args struct {
		serverPort    string
		serverHost    string
		zeroBits      int
		maxIterations int
	}

	tests := []struct {
		name       string
		args       args
		wantErr    error
		testClient func(*testing.T, args, chan struct{})
	}{
		{
			name: "Error - failed to send the challenge header",
			args: args{
				serverPort: defaultserverPort,
				serverHost: defaultserverHost,
				zeroBits:   3,
			},
			wantErr: io.EOF,
			testClient: func(t *testing.T, args args, syncChan chan struct{}) {
				<-syncChan

				conn, err := net.Dial("tcp", args.serverHost+args.serverPort)
				if err != nil {
					t.Fatal(err)
				}

				t.Log("test client connected to the server")

				conn.Close()

				t.Log("test client shutdown")
			},
		},
		{
			name: "Error - failed to read a response. Client close connection",
			args: args{
				serverPort:    defaultserverPort,
				serverHost:    defaultserverHost,
				zeroBits:      3,
				maxIterations: 10000,
			},
			wantErr: io.EOF,
			testClient: func(t *testing.T, args args, syncChan chan struct{}) {
				<-syncChan

				conn, err := net.Dial("tcp", args.serverHost+args.serverPort)
				if err != nil {
					t.Fatal(err)
				}

				t.Log("test client connected to the server")

				buf := make([]byte, 1024)
				n, err := conn.Read(buf)
				if err != nil {
					t.Fatal(err)
				}

				headerPOW := string(buf[:n])

				t.Log("test client received a challenge header:", headerPOW)

				conn.Close()

				t.Log("test client shutdown")
			},
		},
		{
			name: "Error - failed to parse a response header",
			args: args{
				serverPort:    defaultserverPort,
				serverHost:    defaultserverHost,
				zeroBits:      3,
				maxIterations: 10000,
			},
			wantErr: hash.ErrInvalidVersion,
			testClient: func(t *testing.T, args args, syncChan chan struct{}) {
				<-syncChan

				conn, err := net.Dial("tcp", args.serverHost+args.serverPort)
				if err != nil {
					t.Fatal(err)
				}

				t.Log("test client connected to the server")

				buf := make([]byte, 1024)
				n, err := conn.Read(buf)
				if err != nil {
					t.Fatal(err)
				}

				headerPOW := string(buf[:n])

				t.Log("test client received a challenge header:", headerPOW)

				headerResponse := "0:3:230920144708:127.0.0.1:50368::MjQ3:MA="
				if _, err := conn.Write([]byte(headerResponse)); err != nil {
					t.Fatal(err)
				}

				t.Log("test client sent an invalid header:", headerResponse)

				conn.Close()

				t.Log("test client shutdown")
			},
		},
		{
			name: "Error - pow validation failed",
			args: args{
				serverPort:    defaultserverPort,
				serverHost:    defaultserverHost,
				zeroBits:      3,
				maxIterations: 10000,
			},
			wantErr: hash.ErrMaxIterationsExceeded,
			testClient: func(t *testing.T, args args, syncChan chan struct{}) {
				<-syncChan

				conn, err := net.Dial("tcp", args.serverHost+args.serverPort)
				if err != nil {
					t.Fatal(err)
				}

				t.Log("test client connected to the server")

				buf := make([]byte, 1024)
				n, err := conn.Read(buf)
				if err != nil {
					t.Fatal(err)
				}

				headerPOW := string(buf[:n])

				t.Log("test client received a challenge header:", headerPOW)

				headerResponse := "1:3:230920144708:127.0.0.1:50368::MjQ3:MQ=="
				if _, err := conn.Write([]byte(headerResponse)); err != nil {
					t.Fatal(err)
				}

				t.Log("test client sent a header with failed pow:", headerResponse)

				conn.Close()

				t.Log("test client shutdown")
			},
		},
		{
			name: "Success - pow validation successful",
			args: args{
				serverPort:    defaultserverPort,
				serverHost:    defaultserverHost,
				zeroBits:      3,
				maxIterations: 1000000,
			},
			testClient: func(t *testing.T, args args, syncChan chan struct{}) {
				<-syncChan

				conn, err := net.Dial("tcp", args.serverHost+args.serverPort)
				if err != nil {
					t.Fatal(err)
				}

				t.Log("test client connected to the server")

				buf := make([]byte, 1024)
				n, err := conn.Read(buf)
				if err != nil {
					t.Fatal(err)
				}

				headerPOW := string(buf[:n])

				t.Log("test client received a challenge header:", headerPOW)

				var hashcash hash.Hashcash
				if err := hashcash.Parse(headerPOW); err != nil {
					t.Fatal(err)
				}

				if err := hashcash.Calculate(args.maxIterations); err != nil {
					t.Fatal(err)
				}

				if _, err := conn.Write([]byte(hashcash.String())); err != nil {
					t.Fatal(err)
				}

				t.Log("test client sent a solution header:", hashcash.String())

				buf = make([]byte, 1024)
				n, err = conn.Read(buf)
				if err != nil {
					t.Fatal(err)
				}

				response := string(buf[:n])

				t.Log("test client received a response:", response)

				conn.Close()

				t.Log("test client shutdown")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncChan := make(chan struct{})
			defer close(syncChan)

			// Run test server.
			go tt.testClient(t, tt.args, syncChan)

			// Handle connection.
			cfg := getConfig(tt.args.serverHost, tt.args.serverPort, tt.args.zeroBits)
			log := zaptest.NewLogger(t)
			defer log.Sync()

			book, err := book.New(defaultSourceFile)

			server := New(log, cfg, book)
			assert.NoError(t, err)

			ln, err := net.Listen("tcp", cfg.Server.Host+cfg.Server.Port)
			defer ln.Close()

			assert.NoError(t, err)

			syncChan <- struct{}{}

			conn, err := ln.Accept()
			assert.NoError(t, err)

			err = server.handle(conn)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
