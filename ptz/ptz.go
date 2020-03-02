package ptz

import (
	"camera/goonvif/PTZ"
	"camera/goonvif/xsd"
	"camera/goonvif/xsd/onvif"
	"net/http"
)

func (c *Camera) PTZ_RelativeMove(UpOrDown, LeftOrRight, Zoom int8, Angle float64, token onvif.ReferenceToken) (*http.Response, error) {

	X := 0.0
	Y := 0.0
	Z := 0.0

	displacement := 0.5 * Angle
	switch LeftOrRight {
	case -1:
		X = -displacement
	case 0:
	case 1:
		X = displacement
	}

	switch UpOrDown {
	case -1:
		Y = -displacement
	case 0:
	case 1:
		Y = displacement
	}

	switch Zoom {
	case -1:
		Z = -displacement
	case 0:
	case 1:
		Z = displacement
	}

	//PTZ 控制
	RelMove := PTZ.RelativeMove{
		ProfileToken: token,
		Translation: onvif.PTZVector{
			PanTilt: onvif.Vector2D{ // x为负数，表示左转，x为正数，表示右转 y为负数，表示下转，y为正数，表示上转
				X:     X,
				Y:     Y,
				Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/PanTiltSpaces/TranslationGenericSpace"),
			},
			Zoom: onvif.Vector1D{ // x  范围也在0--1之间 x为正数表示拉近，x为负数，表示拉远
				X:     Z, // -1.0 -> 1.0
				Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/ZoomSpaces/TranslationGenericSpace"),
			},
		},
		// 移动速度固定
		Speed: onvif.PTZSpeed{
			PanTilt: onvif.Vector2D{ // X,Y 绝对值表示速度 0~1
				X:     0.5,
				Y:     0.5,
				Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/PanTiltSpaces/GenericSpeedSpace"),
			},
			Zoom: onvif.Vector1D{ // X 绝对值表示速度 0~1
				X:     0.5,
				Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/ZoomSpaces/ZoomGenericSpeedSpace"),
			},
		},
	}

	return c.Call(RelMove)
}

func (c *Camera) PTZ_GetStaus(token onvif.ReferenceToken) (*http.Response, error) {
	GetStatus := PTZ.GetStatus{ProfileToken: token}
	return c.Call(GetStatus)
}

//设置预置位置
func (c *Camera) PTZ_SetPreset(token onvif.ReferenceToken, presetName string, presetToken string) (*http.Response, error) {
	SetPreset := PTZ.SetPreset{ProfileToken: token, PresetName: xsd.String(presetName), PresetToken: onvif.ReferenceToken(presetToken)}
	return c.Call(SetPreset)
}

//获取预置位置
func (c *Camera) PTZ_GetPresets(token onvif.ReferenceToken) (*http.Response, error) {
	GetPresets := PTZ.GetPresets{ProfileToken: token}
	return c.Call(GetPresets)
}

//移除预置位置
func (c *Camera) PTZ_RemovePreset(token onvif.ReferenceToken, presetToken string) (*http.Response, error) {
	RemovePreset := PTZ.RemovePreset{ProfileToken: token, PresetToken: onvif.ReferenceToken(presetToken)}
	return c.Call(RemovePreset)
}

//转到预置位置
func (c *Camera) PTZ_GotoPreset(token onvif.ReferenceToken, presetToken string) (*http.Response, error) {
	GotoPreset := PTZ.GotoPreset{ProfileToken: token, PresetToken: onvif.ReferenceToken(presetToken), Speed: onvif.PTZSpeed{
		PanTilt: onvif.Vector2D{ // X,Y 绝对值表示速度 0~1
			X:     0.5,
			Y:     0.5,
			Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/PanTiltSpaces/GenericSpeedSpace"),
		},
		Zoom: onvif.Vector1D{ // X 绝对值表示速度 0~1
			X:     0.5,
			Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/ZoomSpaces/ZoomGenericSpeedSpace"),
		},
	}}
	return c.Call(GotoPreset)
}

//设置home位置
func (c *Camera) PTZ_SetHomePosition(token onvif.ReferenceToken) (*http.Response, error) {
	SetHomePosition := PTZ.SetHomePosition{ProfileToken: token}
	return c.Call(SetHomePosition)
}

//转到home位置
func (c *Camera) PTZ_GotoHomePosition(token onvif.ReferenceToken) (*http.Response, error) {
	GotoHomePosition := PTZ.GotoHomePosition{ProfileToken: token, Speed: onvif.PTZSpeed{
		PanTilt: onvif.Vector2D{ // X,Y 绝对值表示速度 0~1
			X:     0.5,
			Y:     0.5,
			Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/PanTiltSpaces/GenericSpeedSpace"),
		},
		Zoom: onvif.Vector1D{ // X 绝对值表示速度 0~1
			X:     0.5,
			Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/ZoomSpaces/ZoomGenericSpeedSpace"),
		},
	}}
	return c.Call(GotoHomePosition)
}

//创建巡航
func (c *Camera) PTZ_CreatePresetTour(token onvif.ReferenceToken) (*http.Response, error) {
	CreatePresetTour := PTZ.CreatePresetTour{ProfileToken: token}
	return c.Call(CreatePresetTour)
}

//执行巡航
func (c *Camera) PTZ_OperatePresetTour(token onvif.ReferenceToken, presetToken string) (*http.Response, error) {
	OperatePresetTour := PTZ.OperatePresetTour{ProfileToken: token, PresetTourToken: onvif.ReferenceToken(presetToken)}
	return c.Call(OperatePresetTour)
}

//移除巡航
func (c *Camera) PTZ_RemovePresetTour(token onvif.ReferenceToken, presetToken string) (*http.Response, error) {
	RemovePresetTour := PTZ.RemovePresetTour{ProfileToken: token, PresetTourToken: onvif.ReferenceToken(presetToken)}
	return c.Call(RemovePresetTour)
}
