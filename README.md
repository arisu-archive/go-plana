<div align="center">

# 🎮 Plana Client

[![Go Reference](https://pkg.go.dev/badge/github.com/arisu-archive/go-plana.svg)](https://pkg.go.dev/github.com/arisu-archive/go-plana) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![CI](https://github.com/arisu-archive/go-plana/actions/workflows/ci.yml/badge.svg)](https://github.com/arisu-archive/go-plana/actions/workflows/ci.yml) [![Codecov](https://codecov.io/gh/arisu-archive/go-plana/graph/badge.svg)](https://codecov.io/gh/arisu-archive/go-plana)

**A Go client library for interacting with Blue Archive's Yostar game servers**

[Features](#-features) • [Installation](#-installation) • [Quick Start](#-quick-start) • [Documentation](#-documentation) • [Contributing](#-contributing)

</div>

---

## 📖 Overview

Plana Client provides an idiomatic Go interface for communicating with the
Blue Archive game API used by the Yostar/global release. It handles request
metadata, protocol encoding, packet encryption and obfuscation, compression,
session state, HTTP transport, and response decoding so callers can work with
generated request and response types.

### Key Highlights

- 🔐 **Secure Communication**: RSA-OAEP pre-session encryption and AES-CBC
  session encryption with managed key bundles.
- 🌐 **Yostar/Global Support**: Production gateway and game endpoints for the
  Yostar/global release, with configurable base URLs.
- 🛡️ **Type-Safe API**: Generated Protocol Buffer messages and FlatBuffers
  enums for request and response data.
- 🧩 **Modular Design**: Dedicated services for each supported game feature.
- ⚡ **Efficient Transport**: Reusable HTTP transport, gzip compression,
  checksums, and compact multipart packets.
- 🧪 **CI-Verified**: Automated linting, race detection, atomic coverage, and
  Codecov reporting.

## ✨ Features

### Core Services

- **Account Management** - Session authentication and Yostar account validation
- **Arena** - Competitive ranking lists
- **Raid System** - Lobby access, opponent searches, best teams, and rankings
- **Eliminate Raid** - Lobby access, boss-group searches, best teams, and rankings
- **Clan Operations** - Clan search and member lists
- **Friend System** - Friend-code search and detailed player information
- **Queuing System** - Authentication tickets and waiting-queue processing
- **Cookie Service** - Authentication cookies through an optional cookie provider

### Advanced Capabilities

- Custom client-level JSON serialization
- Flexible request builder with custom headers
- Automatic protocol encoding and CRC32 checksum generation
- RSA and session-based AES encryption/decryption
- Gzip compression and XOR obfuscation
- Multipart form-data packet transport
- Request-scoped context cancellation
- Typed web API and invalid-session errors

## 📦 Installation

```bash
go get github.com/arisu-archive/go-plana
```

### Requirements

- Go 1.25 or newer. The module currently selects the Go 1.26.2 toolchain.
- The protocol RSA public key for pre-session requests that require encryption.
- Valid Yostar credentials and session material for authenticated operations.

Dependencies are managed through Go modules.

## 🚀 Quick Start

Create a client with the protocol RSA public key. Passing `nil` as the HTTP
client uses a new `http.Client` with default settings.

```go
client := plana.NewClient(publicKey, nil)

ranks, err := client.Arena.GetRanks(ctx, 1, 20)
if err != nil {
	log.Fatal(err)
}

fmt.Printf("%+v\n", ranks)
```

`publicKey` should be a `*rsa.PublicKey`. The client uses the production
Yostar/global gateway and game endpoints by default.

### Authenticated Requests

Authenticated services accept a `UserSession` containing the server session
key, the two AES key bundles, and the current request counter.

```go
session := &plana.UserSession{
	SessionKey:      sessionKey,
	ClientKeyBundle: plana.AESKeyBundle{Key: clientKey, IV: clientIV},
	ServerKeyBundle: plana.AESKeyBundle{Key: serverKey, IV: serverIV},
	RequestCount:    0,
}

auth, err := client.Account.Authenticate(ctx, session)
if err != nil {
	log.Fatal(err)
}

fmt.Printf("%+v\n", auth)
```

### Raid Operations

```go
opponents, err := client.Raid.
	WithOpponentRank(100).
	Search(ctx, session)
if err != nil {
	log.Fatal(err)
}

lobby, err := client.Raid.Lobby(ctx, session)
if err != nil {
	log.Fatal(err)
}
```

Raid opponents can also be selected by score with `WithOpponentScore`.
Eliminate raids expose equivalent rank and score searches together with a boss
group selector.

### Friend and Clan Search

```go
friends, err := client.Friend.Search().
	ByCode("ABC123456").
	WithLevelOption(flatdata.FriendSearchLevelOptionAll).
	Execute(ctx, session)
if err != nil {
	log.Fatal(err)
}

clans, err := client.Clan.Search().
	ByName("Arisu Archive").
	Execute(ctx, session)
if err != nil {
	log.Fatal(err)
}
```

Clan searches can use either `ByName` or `ByCode`; setting one clears the
other.

## 📚 Documentation

The complete exported API is available on
[pkg.go.dev](https://pkg.go.dev/github.com/arisu-archive/go-plana/plana).

### Available Services

| Service | Supported operations |
|---|---|
| `Account` | Authenticate a session and validate Yostar account data |
| `Arena` | Retrieve arena ranking lists |
| `Clan` | Search by name or code and retrieve clan members |
| `Cookie` | Request an authentication cookie from a configured cookie service |
| `EliminateRaid` | Retrieve lobby data, search rankings, and inspect best teams |
| `Friend` | Search by friend code and retrieve detailed friend information |
| `Queuing` | Obtain an authentication ticket and process the waiting queue |
| `Raid` | Retrieve lobby data, search rankings, and inspect best teams |

### Client Configuration

The client defaults to the production Yostar/global endpoints:

| API | Default URL |
|---|---|
| Gateway | `https://prod-gateway.bluearchiveyostar.com:5100/` |
| Game | `https://prod-game.bluearchiveyostar.com:5000/` |

Both URLs are exported fields and can be replaced when needed:

```go
gatewayURL, err := url.Parse("https://gateway.example.com")
if err != nil {
	log.Fatal(err)
}

gameURL, err := url.Parse("https://game.example.com")
if err != nil {
	log.Fatal(err)
}

client.GatewayURL = gatewayURL
client.GameURL = gameURL
```

To use the optional cookie service, derive a client with `WithCookie`:

```go
cookieURL, err := url.Parse("https://cookies.example.com")
if err != nil {
	log.Fatal(err)
}

client = client.WithCookie(&plana.CookieJarConfig{
	URL:          cookieURL,
	ClientID:     cloudflareAccessClientID,
	ClientSecret: cloudflareAccessClientSecret,
})

cookie, err := client.Cookie.GetCookie(ctx, plana.GetCookieOptions{
	UserID: yostarUserID,
	Seed:   seed,
})
```

`ClientID` and `ClientSecret` are optional. When both are set, the cookie
request includes the corresponding Cloudflare Access headers.

### Custom Requests

Prefer the dedicated services for supported operations. For lower-level calls,
the request builder accepts any payload implementing `RequestPacketReader`:

```go
payload := plana.RaidLobbyRequestWrapper{
	RaidLobbyRequest: &protos.RaidLobbyRequest{},
}

request, err := client.R().
	WithSession(session).
	WithHeader("X-Custom-Header", "value").
	Game(ctx, protos.Protocol_Raid_Lobby, payload)
if err != nil {
	log.Fatal(err)
}

response := new(protos.RaidLobbyResponse)
_, err = client.Do(request, response)
```

### JSON Serialization

`DefaultJSONSerializer` uses Go's `encoding/json` package and is installed when
the client is initialized. The exported `Client.JSONSerializer` field controls
client-level request and response JSON operations, including cookie request
serialization and response-envelope decoding.

The packet processor captures the configured serializer during client
initialization. Reassigning `Client.JSONSerializer` after `NewClient` returns
does not reconfigure that existing processor, so keep the default serializer
unless the distinction is intentional.

### Error Handling

Invalid-session errors wrap the underlying web API error, so callers can check
the specific condition first and still inspect the original packet when useful.

```go
response, err := client.Raid.Lobby(ctx, session)
if err != nil {
	var sessionErr *plana.InvalidSessionError
	var apiErr *plana.ErrWebAPIError

	switch {
	case errors.As(err, &sessionErr):
		log.Printf("invalid session: code=%d", sessionErr.Code())
	case errors.As(err, &apiErr):
		log.Printf("API error: code=%d reason=%s", apiErr.Code(), apiErr.Packet.Reason)
	default:
		log.Printf("request failed: %v", err)
	}
}
```

## 🏗️ Project Structure

```text
go-plana/
|-- plana/
|   |-- plana.go             # Client, request builder, and HTTP transport
|   |-- processor.go         # Compression, checksums, and encryption
|   |-- request_packet.go    # Packet population and RSA handling
|   |-- response.go          # Response envelope
|   |-- errors.go            # Typed API errors
|   |-- account.go           # Account service
|   |-- arena.go             # Arena service
|   |-- clan.go              # Clan service
|   |-- cookie.go            # Cookie service
|   |-- eliminate_raid.go    # Eliminate raid service
|   |-- friend.go            # Friend service
|   |-- queuing.go           # Queuing service
|   |-- raid.go              # Raid service
|   `-- *_test.go            # Ginkgo test suites
|-- go.mod
|-- go.sum
|-- LICENSE
`-- README.md
```

## 🧪 Testing

The project uses Ginkgo and Gomega. Run the same lint and race-enabled test
checks enforced by CI:

```bash
# Run lint checks
golangci-lint run --timeout=5m ./...

# Run tests with the race detector
go test -race ./...

# Run tests with the CI coverage profile
go test -race -covermode=atomic -coverprofile=coverage.out ./...

# Run the Plana package verbosely
go test -race -v ./plana
```

## 🤝 Contributing

Contributions are welcome. For substantial changes, open an issue first so the
approach can be discussed.

1. Clone the repository:

   ```bash
   git clone https://github.com/arisu-archive/go-plana.git
   cd go-plana
   ```

2. Download dependencies:

   ```bash
   go mod download
   ```

3. Make the change and add or update tests.

4. Run the same verification used by CI:

   ```bash
   golangci-lint run --timeout=5m ./...
   go test -race -covermode=atomic -coverprofile=coverage.out ./...
   ```

Please follow idiomatic Go style, document public API changes, and ensure the
lint and test checks pass before opening a pull request. Conventional commit
messages are encouraged.

## ⚠️ Disclaimer

This project is for educational and research purposes only. Use it at your own
risk, respect the game's terms of service, and do not use it to disrupt the
service or other players. The authors are not responsible for misuse or damage
caused by this library.

## 📄 License

Plana Client is available under the [MIT License](LICENSE).

## 🙏 Acknowledgments

- Blue Archive and its game services are operated by their respective owners.
- The project uses Protocol Buffers, FlatBuffers, Ginkgo, and Gomega.
- Thanks to the Arisu Archive contributors and the Blue Archive community.

## 📞 Support

- 🐛 Issues: [GitHub Issues](https://github.com/arisu-archive/go-plana/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/arisu-archive/go-plana/discussions)
- ⭐ Star this repository if you find it useful.

---

<div align="center">

**Made with ❤️ by the Arisu Archive Team**

[⬆ Back to Top](#-plana-client)

</div>
