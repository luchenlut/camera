package ptz

import (
	"iot-hub/camera/goonvif/Media"
	"iot-hub/camera/goonvif/xsd/onvif"
	"net/http"
)

func (c *Camera) Media_GetStreamUri(setup onvif.StreamSetup, token onvif.ReferenceToken) (*http.Response, error) {
	StreamUri := Media.GetStreamUri{StreamSetup: setup, ProfileToken: token}
	return c.Call(StreamUri)
}

func (c *Camera) Media_GetStreamUri2(token onvif.ReferenceToken) (*http.Response, error) {
	StreamUri := Media.GetStreamUri{ProfileToken: token}
	return c.Call(StreamUri)
}

func (c *Camera) Media_GetStreamUri3(token onvif.ReferenceToken) (*http.Response, error) {
	setup := onvif.StreamSetup{
		Stream: onvif.StreamType("RTP-Unicast"), // Defines if a multicast or unicast stream is requested  enum:{RTP-Unicast,RTP-Multicast}
		Transport: onvif.Transport{
			Protocol: onvif.TransportProtocol("HTTP"), // Defines the network protocol for streaming, either UDP=RTP/UDP, RTSP=RTP/RTSP/TCP or HTTP=RTP/RTSP/HTTP/TCP enum {UDP,TCP,RTSP,HTTP}
			Tunnel:   nil,
		},
	}

	StreamUri := Media.GetStreamUri{StreamSetup: setup, ProfileToken: token}
	return c.Call(StreamUri)
}

func (c *Camera) Media_GetSnapshotUri(token onvif.ReferenceToken) (*http.Response, error) {
	SnapshotUri := Media.GetSnapshotUri{ProfileToken: token}
	return c.Call(SnapshotUri)
}
