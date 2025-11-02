package certstream_test

import (
	"log"
	"time"

	"github.com/d-Rickyy-b/certstream-server-go/pkg/certstream"
)

// ExampleNew demonstrates basic usage with default configuration
func ExampleNew() {
	// Create a certstream instance with defaults
	cs := certstream.New()

	// Start consuming certificates
	certChan := cs.Start()

	// Process certificates - this loop runs at YOUR speed
	// CT workers will automatically slow down to match your processing rate
	for cert := range certChan {
		// Your custom processing logic here
		log.Printf("New certificate for domains: %v\n", cert.Data.LeafCert.AllDomains)

		// The CT log workers will wait until you finish processing
		// before sending the next certificate
	}
}

// ExampleNewFromConfigFile demonstrates usage with a config file
func ExampleNewFromConfigFile() {
	// Load configuration from file
	cs, err := certstream.NewFromConfigFile("./config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Start and consume
	certChan := cs.Start()

	for cert := range certChan {
		processCertificate(cert)
	}
}

// ExampleCertStream_EnableRecovery shows how to enable recovery mode
func ExampleCertStream_EnableRecovery() {
	cs := certstream.New()

	// Enable recovery to resume from last position after restart
	cs.EnableRecovery("./ct_index.json")

	// Adjust buffer sizes for your use case
	cs.SetBufferSizes(1000, 5000)

	certChan := cs.Start()

	for cert := range certChan {
		processCertificate(cert)
	}
}

// Example showing slow processing with automatic backpressure
func ExampleCertStream_slowProcessing() {
	cs := certstream.New()
	certChan := cs.Start()

	for cert := range certChan {
		// Slow processing - maybe saving to database
		log.Printf("Processing: %v\n", cert.Data.LeafCert.AllDomains)

		// Simulate slow operation (e.g., database write, API call)
		time.Sleep(1 * time.Second)

		log.Println("Done!")

		// The CT workers will automatically slow down to 1 cert/second
		// to match your processing speed. No certificates are dropped!
	}
}

// Helper function for examples
func processCertificate(cert certstream.Entry) {
	// Your custom logic here
	log.Printf("Domains: %v\n", cert.Data.LeafCert.AllDomains)
}

