package apiutils

import (
	"net"
	"net/url"
	"strings"
)

// IsTimeout tests if this is a timeout error in the underlying transport.
// This is unbelievably ugly.
// See: http://stackoverflow.com/questions/23494950/specifically-check-for-timeout-error for details
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	switch err := err.(type) {
	case *url.Error:
		if err, ok := err.Err.(net.Error); ok {
			return err.Timeout()
		}
	case net.Error:
		return err.Timeout()
	}

	if strings.Contains(err.Error(), "use of closed network connection") {
		return true
	}
	return false
}

func IsNetError(err error) bool {
	if err == nil {
		return false
	}
	switch err := err.(type) {
	case *url.Error:
		if _, ok := err.Err.(net.Error); ok {
			return true
		}
	case net.Error:
		return true
	}

	if strings.Contains(err.Error(), "use of closed network connection") {
		return true
	}
	return false
}
