// Package common provides shared utilities for custom Seqera resources.
package common

import (
	"context"
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

// PaginatedSearch performs a paginated search through API results.
// It handles pagination automatically, calling fetchPage for each page until the item is found
// or all pages have been searched.
//
// Parameters:
//   - ctx: Context for the search
//   - fetchPage: Function that fetches a single page of results given max and offset parameters.
//     Returns (items, totalSize, error). If totalSize is unknown, return 0.
//   - matchItem: Function that returns true if the item matches the search criteria
//
// Returns the first matching item or nil if not found.
func PaginatedSearch[T any](
	ctx context.Context,
	fetchPage func(ctx context.Context, max, offset int) (items []T, totalSize int64, err error),
	matchItem func(item *T) bool,
) (*T, error) {
	pageSize := 100
	offset := 0

	for {
		items, totalSize, err := fetchPage(ctx, pageSize, offset)
		if err != nil {
			return nil, err
		}

		// Search for matching item in current page
		for i := range items {
			if matchItem(&items[i]) {
				return &items[i], nil
			}
		}

		// Check pagination stopping conditions
		itemsInPage := len(items)

		// Stop if no items in this page
		if itemsInPage == 0 {
			break
		}

		// Stop if we've fetched all items based on totalSize
		if totalSize > 0 && int64(offset)+int64(itemsInPage) >= totalSize {
			break
		}

		// Stop if we got a partial page (less than requested) - indicates last page
		if itemsInPage < pageSize {
			break
		}

		// Move to next page
		offset += pageSize
	}

	return nil, nil
}
