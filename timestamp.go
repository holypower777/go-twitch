package bot

import (
	"strconv"
	"time"
)

const (
	emptyTimeStr                     = `"0001-01-01T00:00:00Z"`
	referenceTimeStr                 = `"2006-01-02T15:04:05Z"`
	referenceTimeStrFractional       = `"2006-01-02T15:04:05.000Z"` // This format was returned by the Projects API before October 1, 2017.
	referenceUnixTimeStr             = `1136214245`
	referenceUnixTimeStrMilliSeconds = `1136214245000` // Millisecond-granular timestamps were introduced in the Audit log API.
)

var (
	referenceTime = time.Date(2006, time.January, 02, 15, 04, 05, 0, time.UTC)
	// unixOrigin    = time.Unix(0, 0).In(time.UTC)
)

type Timestamp struct {
	time.Time
}

func (t Timestamp) String() string {
	return t.Time.String()
}

func (t *Timestamp) UnmarshalJSON(data []byte) (err error) {
	str := string(data)
	i, err := strconv.ParseInt(str, 10, 64)
	if err == nil {
		t.Time = time.Unix(i, 0)
		if t.Time.Year() > 3000 {
			t.Time = time.Unix(0, i*1e6)
		}
	} else {
		t.Time, err = time.Parse(`"`+time.RFC3339+`"`, str)
	}
	return
}
