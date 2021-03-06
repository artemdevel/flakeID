//Package flakeID implements several approaches to generate IDs similar to Twitter's Snowflake IDs.
package flakeID

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strconv"
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

// Flaker interface defines set of methods for flake ID generators.
type Flaker interface {
	Next() uint64
	// Parse returns time when ID was generated, host ID and counter or random bits.
	// For RandomFlake generator host ID is always 0.
	Parse(flakeID uint64) (time.Time, uint32, uint32)
	// Helper methods to convert flake IDs into/from different string representations.
	ConvertTo(flakeID uint64, to string) (string, error)
	ConvertFrom(s string, from string) (uint64, error)
}

// RandomFlake represents a structure which is used to generate an ID as time delta between
// the initial epoch time and current time + some random value.
type RandomFlake struct {
	epochTime time.Time
	value     uint64
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

// ConvertTo implementation of Flaker interface for RandomFlake.
func (rf *RandomFlake) ConvertTo(flakeID uint64, to string) (string, error) {
	if flakeID == 0 {
		return "", fmt.Errorf("Noting to convert.")
	}
	return convertTo(flakeID, to)
}

// ConvertFrom implementation of Flaker interface for RandomFlake.
func (rf *RandomFlake) ConvertFrom(s string, from string) (uint64, error) {
	if s == "" {
		return 0, fmt.Errorf("Noting to convert.")
	}
	return convertFrom(s, from)
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

// ConvertTo implementation of Flaker interface for HostFlake.
func (hf *HostFlake) ConvertTo(flakeID uint64, to string) (string, error) {
	if flakeID == 0 {
		return "", fmt.Errorf("Noting to convert.")
	}
	return convertTo(flakeID, to)
}

// ConvertFrom implementation of Flaker interface for HostFlake.
func (hf *HostFlake) ConvertFrom(s string, from string) (uint64, error) {
	if s == "" {
		return 0, fmt.Errorf("Noting to convert.")
	}
	return convertFrom(s, from)
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


// Helper conversion functions.
func convertTo(flakeID uint64, to string) (string, error) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, flakeID)
	switch to {
	case "hex":
		return fmt.Sprintf("%x", flakeID), nil
	case "base64":
		return base64.RawStdEncoding.EncodeToString(b), nil
	default:
		return "", fmt.Errorf("Unsupported conversion to '%s'.", to)
	}
}

func convertFrom(s string, from string) (uint64, error) {
	switch from {
	case "hex":
		return strconv.ParseUint(s, 16, 64)
	case "base64":
		if b, err := base64.RawStdEncoding.DecodeString(s); err != nil {
			return 0, err
		} else {
			return binary.BigEndian.Uint64(b), nil
		}
	default:
		return 0, fmt.Errorf("Unsupported conversion from '%s'.", from)
	}
}
