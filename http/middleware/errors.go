// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package middleware

import "errors"

// All exported package errors.
var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidRequest = errors.New("request is invalid")
)
