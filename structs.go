package camera

import (
	"encoding/json"
	"time"
)

type AckPacket struct {
	CommandID string `json:"command_id"`
	Status    string `json:"status"`
	Result    string `json:"result"`
}

const (
	Timeout = "timeout"
	Receive = "receive"
	Execute = "execute"

	Reply   = "reply"
	Update  = "update"
	Get     = "get"
	Delete  = "delete"
	Control = "control"
)

type DevicePayload struct {
	Did     string
	Payload []byte
}
type NodePayload struct {
	GwID    string
	Did     string
	Payload []byte
}

/*RequestTwins请求信息格式*/
type RequestTwins struct {
	Method  string `json:"method,omitempty"`
	State   *State `json:"state,omitempty"`
	Version int64  `json:"version,omitempty"`
}

func (request *RequestTwins) UnmarshalJSONText(data []byte) error {
	return json.Unmarshal(data, &request)
}

func (request *RequestTwins) MarshalJSONText() ([]byte, error) {
	return json.Marshal(request)
}

/*数字孪生的结构体*/
type DBTwins struct {
	State     State    `json:"state,omitempty"`
	MetaData  Metadata `json:"meta_data,omitempty"`
	Timestamp int64    `json:"timestamp"` /*影子文档的最新更新时间*/
	Version   int64    `json:"version"`   /*影子文档的版本信息*/
}

type Twins struct {
	Tid          int32     `json:"tid"`
	Pid          int32     `json:"pid"`
	Did          string    `json:"did"`
	Cid          string    `json:"cid"`
	DigitalTwins DBTwins   `json:"digital"`
	Timestamp    time.Time `json:"timestamp"`
}

/*ResponseTwins返回信息*/
type ResponseTwins struct {
	Method    string  `json:"method,omitempty"`
	Payload   Payload `json:"payload,omitempty"`
	Timestamp int64   `json:"timestamp,omitempty"`
	Version   int64   `json:"version,omitempty"`
	CommandID string  `json:"command_id,omitempty"`
}

func (response *ResponseTwins) UnmarshalJSONText(data []byte) error {
	return json.Unmarshal(data, &response)
}

func (response *ResponseTwins) MarshalJSONText() ([]byte, error) {
	return json.Marshal(response)
}

type SendResult struct {
	Code    int32  `json:"code,omitempty"`
	Message string `json:"msg,omitempty"`
	Result  Result `json:"result,omitempty"`
}
type Result struct {
	Fid string `json:"fid,omitempty"`
}

/*Payload消息负载*/
type Payload struct {
	Status   string   `json:"status,omitempty"`
	Content  Content  `json:"content,omitempty"`
	State    State    `json:"state,omitempty"`
	MetaData Metadata `json:"meta_data,omitempty"`
}

/*Content 更新失败返回信息*/
type Content struct {
	ErrorCode    string `json:"errorcode,omitempty"`
	ErrorMessage string `json:"errormessage,omitempty"`
}

/*数字孪生中的状态*/
type State struct {
	Reported map[string]interface{} `json:"reported,omitempty"` /*报告状态*/
	Desired  map[string]interface{} `json:"desired,omitempty"`  /*预期状态*/
}

/*数字孪生中元数据*/
type Metadata struct {
	Reported map[string]Meta `json:"reported,omitempty"`
	Desired  map[string]Meta `json:"desired,omitempty"`
}

/*元数据的时间戳*/
type Meta struct {
	Timestamp int64 `json:"timestamp"`
}

type Message struct {
	Tid    int32                  `json:"tid"`
	Pid    int32                  `json:"pid"`
	Did    string                 `json:"did"`
	Values map[string]interface{} `json:"values"`
	Time   int64                  `json:"time"` // TODO 前端使用s时间戳，等所有人都改过来，以后可以删除改字段
	Now    int64                  `json:"now"`  // 精确到毫秒的时间
}

type FileResponse struct {
	Code   string     `json:"code"`
	Msg    string     `json:"msg"`
	Result FileResult `json:"result"`
}

type FileResult struct {
	Fid string `json:"fid"`
}

var ErrorCode = map[string]string{
	"400": "不正确的json格式",
	"401": "影子json缺少method信息",
	"402": "影子json缺少state字段",
	"403": "影子json reported属性字段为空",
	"404": "影子json method是无效的方法",
	"405": "影子reported属性个数超过128个",
	"406": "影子版本冲突",
	"500": "服务端处理异常",
}
