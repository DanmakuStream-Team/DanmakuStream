package media

import "testing"

func TestValidMessageVideoHeader(t *testing.T) {
	tests := []struct {
		name string
		ext  string
		head []byte
		want bool
	}{
		{name: "mp4", ext: ".mp4", head: []byte{0, 0, 0, 24, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm'}, want: true},
		{name: "short mp4", ext: ".mp4", head: []byte("ftyp"), want: false},
		{name: "fake mp4", ext: ".mp4", head: []byte("0000not-video"), want: false},
		{name: "webm", ext: ".webm", head: []byte{0x1a, 0x45, 0xdf, 0xa3}, want: true},
		{name: "fake webm", ext: ".webm", head: []byte{0, 0, 0, 0}, want: false},
		{name: "unsupported", ext: ".avi", head: []byte{0x1a, 0x45, 0xdf, 0xa3}, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := validMessageVideoHeader(test.ext, test.head); got != test.want {
				t.Fatalf("validMessageVideoHeader(%q) = %v, want %v", test.ext, got, test.want)
			}
		})
	}
}

func TestRandomHex(t *testing.T) {
	value, err := randomHex(16)
	if err != nil {
		t.Fatalf("randomHex() error = %v", err)
	}
	if len(value) != 32 {
		t.Fatalf("len(randomHex(16)) = %d, want 32", len(value))
	}
}
