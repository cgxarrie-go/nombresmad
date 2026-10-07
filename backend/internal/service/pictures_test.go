package service

import "testing"

func TestImageExt(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		ext  string
		ok   bool
	}{
		{"jpeg", []byte{0xFF, 0xD8, 0xFF, 0x00}, ".jpg", true},
		{"png", []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, ".png", true},
		{"gif", []byte("GIF89a"), ".gif", true},
		{"webp", []byte("RIFFxxxxWEBP"), ".webp", true},
		{"text", []byte("not an image"), "", false},
		{"short", []byte{0xFF, 0xD8}, "", false},
	}
	for _, tc := range cases {
		ext, ok := imageExt(tc.data)
		if ext != tc.ext || ok != tc.ok {
			t.Fatalf("%s: ext=%s ok=%v", tc.name, ext, ok)
		}
	}
}
