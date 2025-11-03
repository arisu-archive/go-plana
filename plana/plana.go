package plana

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/arisu-archive/plana-protos/protos"
)

const (
	Version           = "1.82.378581"
	defaultUserAgent  = "BestHTTP/2 v2.4.0"
	defaultXorKey     = 0xD9
	defaultGatewayURL = "https://prod-gateway.bluearchiveyostar.com:5100/"
	defaultGameURL    = "https://prod-game.bluearchiveyostar.com:5000/"
)

type Client struct {
	client *http.Client

	// XorEncryptionKey is the byte used to XOR the payload before sending.
	XorEncryptionKey byte

	// JSONSerializer is used to serialize and deserialize JSON payloads.
	JSONSerializer JSONSerializer

	// User agent used when communicating with the game API.
	UserAgent string

	ProtocolEncoderURL *url.URL // URL of the protocol encoder service.
	GetCookieURL       *url.URL // URL for getting cookies.

	// PublicKey is the RSA public key used for encrypting sensitive data.
	publicKey *rsa.PublicKey

	GatewayURL *url.URL // Base URL for gateway requests. Defaults based on the provided server variable.
	GameURL    *url.URL // Base URL for game requests. Defaults based on the provided server variable.

	processor *Processor // Packet processor for handling encryption and payload building.

	common service // Reuse a single struct instead of allocating one for each service on the heap.
	// Services used for talking to different parts of the game API.
	Account       *AccountService
	Arena         *ArenaService
	Clan          *ClanService
	Cookie        *CookieService
	EliminateRaid *EliminateRaidService
	Friend        *FriendService
	Queuing       *QueuingService
	Raid          *RaidService
}

// service represents a service for interacting with a specific part of the game API.
type service struct {
	client *Client
}

// UserSession holds keys and IVs used for encrypting and forging packets.
type UserSession struct {
	protos.SessionKey // Session key information
	ClientKeyBundle   AESKeyBundle
	ServerKeyBundle   AESKeyBundle
	RequestCount      int64
}

// apiType represents the type of API being accessed.
type apiType int

const (
	gateway apiType = iota
	game
)

// requestParams groups arguments for newRequest so we don't exceed argument limits.
type requestParams struct {
	apiType  apiType
	protocol protos.Protocol
	body     RequestPacketReader
	session  UserSession
}

// Request represents an API request.
type Request struct {
	*http.Request
	apiType    apiType
	SessionKey UserSession
}

// Response represents an API response.
type Response struct {
	*http.Response
}

// JSONSerializer defines methods for serializing and deserializing JSON data.
type JSONSerializer interface {
	Serialize(v any, indent string) ([]byte, error)
	Deserialize(data []byte, v any) error
	DeserializeReader(r io.Reader, v any) error
}

// StdJSONSerializer is the default implementation of JSONSerializer using the encoding/json package.
type DefaultJSONSerializer struct{}

// Serialize serializes a value into JSON.
func (*DefaultJSONSerializer) Serialize(v any, indent string) ([]byte, error) {
	if indent != "" {
		return json.MarshalIndent(v, "", indent) //nolint:wrapcheck // no need to wrap
	}
	return json.Marshal(v) //nolint:wrapcheck // no need to wrap
}

// Deserialize deserializes JSON data into a value.
func (*DefaultJSONSerializer) Deserialize(data []byte, v any) error {
	return json.Unmarshal(data, v) //nolint:wrapcheck // no need to wrap
}

func (*DefaultJSONSerializer) DeserializeReader(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v) //nolint:wrapcheck // no need to wrap
}

type RequestBuilder struct {
	client  *Client
	session UserSession
}

func (c *Client) R() *RequestBuilder {
	return &RequestBuilder{
		client: c,
	}
}

func (rb *RequestBuilder) WithSession(session UserSession) *RequestBuilder {
	rb.session = session
	return rb
}

func (rb *RequestBuilder) Gateway(ctx context.Context, protocol protos.Protocol, body RequestPacketReader) (*Request, error) {
	return rb.client.newRequest(ctx, requestParams{
		apiType:  gateway,
		protocol: protocol,
		body:     body,
		session:  rb.session,
	})
}

func (rb *RequestBuilder) Game(ctx context.Context, protocol protos.Protocol, body RequestPacketReader) (*Request, error) {
	return rb.client.newRequest(ctx, requestParams{
		apiType:  game,
		protocol: protocol,
		body:     body,
		session:  rb.session,
	})
}

// NewClient returns a new Arona API client. If a nil httpClient is
// provided, a new http.Client will be used.
func NewClient(protocolEncoderURL *url.URL, publicKey *rsa.PublicKey, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	httpClient2 := *httpClient
	c := &Client{
		client:             &httpClient2,
		ProtocolEncoderURL: protocolEncoderURL,
		publicKey:          publicKey,
	}
	return c.initialize()
}

// initialize sets up the client with default values.
func (c *Client) initialize() *Client {
	// Set default URLs based on the server
	if c.GatewayURL == nil {
		c.GatewayURL, _ = url.Parse(defaultGatewayURL)
	}
	if c.GameURL == nil {
		c.GameURL, _ = url.Parse(defaultGameURL)
	}
	if c.UserAgent == "" {
		c.UserAgent = defaultUserAgent
	}
	if c.XorEncryptionKey == 0 {
		c.XorEncryptionKey = defaultXorKey
	}
	if c.JSONSerializer == nil {
		c.JSONSerializer = &DefaultJSONSerializer{}
	}
	c.processor = &Processor{
		XorKey:         c.XorEncryptionKey,
		JSONSerializer: c.JSONSerializer,
	}
	c.common.client = c
	c.Account = (*AccountService)(&c.common)
	c.Arena = (*ArenaService)(&c.common)
	c.Clan = (*ClanService)(&c.common)
	c.Cookie = (*CookieService)(&c.common)
	c.EliminateRaid = (*EliminateRaidService)(&c.common)
	c.Friend = (*FriendService)(&c.common)
	c.Queuing = (*QueuingService)(&c.common)
	c.Raid = (*RaidService)(&c.common)
	return c
}

func (c *Client) Do(ctx context.Context, req *Request, v any) (*Response, error) {
	resp, err := c.bareDo(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch v := v.(type) {
	case nil:
	case io.Writer:
		_, err = io.Copy(v, resp.Body)
	default:
		decErr := c.JSONSerializer.DeserializeReader(resp.Body, v)
		if errors.Is(decErr, io.EOF) {
			decErr = nil // ignore EOF errors caused by empty response body
		}
		if decErr != nil {
			err = decErr
		}
	}
	return resp, err
}

func (c *Client) bareDo(ctx context.Context, req *Request) (*Response, error) {
	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	// Determine if response should be decrypted based on session keys
	// If we have server keys, we expect encrypted content that needs decryption
	if len(req.SessionKey.ServerKeyBundle.Key) > 0 && len(req.SessionKey.ServerKeyBundle.IV) > 0 {
		// Decrypt the response
		ciphertext, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("response read failed: %w", err)
		}
		decryptedData, err := decryptPayload(ciphertext, req.SessionKey.ClientKeyBundle.Key, req.SessionKey.ClientKeyBundle.IV)
		if err != nil {
			return nil, fmt.Errorf("response decryption failed: %w", err)
		}
		// Replace response body with decrypted data
		resp.Body = io.NopCloser(bytes.NewReader(decryptedData))
	}
	// If no server keys, treat response as plain JSON (no decryption needed)
	return &Response{Response: resp}, nil
}

// newRequest creates a new API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash.
func (c *Client) newRequest(
	ctx context.Context,
	params requestParams,
) (*Request, error) {
	c.populate(params.body.Packet(), params.protocol, WithSessionKey(params.session))
	// Process payload through crypto pipeline
	payload, err := c.processor.Process(params.body, params.session)
	if err != nil {
		return nil, err
	}
	// Encode protocol with checksum
	checksum := computeHash(payload, 0)
	encodedProtocol, err := c.encodeProtocol(ctx, checksum, params.protocol)
	if err != nil {
		return nil, fmt.Errorf("protocol encoding failed: %w", err)
	}
	// Build final packet
	packetData := c.processor.BuildPacket(payload, checksum, encodedProtocol, params.session)
	// Create multipart form
	mw := &multipartWriter{}
	buf, contentType, err := mw.write(packetData)
	if err != nil {
		return nil, fmt.Errorf("multipart creation failed: %w", err)
	}

	// Build HTTP request
	req, err := c.buildHTTPRequest(params.apiType, buf, contentType)
	if err != nil {
		return nil, err
	}

	return &Request{
		Request:    req,
		apiType:    params.apiType,
		SessionKey: params.session,
	}, nil
}

// buildHTTPRequest creates the HTTP request with proper headers.
func (c *Client) buildHTTPRequest(apiType apiType, body *bytes.Buffer, contentType string) (*http.Request, error) {
	u, err := c.getBaseURL(apiType).Parse("/api/gateway")
	if err != nil {
		return nil, fmt.Errorf("URL parse failed: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), body) //nolint:noctx // context will be added in Do method
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("mx", "2") //nolint:canonicalheader // required by API
	req.Header.Set("Accept-Encoding", "identity")

	return req, nil
}

func (c *Client) getBaseURL(apiType apiType) *url.URL {
	if apiType == gateway {
		return c.GatewayURL
	}
	return c.GameURL
}

type multipartWriter struct{}

func (*multipartWriter) write(packetData []byte) (*bytes.Buffer, string, error) {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	writer.SetBoundary(fmt.Sprintf("BestHTTP_HTTPMultiPartForm_%s", randomBoundary())) //nolint:errcheck // cannot fail

	part, err := writer.CreateFormFile("mx", "mx.dat")
	if err != nil {
		return nil, "", fmt.Errorf("form file creation failed: %w", err)
	}
	if _, err := part.Write(packetData); err != nil {
		return nil, "", fmt.Errorf("form file write failed: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("multipart writer close failed: %w", err)
	}
	return buf, writer.FormDataContentType(), nil
}

// randomBoundary generates a random boundary string for multipart form data.
func randomBoundary() string {
	var buf [4]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%X", buf[:])
}
