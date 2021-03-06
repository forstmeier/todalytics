package evt

import (
	"errors"
	"testing"
)

func TestConvertError(t *testing.T) {
	err := &ConvertError{
		err: errors.New("mock convert error"),
	}

	recieved := err.Error()
	expected := "package evt: mock convert error"

	if recieved != expected {
		t.Errorf("incorrect error message, received: %s, expected: %s", recieved, expected)
	}
}
