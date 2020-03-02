package ptz

import (
	"camera/goonvif/Device"
	"camera/goonvif/xsd"
	"camera/goonvif/xsd/onvif"
	"net/http"
	"time"
)

//	onvif.User{
//		Username:  "TestUser",
//		Password:  "TestPassword",
//		UserLevel: "User",
//	}
func (c *Camera) Device_CreateUsers(user onvif.User) (*http.Response, error) {
	CreateUser := Device.CreateUsers{
		User: user,
	}
	return c.Call(CreateUser)
}

func (c *Camera) Device_GetSystemDateAndTime() (*http.Response, error) {
	SystemDateAndTime := Device.GetSystemDateAndTime{}
	return c.Call(SystemDateAndTime)
}

func (c *Camera) Device_GetCapabilitieAll() (*http.Response, error) {
	getCapabilities := Device.GetCapabilities{Category: "All"}
	return c.Call(getCapabilities)
}

func (c *Camera) Device_GetCapabilities(category onvif.CapabilityCategory) (*http.Response, error) {
	getCapabilities := Device.GetCapabilities{Category: category}
	return c.Call(getCapabilities)
}

func (c *Camera) Device_GetDeviceInformation() (*http.Response, error) {
	getDeviceInformation := Device.GetDeviceInformation{}
	return c.Call(getDeviceInformation)
}

//时间校准
func (c *Camera) Device_SetSystemDateAndTime(dateTimeType onvif.SetDateTimeType, daylightSavings xsd.Boolean, timeZone onvif.TimeZone, now time.Time) (*http.Response, error) {
	SetSystemDateAndTime := Device.SetSystemDateAndTime{
		DateTimeType:    dateTimeType,
		DaylightSavings: daylightSavings,
		TimeZone:        timeZone,
		UTCDateTime: onvif.DateTime{
			Time: onvif.Time{
				Hour:   xsd.Int(now.Hour()),
				Minute: xsd.Int(now.Minute()),
				Second: xsd.Int(now.Second())},
			Date: onvif.Date{
				Year:  xsd.Int(now.Year()),
				Month: xsd.Int(now.Month()),
				Day:   xsd.Int(now.Day())},
		}}
	return c.Call(SetSystemDateAndTime)
}
