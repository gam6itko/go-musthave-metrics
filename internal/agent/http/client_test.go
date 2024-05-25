package http

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	mock_http "github.com/gam6itko/go-musthave-metrics/internal/agent/http/mocks"
	"github.com/gam6itko/go-musthave-metrics/internal/rsautils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"net/http"
	"testing"
)

var (
	_ IClient = (*EncryptDecorator)(nil)
	_ IClient = (*SignDecorator)(nil)
	_ IClient = (*XRealIPDecorator)(nil)
)

// Проверяем что подпись имеется и её можно проверить.
func TestSignDecorator_Do(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	requestBody := "requestBody string"
	signKey := "super-foo-hash"

	inner := mock_http.NewMockIClient(ctrl)
	inner.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			base64str := req.Header.Get("HashSHA256")
			require.NotEmpty(t, base64str)

			hashFromReq, err := base64.StdEncoding.DecodeString(base64str)
			require.NoError(t, err)

			bRequestBody, err := io.ReadAll(req.Body)
			require.NoError(t, err)
			err = req.Body.Close()
			require.NoError(t, err)

			h := hmac.New(sha256.New, []byte(signKey))
			_, err = h.Write(bRequestBody)
			require.NoError(t, err)

			dst := h.Sum(nil)
			require.True(t, bytes.Equal(hashFromReq, dst))

			return &http.Response{}, nil
		})

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("http://example.com/"),
		bytes.NewBuffer([]byte(requestBody)),
	)
	require.NoError(t, err)
	client := NewSignDecorator(inner, signKey)

	_, err = client.Do(req)
	require.NoError(t, err)
}

func TestXRealIPDecorator_Do(t *testing.T) {
	ctrl := gomock.NewController(t)

	inner := mock_http.NewMockIClient(ctrl)
	inner.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			ip := req.Header.Get("X-Real-IP")
			require.NotEmpty(t, ip)

			require.Equal(t, "192.168.1.1", ip)

			return &http.Response{}, nil
		})

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("http://example.com/"),
		nil,
	)
	client := NewXRealIPDecorator(inner, net.ParseIP("192.168.1.1"))
	_, err = client.Do(req)
	require.NoError(t, err)
}

func TestEncryptDecorator_Do(t *testing.T) {
	// generate key pairs
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)

	requestBody := "requestBody string"

	ctrl := gomock.NewController(t)

	inner := mock_http.NewMockIClient(ctrl)
	inner.EXPECT().
		Do(gomock.Any()).
		DoAndReturn(func(req *http.Request) (*http.Response, error) {
			bRequestBody, err := io.ReadAll(req.Body)
			require.NoError(t, err)

			hash := sha256.New()
			b, err := rsautils.DecryptOAEP(hash, rand.Reader, privateKey, bRequestBody, nil)
			require.NoError(t, err)

			require.True(t, bytes.Equal([]byte(requestBody), b))

			return &http.Response{}, nil
		})

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("http://example.com/"),
		bytes.NewBuffer([]byte(requestBody)),
	)
	client := NewEncryptDecorator(inner, &privateKey.PublicKey)
	_, err = client.Do(req)
	require.NoError(t, err)
}
