package filesystem

import (
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkCalculateHash(b *testing.B) {
	dir := b.TempDir()
	files := []string{"file1.kt", "file2.xml", "file3.html"}
	for _, f := range files {
		os.WriteFile(filepath.Join(dir, f), []byte("content data for "+f), 0o644)
	}

	r := &BuildRepository{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.CalculateProjectHash(nil, dir)
	}
}

func BenchmarkStreamingHash(b *testing.B) {
	dir := b.TempDir()
	files := []string{"file1.kt", "file2.xml", "file3.html"}
	for _, f := range files {
		os.WriteFile(filepath.Join(dir, f), []byte("content data for "+f), 0o644)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hasher := sha256.New()
		for _, f := range files {
			fh, _ := os.Open(filepath.Join(dir, f))
			io.Copy(hasher, fh)
			fh.Close()
		}
		hasher.Sum(nil)
	}
}
