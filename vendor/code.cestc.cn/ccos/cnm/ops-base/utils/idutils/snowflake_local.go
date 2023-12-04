package idutils

import (
	"sync"
	"sync/atomic"
	"time"
)

type snowflakeLocal struct {
	epoch      time.Time
	step       int64
	time       int64
	machinedId int32
	count      int64
	mu         sync.Mutex
}

func newSnowflakeLocal(machinedId int32, sceneType int64) IdExec {
	s := &snowflakeLocal{
		machinedId: machinedId,
		mu:         sync.Mutex{},
	}
	var curTime = time.Now()
	// 如果是region内唯一，不需要region位
	if sceneType == SceneTypeRegion {
		// 机器位只需要5bit
		machinedShift = seqBits + regionBits
		// 序列号最大可以到 1<<27-1
		seqMax = -1 ^ (-1 << (machinedShift))
	}
	s.epoch = curTime.Add(time.Unix(Epoch, 0).Sub(curTime))
	return s
}

func (s *snowflakeLocal) generate() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := int64(time.Since(s.epoch).Seconds())

	if now == s.time {
		s.step = (s.step + 1) & seqMax
		count := atomic.AddInt64(&s.count, 1)
		if count == seqMax+1 {
			for now <= s.time {
				now = int64(time.Since(s.epoch).Seconds())
				s.count = 0
			}
		}
	} else {
		s.count = 0
	}

	s.time = now

	return (now << timeShift) | (int64(s.machinedId) << machinedShift) | s.step
}
