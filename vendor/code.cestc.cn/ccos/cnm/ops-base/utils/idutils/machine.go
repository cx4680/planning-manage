package idutils

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"code.cestc.cn/ccos/cnm/ops-base/logging"
	"code.cestc.cn/ccos/cnm/ops-base/opshttp/client"
)

type machined struct {
	serviceName string
	host        string
	machineId   int32
	sceneType   int64
}

func newMachined(serviceName, host string, sceneType int64) *machined {
	m := &machined{
		serviceName: serviceName,
		host:        host,
		sceneType:   sceneType,
	}
	id, err := m.getMachineId()
	if err != nil {
		logging.Fatal(err)
	}
	go m.heartbeat()
	m.machineId = id
	return m
}

func (m *machined) getMachineId() (int32, error) {
	type remoteMachineId struct {
		ID int32 `json:"id"`
	}

	url := getUrl()

	var reply remoteMachineId
	err := client.NewReq(context.Background(), client.DefaultClient()).Get(fmt.Sprintf("%s%s?serviceName=%s&host=%s&sceneType=%d", url, remoteMachinedPath, m.serviceName, m.host, m.sceneType)).Response().ParseJson(&reply)
	if err != nil {
		return 0, err
	}
	return reply.ID, nil
}

func (m *machined) heartbeat() {
	t := time.NewTicker(10 * time.Second)
	for {
		<-t.C
		id, err := m.getMachineId()
		if err != nil {
			logging.Errorw("m.getMachineId error", zap.Error(err))
		} else {
			m.machineId = id
		}
	}
}
