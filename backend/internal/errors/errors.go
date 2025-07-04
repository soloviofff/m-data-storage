package errors

import "errors"

// Common errors
var (
	// ErrNotFound - ошибка, когда сущность не найдена
	ErrNotFound = errors.New("entity not found")
)
