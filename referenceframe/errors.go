package referenceframe

import "github.com/pkg/errors"

// NewParentFrameMissingError returns an error indicating that the parent frame is nil.
func NewParentFrameMissingError() error {
	return errors.New("parent frame is nil")
}

// NewFrameMissingError returns an error indicating that the given frame is missing from the framesystem.
func NewFrameMissingError(frameName string) error {
	return errors.Errorf("frame with name %q not in frame system", frameName)
}

// NewFrameAlreadyExistsError returns an error indicating that a frame of the given name already exists.
func NewFrameAlreadyExistsError(frameName string) error {
	return errors.Errorf("frame with name %q already in frame system", frameName)
}

// NewIncorrectInputLengthError returns an error indicating that the length of the Innput array does not match the DoF of the frame.
func NewIncorrectInputLengthError(actual, expected int) error {
	return errors.Errorf("number of inputs does not match frame DoF, expected %d but got %d", expected, actual)
}
