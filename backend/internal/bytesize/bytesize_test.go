package bytesize

import "testing"

func TestFormatParse(t *testing.T) {
	if Format(512) != "512 B" {
		t.Fatal(Format(512))
	}
	n, err := Parse("64MB")
	if err != nil || n != 64*1024*1024 {
		t.Fatalf("%d %v", n, err)
	}
}
