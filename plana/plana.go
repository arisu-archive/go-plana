package plana

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"mime/multipart"
	"net/http"
	"net/url"
	"sync"

	"github.com/arisu-archive/plana-protos/protos"
)

const (
	Version              = "1.82.390231"
	defaultBundleVersion = "qtmrfsa5k8"
	defaultUserAgent     = "BestHTTP/2 v2.4.0"
	defaultXorKey        = 0xD9
	defaultGatewayURL    = "https://prod-gateway.bluearchiveyostar.com:5100/"
	defaultGameURL       = "https://prod-game.bluearchiveyostar.com:5000/"
)

type Client struct {
	clientMu sync.Mutex
	client   *http.Client

	// XorEncryptionKey is the byte used to XOR the payload before sending.
	XorEncryptionKey byte

	// JSONSerializer is used to serialize and deserialize JSON payloads.
	JSONSerializer JSONSerializer

	// User agent used when communicating with the game API.
	BundleVersion string
	UserAgent     string

	ProtocolEncoderURL   *url.URL // URL of the protocol encoder service.
	ProtocolEncoderToken string   // Token for authenticating protocol encoder requests.

	GetCookieURL   *url.URL // URL for getting cookies.
	GetCookieToken string   // Token for authenticating GetCookie requests.

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
	session  *UserSession
	headers  map[string]string
}

// Request represents an API request.
type Request struct {
	*http.Request
	apiType    apiType
	SessionKey *UserSession
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
	session *UserSession
	headers map[string]string
}

func (c *Client) R() *RequestBuilder {
	return &RequestBuilder{
		client:  c,
		headers: make(map[string]string),
	}
}

func (rb *RequestBuilder) WithSession(session *UserSession) *RequestBuilder {
	rb.session = session
	return rb
}

// WithHeader adds a custom header to the request.
func (rb *RequestBuilder) WithHeader(key, value string) *RequestBuilder {
	rb.headers[key] = value
	return rb
}

// WithHeaders adds multiple custom headers to the request.
func (rb *RequestBuilder) WithHeaders(headers map[string]string) *RequestBuilder {
	maps.Copy(rb.headers, headers)
	return rb
}

func (rb *RequestBuilder) Gateway(
	ctx context.Context,
	protocol protos.Protocol,
	body RequestPacketReader,
	opts ...PacketPopulatorOption,
) (*Request, error) {
	return rb.client.newRequest(ctx, requestParams{
		apiType:  gateway,
		protocol: protocol,
		body:     body,
		session:  rb.session,
		headers:  rb.headers,
	}, opts...)
}

func (rb *RequestBuilder) Game(
	ctx context.Context,
	protocol protos.Protocol,
	body RequestPacketReader,
	opts ...PacketPopulatorOption,
) (*Request, error) {
	return rb.client.newRequest(ctx, requestParams{
		apiType:  game,
		protocol: protocol,
		body:     body,
		session:  rb.session,
		headers:  rb.headers,
	}, opts...)
}

// NewClient returns a new Arona API client. If a nil httpClient is
// provided, a new http.Client will be used.
func NewClient(publicKey *rsa.PublicKey, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	httpClient2 := *httpClient
	c := &Client{
		client:    &httpClient2,
		publicKey: publicKey,
	}
	return c.initialize()
}

func (c *Client) copy() *Client {
	c.clientMu.Lock()
	clone := &Client{
		client:               &http.Client{},
		publicKey:            c.publicKey,
		UserAgent:            c.UserAgent,
		XorEncryptionKey:     c.XorEncryptionKey,
		ProtocolEncoderURL:   c.ProtocolEncoderURL,
		ProtocolEncoderToken: c.ProtocolEncoderToken,
		JSONSerializer:       c.JSONSerializer,
		GetCookieURL:         c.GetCookieURL,
		GetCookieToken:       c.GetCookieToken,
	}
	c.clientMu.Unlock()
	// Shallow copy is sufficient since fields are either value types or pointers
	return clone
}

func (c *Client) WithCookie(jarURL *url.URL, token string) *Client {
	// Copy a new Client to avoid modifying the original
	c2 := c.copy()
	defer c2.initialize()
	c2.GetCookieURL = jarURL
	c2.GetCookieToken = token
	return c2
}

func (c *Client) WithEncoder(encoderURL *url.URL, token string) *Client {
	// Copy a new Client to avoid modifying the original
	c2 := c.copy()
	defer c2.initialize()
	c2.ProtocolEncoderURL = encoderURL
	c2.ProtocolEncoderToken = token
	return c2
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
	if c.BundleVersion == "" {
		c.BundleVersion = defaultBundleVersion
	}
	if c.XorEncryptionKey == 0 {
		c.XorEncryptionKey = defaultXorKey
	}
	if c.JSONSerializer == nil {
		c.JSONSerializer = &DefaultJSONSerializer{}
	}
	c.processor = &Processor{
		PublicKey:      c.publicKey,
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

func (c *Client) Do(ctx context.Context, req *Request, packet any) (*Response, error) {
	resp, err := c.bareDo(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if req.SessionKey != nil {
		req.SessionKey.RequestCount++
	}

	var responseData ResponseData
	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	decErr := c.JSONSerializer.Deserialize(response, &responseData)
	if errors.Is(decErr, io.EOF) {
		decErr = nil // ignore EOF errors caused by empty response body
	}
	if decErr != nil {
		return nil, fmt.Errorf("failed to deserialize response data: %w", decErr)
	}

	// Handle error protocol
	if responseData.Protocol == "Protocol_Error" {
		errPacket, err := c.handleErrorPacket(responseData)
		if err != nil {
			return nil, err
		}
		return nil, c.handleKnownErrorPacket(errPacket)
	}
	if err := c.handleResponsePacket(responseData, packet); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) handleResponsePacket(responseData ResponseData, packet any) error {
	if err := c.JSONSerializer.Deserialize([]byte(responseData.Packet), packet); err != nil {
		return fmt.Errorf("failed to deserialize response packet: %w", err)
	}
	return nil
}

func (*Client) handleKnownErrorPacket(errPacket *protos.ErrorPacket) error {
	err := NewWebAPIError(errPacket)
	switch err.Code() {
	case protos.WebAPIErrorCode_InvalidSession, protos.WebAPIErrorCode_SessionNotFound,
		protos.WebAPIErrorCode_SessionParseFail, protos.WebAPIErrorCode_SessionInvalidInput,
		protos.WebAPIErrorCode_SessionNotAuth, protos.WebAPIErrorCode_SessionDuplicateLogin,
		protos.WebAPIErrorCode_SessionTimeOver, protos.WebAPIErrorCode_SessionInvalidVersion,
		protos.WebAPIErrorCode_SessionChangeDate:
		return NewInvalidSessionError("invalid session", err)
	}
	return err
}

func (c *Client) handleErrorPacket(responseData ResponseData) (*protos.ErrorPacket, error) {
	errorPacket := new(protos.ErrorPacket)
	if err := c.handleResponsePacket(responseData, errorPacket); err != nil {
		return nil, fmt.Errorf("failed to handle error packet: %w", err)
	}
	return errorPacket, nil
}

func (c *Client) bareDo(ctx context.Context, req *Request) (*Response, error) {
	resp, err := c.client.Do(req.WithContext(ctx)) //nolint:bodyclose // response body will be handled by caller
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Determine if gzip uncompression is needed
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzgzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		resp.Body = io.NopCloser(gzgzipReader)
	}

	// Determine if response should be decrypted based on session keys
	// If we have server keys, we expect encrypted content that needs decryption
	if req.SessionKey != nil && len(req.SessionKey.ServerKeyBundle.Key) > 0 && len(req.SessionKey.ServerKeyBundle.IV) > 0 {
		// Decrypt the response
		ciphertext, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("response read failed: %w", err)
		}

		// Decode the response payload from base64 string and decrypt it.
		decodedCiphertext, err := base64.StdEncoding.DecodeString(string(ciphertext))
		if err != nil {
			return nil, fmt.Errorf("response base64 decode failed: %w", err)
		}
		decryptedData, err := decryptPayload(decodedCiphertext, req.SessionKey.ClientKeyBundle.Key, req.SessionKey.ClientKeyBundle.IV)
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
	opts ...PacketPopulatorOption,
) (*Request, error) {
	opts = append(opts, withSessionKey(params.session))
	c.populate(params.body.Packet(), params.protocol, opts...)
	// Process payload through crypto pipeline
	// For now, we assume gateway bypass is false
	payload, err := c.processor.Process(params.body, params.session, false)
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
	req, err := c.buildHTTPRequest(params.apiType, buf, contentType, params.headers)
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
func (c *Client) buildHTTPRequest(apiType apiType, body *bytes.Buffer, contentType string, customHeaders map[string]string) (*http.Request, error) {
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
	req.Header.Set("Bundle-Version", c.BundleVersion)
	req.Header.Set("Accept-Encoding", "identity")

	// Apply custom headers
	for key, value := range customHeaders {
		req.Header.Set(key, value)
	}

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
