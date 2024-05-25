package http

import (
	"bytes"
	"crypto/hmac"
	cryrand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"github.com/gam6itko/go-musthave-metrics/internal/rsautils"
	"io"
	"log"
	"net"
	"net/http"
)

type IClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type EncryptDecorator struct {
	inner     IClient
	publicKey *rsa.PublicKey
}

func NewEncryptDecorator(inner IClient, publicKey *rsa.PublicKey) *EncryptDecorator {
	return &EncryptDecorator{inner: inner, publicKey: publicKey}
}

func (ths EncryptDecorator) Do(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	hash := sha256.New()
	enc, err := rsautils.EncryptOAEP(hash, cryrand.Reader, ths.publicKey, body, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Body = io.NopCloser(bytes.NewBuffer(enc))

	return ths.inner.Do(req)
}

type SignDecorator struct {
	inner IClient
	sign  string
}

func NewSignDecorator(inner IClient, sign string) *SignDecorator {
	return &SignDecorator{inner: inner, sign: sign}
}

func (ths SignDecorator) Do(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	if err = req.Body.Close(); err != nil {
		return nil, err
	}

	// подписываем алгоритмом HMAC, используя SHA-256
	h := hmac.New(sha256.New, []byte(ths.sign))
	if _, err = h.Write(body); err != nil {
		return nil, err
	}
	dst := h.Sum(nil)

	req.Header.Set("HashSHA256", base64.StdEncoding.EncodeToString(dst))
	req.Body = io.NopCloser(bytes.NewBuffer(body))

	return ths.inner.Do(req)
}

type XRealIPDecorator struct {
	inner IClient
	ip    net.IP
}

func NewXRealIPDecorator(inner IClient, ip net.IP) *XRealIPDecorator {
	return &XRealIPDecorator{inner: inner, ip: ip}
}

func (ths XRealIPDecorator) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-Real-IP", ths.ip.String())
	return ths.inner.Do(req)
}
