package handler

import (
	"errors"
	hash "go-pow/pkg/hashcash"
	"go-pow/pkg/pow"
	"go-pow/server/pkg/book"
	"go-pow/server/pkg/cache"
	"go-pow/server/pkg/config"

	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

const (
	defaultserverHost      = "localhost"
	defaultserverPort      = ":8081"
	defaultSourceFile      = "quotes_test.txt"
	defaultPowTimeout      = time.Second * 1
	defaultCacheExpitation = time.Millisecond * 50
)

// getConfig returns a config.Config with the given parameters.
func getConfig(zeroBits int, timeout time.Duration) *config.Config {
	config := &config.Config{}

	config.Server.Host = defaultserverHost
	config.Server.Port = defaultserverPort
	config.Server.Timeout = defaultPowTimeout

	config.Pow.ZeroBits = zeroBits

	config.Cache.Expiration = defaultCacheExpitation

	return config
}

func TestHandleConnection(t *testing.T) {
	initPhaseCacheErr := errors.New("initPhaseCacheErr")
	validPhseCacheErr := errors.New("validPhaseCacheErr")

	headerInvalidVersion := "0:3:231019170010:127.0.0.1:58312::NzUx:MTM4Ng=="
	headerInValidSolution := "1:3:231019170010:127.0.0.1:58312::NzUx:MTM4MA=="
	headerValid := "1:3:231019170010:127.0.0.1:58312::NzUx:MTM4Ng=="

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
			name: "Error - failed to read client request. Failed to read from connection",
			args: args{
				zeroBits:    3,
				phaseNumber: 1,
			},
			wantErr: pow.ErrReadConn,
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
			name: "Error - failed to add rand to the cache",
			args: args{
				zeroBits:    3,
				phaseNumber: 1,
			},
			setupMocks: func(cacheMock *cache.MockCache) {
				cacheMock.On("Add", mock.Anything).Return(initPhaseCacheErr)
			},
			wantErr: initPhaseCacheErr,
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

				conn.Close()

				t.Log("test client shutdown")
			},
		},
		{
			name: "Error - failed to parse a solution header, invalid version",
			args: args{
				zeroBits:    3,
				phaseNumber: 1,
			},
			wantErr: hash.ErrInvalidVersion,
			testClient: func(t *testing.T, args args) {
				conn, err := net.Dial("tcp", defaultserverHost+defaultserverPort)
				if err != nil {
					t.Fatal(err)
				}

				t.Log("test client connected to the server")

				prot := pow.New(conn, defaultPowTimeout)

				prot.SetValidPhase()
				prot.SetPayload([]byte(headerInvalidVersion))

				if err := prot.Write(); err != nil {
					t.Fatal(err)
				}

				t.Log("test client sent a solution header:", headerInvalidVersion)

				conn.Close()

				t.Log("test client shutdown")
			},
		},
		{
			name: "Error - failed to check a cache",
			args: args{
				zeroBits:    3,
				phaseNumber: 1,
			},
			setupMocks: func(cacheMock *cache.MockCache) {
				cacheMock.On("Get", mock.Anything).Return(false, validPhseCacheErr)
			},
			wantErr: validPhseCacheErr,
			testClient: func(t *testing.T, args args) {
				conn, err := net.Dial("tcp", defaultserverHost+defaultserverPort)
				if err != nil {
					t.Fatal(err)
				}

				t.Log("test client connected to the server")

				prot := pow.New(conn, defaultPowTimeout)

				prot.SetValidPhase()
				prot.SetPayload([]byte(headerValid))

				if err := prot.Write(); err != nil {
					t.Fatal(err)
				}

				t.Log("test client sent a solution header:", headerValid)

				conn.Close()

				t.Log("test client shutdown")
			},
		},
		{
			name: "Error - rand is not in the cache",
			args: args{
				zeroBits:    3,
				phaseNumber: 1,
			},
			setupMocks: func(cacheMock *cache.MockCache) {
				cacheMock.On("Get", mock.Anything).Return(false, nil)
			},
			wantErr: ErrUnreconizedRandom,
			testClient: func(t *testing.T, args args) {
				conn, err := net.Dial("tcp", defaultserverHost+defaultserverPort)
				if err != nil {
					t.Fatal(err)
				}

				t.Log("test client connected to the server")

				prot := pow.New(conn, defaultPowTimeout)

				prot.SetValidPhase()
				prot.SetPayload([]byte(headerValid))

				if err := prot.Write(); err != nil {
					t.Fatal(err)
				}

				t.Log("test client sent a solution header:", headerValid)

				conn.Close()

				t.Log("test client shutdown")
			},
		},
		{
			name: "Error - pow validation failed",
			args: args{
				zeroBits:    3,
				phaseNumber: 1,
			},
			setupMocks: func(cacheMock *cache.MockCache) {
				cacheMock.On("Get", mock.Anything).Return(true, nil)
			},
			wantErr: ErrInvalidSolutionHeader,
			testClient: func(t *testing.T, args args) {
				conn, err := net.Dial("tcp", defaultserverHost+defaultserverPort)
				if err != nil {
					t.Fatal(err)
				}

				t.Log("test client connected to the server")

				prot := pow.New(conn, defaultPowTimeout)

				prot.SetValidPhase()
				prot.SetPayload([]byte(headerInValidSolution))

				if err := prot.Write(); err != nil {
					t.Fatal(err)
				}

				t.Log("test client sent a solution header:", headerInValidSolution)

				conn.Close()

				t.Log("test client shutdown")
			},
		},
		{
			name: "Success - pow validation successful",
			args: args{
				zeroBits:      3,
				maxIterations: 1000000,
				phaseNumber:   2,
			},
			setupMocks: func(cacheMock *cache.MockCache) {
				cacheMock.On("Add", mock.Anything).Return(nil)
				cacheMock.On("Get", mock.Anything).Return(true, nil)
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

				prot.SetValidPhase()
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
			// Setup mocks.
			cacheMock := &cache.MockCache{}

			if tt.setupMocks != nil {
				tt.setupMocks(cacheMock)
			}

			// Get config.
			cfg := getConfig(tt.args.zeroBits, defaultPowTimeout)

			// Create handler.
			log := zaptest.NewLogger(t)
			defer log.Sync()

			book, err := book.New(defaultSourceFile)
			assert.NoError(t, err)

			handler := New(tt.args.zeroBits, cfg.Server.Timeout, log, book, cacheMock)

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
					assert.True(t, errors.Is(err, tt.wantErr))
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}
