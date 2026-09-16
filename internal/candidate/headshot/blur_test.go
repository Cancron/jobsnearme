package headshot

import (
	"bytes"
	"testing"
)

func TestBlur_ProducesADifferentValidJPEGOfTheSameSize(t *testing.T) {
	input := encode(t, bands(outputEdge, outputEdge, false, red, green, blue), "jpeg")

	out, err := Blur(input)
	if err != nil {
		t.Fatalf("Blur: %v", err)
	}
	if bytes.Equal(out, input) {
		t.Fatal("Blur returned the input unchanged")
	}

	img := decodeResult(t, out)
	b := img.Bounds()
	if b.Dx() != outputEdge || b.Dy() != outputEdge {
		t.Errorf("blurred size = %dx%d, want %dx%d", b.Dx(), b.Dy(), outputEdge, outputEdge)
	}
}

// TestBlur_SoftensASharpEdge is the actual evidence the transform blurs rather than
// merely re-encoding: a pixel sitting exactly on a hard band boundary in the input must
// come out as a blend of its two neighbouring bands, not one pure band colour.
func TestBlur_SoftensASharpEdge(t *testing.T) {
	input := encode(t, bands(outputEdge, outputEdge, false, red, green, blue), "jpeg")

	out, err := Blur(input)
	if err != nil {
		t.Fatalf("Blur: %v", err)
	}
	img := decodeResult(t, out)

	// The red/green boundary sits at outputEdge/3 on the x axis (see bands()). A pixel a
	// few px to either side of it starts life as pure red or pure green; after a strong
	// blur it must carry a visible amount of BOTH channels.
	boundary := outputEdge / 3
	c := img.At(boundary, outputEdge/2)
	r, g, _, _ := c.RGBA()
	if r == 0 || g == 0 {
		t.Errorf("pixel at the band boundary = %v, want a visible red/green blend (strong blur)", c)
	}
}

func TestBlur_RejectsUndecodableInput(t *testing.T) {
	if _, err := Blur([]byte("not an image")); err == nil {
		t.Fatal("Blur accepted undecodable input")
	}
}
