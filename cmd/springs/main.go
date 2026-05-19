//go:build !appunit

package main

import (
	"springs/internal/app"
)

func main() {
	exitIfError(app.Run())
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-18T21:17:44-05:00","module_hash":"6bd4cabe70016232f0f9fea6b931bf705348a9c0a87f43d9d5cd7c6171d9530e","functions":[{"id":"func/main","name":"main","line":9,"end_line":11,"hash":"dde35a2b7dbfabb1f34e1a27c9c9110f96d610aa5d33290f005d6d33d81e7e5c"},{"id":"func/exitIfError","name":"exitIfError","line":13,"end_line":17,"hash":"36e1340dd723f156d0234b8d8e7933b0a4e878325131217e4e319478fb85adbe"}]}
// mutate4go-manifest-end
