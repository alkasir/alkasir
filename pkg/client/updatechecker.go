package client

import (
	"errors"
	"math/rand"
	"time"

	"github.com/thomasf/lg"
)

// UpdateRequest is sent to the client of an UpdateChecker on a channel.
type UpdateRequest struct {
	ResponseC chan UpdateResult // The client is required to send an UpdateResult back on this channel to
	id        string
}

//go:generate stringer -type=UpdateResult

// UpdateResult is sent by the client of an UpdateChecker on the
// UpdateRequest's Repsonse channel to indicate wehter the Update was
// successful or failed.
type UpdateResult int

// These are the UpdateResult constants
const (
	UpdateError UpdateResult = iota
	UpdateSuccess
)

// UpdateChecker manages triggering UpdateRequests on a channel on intervals orand then handling of UpdateRepsonses
type UpdateChecker struct {
	Interval        time.Duration      // Duration between automatic checks
	LastCheck       time.Time          // The last time update was checked
	LastFailedCheck time.Time          // The last time an update check failed (cleared on succesful update)
	LastUpdate      time.Time          // The last time an update was successful
	RequestC        chan UpdateRequest `json:"-"` // Channel that updatechecker clients receives requests on
	forceRequestC   chan bool          // Channel for triggering update requests manually
	response        chan UpdateResult  // Channel that updatechecker clients sends responses on
	active          bool               // Is the checker active and running
}

// NewUpdateChecker creates and returns an UpdateChecker instance.
// The caller should then listen on the RequestC channel for UpdateRequests.
func NewUpdateChecker(name string) (*UpdateChecker, error) {

	c := &UpdateChecker{
		Interval: time.Duration(1*time.Hour + (time.Minute * (time.Duration(rand.Intn(120))))),
	}
	c.response = make(chan UpdateResult)
	c.RequestC = make(chan UpdateRequest)
	c.forceRequestC = make(chan bool)

	lg.Infof("Setting up update timer for %s every %f minute(s) ",
		name, c.Interval.Minutes())
	ticker := time.NewTicker(c.Interval)
	go func() {
		for {
			select {
			case <-c.forceRequestC:
				if !c.active {
					continue
				}
				c.RequestC <- UpdateRequest{
					ResponseC: c.response,
				}
			case <-ticker.C:
				if !c.active {
					continue
				}
				c.RequestC <- UpdateRequest{
					ResponseC: c.response,
				}
			case response := <-c.response:
				c.LastCheck = time.Now()
				switch response {
				case UpdateSuccess:
					lg.V(5).Infoln("UpdateSuccess")
					c.LastUpdate = c.LastCheck
					c.LastFailedCheck = time.Time{}
				case UpdateError:
					lg.Warningln("update check failed")
					c.LastFailedCheck = c.LastCheck
					<-time.After(3*time.Second + time.Duration(rand.Intn(5)))
					go func() {
						c.forceRequestC <- true
					}()
				}
			}

		}
	}()
	return c, nil
}

// Deactivate pauses triggering of update reuquests
func (u *UpdateChecker) Deactivate() {
	u.active = false
}

// Activate starts or resumes the triggering of update requests
func (u *UpdateChecker) Activate() {
	u.active = true
}

// UpdateNow forces an Update request to be triggered if the
func (u *UpdateChecker) UpdateNow() error {
	if u.active {
		u.forceRequestC <- true
		return nil
	}
	return errors.New("Updatechecker is not active")
}
