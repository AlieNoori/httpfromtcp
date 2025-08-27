package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	h := "Host: localhost:42069\r\nFooFoo:     barbar     \r\n\r\n"
	data := []byte(h)
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)

	host, ok := headers.Get("host")
	assert.Equal(t, "localhost:42069", host)
	assert.True(t, ok)

	fooFoo, ok := headers.Get("FooFoo")
	assert.Equal(t, "barbar", fooFoo)
	assert.True(t, ok)

	missingKey, ok := headers.Get("missingKey")
	assert.Equal(t, "", missingKey)
	assert.False(t, ok)
	assert.Equal(t, len(h), n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	headers = NewHeaders()
	h = "Set-Person: lane-loves-go\r\nSet-Person: prime-loves-zig\r\nSet-Person: tj-loves-ocaml\r\n"
	data = []byte(h)
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)

	person, ok := headers.Get("Set-Person")
	assert.Equal(t, "lane-loves-go, prime-loves-zig, tj-loves-ocaml", person)
	assert.True(t, ok)

	assert.Equal(t, len(h), n)
	assert.False(t, done)
}
