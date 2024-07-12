// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package cache

import "errors"

// Package errors.
var (
	// The value under the given key was not found.
	ErrNotFound = errors.New("not found")

	// The caching backend produced an error that is not reflected by any other
	// error.
	ErrBackend = errors.New("cache backend error")
)
