package p2p

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/sunboyy/lettered/pkg/common"
	"github.com/sunboyy/lettered/pkg/security"
	"golang.org/x/net/proxy"
)

const (
	// headerPublicKey is a HTTP header key for public key
	headerPublicKey = "X-Public-Key"

	// headerSignature is a HTTP header key for signature
	headerSignature = "X-Signature"
)

// Client is an HTTP client which is specialized for use to request peer
// services. When requesting peer services, specialized HTTP headers and HTTP
// proxy is required to ensure privacy and integrity.
type Client struct {
	// privateKey is used for signing the request. In order to communicate
	// with peers, they must ensure that the request is authentic. Public
	// key which represents identity can be derived from the private key and
	// signature proves that the request comes from the valid source.
	privateKey *btcec.PrivateKey

	// httpClient is a client used for sending HTTP requests. It has
	// specialized transport mechanism so that onion addresses can be
	// requested when Tor proxy is used.
	httpClient *http.Client
}

// NewClient is a constructor function for Client. Proxy URL is extracted from
// P2P configuration and used to setup a specialized HTTP client and private key
// is used to sign requests and will never be transmitted to the internet.
func NewClient(commonConfig common.Config, config Config) (*Client, error) {
	proxyURL, err := url.Parse(config.ProxyURL)
	if err != nil {
		return nil, err
	}

	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{Dial: dialer.Dial}

	return &Client{
		privateKey: commonConfig.PrivateKey,
		httpClient: &http.Client{Transport: transport},
	}, nil
}

// GET is a convenience method for requesting peer services using GET method.
// For the detailed explaination, see request.
func (c *Client) GET(hostname string, path string) (*http.Response, error) {
	return c.request(http.MethodGet, hostname, path, nil)
}

// POST is a convenience method for requesting peer services using POST method.
// For the detailed explaination, see request.
func (c *Client) POST(hostname string, path string, reqBody interface{}) (
	*http.Response, error) {

	return c.request(http.MethodPost, hostname, path, reqBody)
}

// request creates an HTTP request to the peer service. It requires the
// following parameters:
//  - method: Must be a valid uppercased HTTP method (e.g. GET, POST).
//  - hostname: peer service host. If Tor proxy is configured when constructing
//    the client, onion addresses can be used.
//  - path: Endpoint of the requesting resource.
//  - body: Request body to send with the request. nil can be used for GET
//    requests.
//
// When calling request, it converts the body parameter into JSON. The
// JSON-encoded string can be both used for sending as a request body and used
// for signing the request. X-Public-Key and X-Signature headers will be set
// as it is the requirement to request peer services.
func (c *Client) request(method string, hostname string, path string,
	body interface{}) (*http.Response, error) {

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		method,
		fmt.Sprintf("http://%s%s", hostname, path),
		bytes.NewBuffer(bodyJSON),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		headerPublicKey,
		security.MarshalPublicKey(c.privateKey.PubKey()),
	)

	bodyHash := sha256.Sum256(bodyJSON)
	fmt.Println("bodyHash: " + hex.EncodeToString(bodyHash[:]))
	signature := ecdsa.Sign(c.privateKey, bodyHash[:])
	req.Header.Set(
		headerSignature,
		base64.StdEncoding.EncodeToString(signature.Serialize()),
	)

	return c.httpClient.Do(req)
}
