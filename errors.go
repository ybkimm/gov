package main

import "errors"

var (
	ErrUnknownCommand  = errors.New("unknown command")
	ErrHelped          = errors.New("helped")
	ErrVersionRequired = errors.New("version is required")
	ErrInvalidVersion  = errors.New("invalid version")
	ErrUnknownArgs     = errors.New("unknown arguments")
)
