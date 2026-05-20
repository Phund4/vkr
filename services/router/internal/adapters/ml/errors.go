package mlclient

import (
	"errors"
	"fmt"
)

// ErrMLHTTPNonSuccess ответ ML с кодом вне 2xx.
var ErrMLHTTPNonSuccess = errors.New("ml http non-success status")

// mlHTTPError добавляет код и фрагмент тела к ErrMLHTTPNonSuccess.
func mlHTTPError(statusCode int, bodySnippet string) error {
	return fmt.Errorf("%w: %d %s", ErrMLHTTPNonSuccess, statusCode, bodySnippet)
}
