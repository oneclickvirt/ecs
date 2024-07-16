package speedtest

import (
	"encoding/json"
	"fmt"
	"time"
)

type fullOutput struct {
	Timestamp outputTime `json:"timestamp"`
	UserInfo  *User      `json:"user_info"`
	Servers   Servers    `json:"servers"`
}

type outputTime time.Time

func (t outputTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 15:04:05.000"))
	return []byte(stamp), nil
}

func (s *Speedtest) JSON(servers Servers) ([]byte, error) {
	return json.Marshal(
		fullOutput{
			Timestamp: outputTime(time.Now()),
			UserInfo:  s.User,
			Servers:   servers,
		},
	)
}
