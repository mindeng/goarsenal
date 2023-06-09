package aweb

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"
)

// TestSignRequest tests the SignRequest function.
func TestSignRequest(t *testing.T) {
	t.Parallel()

	signingKey := "test"
	req := newReqBuilder("GET", "http://example.com").build()
	signer := NewSigner(signingKey)
	err := signer.SignRequest(req, time.Now().Add(1*time.Second), "")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	signTime := req.Header.Get(headerKeySigExpiredTime)
	if signTime == "" {
		t.Errorf("Expected %s header to be set", headerKeySigExpiredTime)
	}

	signature := req.Header.Get(headerKeySignature)
	if signature == "" {
		t.Errorf("Expected %s header to be set", headerKeySignature)
	}

	err = signer.VerifyRequest(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// sleep for 1.1 seconds to make sure the signature has expired
	time.Sleep(1100 * time.Millisecond)
	err = signer.VerifyRequest(req)
	if err != ErrSigExpired {
		t.Errorf("Expected error %v, got %v", ErrSigExpired, err)
	}
}

// TestVerifyRequest tests the VerifyRequest function.
func TestVerifyRequest(t *testing.T) {
	t.Parallel()
	// create a signer with a signing key
	signer := NewSigner("test")

	// create a request with a valid signature
	req := newReqBuilder("GET", "http://example.com").build()
	err := signer.SignRequest(req, time.Now().Add(1*time.Second), "")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	// verify the request
	err = signer.VerifyRequest(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// create a request with an invalid signature
	req = newReqBuilder("GET", "http://example.com").build()
	req.Header.Set(headerKeySignature, "invalid")
	req.Header.Set(headerKeySigExpiredTime, time.Now().Add(1*time.Second).Format(time.RFC3339))
	// verify the request
	err = signer.VerifyRequest(req)
	if err != ErrInvalidSignature {
		t.Errorf("Expected error %v, got %v", ErrInvalidSignature, err)
	}

	// create a request withouth a signature
	req = newReqBuilder("GET", "http://example.com").build()
	// verify the request
	err = signer.VerifyRequest(req)
	if err != ErrNoSignature {
		t.Errorf("Expected error %v, got %v", ErrNoSignature, err)
	}

	// create a request withouth a signature expired time
	req = newReqBuilder("GET", "http://example.com").build()
	req.Header.Set(headerKeySignature, "invalid")
	// verify the request
	err = signer.VerifyRequest(req)
	if err != ErrNoSigExpiredTime {
		t.Errorf("Expected error %v, got %v", ErrNoSigExpiredTime, err)
	}
}

func TestCalcSignature(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		signingKey        string
		req               *http.Request
		headerKeysToSign  []string
		expectedSignature string
		expectedErr       error
	}{

		{
			name:        "no signing key",
			signingKey:  "test",
			req:         newReqBuilder("GET", "http://example.com").build(),
			expectedErr: ErrNoSigExpiredTime,
		},

		{
			name:       "sign GET request",
			signingKey: "test",
			req: newReqBuilder("GET", "http://example.com").
				withHeaders(map[string]string{
					headerKeySigExpiredTime: "2017-01-01T00:00:00Z",
				}).build(),
			expectedSignature: "SQ43bXZFCGmtxTZQ21JPJv0chHp0M9EnRI4YUrtGLFA=",
		},

		{
			name:       "different signing key",
			signingKey: "test2",
			req: newReqBuilder("GET", "http://example.com").
				withHeaders(map[string]string{
					headerKeySigExpiredTime: "2017-01-01T00:00:00Z",
				}).build(),
			expectedSignature: "YkEORS9zKzXN6uJPbAk/5WI+/ASYijMAbbeHEXtnpIA=",
		},

		{
			name:       "different signing time",
			signingKey: "test",
			req: newReqBuilder("GET", "http://example.com").
				withHeaders(map[string]string{
					headerKeySigExpiredTime: "2017-01-01T00:00:01Z",
				}).build(),
			expectedSignature: "JboiDEACIbS4L8Eh3D3au5/7p8s0ohTf2v7zBnwhgNQ=",
		},

		{
			name:       "sign request with query string",
			signingKey: "test",
			req: newReqBuilder("GET", "http://example.com?a=1").
				withHeaders(map[string]string{
					headerKeySigExpiredTime: "2017-01-01T00:00:00Z",
				}).build(),
			expectedSignature: "Zp6peh1pkk6VOPV1su87YR5cEVg9m4VUxhWcC+6qYa8=",
		},

		{
			name:       "sign POST request",
			signingKey: "test",
			req: newReqBuilder("POST", "http://example.com").
				withHeaders(map[string]string{
					headerKeySigExpiredTime: "2017-01-01T00:00:00Z",
				}).build(),
			expectedSignature: "wiqZ0JsPlbAzedvP8sYpX0oTzIzXHTv2DntkafcJhd4=",
		},

		{
			name:       "sign POST request with body",
			signingKey: "test",
			req: newReqBuilder("POST", "http://example.com").
				withHeaders(map[string]string{
					headerKeySigExpiredTime: "2017-01-01T00:00:00Z",
				}).withBody("test body").build(),
			expectedSignature: "ENz9ZQxSUwlCR4euqhIpSWJYWozS+PgDniFk+Z7hU2k=",
		},

		{
			name:       "sign POST request with different body",
			signingKey: "test",
			req: newReqBuilder("POST", "http://example.com").
				withHeaders(map[string]string{
					headerKeySigExpiredTime: "2017-01-01T00:00:00Z",
				}).withBody("test body 2").build(),
			expectedSignature: "PeSp+m8MxaQpAcWtGtj8tGYrdzVqhrt3NnNW3dca7ac=",
		},
	}

	for _, test := range tests {
		t.Logf("Running test: [%s]", test.name)
		signature, err := calcSignature(test.signingKey, test.req, "", test.headerKeysToSign...)
		if err != test.expectedErr {
			t.Errorf("[%s] Expected error %v, got %v", test.name, test.expectedErr, err)
		}

		if signature != test.expectedSignature {
			t.Errorf("[%s] Expected signature %s, got %s", test.name, test.expectedSignature, signature)
		}
	}
}

type reqBuilder struct {
	req http.Request
}

// newReqBuilder returns a new reqBuilder.
func newReqBuilder(method, urlStr string) *reqBuilder {
	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		panic(err)
	}
	return &reqBuilder{req: *req}
}

// build returns the built http.Request.
func (rb *reqBuilder) build() *http.Request {
	return &rb.req
}

// withHeaders returns a new http.Request with the given headers.
func (rb *reqBuilder) withHeaders(headers map[string]string) *reqBuilder {
	for k, v := range headers {
		rb.req.Header.Set(k, v)
	}
	return rb
}

// withBody returns a new http.Request with the given body.
func (rb *reqBuilder) withBody(body string) *reqBuilder {
	rb.req.Body = io.NopCloser(bytes.NewBuffer([]byte(body)))
	return rb
}
