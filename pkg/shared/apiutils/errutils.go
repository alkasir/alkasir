package apiutils

import (
	"fmt"
	"net/http"

	"github.com/thomasf/lg"
	api "github.com/alkasir/alkasir/pkg/shared"
)

// statusError is an object that can be converted into an api.Status
type statusError interface {
	Status() api.Status
}

// errToAPIStatus converts an error to an api.Status object.
func ErrToAPIStatus(err error) *api.Status {
	switch t := err.(type) {
	case statusError:
		status := t.Status()
		if len(status.Status) == 0 {
			status.Status = api.StatusFailure
		}
		if status.Code == 0 {
			switch status.Status {
			case api.StatusSuccess:
				status.Code = http.StatusOK
			case api.StatusFailure:
				status.Code = http.StatusInternalServerError
			}
		}
		//TODO: check for invalid responses
		return &status
	default:
		status := http.StatusInternalServerError

		// Log errors that were not converted to an error status
		// by REST storage - these typically indicate programmer
		// error by not using pkg/api/errors, or unexpected failure
		// cases.
		lg.Warningln(fmt.Errorf("apiserver received an error that is not an api.Status: %v", err))

		return &api.Status{
			Status:  api.StatusFailure,
			Code:    status,
			Reason:  api.StatusReasonUnknown,
			Message: err.Error(),
		}
	}
}
