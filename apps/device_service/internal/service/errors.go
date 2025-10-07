package service

import "errors"

var (
	ErrInvalidInput           = errors.New("invalid input")
	ErrParentLocationNotFound = errors.New("parent location not found")
	ErrLocationNotFound       = errors.New("location not found")
	ErrLocationInUse          = errors.New("location is in use by devices")
	ErrDeviceNotFound         = errors.New("device not found")
	ErrUnsupportedDeviceType  = errors.New("unsupported device type")
)
