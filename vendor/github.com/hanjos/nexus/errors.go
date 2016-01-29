package nexus

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// Error is returned when there's an error on an attempt to access Nexus.
type Error struct {
	URL        string // e.g. http://somewhere.com
	StatusCode int    // e.g. 400
	Status     string // e.g. 400 Bad response
	Message    string // e.g. Error (400 Bad response) from http://somewhere.com
}

// Error implements the error interface.
func (err Error) Error() string {
	return err.Message
}

// Nexus' API returns error messages sometimes; this function is an attempt to
// capture and return them to the caller.
func (nexus Nexus2x) errorFromResponse(response *http.Response) Error {
	e := Error{
		URL:        nexus.URL,
		StatusCode: response.StatusCode,
		Status:     response.Status,
		Message:    fmt.Sprintf("Error (%v) from %v", response.Status, nexus.URL),
	}

	body, err := bodyToBytes(response.Body)
	if err != nil {
		return e // problems getting the response shouldn't mask the original error
	}

	contentType := response.Header.Get("Content-Type")
	switch {
	case strings.Contains(contentType, "application/json"):
		msg, err := tryFromJSON(body)
		if err != nil {
			return e // couldn't determine a message; use the default one then
		}

		e.Message = msg
	case strings.Contains(contentType, "text/html"):
		msg, err := tryFromHTML(body)
		if err != nil {
			return e // couldn't determine a message; use the default one then
		}

		e.Message = msg
	}

	return e
}

// assuming body is a JSON object with a certain schema.
func tryFromJSON(body []byte) (string, error) {
	// try to extract a message from the response
	var errorResponse struct {
		Errors []struct {
			Msg string
		}
	}

	err := json.Unmarshal(body, &errorResponse)
	if err != nil {
		return "", err // the response doesn't have the expected format
	}

	if len(errorResponse.Errors) != 1 { // can't find the actual message
		return "", fmt.Errorf("Can't determine the message in %v", errorResponse)
	}

	// found a message; use it
	return errorResponse.Errors[0].Msg, nil
}

// yes, I'm using a regex instead of a parser, sue me :)
// All I wanted was an html.Unmarshaller...
var pRe = regexp.MustCompile(`<p>([^<]*)</p>`)

// Tries to get the error message from an HTML response. Sometimes it happens...
func tryFromHTML(body []byte) (string, error) {
	matches := pRe.FindStringSubmatch(string(body))
	if matches == nil {
		return "", fmt.Errorf("Can't find a message in %v", string(body))
	}

	// matches should have the format [ "<p>message</p>" "message" ], so we want
	// the second element
	return matches[1], nil
}
