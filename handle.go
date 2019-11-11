package camera

import (
	"github.com/sirupsen/logrus"
)

const (
	keyPrefix = "hub:digital:%d:%d:%s"
)

func setMQTT(key string, value interface{}) error {
	reported := make(map[string]interface{})
	reported[key]=value
	state := &State{Reported: reported}
	twins := RequestTwins{Method: Update, State: state, Version: 1}
	//fmt.Printf("%+v",twins)
	jsonText, err := twins.MarshalJSONText()
	if err != nil{
		return err
	}
	err = pubSub.publish(pubSub.rxTopic, jsonText)
	if err != nil{
		return err
	}

	return nil
}

func handleSetSystemDateAndTime(time interface{}) error {
	logrus.Println("time: ", time)
	return setMQTT(TimeCalibrationData, time)
	//return nil
}

func handleGetSnapshot(url interface{}) error {
	logrus.Println("url:  ", url)
	return setMQTT(SnapshotURL, url)
}

func HandleInterval(state interface{}) error {
	logrus.Println("state:  ", state)
	return setMQTT(CameraStatus, state)
}
