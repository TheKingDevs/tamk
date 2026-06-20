package watcher

import (
	"testing"
)

func BenchmarkQueueEvent(b *testing.B) {
	w := &Watcher{
		events:  make(chan string, 256),
		pending: make(map[string]struct{}),
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w.queueEvent("/path/to/file.html")
	}
}

func BenchmarkWatchExtensions(b *testing.B) {
	ext := ".html"
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = WatchExtensions[ext]
	}
}
