package repo

import "testing"

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
	store := NewPictureRepo("pictures")
	if _, ok := store.Path("../12-99.jpg"); ok {
		t.Fatal("traversal name accepted")
	}
	if _, ok := store.Path("12-99.jpg"); !ok {
		t.Fatal("valid name rejected")
	}
}
