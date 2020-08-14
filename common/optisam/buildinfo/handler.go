// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package buildinfo

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

// Handler returns an HTTP handler for version information.
func Handler(buildInfo BuildInfo) http.Handler {
	var body []byte

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if body == nil {
			var err error

			body, err = json.Marshal(buildInfo)
			if err != nil {
				panic(errors.Wrap(err, "failed to render version information"))
			}
		}

		_, _ = w.Write(body)
	})
}
