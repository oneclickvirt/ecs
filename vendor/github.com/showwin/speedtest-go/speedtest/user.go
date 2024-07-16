package speedtest

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
)

const speedTestConfigUrl = "https://www.speedtest.net/speedtest-config.php"

// User represents information determined about the caller by speedtest.net
type User struct {
	IP  string `xml:"ip,attr"`
	Lat string `xml:"lat,attr"`
	Lon string `xml:"lon,attr"`
	Isp string `xml:"isp,attr"`
}

// Users for decode xml
type Users struct {
	Users []User `xml:"client"`
}

// FetchUserInfo returns information about caller determined by speedtest.net
func (s *Speedtest) FetchUserInfo() (*User, error) {
	return s.FetchUserInfoContext(context.Background())
}

// FetchUserInfo returns information about caller determined by speedtest.net
func FetchUserInfo() (*User, error) {
	return defaultClient.FetchUserInfo()
}

// FetchUserInfoContext returns information about caller determined by speedtest.net, observing the given context.
func (s *Speedtest) FetchUserInfoContext(ctx context.Context) (*User, error) {
	dbg.Printf("Retrieving user info: %s\n", speedTestConfigUrl)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, speedTestConfigUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.doer.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// Decode xml
	decoder := xml.NewDecoder(resp.Body)

	var users Users
	if err = decoder.Decode(&users); err != nil {
		return nil, err
	}

	if len(users.Users) == 0 {
		return nil, errors.New("failed to fetch user information")
	}

	s.User = &users.Users[0]
	return s.User, nil
}

// FetchUserInfoContext returns information about caller determined by speedtest.net, observing the given context.
func FetchUserInfoContext(ctx context.Context) (*User, error) {
	return defaultClient.FetchUserInfoContext(ctx)
}

// String representation of User
func (u *User) String() string {
	extInfo := ""
	return fmt.Sprintf("%s (%s) [%s, %s] %s", u.IP, u.Isp, u.Lat, u.Lon, extInfo)
}
