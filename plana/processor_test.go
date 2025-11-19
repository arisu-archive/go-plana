package plana_test

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/arisu-archive/go-plana/plana"
)

var _ = Describe("Processor", func() {
	Describe("JSON Serialization", func() {
		var serializer plana.JSONSerializer

		BeforeEach(func() {
			serializer = &plana.DefaultJSONSerializer{}
		})

		It("should serialize basic types", func() {
			data := map[string]any{"key": "value", "number": 42}
			result, err := serializer.Serialize(data, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeEmpty())
		})

		It("should serialize with indentation", func() {
			data := map[string]any{"key": "value"}
			result, err := serializer.Serialize(data, "  ")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeEmpty())
		})

		It("should deserialize JSON", func() {
			original := map[string]any{"key": "value"}
			serialized, err := serializer.Serialize(original, "")
			Expect(err).NotTo(HaveOccurred())

			var deserialized map[string]any
			err = serializer.Deserialize(serialized, &deserialized)
			Expect(err).NotTo(HaveOccurred())
			Expect(deserialized["key"]).To(Equal("value"))
		})

		It("should handle empty data", func() {
			result, err := serializer.Serialize(map[string]any{}, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal([]byte("{}")))
		})
	})

	Describe("Processor.Process", func() {
		var processor *plana.Processor

		BeforeEach(func() {
			processor = &plana.Processor{
				XorKey:         0xD9,
				JSONSerializer: &plana.DefaultJSONSerializer{},
			}
		})

		It("should process without encryption", func() {
			body := map[string]any{"key": "value"}
			key := plana.UserSession{}

			result, err := processor.Process(body, &key, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeEmpty())

			// Verify that the result can be decrypted back to original
			for i := range result {
				result[i] ^= processor.XorKey
			}
			// Decompress it
			gz, err := gzip.NewReader(bytes.NewReader(result[4:])) // Skip checksum
			Expect(err).NotTo(HaveOccurred())

			decompressed, err := io.ReadAll(gz)
			Expect(err).NotTo(HaveOccurred())
			Expect(gz.Close()).To(Succeed())
			var deserialized map[string]any
			err = processor.JSONSerializer.Deserialize(decompressed, &deserialized)
			Expect(err).NotTo(HaveOccurred())
			Expect(deserialized["key"]).To(Equal("value"))
		})

		It("should process with encryption", func() {
			aesKey := []byte{}
			iv := []byte{}

			body := map[string]any{"key": "value"}
			key := plana.UserSession{
				ClientKeyBundle: plana.AESKeyBundle{
					Key: aesKey,
					IV:  iv,
				},
			}

			result, err := processor.Process(body, &key, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeEmpty())

			// Verify that the result can be decrypted back to original
			for i := range result {
				result[i] ^= processor.XorKey
			}
			// Decompress it
			gz, err := gzip.NewReader(bytes.NewReader(result[4:])) // Skip checksum
			Expect(err).NotTo(HaveOccurred())

			decompressed, err := io.ReadAll(gz)
			Expect(err).NotTo(HaveOccurred())
			Expect(gz.Close()).To(Succeed())

			// Decrypt AES
			block, err := aes.NewCipher(aesKey)
			Expect(err).NotTo(HaveOccurred())

			decrypted := make([]byte, len(decompressed))
			mode := cipher.NewCBCDecrypter(block, iv)
			mode.CryptBlocks(decrypted, decompressed)

			// Remove PKCS7 padding
			paddingLen := int(decrypted[len(decrypted)-1])
			decrypted = decrypted[:len(decrypted)-paddingLen]

			var deserialized map[string]any
			err = processor.JSONSerializer.Deserialize(decrypted, &deserialized)
			Expect(err).NotTo(HaveOccurred())
			Expect(deserialized["key"]).To(Equal("value"))
		})

		It("should produce different results with different inputs", func() {
			body1 := map[string]any{"key": "value1"}
			body2 := map[string]any{"key": "value2"}
			key := plana.UserSession{}

			result1, err1 := processor.Process(body1, &key, false)
			Expect(err1).NotTo(HaveOccurred())

			result2, err2 := processor.Process(body2, &key, false)
			Expect(err2).NotTo(HaveOccurred())

			Expect(result1).NotTo(Equal(result2))
		})

		It("should produce consistent results for same input", func() {
			body := map[string]any{"key": "value"}
			key := plana.UserSession{}

			result1, err1 := processor.Process(body, &key, false)
			Expect(err1).NotTo(HaveOccurred())

			result2, err2 := processor.Process(body, &key, false)
			Expect(err2).NotTo(HaveOccurred())

			Expect(result1).To(Equal(result2))
		})

		It("should handle empty request body", func() {
			body := map[string]any{}
			key := plana.UserSession{}

			result, err := processor.Process(body, &key, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeEmpty())
		})
	})

	Describe("Processor.BuildPacket", func() {
		var processor *plana.Processor

		BeforeEach(func() {
			processor = &plana.Processor{}
		})

		It("should build packet with empty payload", func() {
			payload := []byte{}
			protocol := uint32(0x12345678)
			checksum := uint32(0xDEADBEEF)
			key := plana.UserSession{
				ServerKeyBundle: plana.AESKeyBundle{
					Key: []byte{0x01, 0x02},
					IV:  []byte{0x03, 0x04},
				},
			}

			result := processor.BuildPacket(payload, checksum, protocol, &key)
			Expect(len(result)).To(Equal(4 + 4 + 1 + 1 + 2 + 2)) // protocol + keyLen + ivLen + key + iv
		})

		It("should build packet with payload", func() {
			payload := []byte{0x01, 0x02, 0x03}
			checksum := uint32(0xDEADBEEF)
			protocol := uint32(0x12345678)
			key := plana.UserSession{
				ServerKeyBundle: plana.AESKeyBundle{
					Key: []byte{0xAA, 0xBB},
					IV:  []byte{0xCC, 0xDD, 0xEE},
				},
			}

			result := processor.BuildPacket(payload, checksum, protocol, &key)
			Expect(len(result)).To(Equal(4 + 4 + 1 + 1 + 2 + 3 + 3))
			// Check the format: 4 bytes checksum + 4 bytes protocol + lengths + keys + iv + payload
			Expect(result[0:4]).To(Equal([]byte{0xEF, 0xBE, 0xAD, 0xDE}))             // Checksum
			Expect(result[4:8]).To(Equal([]byte{0x78, 0x56, 0x34, 0x12}))             // Protocol
			Expect(result[8:9]).To(Equal([]byte{byte(len(key.ServerKeyBundle.Key))})) // ServerKey length
			Expect(result[9:10]).To(Equal([]byte{byte(len(key.ServerKeyBundle.IV))})) // ServerIV length
			Expect(result[10:12]).To(Equal(key.ServerKeyBundle.Key))                  // ServerKey
			Expect(result[12:15]).To(Equal(key.ServerKeyBundle.IV))                   // ServerIV
			Expect(result[15:]).To(Equal(payload))                                    // Payload
		})

		It("should handle large payloads", func() {
			payload := make([]byte, 10000)
			for i := range payload {
				payload[i] = 'A'
			}
			checksum := uint32(0xDEADBEEF)
			protocol := uint32(0xDEADBEEF)
			key := plana.UserSession{
				ServerKeyBundle: plana.AESKeyBundle{
					Key: []byte{0x01},
					IV:  []byte{0x02},
				},
			}

			result := processor.BuildPacket(payload, checksum, protocol, &key)
			expectedLen := 4 + 4 + 1 + 1 + 1 + 1 + len(payload)
			Expect(len(result)).To(Equal(expectedLen))
		})

		It("should handle different protocol values", func() {
			payload := []byte{}
			checksum := uint32(0)
			key := plana.UserSession{
				ServerKeyBundle: plana.AESKeyBundle{
					Key: []byte{},
					IV:  []byte{},
				},
			}

			tests := []uint32{0x00000000, 0xFFFFFFFF, 0x12345678, 0xAABBCCDD}

			for _, protocol := range tests {
				result := processor.BuildPacket(payload, checksum, protocol, &key)
				Expect(len(result) > 0).To(BeTrue())
			}
		})
	})
})
