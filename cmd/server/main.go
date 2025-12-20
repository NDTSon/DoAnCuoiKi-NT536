// Copyright 2023 LiveKit, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/urfave/cli/v3"

	appauth "github.com/livekit/livekit-server/pkg/auth"
	apphandler "github.com/livekit/livekit-server/pkg/handler"
	"github.com/livekit/livekit-server/pkg/storage"

	"github.com/livekit/livekit-server/pkg/config"
	"github.com/livekit/livekit-server/pkg/routing"
	"github.com/livekit/livekit-server/pkg/rtc"
	"github.com/livekit/livekit-server/pkg/service"
	"github.com/livekit/livekit-server/pkg/telemetry/prometheus"
	"github.com/livekit/livekit-server/version"
	"github.com/livekit/protocol/logger"
)

var baseFlags = []cli.Flag{
	&cli.StringSliceFlag{
		Name:  "bind",
		Usage: "IP address to listen on, use flag multiple times to specify multiple addresses",
	},
	&cli.StringFlag{
		Name:  "config",
		Usage: "path to LiveKit config file",
	},
	&cli.StringFlag{
		Name:    "config-body",
		Usage:   "LiveKit config in YAML, typically passed in as an environment var in a container",
		Sources: cli.EnvVars("LIVEKIT_CONFIG"),
	},
	&cli.StringFlag{
		Name:  "key-file",
		Usage: "path to file that contains API keys/secrets",
	},
	&cli.StringFlag{
		Name:    "keys",
		Usage:   "api keys (key: secret\n)",
		Sources: cli.EnvVars("LIVEKIT_KEYS"),
	},
	&cli.StringFlag{
		Name:    "region",
		Usage:   "region of the current node. Used by regionaware node selector",
		Sources: cli.EnvVars("LIVEKIT_REGION"),
	},
	&cli.StringFlag{
		Name:    "node-ip",
		Usage:   "IP address of the current node, used to advertise to clients. Automatically determined by default",
		Sources: cli.EnvVars("NODE_IP"),
	},
	&cli.StringFlag{
		Name:    "udp-port",
		Usage:   "UDP port(s) to use for WebRTC traffic",
		Sources: cli.EnvVars("UDP_PORT"),
	},
	&cli.StringFlag{
		Name:    "redis-host",
		Usage:   "host (incl. port) to redis server",
		Sources: cli.EnvVars("REDIS_HOST"),
	},
	&cli.StringFlag{
		Name:    "redis-password",
		Usage:   "password to redis",
		Sources: cli.EnvVars("REDIS_PASSWORD"),
	},
	&cli.StringFlag{
		Name:    "turn-cert",
		Usage:   "tls cert file for TURN server",
		Sources: cli.EnvVars("LIVEKIT_TURN_CERT"),
	},
	&cli.StringFlag{
		Name:    "turn-key",
		Usage:   "tls key file for TURN server",
		Sources: cli.EnvVars("LIVEKIT_TURN_KEY"),
	},
	&cli.StringFlag{
		Name:  "cpuprofile",
		Usage: "write CPU profile to `file`",
	},
	&cli.StringFlag{
		Name:  "memprofile",
		Usage: "write memory profile to `file`",
	},
	&cli.BoolFlag{
		Name:  "dev",
		Usage: "sets log-level to debug, console formatter, and /debug/pprof. insecure for production",
	},
	&cli.BoolFlag{
		Name:   "disable-strict-config",
		Usage:  "disables strict config parsing",
		Hidden: true,
	},
}

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	defer func() {
		if rtc.Recover(logger.GetLogger()) != nil {
			os.Exit(1)
		}
	}()

	generatedFlags, err := config.GenerateCLIFlags(baseFlags, true)
	if err != nil {
		fmt.Println(err)
	}

	cmd := &cli.Command{
		Name:        "livekit-server",
		Usage:       "High performance WebRTC server",
		Description: "run without subcommands to start the server",
		Flags:       append(baseFlags, generatedFlags...),
		Action:      startServer,
		Commands: []*cli.Command{
			{
				Name:   "generate-keys",
				Usage:  "generates an API key and secret pair",
				Action: generateKeys,
			},
			{
				Name:   "ports",
				Usage:  "print ports that server is configured to use",
				Action: printPorts,
			},
			{
				// this subcommand is deprecated, token generation is provided by CLI
				Name:   "create-join-token",
				Hidden: true,
				Usage:  "create a room join token for development use",
				Action: createToken,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "room",
						Usage:    "name of room to join",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "identity",
						Usage:    "identity of participant that holds the token",
						Required: true,
					},
					&cli.BoolFlag{
						Name:     "recorder",
						Usage:    "creates a hidden participant that can only subscribe",
						Required: false,
					},
				},
			},
			{
				Name:   "list-nodes",
				Usage:  "list all nodes",
				Action: listNodes,
			},
			{
				Name:   "help-verbose",
				Usage:  "prints app help, including all generated configuration flags",
				Action: helpVerbose,
			},
		},
		Version: version.Version,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Println(err)
	}
}

func getConfig(c *cli.Command) (*config.Config, error) {
	confString, err := getConfigString(c.String("config"), c.String("config-body"))
	if err != nil {
		return nil, err
	}

	strictMode := true
	if c.Bool("disable-strict-config") {
		strictMode = false
	}

	conf, err := config.NewConfig(confString, strictMode, c, baseFlags)
	if err != nil {
		return nil, err
	}
	config.InitLoggerFromConfig(&conf.Logging)

	if conf.Development {
		logger.Infow("starting in development mode")

		if len(conf.Keys) == 0 {
			logger.Infow("no keys provided, using placeholder keys",
				"API Key", "devkey",
				"API Secret", "secret",
			)
			conf.Keys = map[string]string{
				"devkey": "secret",
			}
			shouldMatchRTCIP := false
			// when dev mode and using shared keys, we'll bind to localhost by default
			if conf.BindAddresses == nil {
				conf.BindAddresses = []string{
					"127.0.0.1",
					"::1",
				}
			} else {
				// if non-loopback addresses are provided, then we'll match RTC IP to bind address
				// our IP discovery ignores loopback addresses
				for _, addr := range conf.BindAddresses {
					ip := net.ParseIP(addr)
					if ip != nil && !ip.IsLoopback() && !ip.IsUnspecified() {
						shouldMatchRTCIP = true
					}
				}
			}
			if shouldMatchRTCIP {
				for _, bindAddr := range conf.BindAddresses {
					conf.RTC.IPs.Includes = append(conf.RTC.IPs.Includes, bindAddr+"/24")
				}
			}
		}
	}
	return conf, nil
}

func startServer(_ context.Context, c *cli.Command) error {
	conf, err := getConfig(c)
	if err != nil {
		return err
	}

	// validate API key length
	if err = conf.ValidateKeys(); err != nil {
		return err
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Try to load from local env file as a fallback for local development
		if err := loadEnvFromFile("config/local.env"); err == nil {
			dbURL = os.Getenv("DATABASE_URL")
		}
	}
	if dbURL == "" {
		return errors.New("DATABASE_URL not set")
	}

	db, err := storage.NewDB(dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	userRepo := storage.NewUserRepository(db)
	authService := appauth.NewService(userRepo)

	issuer := "livekit-local"
	secret := conf.Keys["key1"]
	if secret == "" {
		// fall back to any available secret
		for _, s := range conf.Keys {
			secret = s
			break
		}
	}
	if secret == "" {
		return errors.New("no API key secret configured")
	}
	tokenGenerator := appauth.NewTokenGenerator(issuer, secret)

	authHandler := apphandler.NewAuthHandler(authService, tokenGenerator)
	authMiddleware := apphandler.NewAuthMiddleware(tokenGenerator)

	if cpuProfile := c.String("cpuprofile"); cpuProfile != "" {
		if f, err := os.Create(cpuProfile); err != nil {
			return err
		} else {
			if err := pprof.StartCPUProfile(f); err != nil {
				f.Close()
				return err
			}
			defer func() {
				pprof.StopCPUProfile()
				f.Close()
			}()
		}
	}

	if memProfile := c.String("memprofile"); memProfile != "" {
		if f, err := os.Create(memProfile); err != nil {
			return err
		} else {
			defer func() {
				// run memory profile at termination
				runtime.GC()
				_ = pprof.WriteHeapProfile(f)
				_ = f.Close()
			}()
		}
	}

	currentNode, err := routing.NewLocalNode(conf)
	if err != nil {
		return err
	}

	if err := prometheus.Init(string(currentNode.NodeID()), currentNode.NodeType()); err != nil {
		return err
	}

	server, err := service.InitializeServer(conf, currentNode)
	if err != nil {
		return err
	}

	server.RegisterHTTPHandler("/api/register", http.HandlerFunc(authHandler.Register))
	server.RegisterHTTPHandler("/api/login", http.HandlerFunc(authHandler.Login))
	server.RegisterHTTPHandler("/api/profile", authMiddleware.Authorize(http.HandlerFunc(handleProfile)))

	// Stream Registry API
	server.RegisterHTTPHandler("/api/streaming/list", http.HandlerFunc(handleListStreams))
	server.RegisterHTTPHandler("/api/streaming/register", http.HandlerFunc(handleRegisterStream))
	server.RegisterHTTPHandler("/api/streaming/unregister", http.HandlerFunc(handleUnregisterStream))

	// Serve static files from examples directory
	fs := http.FileServer(http.Dir("examples"))
	server.RegisterHTTPHandler("/examples/", http.StripPrefix("/examples/", fs))
	// Also serve auth.html via /auth/ for convenience
	server.RegisterHTTPHandler("/auth/", http.StripPrefix("/auth/", fs))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		for i := 0; i < 2; i++ {
			sig := <-sigChan
			force := i > 0
			logger.Infow("exit requested, shutting down", "signal", sig, "force", force)
			go server.Stop(force)
		}
	}()

	return server.Start()
}

// --- Stream Registry ---

type StreamInfo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Streamer  string `json:"streamer"`
	Avatar    string `json:"avatar"`
	Viewers   int    `json:"viewers"`
	StartTime int64  `json:"startTime"`
}

var (
	streamMutex   sync.RWMutex
	activeStreams = make(map[string]StreamInfo)
)

func handleListStreams(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	streamMutex.RLock()
	defer streamMutex.RUnlock()

	streams := make([]StreamInfo, 0, len(activeStreams))
	for _, s := range activeStreams {
		streams = append(streams, s)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(streams)
}

func handleRegisterStream(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var info StreamInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if info.ID == "" {
		http.Error(w, "Missing stream ID", http.StatusBadRequest)
		return
	}

	streamMutex.Lock()
	activeStreams[info.ID] = info
	streamMutex.Unlock()

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"registered"}`))
}

func handleUnregisterStream(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	streamMutex.Lock()
	delete(activeStreams, req.ID)
	streamMutex.Unlock()

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"unregistered"}`))
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// --- End Stream Registry ---

// loadEnvFromFile loads KEY=VALUE lines from a .env-like file (no export, no quotes)
func loadEnvFromFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.IndexByte(line, '='); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			_ = os.Setenv(key, val)
		}
	}
	return scanner.Err()
}

func handleProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := apphandler.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp := map[string]string{
		"userId": userID,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func getConfigString(configFile string, inConfigBody string) (string, error) {
	if inConfigBody != "" || configFile == "" {
		return inConfigBody, nil
	}

	outConfigBody, err := os.ReadFile(configFile)
	if err != nil {
		return "", err
	}

	return string(outConfigBody), nil
}
