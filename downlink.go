package camera

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
	"iot-hub/api"
	"iot-hub/camera/config"
	"iot-hub/camera/goonvif/Device"
	"iot-hub/camera/goonvif/Media"
	"iot-hub/camera/goonvif/PTZ"
	"iot-hub/camera/goonvif/xsd"
	"iot-hub/camera/goonvif/xsd/onvif"
	"iot-hub/camera/ptz"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"
)

const (
	SUCCESS = "SUCCESS"
	ACK     = "ACK"
	FAILED  = "FAILED"
	TIMEOUT = "TIMEOUT"
)

func HandleIntervalCheck(value interface{}, callback func(value interface{}) error) {
	// 当该条命令执行成功或超时后再释放锁
	go func() {
		callback(value)
	}()
}
func handleResponse(value interface{}, callback func(value interface{}) error) {
	// 当该条命令执行成功或超时后再释放锁
	go func() {
		callback(value)
	}()
}

// 数据下行处理逻辑
func handlerCameraChan(b *Backend) {
	defer func() {
		if p := recover(); p != nil {
			logrus.Errorf("handler camera chan panics: %v\n", p)
		}
	}()

	for {
		select {
		case node := <-b.deviceChan:
			if node.Payload != nil {
				collectCameraPacket(node)
			}
		}
	}
}

// 采集下行命令处理报文
func collectCameraPacket(p DevicePayload) {
	logrus.Println("采集下行命令处理报文")
	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	resp, err := asset.GetDevice(ctx, &api.GetDeviceRequest{Did: p.Did, Offline: true})
	//resp, err := command.GetDeviceForID(ctx, &api.GetDeviceForIDRequest{ID: p.Did})

	if err != nil {
		logrus.Errorf("get device rpc error %v", err)
		return
	}
	entry := NewEntry(Fields{"tid": resp.Tid, "pid": resp.Pid, "did": resp.Did})
	//entry := NewEntry(Fields{"did": p.Did})

	responseTwins := &ResponseTwins{}
	if err := json.Unmarshal(p.Payload, &responseTwins); err != nil {
		logrus.Errorf("json decode response twins error %v", err)
		return
	}

	entry.DownLink("receive down data from mqtt this shuncom gateway %s", string(p.Payload))

	command.UpdateCommandState(ctx, &api.UpdateCommandStateRequest{Cid: responseTwins.CommandID, State: ACK})
	if err := handlerCameraDownLink(responseTwins, p.Did); err != nil {
		entry.Error("send data to shuncom gateway %v", err)
		command.UpdateCommandState(ctx, &api.UpdateCommandStateRequest{Cid: responseTwins.CommandID, State: FAILED})
	} else {
		command.UpdateCommandState(ctx, &api.UpdateCommandStateRequest{Cid: responseTwins.CommandID, State: SUCCESS})
	}
}

// 处理下行命令
func handlerCameraDownLink(resp *ResponseTwins, gwID string) error {
	entry := logrus.WithFields(logrus.Fields{"Did": gwID})
	var send error
	if  resp.CommandID != "" {
		for desK, desV := range resp.Payload.State.Desired {
			entry.Debugf("接收到下发命令 %v:%v", desK, desV)
			switch desK {
			case PTZControl, Angle, Zoom:
				send = getPTZControlResult(resp.Payload.State.Desired)
				entry.Debug("云台控制", send)
			case Snapshot:
				send = SnapshotUri()
				entry.Debug("快照", send)
			case SetPreset:
				send = PTZSetPreset(resp.Payload.State.Desired[SetPreset].(float64))
				entry.Debug("设置预置位置", send)
			case GetPresets:
				send = PTZGetPresets()
				entry.Debug("获取所有预置位置", send)
			case GotoPreset:
				send = PTZGotoPreset(resp.Payload.State.Desired[GotoPreset].(float64))
				entry.Debug("转到预置位置", send)
			case RemovePreset:
				send = PTZRemovePresets(resp.Payload.State.Desired[RemovePreset].(float64))
				entry.Debug("移除预置位置", send)
			case SetHomePosition:
				send = PTZSetHomePosition()
				entry.Debug("设置Home位置", send)
			case GotoHomePosition:
				send = PTZGotoHomePosition()
				entry.Debug("转到Home位置", send)
			case TimeCalibration:
				send = DeviceSetSystemDateAndTime()
				entry.Debug("时间校准", send)
			default:
				entry.Debug("命令不存在")
			}
		}
	}
	return nil
}

// PTZ控制结果
func getPTZControlResult(Desired map[string]interface{}) error {
	id := Desired[PTZControl]
	if id == "9" {
		// 归位
		err := PTZGotoHomePosition()
		return err
	} else {
		angle := getAngle(Desired[Angle])
		upOrDown, leftOrRight, zoom := getPointer(id)
		if Desired[Zoom] == nil {
			//转向控制
			err := PTZControlMove(upOrDown, leftOrRight, zoom, angle)
			return err
		} else {
			// 缩放
			zoom, _ := strconv.Atoi(Desired[Zoom].(string))
			err := PTZControlMove(0, 0, int8(zoom), angle)
			return err
		}
	}
	return nil
}

// 移动参数
func getPointer(id interface{}) (int8, int8, int8) {
	if id != nil && reflect.TypeOf(id).String() == "string" {
		switch id {
		case "1": //左
			return 0, -1, 0
		case "2": //右
			return 0, 1, 0
		case "3": //上
			return 1, 0, 0
		case "4": //下
			return -1, 0, 0
		case "5": //左上
			return 1, -1, 0
		case "6": //左下
			return -1, -1, 0
		case "7": //右上
			return 1, 1, 0
		case "8": //右下
			return -1, 1, 0
		default:
			return 0, 0, 0
		}
	} else {
		return 0, 0, 0
	}

}

// 移动速度
func getAngle(speed interface{}) float64 {
	//1: 22.5° 2: 45° 3: 90° 4: 180°
	if speed != nil && reflect.TypeOf(speed).String() == "float64" {
		return speed.(float64) / 90.0
	} else {
		return 0.015625
	}
}

// 云台控制
func PTZControlMove(UpOrDown, LeftOrRight, Zoom int8, Angle float64) error {

	camera := &ptz.Camera{Addr: config.C.General.Addr, Username: config.C.General.Username, Password: config.C.General.Password}
	start := time.Now()
	profiles, err := camera.GetProfiles()
	if err != nil {
		return errors.Wrap(err,"PTZControlMove err")
	}
	logrus.Println("camera.GetProfiles(): ", time.Now().Sub(start))

	start = time.Now()
	resp, err := camera.PTZ_RelativeMove(UpOrDown, LeftOrRight, Zoom, Angle, profiles.Profiles.Token)
	logrus.Println("camera.PTZ_RelativeMove(): ", time.Now().Sub(start))

	res := PTZ.RelativeMoveResponse{}
	err = ptz.ParseResponse(resp, &res)
	if err != nil {
		logrus.Println(err)
		return errors.Wrap(err,"PTZControlMove err")
	}

	b, _ := json.Marshal(res)
	logrus.Println("RelativeMoveResponse:", string(b))
	return nil
}

// 快照Uri
func SnapshotUri() error {
	camera := &ptz.Camera{Addr: config.C.General.Addr, Username: config.C.General.Username, Password: config.C.General.Password}
	profiles, err := camera.GetProfiles()
	if err != nil {
		return errors.Wrap(err,"SnapshotUri err")
	}

	resp, err := camera.Media_GetSnapshotUri(profiles.Profiles.Token)
	res := Media.GetSnapshotUriResponse{}
	err = ptz.ParseResponse(resp, &res)
	if err != nil {
		return errors.Wrap(err,"SnapshotUri err")
	}

	b, _ := json.Marshal(res)
	uri := res.MediaUri.Uri
	logrus.Println("SnapshotUriResponse:", string(b))

	getSnapshot(string(uri))
	return nil
}

// 获取快照
func getSnapshot(uri string) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	request, err := http.NewRequest("GET", uri, nil)
	request.SetBasicAuth(config.C.General.Username, config.C.General.Password)
	response, err := client.Do(request)
	logrus.Println("response: ", response)
	if err != nil {
		logrus.Println("response Error,", err)
	}
	defer response.Body.Close()
	result, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logrus.Println("read response Error,", err)
	}

	fileName := fmt.Sprintf("%s_%s", config.C.General.SnapshotPath, time.Now().Format("20060102150405")+".png")
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}

	file.Write(result)
	defer file.Close()

	fid := sendSnapshot(fileName)
	if fid == "" {
		fid = "快照上传异常"
	}
	go handleResponse(fid, handleGetSnapshot)
}

// 上传快照
func sendSnapshot(fileName string) string {
	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)
	fileWriter, _ := bodyWriter.CreateFormFile("file", fileName)

	file, _ := os.Open(fileName)
	defer file.Close()

	io.Copy(fileWriter, file)

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	response, _ := http.Post(config.C.File.URL, contentType, bodyBuffer)
	defer response.Body.Close()
	result, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("read response Error,", err)
	}
	fmt.Println(string(result))
	data := FileResponse{}
	err = json.Unmarshal(result, &data)
	if err != nil {
		return data.Result.Fid
	}
	return ""
}

// 设置预置位置
func PTZSetPreset(presetToken float64) error {
	camera := &ptz.Camera{Addr: config.C.General.Addr, Username: config.C.General.Username, Password: config.C.General.Password}
	profiles, err := camera.GetProfiles()
	if err != nil {
		logrus.Println(err)
		return errors.Wrap(err,"PTZSetPreset err")
	}

	presetToken4Float := strconv.FormatFloat(presetToken, 'f', -1, 64)
	preserName := fmt.Sprintf("预置点 %s", presetToken4Float)
	resp, err := camera.PTZ_SetPreset(profiles.Profiles.Token, preserName, presetToken4Float)

	res := PTZ.SetPresetResponse{}
	err = ptz.ParseResponse(resp, &res)
	if err != nil {
		logrus.Println(err)
		return errors.Wrap(err,"PTZSetPreset err")
	}

	b, _ := json.Marshal(res)
	logrus.Println("SetPresetResponse:", string(b))
	return nil
}

// 获取预置位置
func PTZGetPresets() error {
	camera := &ptz.Camera{Addr: config.C.General.Addr, Username: config.C.General.Username, Password: config.C.General.Password}
	profiles, err := camera.GetProfiles()
	if err != nil {
		return errors.Wrap(err,"PTZGetPresets err")
	}

	resp, err := camera.PTZ_GetPresets(profiles.Profiles.Token)
	message, _ := ptz.GetSoapMessage(resp)
	logrus.Println(message)

	res := PTZ.GetPresetsResponse{}
	err = ptz.ParseResponse(resp, &res)
	if err != nil {
		logrus.Println(err)
		return errors.Wrap(err,"PTZGetPresets err")
	}

	b, _ := json.Marshal(res)
	logrus.Println("PTZGetPresets:", string(b))
	return nil
}

// 移除预置位置
func PTZRemovePresets(presetToken float64) error {
	camera := &ptz.Camera{Addr: config.C.General.Addr, Username: config.C.General.Username, Password: config.C.General.Password}
	profiles, err := camera.GetProfiles()
	if err != nil {
		logrus.Println(err)
		return errors.Wrap(err,"PTZRemovePresets err")
	}

	resp, err := camera.PTZ_RemovePreset(profiles.Profiles.Token, strconv.FormatFloat(presetToken, 'f', -1, 64))

	res := PTZ.RemovePresetResponse{}
	err = ptz.ParseResponse(resp, &res)
	if err != nil {
		logrus.Println(err)
		return errors.Wrap(err,"PTZRemovePresets err")
	}

	b, _ := json.Marshal(res)
	logrus.Println("PTZRemovePresets:", string(b))
	return nil
}

// 回到预置位置
func PTZGotoPreset(presetToken float64) error {
	camera := &ptz.Camera{Addr: config.C.General.Addr, Username: config.C.General.Username, Password: config.C.General.Password}
	profiles, err := camera.GetProfiles()
	if err != nil {
		return errors.Wrap(err,"GetProfiles err")
	}

	resp, err := camera.PTZ_GotoPreset(profiles.Profiles.Token, strconv.FormatFloat(presetToken, 'f', -1, 64))
	res := PTZ.GotoPresetResponse{}
	err = ptz.ParseResponse(resp, &res)
	if err != nil {
		return errors.Wrap(err,"PTZ_GotoPreset err")
	}
	b, _ := json.Marshal(res)
	logrus.Println("PTZGotoPresetResponse:", string(b))
	return nil
}

// 设置Home位置
func PTZSetHomePosition() error {
	camera := &ptz.Camera{Addr: config.C.General.Addr, Username: config.C.General.Username, Password: config.C.General.Password}
	profiles, err := camera.GetProfiles()
	if err != nil {
		return errors.Wrap(err,"PTZSetHomePosition err")
	}

	resp, err := camera.PTZ_SetHomePosition(profiles.Profiles.Token)

	res := PTZ.SetHomePositionResponse{}
	err = ptz.ParseResponse(resp, &res)
	if err != nil {
		return errors.Wrap(err,"PTZSetHomePosition err")
	}

	b, _ := json.Marshal(res)
	logrus.Println("SetHomePosition:", string(b))
	return nil
}

// 转到Home位置
func PTZGotoHomePosition() error {
	camera := &ptz.Camera{Addr: config.C.General.Addr, Username: config.C.General.Username, Password: config.C.General.Password}
	profiles, err := camera.GetProfiles()
	if err != nil {
		return errors.Wrap(err,"PTZGotoHomePosition err")
	}

	resp, err := camera.PTZ_GotoHomePosition(profiles.Profiles.Token)

	res := PTZ.GotoHomePositionResponse{}
	err = ptz.ParseResponse(resp, &res)
	if err != nil {
		return errors.Wrap(err,"PTZGotoHomePosition err")
	}

	b, _ := json.Marshal(res)
	logrus.Println("GotoHomePositionResponse:", string(b))
	return nil
}

// 时间校准
func DeviceSetSystemDateAndTime() error {
	camera := &ptz.Camera{Addr: config.C.General.Addr, Username: config.C.General.Username, Password: config.C.General.Password}
	//设置时区： CST-8 东八区
	timeZone := onvif.TimeZone{TZ: xsd.Token("CST-0")}
	//设置时间
	now := time.Now()
	resp, err := camera.Device_SetSystemDateAndTime("Manual", false, timeZone, now)
	res := Device.SetSystemDateAndTimeResponse{}

	err = ptz.ParseResponse(resp, &res)
	if err != nil {
		return errors.Wrap(err,"SetSystemDateAndTime err")
	}

	b, _ := json.Marshal(res)
	logrus.Println("SetSystemDateAndTimeResponse:", string(b))

	formatTime := now.Format("2006-01-02 15:04:05")
	go handleResponse(formatTime, handleSetSystemDateAndTime)

	return nil
}

