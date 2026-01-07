<div align="center">

# 🎮 Plana Client

[![Go Version](https://pkg.go.dev/badge/github.com/arisu-archive/go-plana.svg)](https://pkg.go.dev/github.com/arisu-archive/go-plana) [![License](https://img.shields.io/github/license/arisu-archive/go-plana)](LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/arisu-archive/go-plana)](https://goreportcard.com/report/github.com/arisu-archive/go-plana)

**A Go client library for interacting with Blue Archive game API**

[Features](#features) • [Installation](#installation) • [Quick Start](#quick-start) • [Documentation](#documentation) • [Contributing](#contributing) • [License](#license)

</div>

---

## 📖 Overview

`go-plana` is a robust Go client library for programmatically interacting with the Blue Archive game API. It provides a type-safe, idiomatic Go interface for authentication, game data retrieval, and various game operations including raids, friends, clans, and more.

### ✨ Features

- 🔐 **Authentication & Session Management** - Handle account authentication and session key management
- 🛡️ **Encryption & Security** - Built-in support for RSA and AES encryption with XOR obfuscation
- 🎯 **Comprehensive API Coverage** - Support for multiple game services:
  - Account authentication
  - Raid operations and rankings
  - Friend management and search
  - Clan operations
  - Arena battles
  - Eliminate raid events
  - Queuing systems
- 🔧 **Flexible Configuration** - Customizable URLs, encryption keys, and HTTP clients
- 📦 **Protocol Buffer Support** - Efficient serialization using FlatBuffers and Protocol Buffers
- 🧪 **Well-Tested** - Comprehensive test suite using Ginkgo/Gomega
- 🎨 **Fluent API Design** - Builder patterns for intuitive request construction

## 📦 Installation

```bash
go get github.com/arisu-archive/go-plana
```

### Requirements

- Go 1.24.3 or higher
- RSA public key for encryption (game-specific)
- Protocol encoder service URL

## 🚀 Quick Start

### Basic Setup

```go
package main

import (
    "context"
    "crypto/rsa"
    "net/url"
    
    "github.com/arisu-archive/go-plana/plana"
)

func main() {
    // Parse protocol encoder URL
    encoderURL, _ := url.Parse("https://your-protocol-encoder.example.com")
    
    // Load your RSA public key
    var publicKey *rsa.PublicKey
    // ... load your public key
    
    // Create a new client
    client := plana.NewClient(encoderURL, publicKey, nil)
    
    // Optional: Configure custom URLs
    client.GatewayURL, _ = url.Parse("https://custom-gateway.example.com")
    client.GameURL, _ = url.Parse("https://custom-game.example.com")
}
```

### Authentication

```go
ctx := context.Background()

// Create a session with your credentials
session := &plana.UserSession{
    // Initialize with your session keys
}

// Authenticate
authResponse, err := client.Account.Authenticate(ctx, session)
if err != nil {
    log.Fatalf("Authentication failed: %v", err)
}

fmt.Printf("Logged in as: %s\n", authResponse.AccountNickname)
```

### Making API Requests

```go
// Get raid lobby information
raidLobby, err := client.Raid.Lobby(ctx, session)
if err != nil {
    log.Fatalf("Failed to get raid lobby: %v", err)
}

// Search for friends by friend code
friendResult, err := client.Friend.Search().
    ByCode("ABC123456").
    Execute(ctx, session)
if err != nil {
    log.Fatalf("Friend search failed: %v", err)
}

// Get raid opponents by rank
opponents, err := client.Raid.WithOpponentRank(100).
    Search(ctx, session)
if err != nil {
    log.Fatalf("Failed to get raid opponents: %v", err)
}
```

### Using Request Builder

```go
// Create requests with custom headers
req, err := client.R().
    WithSession(session).
    WithAuthToken("your-token").
    WithHeader("Custom-Header", "value").
    Game(ctx, protos.Protocol_Raid_Lobby, requestBody)
if err != nil {
    log.Fatal(err)
}

// Execute the request
var response protos.RaidLobbyResponse
_, err = client.Do(ctx, req, &response)
```

## 📚 Documentation

### Core Components

#### Client

The `Client` is the main entry point for all API operations:

```go
type Client struct {
    // HTTP client for requests
    client *http.Client
    
    // Encryption configuration
    XorEncryptionKey byte
    publicKey *rsa.PublicKey
    
    // Service URLs
    ProtocolEncoderURL *url.URL
    GetCookieURL *url.URL
    GatewayURL *url.URL
    GameURL *url.URL
    
    // Game services
    Account *AccountService
    Arena *ArenaService
    Clan *ClanService
    Cookie *CookieService
    EliminateRaid *EliminateRaidService
    Friend *FriendService
    Queuing *QueuingService
    Raid *RaidService
}
```

#### Session Management

User sessions maintain encryption keys and request state:

```go
type UserSession struct {
    SessionKey protos.SessionKey
    ClientKeyBundle AESKeyBundle
    ServerKeyBundle AESKeyBundle
    RequestCount int64
}

type AESKeyBundle struct {
    Key []byte
    IV  []byte
}
```

### Available Services

#### Account Service

```go
// Authenticate with game servers
authResp, err := client.Account.Authenticate(ctx, session)

// Check Yostar account
checkResp, err := client.Account.CheckYostar(ctx, plana.YostarCheckOption{
    Cookie: "your-cookie",
    EnterTicket: "your-ticket",
})
```

#### Raid Service

```go
// Get raid lobby
lobby, err := client.Raid.Lobby(ctx, session)

// Search opponents by rank
opponents, err := client.Raid.WithOpponentRank(100).Search(ctx, session)

// Search opponents by score
opponents, err := client.Raid.WithOpponentScore(50000).Search(ctx, session)

// Get best team for an account
team, err := client.Raid.GetBestTeam(ctx, session, accountID)

// Get ranking index
ranking, err := client.Raid.GetRankingIndex(ctx, session)
```

#### Friend Service

```go
// Search friends with builder pattern
result, err := client.Friend.Search().
    ByCode("ABC123456").
    WithLevelOption(flatdata.FriendSearchLevelOptionAll).
    Execute(ctx, session)

// Get friend detailed information
detail, err := client.Friend.GetDetail(ctx, session, friendAccountID)
```

#### Cookie Service

```go
// Get authentication cookie
cookie, err := client.Cookie.GetCookie(ctx, plana.GetCookieOptions{
    UserID: "your-user-id",
    Seed: "your-seed",
    AuthToken: "your-auth-token",
})
```

### Custom JSON Serialization

You can provide your own JSON serializer:

```go
type CustomSerializer struct{}

func (s *CustomSerializer) Serialize(v any, indent string) ([]byte, error) {
    // Your custom serialization logic
}

func (s *CustomSerializer) Deserialize(data []byte, v any) error {
    // Your custom deserialization logic
}

func (s *CustomSerializer) DeserializeReader(r io.Reader, v any) error {
    // Your custom deserialization logic
}

// Use custom serializer
client.JSONSerializer = &CustomSerializer{}
```

### Error Handling

The library provides structured error types:

```go
resp, err := client.Raid.Lobby(ctx, session)
if err != nil {
    // Check for specific error types
    if errors.Is(err, plana.ErrInvalidSession) {
        // Handle invalid session
    }
    
    // Check for API errors
    var apiErr *plana.WebAPIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API Error: %s (code: %d)\n", apiErr.Message(), apiErr.Code())
    }
}
```

## 🧪 Testing

The project uses Ginkgo and Gomega for testing:

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test suite
go test -v ./plana/...
```

## 🏗️ Project Structure

```
go-plana/
├── plana/                    # Main package
│   ├── plana.go             # Core client and types
│   ├── processor.go         # Payload processing and encryption
│   ├── account.go           # Account service
│   ├── raid.go              # Raid service
│   ├── friend.go            # Friend service
│   ├── clan.go              # Clan service
│   ├── arena.go             # Arena service
│   ├── cookie.go            # Cookie service
│   ├── eliminate_raid.go    # Eliminate raid service
│   ├── queuing.go           # Queuing service
│   ├── encoder.go           # Protocol encoding
│   ├── errors.go            # Error types
│   └── *_test.go            # Test files
├── go.mod                   # Go module definition
├── go.sum                   # Dependency checksums
├── LICENSE                  # MIT License
└── README.md               # This file
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/arisu-archive/go-plana.git
   cd go-plana
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Run tests**
   ```bash
   go test ./...
   ```

### Guidelines

- Write tests for new features
- Follow Go best practices and idioms
- Update documentation for API changes
- Ensure all tests pass before submitting PR
- Use meaningful commit messages

## ⚠️ Disclaimer

This library is for educational and research purposes only. Use at your own risk. The authors are not responsible for any misuse or damage caused by this library.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

```
Copyright (c) 2024 Arisu Archive
```

## 🙏 Acknowledgments

- Built for the Blue Archive community
- Uses Protocol Buffers and FlatBuffers for efficient serialization
- Inspired by the need for programmatic game API access

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/arisu-archive/go-plana/issues)
- **Discussions**: [GitHub Discussions](https://github.com/arisu-archive/go-plana/discussions)

---

<div align="center">

**Made with ❤️ by the Arisu Archive team**

[⬆ Back to Top](#-go-plana)

</div>
