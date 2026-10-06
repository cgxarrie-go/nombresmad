package main

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

func TestPictureNameRe(t *testing.T) {
	if !pictureNameRe.MatchString("12-99.jpg") {
		t.Fatal("expected stored name to match")
	}
	rejected := []string{"../12-99.jpg", "12-99.exe", "12.jpg", "photo.jpg", "12-99.JPG"}
	for _, name := range rejected {
		if pictureNameRe.MatchString(name) {
			t.Fatalf("accepted %s", name)
		}
	}
}

func TestPicturePathRejectsTraversal(t *testing.T) {
	if _, ok := picturePath("../12-99.jpg"); ok {
		t.Fatal("traversal name accepted")
	}
	if _, ok := picturePath("12-99.jpg"); !ok {
		t.Fatal("valid name rejected")
	}
}
