package idutils

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"code.cestc.cn/ccos/cnm/ops-base/logging"
	"code.cestc.cn/ccos/cnm/ops-base/opshttp/client"
)

const (
	envRegionManagementCenterUrl = "REGION_MANAGEMENT_URL"
	getRegionListPath            = "/api/region-management/v1/region?size=10000"
	maxStepSize                  = 1e5

	maxRetryTimes = 100 // 重试100次还是获取不到id，抛出异常
)

var (
	regionManagementCenterUrl = "https://region-management"
)

var (
	StepSize int64 = 1000
)

const (
	segmentRegionBits = 8
	segmentSeqBits    = 24
)

var (
	segmentRegionShift = segmentSeqBits
	segmentTimeShift   = segmentRegionBits + segmentSeqBits
)

type idSegment struct {
	serviceName     string
	stepSize        int64
	signallingQueue chan int64
	cacheQueue      chan int64
	regionId        int64
	epoch           time.Time
	isFillCache     int32
	sceneType       int64
}

func newIdSegment(serviceName string, sceneType int64) IdExec {
	if StepSize > maxStepSize {
		logging.Fatal("step size too long")
	}
	i := &idSegment{
		serviceName:     serviceName,
		stepSize:        StepSize,
		signallingQueue: make(chan int64, StepSize),
		cacheQueue:      make(chan int64, StepSize),
		sceneType:       sceneType,
	}

	var curTime = time.Now()
	i.epoch = curTime.Add(time.Unix(Epoch/1000, (Epoch%1000)*1000000).Sub(curTime))

	i.init()
	if sceneType == SceneTypeOverall {
		i.regionId = i.getRegion()
	}

	// 启动从备用往当前channel中填充
	go i.fillSignallingQueue()
	return i
}

func (i *idSegment) generate() int64 {
	var id int64

	// channel中获取生成的id
	for k := 0; k < maxRetryTimes; k++ {
		select {
		case id = <-i.signallingQueue:
			goto next
		default:
			// 兜底，如果号段不够，等待10ms
			time.Sleep(10 * time.Millisecond)
		}
	}

next:

	if id == 0 {
		logging.Fatal("get id empty")
	}

	// 当前channel剩一半的时候去填充备用channel
	if len(i.signallingQueue) <= int(i.stepSize)/2 && len(i.cacheQueue) <= 0 && atomic.CompareAndSwapInt32(&i.isFillCache, 0, 1) {
		go i.fillCacheQueue()
	}
	now := int64(time.Since(i.epoch).Seconds())

	return (now << segmentTimeShift) | (i.regionId << segmentRegionShift) | id
}

func (i *idSegment) init() {
	ids := i.getSegment()
	for _, v := range ids {
		i.signallingQueue <- v
	}
}

func (i *idSegment) fillCacheQueue() {
	ids := i.getSegment()
	for _, v := range ids {
		i.cacheQueue <- v
	}
	i.isFillCache = 0
}

func (i *idSegment) fillSignallingQueue() {
	for {
		id := <-i.cacheQueue
		i.signallingQueue <- id
	}
}

func (i *idSegment) getSegment() []int64 {
	type remoteId struct {
		MinNum int64 `json:"minNum"`
		MaxNum int64 `json:"maxNum"`
	}

	url := getUrl()

	var reply remoteId
	err := client.NewReq(context.Background(), client.DefaultClient()).Get(fmt.Sprintf("%s%s?serviceName=%s&stepSize=%d&sceneType=%d", url, IdSegmentPath, i.serviceName, i.stepSize, i.sceneType)).Response().ParseJson(&reply)
	if err != nil {
		logging.Fatal(err)
	}

	var list []int64
	for k := reply.MinNum; k <= reply.MaxNum; k++ {
		list = append(list, k)
	}

	return list
}

func (i *idSegment) getRegion() int64 {

	type getRegionListReply struct {
		List []struct {
			RegionId      string `json:"regionId"`
			RegionName    string `json:"regionName"`
			Domain        string `json:"domain"`
			CentralRegion bool   `json:"isCenterRegion"`
		} `json:"list"`
	}

	url := os.Getenv(envRegionManagementCenterUrl)
	cm := getInstallConfig()
	if len(url) <= 0 {
		url = fmt.Sprintf("%s.%s", regionManagementCenterUrl, cm.GlobalBaseDomain)
	}

	var reply getRegionListReply

	err := client.NewReq(context.Background(), client.DefaultClient()).Get(fmt.Sprintf("%s%s", url, getRegionListPath)).Response().ParseJson(&reply)
	if err != nil {
		logging.Fatal(err)
	}

	for k, v := range reply.List {
		if v.RegionId == cm.Region {
			return int64(k)
		}
	}

	return 0
}
