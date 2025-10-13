package errorhandling

import "errors"

// Adds text to the beginning of an existing error (for sanity)
func ErrorAnd(errorText string, existingError error) error {
	return errors.Join(errors.New(errorText), existingError)
}
