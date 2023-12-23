package ffmpeg

import "C"
import "fmt"

type AVError struct {
	Code int
}

func (e *AVError) Error() string {
	return fmt.Sprintf("AVError %v", e.Code)
}

func wrapError(code C.int) error {
	if code >= 0 {
		return nil
	}

	return &AVError{Code: int(code)}
}
