package hasher

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"empty key", ""},
		{"simple key", "secret"},
		{"long key", "very-long-secret-key-for-testing"},
		{"special chars", "!@#$%^&*()"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hasher := New(tc.key)
			assert.NotNil(t, hasher)
			assert.Equal(t, []byte(tc.key), hasher.key)
		})
	}
}

func TestHasher_Hash(t *testing.T) {
	key := "test-secret-key"
	hasher := New(key)

	tests := []struct {
		name string
		data []byte
	}{
		{"empty data", []byte{}},
		{"simple text", []byte("hello world")},
		{"json data", []byte(`{"id":"test","type":"gauge","value":123.45}`)},
		{"binary data", []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}},
		{"large data", make([]byte, 1000)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hash := hasher.Hash(tc.data)

			assert.NotEmpty(t, hash)

			hash2 := hasher.Hash(tc.data)
			assert.Equal(t, hash, hash2)

			if len(tc.data) > 0 {
				differentData := append(tc.data, 0x00)
				differentHash := hasher.Hash(differentData)
				assert.NotEqual(t, hash, differentHash)
			}
		})
	}
}

func TestHasher_Hash_DifferentKeys(t *testing.T) {
	data := []byte("same data")

	hasher1 := New("key1")
	hasher2 := New("key2")

	hash1 := hasher1.Hash(data)
	hash2 := hasher2.Hash(data)

	assert.NotEqual(t, hash1, hash2)
}

func TestHasher_Validate(t *testing.T) {
	key := "test-secret-key"
	hasher := New(key)

	tests := []struct {
		name          string
		data          []byte
		expectedValid bool
		modifyHash    bool
	}{
		{"valid hash", []byte("test data"), true, false},
		{"empty data valid", []byte{}, true, false},
		{"json data valid", []byte(`{"key":"value"}`), true, false},
		{"invalid hash", []byte("test data"), false, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			correctHash := hasher.Hash(tc.data)

			var hashToValidate string
			if tc.modifyHash {
				hashToValidate = correctHash[:len(correctHash)-2] + "00"
			} else {
				hashToValidate = correctHash
			}

			valid, err := hasher.Validate(tc.data, hashToValidate)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedValid, valid)
		})
	}
}

func TestHasher_Validate_DifferentData(t *testing.T) {
	hasher := New("secret")

	originalData := []byte("original data")
	hash := hasher.Hash(originalData)

	modifiedData := []byte("modified data")
	valid, err := hasher.Validate(modifiedData, hash)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestHasher_Validate_WrongKey(t *testing.T) {
	data := []byte("test data")

	hasher1 := New("key1")
	hasher2 := New("key2")

	hash := hasher1.Hash(data)

	valid, err := hasher2.Validate(data, hash)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestHashHeaderKey(t *testing.T) {
	assert.Equal(t, "HashSHA256", HashHeaderKey)
	assert.NotEmpty(t, HashHeaderKey)
}

func TestHasher_ConsistentResults(t *testing.T) {
	key := "consistent-test-key"
	data := []byte("consistent test data")

	hasher1 := New(key)
	hasher2 := New(key)

	hash1 := hasher1.Hash(data)
	hash2 := hasher2.Hash(data)

	assert.Equal(t, hash1, hash2)

	valid, err := hasher1.Validate(data, hash2)
	require.NoError(t, err)
	assert.True(t, valid)

	valid, err = hasher2.Validate(data, hash1)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestHasher_EmptyKey(t *testing.T) {
	hasher := New("")
	data := []byte("test with empty key")

	hash := hasher.Hash(data)
	assert.NotEmpty(t, hash)

	valid, err := hasher.Validate(data, hash)
	require.NoError(t, err)
	assert.True(t, valid)
}

func BenchmarkHasher_Hash(b *testing.B) {
	hasher := New("benchmark-key")
	data := []byte("benchmark test data that is reasonably long for testing performance")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher.Hash(data)
	}
}

func BenchmarkHasher_Validate(b *testing.B) {
	hasher := New("benchmark-key")
	data := []byte("benchmark test data that is reasonably long for testing performance")
	hash := hasher.Hash(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher.Validate(data, hash)
	}
}
