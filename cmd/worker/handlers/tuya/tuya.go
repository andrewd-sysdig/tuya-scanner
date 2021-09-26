package tuya

import (
	"encoding/json"
	"log"
	"time"

	"github.com/SysdigDan/tuya-scanner/cmd/worker/handlers/exporter"
	"github.com/SysdigDan/tuya-scanner/cmd/worker/handlers/mqtt"
	"github.com/SysdigDan/tuya-scanner/cmd/worker/models"
	"github.com/SysdigDan/tuya-scanner/pkg/application"
	"github.com/SysdigDan/tuya-scanner/pkg/tuya-api"
)

type Device struct {
	Dps struct {
		Num1  bool    `json:"1"`
		Num9  int     `json:"9"`
		Num18 float64 `json:"18"`
		Num19 float64 `json:"19"`
		Num20 float64 `json:"20"`
		Num21 int     `json:"21"`
		Num22 int     `json:"22"`
		Num23 int     `json:"23"`
		Num24 int     `json:"24"`
		Num25 int     `json:"25"`
		Num26 int     `json:"26"`
		Num41 string  `json:"41"`
		Num42 string  `json:"42"`
	} `json:"dps"`
}

type DeviceConfig []struct {
	GwID string `json:"gwId"`
	Key  string `json:"key"`
	Type string `json:"type"`
	Name string `json:"name"`
}

var dm *tuya.DeviceManager
var devList []tuya.Device

func TuyaScanner(app *application.Application) {
	log.Println("[tuya] Loading devices from configuration...")
	devices := &app.Devices
	d, _ := json.Marshal(devices)

	log.Println("[tuya] Scanning for Tuya devices...")
	dm = tuya.NewDeviceManager(string(d))
	keys := dm.DeviceKeys()
	devList = make([]tuya.Device, 0)

	for _, k := range keys {
		d, _ := dm.GetDevice(k)
		devList = append(devList, d)
	}

	for {
		for _, b := range devList {
			s, _ := b.(tuya.Switch)

			// need to tell devices to refresh dps
			_, err := s.TuyaRefresh(1 * time.Second)
			if err != nil {
				log.Printf("[tuya] Device %s - %s", b.Name(), err)
			}
			// get status on devices
			status, err := s.TuyaGetStatus(5 * time.Second)
			if err != nil {
				log.Printf("[tuya] Device %s - %s", b.Name(), err)
			}

			if len(status) != 0 {
				// parse data for processing
				data, err := parseTuyaData(app, b.Name(), status)
				if err != nil {
					log.Println("[tuya] error parsing sensor data:", err.Error())
				}
				mqtt.Publish(app, data)
				exporter.LogPrometheusData(data.Name, data.Switch, data.Power_mA, data.Power_W, data.Power_V)
			}
		}
		time.Sleep(time.Duration(app.Cfg.Frequency) * time.Second)
	}
}

func parseTuyaData(app *application.Application, name string, data []byte) (*models.SensorData, error) {
	var d Device
	if err := json.Unmarshal(data, &d); err != nil {
		log.Println("[debug] Error in parsing data", err)
		return nil, err
	}
	return &models.SensorData{
		Name:     name,
		Switch:   d.Dps.Num1,
		Power_mA: d.Dps.Num18,
		Power_W:  d.Dps.Num19 / 10,
		Power_V:  d.Dps.Num20 / 10,
		State:    string(data),
	}, nil
}
