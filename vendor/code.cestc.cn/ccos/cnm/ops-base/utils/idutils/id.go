package idutils

import (
	"fmt"
	"os"
	"sync"

	"code.cestc.cn/ccos/cnm/ops-base/logging"
)

const (
	// region位
	regionBits uint8 = 10

	// pod位
	podBits uint8 = 5

	// 序号位
	seqBits uint8 = 17
)

const (
	TypeSnowflakeLocal  = iota + 1 // 本地获取雪花
	TypeSnowflakeRemote            // 远程获取雪花（基于redis解决了时钟回拨问题）
	TypeIdSegment                  // 号段模式
)

const (
	SceneTypeOverall = iota + 1 // 全局唯一
	SceneTypeRegion             // region内唯一
)

var (
	// Epoch 基础时间， 可以更改
	Epoch int64 = 1288834974

	// 最大序列号
	seqMax int64 = -1 ^ (-1 << seqBits)

	// 机器左移位数
	machinedShift = seqBits

	// 时间位
	timeShift = regionBits + podBits + seqBits

	once = sync.Once{}

	mu = sync.Mutex{}

	instances = idCache{}
)

type idCache map[string]*Id

func (i idCache) Get(serviceName string, sceneType int64) (*Id, bool) {
	value, ok := i[fmt.Sprintf("%s-%d", serviceName, sceneType)]
	return value, ok
}

func (i idCache) Insert(serviceName string, sceneType int64, value *Id) {
	i[fmt.Sprintf("%s-%d", serviceName, sceneType)] = value
}

type IdExec interface {
	generate() int64
}

type Id struct {
	machined     *machined
	idExec       IdExec
	serviceName  string
	sceneType    int64
	generateType int64
	host         string
}

func New(serviceName string, sceneType int64) *Id {
	mu.Lock()
	defer mu.Unlock()

	i, ok := instances.Get(serviceName, sceneType)
	if ok {
		return i
	}

	host, _ := os.Hostname()
	i = &Id{
		serviceName:  serviceName,
		sceneType:    sceneType,
		generateType: TypeSnowflakeLocal, // 写死,现在只需要本地雪花模式
		host:         host,
	}

	// 设置机器位（号段模式下不设置）
	i.setMachined()

	// 获取执行器
	i.idExec = i.getExec()

	// 保存实例
	instances.Insert(serviceName, sceneType, i)

	return i
}

func (i *Id) Generate() int64 {
	return i.idExec.generate()
}

func (i *Id) setMachined() {
	// 号段模式不需要获取机器位
	if i.generateType != TypeIdSegment {
		once.Do(func() {
			i.machined = newMachined(i.serviceName, i.host, i.sceneType)
		})
	}
}

func (i *Id) getExec() IdExec {
	switch i.generateType {
	case TypeSnowflakeLocal:
		if i.machined == nil {
			logging.Fatalf("i.machined is nil", i.generateType)
		}
		return newSnowflakeLocal(i.machined.machineId, i.sceneType)
	case TypeSnowflakeRemote:
		return newSnowflakeRemote(i.serviceName, i.host)
	case TypeIdSegment:
		return newIdSegment(i.serviceName, i.sceneType)
	default:
		logging.Fatalf("generate type error:%d", i.generateType)
	}
	return nil
}
