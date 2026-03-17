package apperrors

import "errors"

var ErrNotFound = errors.New("short code not found")
var ErrConflict = errors.New("Duplicate code found")
var ErrExpiredCode = errors.New("Code is expired")
