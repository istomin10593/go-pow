package handler

import (
	"context"
	"fmt"
	hash "go-pow/pkg/hashcash"
	"go-pow/pkg/pow"
	"go-pow/server/pkg/book"
	"go-pow/server/pkg/cache"
	"go-pow/server/pkg/config"

	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

// getConfig returns a config.Config with the given parameters.
func getConfig(zeroBits int, timeout time.Duration) *config.Config {
	config := &config.Config{}

	config.Server.Host = "localhost"
	config.Server.Port = ":8081"
	config.Server.Timeout = timeout

	config.Pow.ZeroBits = zeroBits

	config.Cache.Expiration = time.Second * 1

	return config
}

func TestHandleConnection(t *testing.T) {
	defaultserverHost := "localhost"
	defaultserverPort := ":8081"
	defaultSourceFile := "quotes_test.txt"
	defaultPowTimeout := time.Second * 1

	type args struct {
		zeroBits      int
		maxIterations int
		phaseNumber   int
	}

	tests := []struct {
		name       string
		args       args
		setupMocks func(*cache.MockCache)
		wantErr    error
		testClient func(*testing.T, args)
	}{
		{
			name: "Error - failed to read client request",
			args: args{
				zeroBits:    3,
				phaseNumber: 1,
			},
			wantErr: fmt.Errorf("failed to read request: failed to read from connection"),
			testClient: func(t *testing.T, args args) {
				conn, err := net.Dial("tcp", defaultserverHost+defaultserverPort)
				if err != nil {
					t.Fatal(err)
				}

				t.Log("test client connected to the server")

				conn.Close()

				t.Log("test client shutdown")
			},
		},
		{
			name: "Success - pow validation successful",
			args: args{
				zeroBits:    3,
				phaseNumber: 1,
			},
			setupMocks: func(cacheMock *cache.MockCache) {
				cacheMock.On("GetUniqueID", context.Background()).Return(int64(123), nil)
			},
			testClient: func(t *testing.T, args args) {
				conn, err := net.Dial("tcp", defaultserverHost+defaultserverPort)
				if err != nil {
					t.Fatal(err)
				}

				t.Log("test client connected to the server")

				prot := pow.New(conn, defaultPowTimeout)

				if err := prot.Write(); err != nil {
					t.Fatal(err)
				}

				t.Log("test client requested a challenge header")

				if err := prot.Read(); err != nil {
					t.Fatal(err)
				}

				t.Log("test client received a challenge header:", string(prot.Payload()))

				var hashcash hash.Hashcash
				if err := hashcash.Parse(prot.Payload()); err != nil {
					t.Fatal(err)
				}

				if err := hashcash.Calculate(args.maxIterations); err != nil {
					t.Fatal(err)
				}

				conn, err = net.Dial("tcp", defaultserverHost+defaultserverPort)
				if err != nil {
					t.Fatal(err)
				}

				t.Log("test client connected to the server")

				prot = pow.New(conn, defaultPowTimeout)

				prot.SetPhase(pow.ValidPhase)
				prot.SetPayload(hashcash.Header())

				if err := prot.Write(); err != nil {
					t.Fatal(err)
				}

				t.Log("test client sent a solution header:", string(hashcash.Header()))

				if err := prot.Read(); err != nil {
					t.Fatal(err)
				}

				t.Log("test client received a response:", string(prot.Payload()))

				conn.Close()

				t.Log("test client shutdown")
			},
		},
		// {
		// 	name: "Error - failed to parse a response header",
		// 	args: args{
		// 		serverPort:    defaultserverPort,
		// 		serverHost:    defaultserverHost,
		// 		zeroBits:      3,
		// 		maxIterations: 10000,
		// 	},
		// 	wantErr: hash.ErrInvalidVersion,
		// 	testClient: func(t *testing.T, args args, syncChan chan struct{}) {
		// 		<-syncChan

		// 		conn, err := net.Dial("tcp", args.serverHost+args.serverPort)
		// 		if err != nil {
		// 			t.Fatal(err)
		// 		}

		// 		t.Log("test client connected to the server")

		// 		buf := make([]byte, 1024)
		// 		n, err := conn.Read(buf)
		// 		if err != nil {
		// 			t.Fatal(err)
		// 		}

		// 		headerPOW := string(buf[:n])

		// 		t.Log("test client received a challenge header:", headerPOW)

		// 		headerResponse := "0:3:230920144708:127.0.0.1:50368::MjQ3:MA="
		// 		if _, err := conn.Write([]byte(headerResponse)); err != nil {
		// 			t.Fatal(err)
		// 		}

		// 		t.Log("test client sent an invalid header:", headerResponse)

		// 		conn.Close()

		// 		t.Log("test client shutdown")
		// 	},
		// },
		// {
		// 	name: "Error - pow validation failed",
		// 	args: args{
		// 		serverPort:    defaultserverPort,
		// 		serverHost:    defaultserverHost,
		// 		zeroBits:      3,
		// 		maxIterations: 10000,
		// 	},
		// 	wantErr: hash.ErrMaxIterationsExceeded,
		// 	testClient: func(t *testing.T, args args, syncChan chan struct{}) {
		// 		<-syncChan

		// 		conn, err := net.Dial("tcp", args.serverHost+args.serverPort)
		// 		if err != nil {
		// 			t.Fatal(err)
		// 		}

		// 		t.Log("test client connected to the server")

		// 		buf := make([]byte, 1024)
		// 		n, err := conn.Read(buf)
		// 		if err != nil {
		// 			t.Fatal(err)
		// 		}

		// 		headerPOW := string(buf[:n])

		// 		t.Log("test client received a challenge header:", headerPOW)

		// 		headerResponse := "1:3:230920144708:127.0.0.1:50368::MjQ3:MQ=="
		// 		if _, err := conn.Write([]byte(headerResponse)); err != nil {
		// 			t.Fatal(err)
		// 		}

		// 		t.Log("test client sent a header with failed pow:", headerResponse)

		// 		conn.Close()

		// 		t.Log("test client shutdown")
		// 	},
		// },
		{
			name: "Success - pow validation successful",
			args: args{
				zeroBits:      3,
				maxIterations: 1000000,
				phaseNumber:   2,
			},
			testClient: func(t *testing.T, args args) {
				conn, err := net.Dial("tcp", defaultserverHost+defaultserverPort)
				if err != nil {
					t.Fatal(err)
				}

				t.Log("test client connected to the server")

				prot := pow.New(conn, defaultPowTimeout)

				if err := prot.Write(); err != nil {
					t.Fatal(err)
				}

				t.Log("test client requested a challenge header")

				if err := prot.Read(); err != nil {
					t.Fatal(err)
				}

				t.Log("test client received a challenge header:", string(prot.Payload()))

				var hashcash hash.Hashcash
				if err := hashcash.Parse(prot.Payload()); err != nil {
					t.Fatal(err)
				}

				if err := hashcash.Calculate(args.maxIterations); err != nil {
					t.Fatal(err)
				}

				conn, err = net.Dial("tcp", defaultserverHost+defaultserverPort)
				if err != nil {
					t.Fatal(err)
				}

				t.Log("test client connected to the server")

				prot = pow.New(conn, defaultPowTimeout)

				prot.SetPhase(pow.ValidPhase)
				prot.SetPayload(hashcash.Header())

				if err := prot.Write(); err != nil {
					t.Fatal(err)
				}

				t.Log("test client sent a solution header:", string(hashcash.Header()))

				if err := prot.Read(); err != nil {
					t.Fatal(err)
				}

				t.Log("test client received a response:", string(prot.Payload()))

				conn.Close()

				t.Log("test client shutdown")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get config.
			cfg := getConfig(tt.args.zeroBits, defaultPowTimeout)

			// Create handler.
			log := zaptest.NewLogger(t)
			defer log.Sync()

			book, err := book.New(defaultSourceFile)
			assert.NoError(t, err)

			cache := &cache.MockCache{}

			handler := New(tt.args.zeroBits, cfg.Server.Timeout, log, book, cache)

			// Handle connection.
			ln, err := net.Listen("tcp", cfg.Server.Host+cfg.Server.Port)
			defer ln.Close()

			assert.NoError(t, err)

			// Run test client.
			go tt.testClient(t, tt.args)

			for i := 0; i < tt.args.phaseNumber; i++ {
				conn, err := ln.Accept()
				assert.NoError(t, err)

				err = handler.Handle(conn)

				if tt.wantErr != nil {
					assert.EqualError(t, err, tt.wantErr.Error())
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}
