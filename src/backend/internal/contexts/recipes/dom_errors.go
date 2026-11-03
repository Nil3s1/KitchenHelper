package recipes

import "errors"

var (
	ErrNotFound   = errors.New("recipe not found")
	ErrValidation = errors.New("recipe validation failed")
)
