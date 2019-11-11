package main

import (
	"iot-hub/camera/goonvif"
	"net/http"
	"io/ioutil"
	"iot-hub/camera/goonvif/Device"
	"iot-hub/camera/goonvif/xsd/onvif"
	"fmt"
	"log"
	"iot-hub/camera/gosoap"
)

const (
	login = "admin"
	password = "Supervisor"
)

func readResponse(resp *http.Response) string {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func main() {
	//Getting an camera instance
	dev, err := goonvif.NewDevice("192.168.13.14:80","","")
	if err != nil {
		panic(err)
	}
	//Authorization
	dev.Authenticate(login, password)
	
	//Preparing commands
	systemDateAndTyme := Device.GetSystemDateAndTime{}
	getCapabilities := Device.GetCapabilities{Category:"All"}
	createUser := Device.CreateUsers{User:
			onvif.User{
				Username:  "TestUser",
				Password:  "TestPassword",
				UserLevel: "User",
			},
		}
	
	//Commands execution
	systemDateAndTymeResponse, err := dev.CallMethod(systemDateAndTyme)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(readResponse(systemDateAndTymeResponse))
	}
	getCapabilitiesResponse, err := dev.CallMethod(getCapabilities)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(readResponse(getCapabilitiesResponse))
	}
	createUserResponse, err := dev.CallMethod(createUser)
	if err != nil {
		log.Println(err)
	} else {
		/*
		You could use https://iot-hub/camera/gosoap for pretty printing response
		 */
		fmt.Println(gosoap.SoapMessage(readResponse(createUserResponse)).StringIndent())
	}



}
