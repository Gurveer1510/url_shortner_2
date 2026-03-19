package domain

import "errors"

var ErrNotFound = errors.New("short code not found")
var ErrConflict = errors.New("duplicate code found")
var ErrExpiredCode = errors.New("code is expired")
