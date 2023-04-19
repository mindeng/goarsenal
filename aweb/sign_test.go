package aweb

import (
	"net/http"
	"testing"
)

func TestCalcSignature(t *testing.T) {
	tests := []struct {
		signingKey        string
		req               *http.Request
		headerKeysToSign  []string
		expectedSignature string
		expectedErr       error
	}{

		{
			signingKey:  "test",
			req:         newReqBuilder("GET", "http://example.com").build(),
			expectedErr: ErrNoSignTime,
		},
		{
			signingKey: "test",
			req: newReqBuilder("GET", "http://example.com").
				withHeaders(map[string]string{
					"X-Sign-Time": "2017-01-01T00:00:00Z",
				}).build(),
			expectedSignature: "OWPlFpd22vMoUOBNjZALBtGpSWN1t40+yVCDQe502eo=",
		},
	}

	for _, test := range tests {
		signature, err := calcSignature(test.signingKey, test.req, test.headerKeysToSign...)
		if err != test.expectedErr {
			t.Errorf("Expected error %v, got %v", test.expectedErr, err)
		}

		if signature != test.expectedSignature {
			t.Errorf("Expected signature %s, got %s", test.expectedSignature, signature)
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

// reqWithHeaders returns a new http.Request with the given headers.
func (rb *reqBuilder) withHeaders(headers map[string]string) *reqBuilder {
	for k, v := range headers {
		rb.req.Header.Set(k, v)
	}
	return rb
}
