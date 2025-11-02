# Certstream Library

Use certstream-server-go as a library in your own Go applications. This allows you to consume Certificate Transparency logs directly without needing WebSocket connections.

## Features

✅ **Direct Channel Access** - Consume CT logs via Go channels  
✅ **Automatic Backpressure** - CT workers slow down to match your processing speed  
✅ **Zero Data Loss** - No certificates are dropped if you're slow  
✅ **Type Safety** - Work with native Go types, no JSON parsing needed  
✅ **Recovery Support** - Resume from last position after restart  
✅ **Simple API** - Just read from a channel  

## Installation

```bash
go get github.com/d-Rickyy-b/certstream-server-go
```

## Quick Start

```go
package main

import (
    "log"
    "github.com/d-Rickyy-b/certstream-server-go/pkg/certstream"
)

func main() {
    // Create certstream instance
    cs := certstream.New()
    
    // Start consuming certificates
    certChan := cs.Start()
    
    // Process certificates at your own pace
    for cert := range certChan {
        log.Printf("New cert for: %v\n", cert.Data.LeafCert.AllDomains)
        
        // Your processing logic here
        // The CT workers will wait for you to finish!
    }
}
```

## Usage Examples

### Basic Usage

```go
cs := certstream.New()
certChan := cs.Start()

for cert := range certChan {
    processCertificate(cert)
}
```

### With Config File

```go
cs, err := certstream.NewFromConfigFile("./config.yaml")
if err != nil {
    log.Fatal(err)
}

certChan := cs.Start()
for cert := range certChan {
    processCertificate(cert)
}
```

### With Recovery (Resume from last position)

```go
cs := certstream.New()
cs.EnableRecovery("./ct_index.json")

certChan := cs.Start()
for cert := range certChan {
    processCertificate(cert)
}
```

### Custom Buffer Sizes

```go
cs := certstream.New()

// Set buffer sizes for CT log fetching and processing
cs.SetBufferSizes(
    1000,  // CT log buffer
    5000,  // Broadcast buffer
)

certChan := cs.Start()
for cert := range certChan {
    processCertificate(cert)
}
```

### Slow Processing with Backpressure

```go
cs := certstream.New()
certChan := cs.Start()

for cert := range certChan {
    log.Printf("Processing: %v\n", cert.Data.LeafCert.AllDomains)
    
    // Slow operation (database write, API call, etc.)
    saveToDatabase(cert)
    time.Sleep(1 * time.Second)
    
    // The CT workers automatically slow down to match your speed!
    // No certificates are dropped!
}
```

## How Backpressure Works

When you process certificates slowly, the library automatically slows down the CT log workers:

```
Your Processing Speed → Controls → Channel Buffer → Controls → CT Workers
```

**Example:**
- You process at 1 cert/second (slow database writes)
- Buffer fills with 5000 certs (~5 seconds)
- CT workers block and slow down to 1 cert/second
- **Zero certificates dropped!**

## Certificate Structure

The `Entry` type contains all certificate information:

```go
type Entry struct {
    Data struct {
        LeafCert struct {
            AllDomains []string  // All domains in the certificate
            Subject    Subject   // Certificate subject
            Issuer     Issuer    // Certificate issuer
            NotBefore  int64     // Valid from timestamp
            NotAfter   int64     // Valid until timestamp
            // ... more fields
        }
        CertIndex  uint64     // Index in CT log
        CertLink   string     // Link to view certificate
        Source     struct {
            Name string        // CT log name
            URL  string        // CT log URL
        }
        UpdateType string     // "X509LogEntry" or "PrecertLogEntry"
    }
    MessageType string        // "certificate_update"
}
```

## Configuration

### Using Config File

Create a `config.yaml` file:

```yaml
general:
  buffer_sizes:
    ctlog: 1000
    broadcastmanager: 5000
  
  recovery:
    enabled: true
    ct_index_file: "./ct_index.json"
```

Then load it:

```go
cs, err := certstream.NewFromConfigFile("./config.yaml")
```

### Programmatic Configuration

```go
cs := certstream.New()
cs.EnableRecovery("./ct_index.json")
cs.SetBufferSizes(1000, 5000)
```

## Complete Example

See the [complete example](../../examples/library-consumer/main.go) for a full working application.

Run it:

```bash
cd examples/library-consumer
go run main.go
```

With options:

```bash
# With config file
go run main.go -config ../../config.yaml

# With recovery
go run main.go -recovery ./ct_index.json

# With slow processing (to see backpressure in action)
go run main.go -slow
```

## Use Cases

- **Phishing Detection** - Monitor for suspicious domain names
- **Brand Protection** - Alert on certificates for your domains
- **Security Research** - Analyze certificate issuance patterns
- **Compliance Monitoring** - Track certificate usage in your organization
- **Threat Intelligence** - Build domain/certificate databases

## Performance

- **Buffer overhead**: Minimal (~10-50MB RAM for typical buffers)
- **CPU usage**: Depends on your processing logic
- **Network**: ~14.5 Mbit/s for real-time CT log consumption
- **Processing rate**: Unlimited - system matches YOUR speed

## Advantages Over WebSocket

| Feature | WebSocket | Library |
|---------|-----------|---------|
| **Overhead** | High (JSON, TCP, protocol) | None (in-memory) |
| **Speed** | ~10-20% slower | Full speed |
| **Type Safety** | JSON parsing required | Native Go types |
| **Backpressure** | Manual (ACKs) | Automatic |
| **Setup** | Server + Client | Just import |

## FAQ

### Does this drop certificates if I'm slow?

**No!** Unlike the WebSocket server, the library version never drops certificates. It uses blocking channel sends, so CT workers automatically wait for you to finish processing.

### How fast can I process certificates?

As fast as you want! The system will match your speed:
- Fast processor: CT workers run at full speed
- Slow processor: CT workers slow down to match
- Zero certificates lost either way

### What if I want to process 1.5 years of historical data?

Enable recovery mode and manually set the starting indices in `ct_index.json`. The system will process at whatever speed you can handle.

### Can I use this in production?

Yes! This library wraps the same battle-tested CT log processing code used by the WebSocket server.

## License

Same as certstream-server-go (see root LICENSE file)

