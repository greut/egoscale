package error

import "errors"

// ErrMissingAPICredentials represent an error due to missing API credentials.
var ErrMissingAPICredentials = errors.New("missing API key/secret")

// ErrResourceNotFound represents an error due to the requested resource not found.
var ErrResourceNotFound = errors.New("resource not found")
