package flakeID

import (
	"testing"
	"time"
)

const (
	hostID        = 789
	hostFlakeID   = 4290444760684712963
	randomFlakeID = 4290444552448220549
	randomFlakeIDHex = "3b8ab896b52f9d85"
	randomFlakeIDBase64 = "O4q4lrUvnYU"
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

func TestConvertTo(t *testing.T) {
	flake := NewRandomFlake(time.Time{})

	if _, err := flake.ConvertTo(0, "hex"); err == nil {
		t.Error("Failed to convert to hex string, expect error.")
	}
	if _, err := flake.ConvertTo(randomFlakeID, "unsupported"); err == nil {
		t.Error("Failed to convert to unsupported string, expect error.")
	}

	if hex, err := flake.ConvertTo(randomFlakeID, "hex"); err != nil {
		t.Error("Failed to convert to hex string", err)
	} else if hex != randomFlakeIDHex {
		t.Error("Failed to convert to hex string", hex)
	}

	if b64, err := flake.ConvertTo(randomFlakeID, "base64"); err != nil {
		t.Error("Failed to convert to base64 string", err)
	} else if b64 != randomFlakeIDBase64 {
		t.Error("Failed to convert to base64 string", b64)
	}
}

func TestConvertFrom(t *testing.T) {
	flake := NewRandomFlake(time.Time{})

	if _, err := flake.ConvertFrom("", "hex"); err == nil {
		t.Error("Failed to convert from hex string, expect error.")
	}
	if _, err := flake.ConvertFrom("unsupported", "unsupported"); err == nil {
		t.Error("Failed to convert from unsupported string, expect error.")
	}

	if id, err := flake.ConvertFrom(randomFlakeIDHex, "hex"); err != nil {
		t.Error("Failed to convert from hex string", err)
	} else if id != randomFlakeID {
		t.Error("Failed to convert to hex string", id)
	}

	if id, err := flake.ConvertFrom(randomFlakeIDBase64, "base64"); err != nil {
		t.Error("Failed to convert from base64 string", err)
	} else if id != randomFlakeID {
		t.Error("Failed to convert to base64 string", id)
	}
}
