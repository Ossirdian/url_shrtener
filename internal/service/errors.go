package service

import "errors"

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrInvalidURL      = errors.New("invalid url")
)
