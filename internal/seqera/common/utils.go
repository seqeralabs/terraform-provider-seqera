// Package common provides shared utilities for custom Seqera resources.
package common

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

// DebugResponse formats an HTTP response for error messages.
func DebugResponse(response *http.Response) string {
	if v := response.Request.Header.Get("Authorization"); v != "" {
		response.Request.Header.Set("Authorization", "(sensitive)")
	}
	dumpReq, err := httputil.DumpRequest(response.Request, true)
	if err != nil {
		dumpReq, err = httputil.DumpRequest(response.Request, false)
		if err != nil {
			return err.Error()
		}
	}
	dumpRes, err := httputil.DumpResponse(response, true)
	if err != nil {
		dumpRes, err = httputil.DumpResponse(response, false)
		if err != nil {
			return err.Error()
		}
	}
	return fmt.Sprintf("**Request**:\n%s\n**Response**:\n%s", string(dumpReq), string(dumpRes))
}
