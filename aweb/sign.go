package aweb

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	ErrNoSignature      = errors.New("no signature")
	ErrNoSigExpiredTime = errors.New("no sig expired time")
	ErrSigExpired       = errors.New("signature expired")
)

const (
	headerKeySignature      = "X-Signature"
	headerKeySigExpiredTime = "X-Sig-Expired"
)

// Signer 签名器
// 用于对 http.Request 进行签名和验证
// 签名时，会增加 X-Signature 和 X-Sig-Expired 两个 Header
// 验证时，会从 X-Signature 中获取签名，并与计算出的签名进行比较，同时会验证 X-Sig-Expired 是否过期
type Signer interface {
	SignRequest(r *http.Request, expiredTime time.Time, headerKeysNeedToSign ...string) error
	VerifyRequest(r *http.Request, headerKeysNeedToSign ...string) error
}

// signer 签名器实现
type signer struct {
	hmacKey string
}

// NewSigner 创建一个签名器
func NewSigner(signingKey string) Signer {
	return &signer{hmacKey: signingKey}
}

// SignRequest 签名 http.Request
func (s *signer) SignRequest(r *http.Request, expiredTime time.Time, headerKeysNeedToSign ...string) error {
	// 为 Header 增加过期时间戳，采用 ISO8601 格式
	r.Header.Add(headerKeySigExpiredTime, expiredTime.UTC().Format(time.RFC3339))

	// 计算签名
	sig, err := calcSignature(s.hmacKey, r, headerKeysNeedToSign...)
	if err != nil {
		return err
	}
	r.Header.Set(headerKeySignature, sig)
	return nil
}

// VerifyRequest 验证 http.Request 的签名
func (s *signer) VerifyRequest(r *http.Request, headerKeysNeedToSign ...string) error {
	// 利用 HMAC 签名验证请求
	// 1. 从请求头中获取签名
	sig := r.Header.Get(headerKeySignature)
	if sig == "" {
		return ErrNoSignature
	}

	expectedSig, err := calcSignature(s.hmacKey, r, headerKeysNeedToSign...)
	if err != nil {
		return err
	}

	// 2. 比较签名
	if sig != expectedSig {
		return fmt.Errorf("signature mismatch: %s != %s", sig, expectedSig)
	}

	// 3. 验证时间戳
	t := r.Header.Get(headerKeySigExpiredTime)
	if t == "" {
		return ErrNoSigExpiredTime
	}
	tm, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return err
	}
	if time.Now().After(tm) {
		return ErrSigExpired
	}

	return nil
}

// calcSignature 计算 HMAC 签名并返回
func calcSignature(key string, r *http.Request, headerKeysToSign ...string) (string, error) {
	// 1. 计算签名
	mac := hmac.New(sha256.New, []byte(key))

	// 增加 URL 的签名
	mac.Write([]byte(r.URL.String()))
	// 增加 Method 的签名
	mac.Write([]byte(r.Method))

	// 对 X-Sig-Expired 进行签名
	signTime := r.Header.Get(headerKeySigExpiredTime)
	if signTime == "" {
		return "", ErrNoSigExpiredTime
	}
	mac.Write([]byte(signTime))

	// 对 headerKeysToSign 中的 Header 进行签名
	for _, k := range headerKeysToSign {
		v := r.Header.Get(k)
		if v != "" {
			mac.Write([]byte(k))
			mac.Write([]byte(v))
		}
	}

	if r.Body != nil {
		// 从请求体中读取数据, 最大 1MB
		r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return "", err
		}

		// 重新设置请求体
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		// 增加 Body 的签名
		mac.Write(body)
	}

	expectedMAC := mac.Sum(nil)
	sig := base64.StdEncoding.EncodeToString(expectedMAC)

	// 2. 返回签名
	return sig, nil
}
