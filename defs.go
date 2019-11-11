package camera

import (
	"errors"
)

// Error codes returned by failures dealing with server or connection.
var (
	ErrWouldBlock   = errors.New("would block")
	ErrServerClosed = errors.New("server has been closed")
)

// definitions about some constants.
const (
	bufferSize1024 = 1024
	seriesKey      = "hub:series"
)
