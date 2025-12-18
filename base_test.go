package shortid

import (
	"testing"
)

func TestEncodeBase62(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{"0", 0, "0"},
		{"1", 1, "1"},
		{"10", 10, "a"},
		{"61", 61, "Z"},
		{"62", 62, "10"},
		{"3844", 3844, "100"},
		{"238328", 238328, "1000"},
		{"123456789", 123456789, "8m0Kx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeBase62(tt.input)
			if result != tt.expected {
				t.Errorf("EncodeBase62(%d) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDecodeBase62(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint64
	}{
		{"0", "0", 0},
		{"1", "1", 1},
		{"a", "a", 10},
		{"Z", "Z", 61},
		{"10", "10", 62},
		{"100", "100", 3844},
		{"1000", "1000", 238328},
		{"8m0Kx", "8m0Kx", 123456789},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeBase62ToUint(tt.input)
			if err != nil {
				t.Errorf("DecodeBase62ToUint(%s) returned error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("DecodeBase62ToUint(%s) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEncodeDecodeBase62(t *testing.T) {
	testNumbers := []uint64{
		//0, 1, 10, 100, 1000, 12345, 999999, 1234567890,
		//18446744073709551615,
		// 832092015965291904,
		832092516194764160,
	}
	encoded := "ZsZxmWnwNW"
	id, err := DecodeBase62ToUint(encoded)
	if err != nil {
		t.Errorf("DecodeBase62ToUint failed for encoded value %s: %v", encoded, err)
	}
	t.Logf(" id: %d, encoded: %s", id, encoded)

	for _, num := range testNumbers {
		encoded := EncodeBase62(num)
		decoded, err := DecodeBase62ToUint(encoded)
		if err != nil {
			t.Errorf("DecodeBase62ToUint failed for encoded value %s: %v", encoded, err)
		}
		t.Logf("num: %d, encoded: %s, decoded: %d", num, encoded, decoded)
		if decoded != num {
			t.Errorf("Round-trip failed: %d -> %s -> %d", num, encoded, decoded)
		}
	}
}

func TestEncodeBase58(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{"0", 0, "1"},
		{"1", 1, "2"},
		{"57", 57, "z"},
		{"58", 58, "21"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeBase58(tt.input)
			if result != tt.expected {
				t.Errorf("EncodeBase58(%d) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsValidBase62(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty", "", false},
		{"valid", "abcXYZ123", true},
		{"invalid", "abc#123", false},
		{"single valid", "a", true},
		{"single invalid", "#", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidBase62(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidBase62(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBase62Length(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected int
	}{
		{"0", 0, 1},
		{"1", 1, 1},
		{"61", 61, 1},
		{"62", 62, 2},
		{"3843", 3843, 2},
		{"3844", 3844, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Base62Length(tt.input)
			if result != tt.expected {
				t.Errorf("Base62Length(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func BenchmarkEncodeBase62(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EncodeBase62(123456789)
	}
}

func BenchmarkDecodeBase62(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DecodeBase62ToUint("8m0Kx")
	}
}
