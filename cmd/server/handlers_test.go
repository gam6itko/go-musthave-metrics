package main

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

// Если несколько раз вызвать метод, то ссылки должны быть на разные области памяти.
func Test_decodeJsonRequest_metricNotSameRef(t *testing.T) {
	req, err := http.NewRequest(
		"POST",
		"/update",
		bytes.NewBufferString(`{"id": "foo", "type": "counter"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	require.NoError(t, err)
	m1, err := decodeMetricsRequest(req)
	require.NoError(t, err)
	req = nil

	req2, err := http.NewRequest(
		"POST",
		"/update",
		bytes.NewBufferString(`{"id": "foo", "type": "counter"}`),
	)
	req2.Header.Set("Content-Type", "application/json")
	require.NoError(t, err)
	m2, err := decodeMetricsRequest(req2)
	require.NoError(t, err)
	req2 = nil

	require.NotSame(t, *m1, *m2)
}
