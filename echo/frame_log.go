package echo

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"iot-hub/api/gw"
	"iot-hub/internal/lorawan/band"
	"time"
)

type FrameType string

const (
	Join FrameType = "Join"
	Up   FrameType = "Up"
	Down FrameType = "Down"
)

const DeLiveKey = "lora:de:live:frame"

type UpLink struct {
	TxInfo  UpLinkTxInfo `json:"tx_info"`
	RxInfo  RXInfoSet    `json:"rx_info"`
	Payload string       `json:"payload"`
}

type DownLink struct {
	TxInfo  DownLinkTxInfo `json:"tx_info"`
	Payload string         `json:"payload"`
}

type Frame struct {
	Type       FrameType `json:"type"`
	DevEUI     string    `json:"dev_eui"`
	Uplink     UpLink    `json:"uplink"`
	Downlink   DownLink  `json:"downlink"`
	PHYPayload []byte    `json:"phy_payload"`
}

type UpLinkTxInfo struct {
	MAC       string        `json:"mac"`
	Timestamp uint32        `json:"timestamp"`
	Frequency int           `json:"frequency"` // frequency in Hz
	DataRate  band.DataRate `json:"dataRate"`  // TX datarate (either LoRa or FSK)
	CodeRate  string        `json:"codeRate"`  // ECC code rate
}

type DownLinkTxInfo struct {
	MAC         string        `json:"mac"`
	Immediately bool          `json:"immediately"`
	Timestamp   uint32        `json:"timestamp"`
	Frequency   int           `json:"frequency"`
	Power       int           `json:"power"`
	DataRate    band.DataRate `json:"dataRate"`
	CodeRate    string        `json:"codeRate"`
	IPol        *bool         `json:"i_pol"`
}

func LogUpLinkFrame(pool *redis.Pool, rxPacket RXPacket, frameType FrameType, payload string) error {
	c := pool.Get()
	defer c.Close()

	phy, _ := rxPacket.PHYPayload.MarshalBinary()
	frame := Frame{
		DevEUI: rxPacket.DevEUI.String(),
		Type:   frameType,
		Uplink: UpLink{
			TxInfo: UpLinkTxInfo{
				MAC:       rxPacket.RXInfoSet[0].MAC.String(),
				Timestamp: uint32(time.Now().Unix()),
				Frequency: rxPacket.RXInfoSet[0].Frequency,
				CodeRate:  rxPacket.RXInfoSet[0].CodeRate,
				DataRate: band.DataRate{
					Modulation:   rxPacket.RXInfoSet[0].DataRate.Modulation,
					Bandwidth:    rxPacket.RXInfoSet[0].DataRate.Bandwidth,
					SpreadFactor: rxPacket.RXInfoSet[0].DataRate.SpreadFactor,
					BitRate:      rxPacket.RXInfoSet[0].DataRate.BitRate,
				},
			},
			RxInfo:  rxPacket.RXInfoSet,
			Payload: payload,
		},
		PHYPayload: phy,
	}
	bytes, err := json.Marshal(frame)
	if err != nil {
		return errors.Wrap(err, "json encode error")
	}
	if err := c.Send("PUBLISH", DeLiveKey, bytes); err != nil {
		logrus.Errorf("redis publish %s,%s", DeLiveKey, string(bytes))
		return errors.Wrap(err, "publish frame to gateway channel error")
	}
	return nil
}

func LogDownLinkFrame(pool *redis.Pool, devEUI string, txPacket gw.TXPacket, frameType FrameType, payload string) error {
	c := pool.Get()
	defer c.Close()

	phy, _ := txPacket.PHYPayload.MarshalBinary()
	frame := Frame{
		Type:   frameType,
		DevEUI: devEUI,
		Downlink: DownLink{
			TxInfo: DownLinkTxInfo{
				MAC:         txPacket.TXInfo.MAC.String(),
				Immediately: txPacket.TXInfo.Immediately,
				Timestamp:   uint32(time.Now().Unix()),
				Frequency:   txPacket.TXInfo.Frequency,
				Power:       txPacket.TXInfo.Power,
				DataRate: band.DataRate{
					Modulation:   txPacket.TXInfo.DataRate.Modulation,
					Bandwidth:    txPacket.TXInfo.DataRate.Bandwidth,
					SpreadFactor: txPacket.TXInfo.DataRate.SpreadFactor,
					BitRate:      txPacket.TXInfo.DataRate.BitRate,
				},
				CodeRate: txPacket.TXInfo.CodeRate,
				IPol:     txPacket.TXInfo.IPol,
			},
			Payload: payload,
		},
		PHYPayload: phy,
	}
	bytes, err := json.Marshal(frame)
	if err != nil {
		return errors.Wrap(err, "json encode error")
	}
	if err := c.Send("PUBLISH", DeLiveKey, bytes); err != nil {
		logrus.Errorf("redis publish %s,%s", DeLiveKey, string(bytes))
		return errors.Wrap(err, "publish frame to gateway channel error")
	}

	return nil
}
