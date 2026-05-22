package main

import "log"

func exitIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// mutate4go-manifest-begin
// {"version":1,"tested_at":"2026-05-22T10:55:36-05:00","module_hash":"0392a3214e180cf84fa40727e0271edec334f26b7cb9be2db3d163e4583997dc","functions":[{"id":"func/exitIfError","name":"exitIfError","line":5,"end_line":9,"hash":"36e1340dd723f156d0234b8d8e7933b0a4e878325131217e4e319478fb85adbe"}]}
// mutate4go-manifest-end
