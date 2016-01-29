package sampleorigins

type SampleOrigin int

//go:generate stringer -type=SampleOrigin
const (
	None SampleOrigin = iota
	Client
	Central
)
