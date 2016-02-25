package apiutils

import (
	"fmt"

	"github.com/alkasir/alkasir/pkg/shared"
	"github.com/alkasir/alkasir/pkg/shared/apierrors"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/thomasf/lg"
)

func WriteRestError(w rest.ResponseWriter, err error) {
	var status shared.Status
	if _, ok := err.(*apierrors.StatusError); !ok {
		err = apierrors.NewInternalError(err)
	}
	switch err := err.(type) {
	case *apierrors.StatusError:
		status = err.Status()
	default:
		panic(err)
	}
	if status.Code > 299 && lg.V(19) {
		lg.WarningDepth(1, fmt.Sprintf("apierror %d: %+v", status.Code, status))
	}
	w.WriteHeader(status.Code)
	err = w.WriteJson(&status)
	if err != nil {
		lg.Warning(err)
	}
}
