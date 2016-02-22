package analysis

import (
	"fmt"

	"github.com/alkasir/alkasir/pkg/measure"
	"github.com/thomasf/lg"
)

type scorer interface {
	Score() float64
}

type HTTPHeaderScore struct {
	// based on http measurement status code field
	StatusCode float64
	// based on http redirects field
	Redirects float64
	// based on http measurement error field
	Error float64
}

func (h HTTPHeaderScore) String() string {
	return fmt.Sprintf(
		"score:%.1f (statusCode:%.1f redirects:%.1f error:%.1f)",
		h.Score(), h.StatusCode, h.Redirects, h.Error)
}

func (h *HTTPHeaderScore) Score() float64 {
	result := 1.0
	result = (result + h.StatusCode) / 2
	result = (result + h.Error) / 2
	result = (result + h.Redirects) / 2
	return result
}

func scoreStatusCode(client, central int) float64 {
	if central < 400 && client > 400 {
		lg.Infof("status ranges %d != %d", client, central)
		return 1.0
	} else if client != central {
		lg.Infof("status %d != %d", client, central)
		return 0.7
	} else if client != central {
		return 0.0
	}
	return 0.5
}

func scoreError(client, central string) float64 {
	if central == client {
		return 0.5
	} else if central == "" && client != "" {
		return 1
	} else if central != "" && client == "" {
		return 0.5
	}
	return 0.5
}

func scoreRedirects(client, central []measure.Redirect) float64 {
	if len(client) == len(central) {
		return 0.5
	}
	return 0.6

}

func scoreHTTPHeaders(client, central measure.HTTPHeaderResult) HTTPHeaderScore {
	score := HTTPHeaderScore{
		StatusCode: scoreStatusCode(client.StatusCode, central.StatusCode),
		Error:      scoreError(client.Error, central.Error),
		Redirects:  scoreRedirects(client.Redirects, central.Redirects),
	}
	return score
}
