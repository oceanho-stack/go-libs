// Package xid 唯一 ID 生成器：接口化设计，提供雪花算法实现。
package xid

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

// Generator 唯一 ID 生成接口。
// 业务代码应依赖此接口，便于切换实现（雪花 / UUID / 自增等）。
type Generator interface {
	// NextID 返回下一个唯一 ID
	NextID() uint64
	// NextIDString 返回下一个唯一 ID 的十进制字符串
	NextIDString() string
}

// 雪花算法常量
const (
	snowEpoch     int64 = 1704067200000 // 起始时间戳：2024-01-01 00:00:00 UTC
	timeBits      uint8 = 41            // 时间戳位数
	machineBits   uint8 = 10            // 机器ID位数
	sequenceBits  uint8 = 12            // 序列号位数
	maxMachine    int64 = -1 ^ (-1 << machineBits)
	maxSequence   int64 = -1 ^ (-1 << sequenceBits)
	timeShift           = machineBits + sequenceBits
	machineShift        = sequenceBits
)

// Snowflake 雪花算法 ID 生成器。
// 不依赖全局时间；多实例通过不同 machineID 保证不冲突。
type Snowflake struct {
	mu        sync.Mutex
	machineID int64
	lastTime  int64
	sequence  int64
}

// NewSnowflake 创建雪花生成器。
// machineID 范围 [0, 1023]，同一时刻不同实例必须不同。
func NewSnowflake(machineID int64) (*Snowflake, error) {
	if machineID < 0 || machineID > maxMachine {
		return nil, errors.New("xid: machineID 必须在 [0,1023] 范围")
	}
	return &Snowflake{
		machineID: machineID,
		lastTime:  -1,
	}, nil
}

// NextID 生成下一个雪花 ID
func (s *Snowflake) NextID() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli() - snowEpoch

	if now < s.lastTime {
		// 时钟回拨：等系统时钟追平（简化处理，避免产生重复ID）
		for now < s.lastTime {
			now = time.Now().UnixMilli() - snowEpoch
		}
	}

	if now == s.lastTime {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			// 当前毫秒序列号耗尽，等待下一毫秒
			for now <= s.lastTime {
				now = time.Now().UnixMilli() - snowEpoch
			}
		}
	} else {
		s.sequence = 0
	}

	s.lastTime = now

	id := (now << timeShift) |
		(s.machineID << machineShift) |
		s.sequence
	return uint64(id)
}

// NextIDString 返回十进制字符串
func (s *Snowflake) NextIDString() string {
	return strconv.FormatUint(s.NextID(), 10)
}
