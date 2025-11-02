package certstream

// Package certstream provides a library interface for consuming Certificate Transparency logs
// directly in Go code without needing WebSocket connections.

import (
	"github.com/letrics/certstream-server-go/pkg/config"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/letrics/certstream-server-go/internal/certificatetransparency"
	"github.com/letrics/certstream-server-go/internal/models"
)

// CertStream is a library interface for consuming CT logs directly
type CertStream struct {
	watcher  *certificatetransparency.Watcher
	certChan chan models.Entry
	config   config.Config
	doneChan chan struct{}
}

// Entry re-exports the internal Entry type for public use
type Entry = models.Entry

// NewFromConfig creates a certstream library instance with the provided config
func NewFromConfig(conf config.Config) *CertStream {
	certChan := make(chan models.Entry, conf.General.BufferSizes.BroadcastManager)

	return &CertStream{
		certChan: certChan,
		config:   conf,
		doneChan: make(chan struct{}),
	}
}

// NewFromConfigFile creates a certstream library instance from a config file
func NewFromConfigFile(configPath string) (*CertStream, error) {
	conf, err := config.ReadConfig(configPath)
	if err != nil {
		return nil, err
	}

	return NewFromConfig(conf), nil
}

// New creates a certstream library instance with default configuration
func New() *CertStream {
	conf := config.Config{}

	// Set reasonable defaults
	conf.General.BufferSizes.CTLog = 1000
	conf.General.BufferSizes.BroadcastManager = 5000
	conf.General.Recovery.Enabled = false

	// Set default CT log fetcher options
	conf.General.CTLogFetcher.BatchSize = 100
	conf.General.CTLogFetcher.ParallelFetch = 1
	conf.General.CTLogFetcher.NumWorkers = 1
	conf.General.CTLogFetcher.HTTPTimeout = 30

	dropOldLogs := true
	conf.General.DropOldLogs = &dropOldLogs

	return NewFromConfig(conf)
}

// Start begins consuming CT logs. Returns a read-only channel you can consume from.
// This is non-blocking - the watcher runs in the background.
//
// Usage:
//
//	cs := certstream.New()
//	certChan := cs.Start()
//	for cert := range certChan {
//	    processCertificate(cert)
//	}
func (cs *CertStream) Start() <-chan Entry {
	log.Printf("Starting certstream library v%s\n", config.Version)

	// Handle signals for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signals
		log.Printf("Received signal %v. Shutting down...\n", sig)
		cs.Stop()
	}()

	// Apply effective config globally so the watcher uses these values
	config.AppConfig = cs.config

	// Create and start watcher
	cs.watcher = certificatetransparency.NewWatcher(cs.certChan)

	// Start watcher in background and signal completion
	go func() {
		cs.watcher.Start()
		close(cs.doneChan)
	}()

	return cs.certChan
}

// Stop gracefully stops the certstream and closes the certificate channel
func (cs *CertStream) Stop() {
	log.Println("Stopping certstream library...")
	if cs.watcher != nil {
		cs.watcher.Stop()
	}
}

// Wait blocks until the certstream is stopped
func (cs *CertStream) Wait() {
	<-cs.doneChan
}

// EnableRecovery enables the recovery feature which allows resuming from the last processed certificate
func (cs *CertStream) EnableRecovery(indexFilePath string) {
	cs.config.General.Recovery.Enabled = true
	cs.config.General.Recovery.CTIndexFile = indexFilePath
}

// SetBufferSizes configures the buffer sizes for the CT log fetching and certificate processing
func (cs *CertStream) SetBufferSizes(ctLogBuffer, broadcastBuffer int) {
	cs.config.General.BufferSizes.CTLog = ctLogBuffer
	cs.config.General.BufferSizes.BroadcastManager = broadcastBuffer
}

// SetCTLogFetcherOptions configures the CT log fetcher performance parameters.
// These control how fast certificates are downloaded from CT logs.
//
// Parameters:
//   - batchSize: Number of certificates to fetch per request (recommended: 500-1000 for high throughput)
//   - parallelFetch: Number of parallel fetches per CT log (recommended: 4-8 for high throughput)
//   - numWorkers: Number of workers processing entries per CT log (recommended: 2-4 for high throughput)
//   - httpTimeout: HTTP timeout in seconds for CT log requests (increase if using large batch sizes)
func (cs *CertStream) SetCTLogFetcherOptions(batchSize, parallelFetch, numWorkers, httpTimeout int) {
	cs.config.General.CTLogFetcher.BatchSize = batchSize
	cs.config.General.CTLogFetcher.ParallelFetch = parallelFetch
	cs.config.General.CTLogFetcher.NumWorkers = numWorkers
	cs.config.General.CTLogFetcher.HTTPTimeout = httpTimeout
}
