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
	err := signer.SignRequest(req, time.Now().Add(1*time.Second))
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
			expectedSignature: "OWPlFpd22vMoUOBNjZALBtGpSWN1t40+yVCDQe502eo=",
		},

		{
			name:       "different signing key",
			signingKey: "test2",
			req: newReqBuilder("GET", "http://example.com").
				withHeaders(map[string]string{
					headerKeySigExpiredTime: "2017-01-01T00:00:00Z",
				}).build(),
			expectedSignature: "P1Xo+GDOaOushtZiuc8LEIFSaGxsdR3we2hYYgpDLLI=",
		},

		{
			name:       "different signing time",
			signingKey: "test",
			req: newReqBuilder("GET", "http://example.com").
				withHeaders(map[string]string{
					headerKeySigExpiredTime: "2017-01-01T00:00:01Z",
				}).build(),
			expectedSignature: "dw3Wi9ZWx2OYz9tkUr25suKL9QtAp594LZnCnfZ1JsE=",
		},

		{
			name:       "sign request with query string",
			signingKey: "test",
			req: newReqBuilder("GET", "http://example.com?a=1").
				withHeaders(map[string]string{
					headerKeySigExpiredTime: "2017-01-01T00:00:00Z",
				}).build(),
			expectedSignature: "v6R49ovVcMB5EH+EbGRGf7H9ceKIWOp/WS12QQbBYp4=",
		},

		{
			name:       "sign POST request",
			signingKey: "test",
			req: newReqBuilder("POST", "http://example.com").
				withHeaders(map[string]string{
					headerKeySigExpiredTime: "2017-01-01T00:00:00Z",
				}).build(),
			expectedSignature: "UY5VxO5habdzHfcKYp+6p+QN3A73flQKk1iqe+VnJFw=",
		},

		{
			name:       "sign POST request with body",
			signingKey: "test",
			req: newReqBuilder("POST", "http://example.com").
				withHeaders(map[string]string{
					headerKeySigExpiredTime: "2017-01-01T00:00:00Z",
				}).withBody("test body").build(),
			expectedSignature: "gtI59p0Mf0bsFLSfZOYAT7XpnQCN78VoKnMfUCbRX5s=",
		},

		{
			name:       "sign POST request with different body",
			signingKey: "test",
			req: newReqBuilder("POST", "http://example.com").
				withHeaders(map[string]string{
					headerKeySigExpiredTime: "2017-01-01T00:00:00Z",
				}).withBody("test body 2").build(),
			expectedSignature: "TF/ndCBoH7QpAsIB+tV3D/dMGkc92ugMaczOszckTsM=",
		},
	}

	for _, test := range tests {
		t.Logf("Running test: [%s]", test.name)
		signature, err := calcSignature(test.signingKey, test.req, test.headerKeysToSign...)
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
