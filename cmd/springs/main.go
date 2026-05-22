//go:build !appunit

package main

import (
	"springs/internal/app"
)

func main() {
	exitIfError(app.Run())
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T10:55:40-05:00","module_hash":"9d84591eb7188c9fb06ea7dcab14b335f6ae6f31642c0969368df94324f04dc4","functions":[{"id":"func/main","name":"main","line":9,"end_line":11,"hash":"dde35a2b7dbfabb1f34e1a27c9c9110f96d610aa5d33290f005d6d33d81e7e5c"}]}
// mutate4go-manifest-end
