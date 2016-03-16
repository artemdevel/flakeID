//Package flakeID implements several approaches to generate IDs similar to Twitter's Snowflake IDs.
package flakeID

import (
	"math/rand"
	"sync"
	"time"
)

const (
	timestampShift  = 23
	hostIDShift     = 10
	randomBitsMask  = 0x7FFFFF
	counterBitsMask = 0x3FF
	hostIDMask      = 0x3FF
)

var (
	defaultEpoch = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	_rand        *rand.Rand
)

// Flaker interface represents set of methods for flake ID generators.
type Flaker interface {
	Next() uint64
	// Parse returns time when ID was generated, host ID and counter or random bits.
	// For RandomFlake generator host ID is always 0.
	Parse(flakeID uint64) (time.Time, uint32, uint32)
}

// RandomFlake represents a structure which is used to generate an ID as time delta between
// the initial epoch time and current time + some random value.
type RandomFlake struct {
	epochTime time.Time
	value     uint64
}

// Next implementation of Flaker interface for RandomFlake generator.
func (rf *RandomFlake) Next() uint64 {
	timeDelta := uint64(time.Since(rf.epochTime).Nanoseconds() / 1e6) << timestampShift
	randomBits := uint64(_rand.Uint32() & randomBitsMask)
	rf.value = timeDelta | randomBits
	return rf.value
}

// Parse implementation of Flaker interface for RandomFlake.
func (rf *RandomFlake) Parse(flakeID uint64) (flakeTime time.Time, _ uint32, randomBits uint32) {
	if flakeID == 0 {
		flakeTime = rf.epochTime.Add(time.Duration(rf.value >> timestampShift) * 1e6)
		randomBits = uint32(rf.value & randomBitsMask)
		return
	}
	flakeTime = rf.epochTime.Add(time.Duration(flakeID >> timestampShift) * 1e6)
	randomBits = uint32(flakeID & randomBitsMask)
	return
}

// HostFlake represents a structure which is used to generate an ID as time delta between
// the initial epoch time + some host ID + counter value. Counter value is reset to 0 each millisecond.
type HostFlake struct {
	epochTime time.Time
	lock      sync.Mutex
	timeDelta uint64
	hostID    uint32
	counter   uint16
	value     uint64
}

// Next implements Flaker interface for RandomFlake.
func (hf *HostFlake) Next() uint64 {
	hf.lock.Lock()
	defer hf.lock.Unlock()

	timeDelta := uint64(time.Since(hf.epochTime).Nanoseconds() / 1e6) << timestampShift
	if hf.timeDelta < timeDelta {
		hf.timeDelta = timeDelta
		hf.counter = 0
	} else {
		hf.counter++
	}
	hf.value = hf.timeDelta | uint64(hf.hostID << hostIDShift) | uint64(hf.counter & counterBitsMask)
	return hf.value
}

// Parse implementation of Flaker interface for HostFlake generator.
func (hf *HostFlake) Parse(flakeID uint64) (flakeTime time.Time, hostID uint32, counter uint32) {
	if flakeID == 0 {
		// Parse flakeID stored in HostFlake.value
		flakeTime = hf.epochTime.Add(time.Duration(hf.value >> timestampShift) * 1e6)
		hostID = uint32((hf.value >> hostIDShift) & hostIDMask)
		counter = uint32(hf.value & counterBitsMask)
		return
	}
	flakeTime = hf.epochTime.Add(time.Duration(flakeID >> timestampShift) * 1e6)
	hostID = uint32((flakeID >> hostIDShift) & hostIDMask)
	counter = uint32(flakeID & counterBitsMask)
	return
}

func init() {
	_rand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// NewRandomFlake constructs a new random-based ID generator.
func NewRandomFlake(epoch time.Time) Flaker {
	var flake RandomFlake
	if epoch.IsZero() {
		flake.epochTime = defaultEpoch
	} else {
		flake.epochTime = epoch
	}
	return &flake
}

// NewHostFlake constructs a new host ID-based ID generator.
func NewHostFlake(hostID uint32, epoch time.Time) Flaker {
	var flake HostFlake
	if epoch.IsZero() {
		flake.epochTime = defaultEpoch
	} else {
		flake.epochTime = epoch
	}
	flake.hostID = hostID
	return &flake
}
