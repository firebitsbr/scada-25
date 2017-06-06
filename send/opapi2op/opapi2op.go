package opapi2op

import (
	"errors"
	"global"
	"opapi4"
	"scada/points"
	"time"

	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
)

const (
	DefaultValueType    = 4 //R8_TYPE
	DefaultSendInterval = 15
	MaxPointSend        = 1000
)

func Run(conf config.Configer) (err error) {
	if conf.String("destination_address") == "" {
		err = errors.New("Empty adress for destination")
		return
	}

	go func() {
		conn := opapi4.NewConnect(
			conf.String("destination_address"),
			conf.String("destination_user_name"),
			conf.String("destination_user_password"),
			conf.DefaultInt("destination_option", 0))

		idx := 0
		ids := make([]int, MaxPointSend)
		vals := make([]float64, MaxPointSend)
		status := make([]uint16, MaxPointSend)
		times := make([]int, MaxPointSend)
		for global.IsRunning() {
			select {
			case <-time.After(time.Second):
				if idx > 0 {
					err := conn.WriteValueTimeArray(DefaultValueType, ids[:idx], vals[:idx], status[:idx], times[:idx])
					//log.Println("Send number:", idx)
					idx = 0
					if err != nil {
						logs.Error("Send error", err)
					}
				}
			case point := <-points.GetQueuePoint():
				//logs.Info(point)
				ids[idx] = int(point.ID)
				vals[idx] = point.Value
				status[idx] = point.Status
				times[idx] = int(point.Time)
				idx++
				if idx >= MaxPointSend {
					err := conn.WriteValueTimeArray(DefaultValueType, ids[:idx], vals[:idx], status[:idx], times[:idx])
					//log.Println("Send number:", idx)
					idx = 0
					if err != nil {
						logs.Error("Send error", err)
					}
				}
			}
		}
	}()
	return
}
