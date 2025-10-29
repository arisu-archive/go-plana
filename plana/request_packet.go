package plana

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1" // #nosec G505 -- RSA OAEP with SHA-1 is used by the protocol
	"fmt"
	"reflect"

	"github.com/arisu-archive/plana-protos/protos"
)

type RequestPacketReader interface {
	Packet() *protos.RequestPacket
}

type PacketPopulatorOption func(*protos.RequestPacket)

func WithSessionKey(session UserSession) PacketPopulatorOption {
	return func(packet *protos.RequestPacket) {
		if reflect.ValueOf(session.SessionKey).IsZero() {
			// Skip if session key is empty
			return
		}
		packet.SessionKey = session.SessionKey
		packet.AccountId = session.AccountServerId
		packet.Hash = session.RequestCount | (int64(packet.Protocol) << 32)
	}
}

func WithRequestCount(requestCount int64) PacketPopulatorOption {
	return func(packet *protos.RequestPacket) {
		packet.Hash = requestCount | (int64(packet.Protocol) << 32)
	}
}

func (*Client) populate(packet *protos.RequestPacket, protocol protos.Protocol, opts ...PacketPopulatorOption) {
	packet.Protocol = protocol
	packet.Hash = 1 | (int64(packet.Protocol) << 32)
	for _, opt := range opts {
		opt(packet)
	}
}

func rsaEncrypt(data []byte, publicKey *rsa.PublicKey) []byte {
	// Use RSA OAEP SHA-1 for encryption
	encryptedData, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, publicKey, data, nil) //nolint:gosec // RSA OAEP with SHA-1 is used by the protocol
	if err != nil {
		panic(fmt.Sprintf("RSA encryption failed: %v", err))
	}
	return encryptedData
}
