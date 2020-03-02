package ptz

import (
	"camera/goonvif"
	"camera/goonvif/Media"
	"camera/gosoap"
	"encoding/xml"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func ReadResponse(resp *http.Response) (string, error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return "", err
	}
	return string(b), nil
}

func GetSoapMessage(resp *http.Response) (string, error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return "", err
	}
	body := gosoap.SoapMessage(string(b)).Body()
	return body, nil
}

func ParseResponse(resp *http.Response, v interface{}) error {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return err
	}
	body := gosoap.SoapMessage(string(b)).Body()
	return xml.Unmarshal([]byte(body), v)
}

type Camera struct {
	Addr     string // 192.168.1.64:80
	Username string // admin
	Password string // ADMIN123
}

func (c *Camera) Call(method interface{}) (*http.Response, error) {
	//Getting an camera instance
	dev, err := goonvif.NewDevice(c.Addr)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	//Authorization
	dev.Authenticate(c.Username, c.Password)
	return dev.CallMethod(method)
}

func (c *Camera) GetProfiles() (*Media.GetProfilesResponse, error) {
	getProfiles := Media.GetProfiles{}
	res, err := c.Call(getProfiles)
	if err != nil {
		log.WithError(err).Error("GetProfiles Call Error")
		return nil, err
	}
	getProfilesResponse := &Media.GetProfilesResponse{}
	err = ParseResponse(res, getProfilesResponse)
	if err != nil {
		log.WithError(err).Error("GetProfiles ParseResponse Error")
		return nil, err
	}
	return getProfilesResponse, nil
}

func (c *Camera) PTZ() {

}

func (c *Camera) Media() {

}

func (c *Camera) Imaging() {

}

func (c *Camera) Event() {

}

func (c *Camera) Device() {

}
