package flakeID

import (
	"testing"
	"time"
)

const (
	hostID        = 789
	randomFlakeID = 4290444552448220549
	hostFlakeID   = 4290444760684712963
)

func TestRandomFlakeCreation(t *testing.T) {
	flake := NewRandomFlake(time.Time{})

	id1 := flake.Next()
	if id1 <= 0 {
		t.Error("Incorrect ID", id1)
	}

	// sleep for 1 second
	time.Sleep(1e6)

	id2 := flake.Next()
	if id1 == id2 {
		t.Error("Incorrect ID", id1, id2)
	}

	if id1 > id2 {
		t.Error("Incorrect ID", id1, id2)
	}
}

func TestRandomFlakeParse(t *testing.T) {
	now := time.Now().UTC()
	flake := NewRandomFlake(time.Time{})
	id1 := flake.Next()
	flakeTime1, flakeHostID1, flakeRandomBits1 := flake.Parse(0)

	if now.Hour() != flakeTime1.Hour() || now.Minute() != flakeTime1.Minute() ||
		now.Second() != flakeTime1.Second() {
		t.Error("Date parsing failed", id1)
	}

	if flakeHostID1 != 0 {
		t.Error("Host ID is not 0", id1)
	}

	if flakeRandomBits1 <= 0 {
		t.Error("Random bits parsing failed", id1)
	}

	flakeTime2, flakeHostID2, flakeRandomBits2 := flake.Parse(randomFlakeID)

	if flakeTime2.Year() != 2016 || flakeTime2.Month() != 3 || flakeTime2.Day() != 16 ||
		flakeTime2.Hour() != 16 || flakeTime2.Minute() != 27 || flakeTime2.Second() != 26 {
		t.Error("Date parsing failed", id1)
	}

	if flakeHostID2 != 0 {
		t.Error("Host ID is not 0", randomFlakeID)
	}

	if flakeRandomBits2 != 3120517 {
		t.Error("Random bits parsing failed", randomFlakeID)
	}
}

func TestHostFlakeCreation(t *testing.T) {
	flake := NewHostFlake(hostID, time.Time{})

	id1 := flake.Next()
	if id1 <= 0 {
		t.Error("Incorrect ID", id1)
	}

	id2 := flake.Next()
	if id2-id1 != 1 {
		t.Error("Incorrect ID", id1, id2)
	}

	// sleep for 1 second
	time.Sleep(1e6)

	id3 := flake.Next()
	if id1 > id3 {
		t.Error("Incorrect ID", id1, id3)
	}
}

func TestHostFlakeParse(t *testing.T) {
	now := time.Now().UTC()
	flake := NewHostFlake(hostID, time.Time{})
	id1 := flake.Next()
	flakeTime1, flakeHostID1, flakeCounter1 := flake.Parse(0)

	if now.Minute() != flakeTime1.Minute() || now.Second() != flakeTime1.Second() {
		t.Error("Date parsing failed", id1)
	}

	if flakeHostID1 != hostID {
		t.Error("Host ID parsing failed", id1)
	}

	if flakeCounter1 != 0 {
		t.Error("Counter parsing failed", id1)
	}

	flakeTime2, flakeHostID2, flakeCounter2 := flake.Parse(hostFlakeID)

	if flakeTime2.Year() != 2016 || flakeTime2.Month() != 3 || flakeTime2.Day() != 16 ||
		flakeTime2.Hour() != 16 || flakeTime2.Minute() != 27 || flakeTime2.Second() != 51 {
		t.Error("Date parsing failed", id1)
	}

	if flakeHostID2 != hostID {
		t.Error("Host ID parsing failed", id1)
	}

	if flakeCounter2 != 3 {
		t.Error("Counter parsing failed", id1)
	}
}
