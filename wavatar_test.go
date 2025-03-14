package wavatar

import (
	"bytes"
	"crypto/md5"
	"image"
	"image/png"
	"testing"
)

func TestNewCreatesCorrectSizeImage(t *testing.T) {
	hash := []byte("test@example.com")
	img := New(hash)

	bounds := img.Bounds()
	if bounds.Dx() != AvatarSize || bounds.Dy() != AvatarSize {
		t.Errorf("Expected image size %dx%d, got %dx%d", AvatarSize, AvatarSize, bounds.Dx(), bounds.Dy())
	}
}

func TestDifferentHashesProduceDifferentImages(t *testing.T) {
	hash1 := []byte("user1@example.com")
	hash2 := []byte("user2@example.com")

	img1 := New(hash1)
	img2 := New(hash2)

	buf1 := new(bytes.Buffer)
	buf2 := new(bytes.Buffer)

	if err := png.Encode(buf1, img1); err != nil {
		t.Fatalf("Failed to encode image 1: %v", err)
	}
	if err := png.Encode(buf2, img2); err != nil {
		t.Fatalf("Failed to encode image 2: %v", err)
	}

	if bytes.Equal(buf1.Bytes(), buf2.Bytes()) {
		t.Error("Different hashes should produce different images")
	}
}

func TestSameHashProducesSameImage(t *testing.T) {
	hash := []byte("same@example.com")

	img1 := New(hash)
	img2 := New(hash)

	buf1 := new(bytes.Buffer)
	buf2 := new(bytes.Buffer)

	if err := png.Encode(buf1, img1); err != nil {
		t.Fatalf("Failed to encode image 1: %v", err)
	}
	if err := png.Encode(buf2, img2); err != nil {
		t.Fatalf("Failed to encode image 2: %v", err)
	}

	if !bytes.Equal(buf1.Bytes(), buf2.Bytes()) {
		t.Error("Same hash should produce identical images")
	}
}

func TestEmptyHash(t *testing.T) {
	hash := []byte{}
	img := New(hash)

	if img == nil {
		t.Fatal("Empty hash should still produce an image, not nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() != AvatarSize || bounds.Dy() != AvatarSize {
		t.Errorf("Expected image size %dx%d, got %dx%d", AvatarSize, AvatarSize, bounds.Dx(), bounds.Dy())
	}
}

func TestLargeHash(t *testing.T) {
	// Create a large hash
	largeHash := make([]byte, 1024*1024) // 1MB hash
	for i := range largeHash {
		largeHash[i] = byte(i % 256)
	}

	img := New(largeHash)

	if img == nil {
		t.Fatal("Large hash should still produce an image, not nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() != AvatarSize || bounds.Dy() != AvatarSize {
		t.Errorf("Expected image size %dx%d, got %dx%d", AvatarSize, AvatarSize, bounds.Dx(), bounds.Dy())
	}
}

func TestMD5HashInput(t *testing.T) {
	email := "test@example.com"
	hash := md5.Sum([]byte(email))

	img := New(hash[:])

	if img == nil {
		t.Fatal("MD5 hash should produce a valid image")
	}

	bounds := img.Bounds()
	if bounds.Dx() != AvatarSize || bounds.Dy() != AvatarSize {
		t.Errorf("Expected image size %dx%d, got %dx%d", AvatarSize, AvatarSize, bounds.Dx(), bounds.Dy())
	}
}

func TestImageIsNotEmpty(t *testing.T) {
	hash := []byte("test@example.com")
	img := New(hash)

	// Convert to RGBA to check pixel values
	rgba := image.NewRGBA(img.Bounds())
	for y := 0; y < AvatarSize; y++ {
		for x := 0; x < AvatarSize; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	// Check for some non-zero pixels (not all transparent)
	hasNonZeroPixel := false
	for y := 0; y < AvatarSize; y++ {
		for x := 0; x < AvatarSize; x++ {
			r, g, b, a := rgba.At(x, y).RGBA()
			if r > 0 || g > 0 || b > 0 || a > 0 {
				hasNonZeroPixel = true
				break
			}
		}
		if hasNonZeroPixel {
			break
		}
	}

	if !hasNonZeroPixel {
		t.Error("Generated image appears to be empty")
	}
}
