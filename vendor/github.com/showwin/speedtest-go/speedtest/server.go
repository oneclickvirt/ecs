package speedtest

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/showwin/speedtest-go/speedtest/transport"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	speedTestServersUrl            = "https://www.speedtest.net/api/js/servers"
	speedTestServersAlternativeUrl = "https://www.speedtest.net/speedtest-servers-static.php"
	speedTestServersAdvanced       = "https://www.speedtest.net/api/ios-config.php"
)

type payloadType int

const (
	typeJSONPayload payloadType = iota
	typeXMLPayload
)

var (
	ErrServerNotFound = errors.New("no server available or found")
)

// Server information
type Server struct {
	URL          string          `xml:"url,attr" json:"url"`
	Lat          string          `xml:"lat,attr" json:"lat"`
	Lon          string          `xml:"lon,attr" json:"lon"`
	Name         string          `xml:"name,attr" json:"name"`
	Country      string          `xml:"country,attr" json:"country"`
	Sponsor      string          `xml:"sponsor,attr" json:"sponsor"`
	ID           string          `xml:"id,attr" json:"id"`
	Host         string          `xml:"host,attr" json:"host"`
	Distance     float64         `json:"distance"`
	Latency      time.Duration   `json:"latency"`
	MaxLatency   time.Duration   `json:"max_latency"`
	MinLatency   time.Duration   `json:"min_latency"`
	Jitter       time.Duration   `json:"jitter"`
	DLSpeed      ByteRate        `json:"dl_speed"`
	ULSpeed      ByteRate        `json:"ul_speed"`
	TestDuration TestDuration    `json:"test_duration"`
	PacketLoss   transport.PLoss `json:"packet_loss"`

	Context *Speedtest `json:"-"`
}

type TestDuration struct {
	Ping     *time.Duration `json:"ping"`
	Download *time.Duration `json:"download"`
	Upload   *time.Duration `json:"upload"`
	Total    *time.Duration `json:"total"`
}

// CustomServer use defaultClient, given a URL string, return a new Server object, with as much
// filled in as we can
func CustomServer(host string) (*Server, error) {
	return defaultClient.CustomServer(host)
}

// CustomServer given a URL string, return a new Server object, with as much
// filled in as we can
func (s *Speedtest) CustomServer(host string) (*Server, error) {
	u, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	parseHost := fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, "/speedtest/upload.php")
	return &Server{
		ID:      "Custom",
		Lat:     "?",
		Lon:     "?",
		Country: "?",
		URL:     parseHost,
		Name:    u.Host,
		Host:    u.Host,
		Sponsor: "?",
		Context: s,
	}, nil
}

// ServerList list of Server
// Users(Client) also exists with @param speedTestServersAdvanced
type ServerList struct {
	Servers []*Server `xml:"servers>server"`
	Users   []User    `xml:"client"`
}

// Servers for sorting servers.
type Servers []*Server

// ByDistance for sorting servers.
type ByDistance struct {
	Servers
}

func (servers Servers) Available() *Servers {
	retServer := Servers{}
	for _, server := range servers {
		if server.Latency != PingTimeout {
			retServer = append(retServer, server)
		}
	}
	for i := 0; i < len(retServer)-1; i++ {
		for j := 0; j < len(retServer)-i-1; j++ {
			if retServer[j].Latency > retServer[j+1].Latency {
				retServer[j], retServer[j+1] = retServer[j+1], retServer[j]
			}
		}
	}
	return &retServer
}

// Len finds length of servers. For sorting servers.
func (servers Servers) Len() int {
	return len(servers)
}

// Swap swaps i-th and j-th. For sorting servers.
func (servers Servers) Swap(i, j int) {
	servers[i], servers[j] = servers[j], servers[i]
}

// Hosts return hosts of servers
func (servers Servers) Hosts() []string {
	var retServer []string
	for _, server := range servers {
		retServer = append(retServer, server.Host)
	}
	return retServer
}

// Less compares the distance. For sorting servers.
func (b ByDistance) Less(i, j int) bool {
	return b.Servers[i].Distance < b.Servers[j].Distance
}

// FetchServerByID retrieves a server by given serverID.
func (s *Speedtest) FetchServerByID(serverID string) (*Server, error) {
	return s.FetchServerByIDContext(context.Background(), serverID)
}

// FetchServerByID retrieves a server by given serverID.
func FetchServerByID(serverID string) (*Server, error) {
	return defaultClient.FetchServerByID(serverID)
}

// FetchServerByIDContext retrieves a server by given serverID, observing the given context.
func (s *Speedtest) FetchServerByIDContext(ctx context.Context, serverID string) (*Server, error) {
	u, err := url.Parse(speedTestServersAdvanced)
	if err != nil {
		return nil, err
	}
	query := u.Query()
	query.Set(strings.ToLower("serverID"), serverID)
	u.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.doer.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var list ServerList
	decoder := xml.NewDecoder(resp.Body)
	if err = decoder.Decode(&list); err != nil {
		return nil, err
	}

	for i := range list.Servers {
		if list.Servers[i].ID == serverID {
			list.Servers[i].Context = s
			if len(list.Users) > 0 {
				sLat, _ := strconv.ParseFloat(list.Servers[i].Lat, 64)
				sLon, _ := strconv.ParseFloat(list.Servers[i].Lon, 64)
				uLat, _ := strconv.ParseFloat(list.Users[0].Lat, 64)
				uLon, _ := strconv.ParseFloat(list.Users[0].Lon, 64)
				list.Servers[i].Distance = distance(sLat, sLon, uLat, uLon)
			}
			return list.Servers[i], err
		}
	}
	return nil, ErrServerNotFound
}

// FetchServers retrieves a list of available servers
func (s *Speedtest) FetchServers() (Servers, error) {
	return s.FetchServerListContext(context.Background())
}

// FetchServers retrieves a list of available servers
func FetchServers() (Servers, error) {
	return defaultClient.FetchServers()
}

// FetchServerListContext retrieves a list of available servers, observing the given context.
func (s *Speedtest) FetchServerListContext(ctx context.Context) (Servers, error) {
	u, err := url.Parse(speedTestServersUrl)
	if err != nil {
		return Servers{}, err
	}
	query := u.Query()
	if len(s.config.Keyword) > 0 {
		query.Set("search", s.config.Keyword)
	}
	if s.config.Location != nil {
		query.Set("lat", strconv.FormatFloat(s.config.Location.Lat, 'f', -1, 64))
		query.Set("lon", strconv.FormatFloat(s.config.Location.Lon, 'f', -1, 64))
	}
	u.RawQuery = query.Encode()
	dbg.Printf("Retrieving servers: %s\n", u.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return Servers{}, err
	}

	resp, err := s.doer.Do(req)
	if err != nil {
		return Servers{}, err
	}

	_payloadType := typeJSONPayload

	if resp.ContentLength == 0 {
		_ = resp.Body.Close()

		req, err = http.NewRequestWithContext(ctx, http.MethodGet, speedTestServersAlternativeUrl, nil)
		if err != nil {
			return Servers{}, err
		}

		resp, err = s.doer.Do(req)
		if err != nil {
			return Servers{}, err
		}

		_payloadType = typeXMLPayload
	}

	defer resp.Body.Close()

	var servers Servers

	switch _payloadType {
	case typeJSONPayload:
		// Decode xml
		decoder := json.NewDecoder(resp.Body)

		if err = decoder.Decode(&servers); err != nil {
			return servers, err
		}
	case typeXMLPayload:
		var list ServerList
		// Decode xml
		decoder := xml.NewDecoder(resp.Body)

		if err = decoder.Decode(&list); err != nil {
			return servers, err
		}

		servers = list.Servers
	default:
		return servers, errors.New("response payload decoding not implemented")
	}

	dbg.Printf("Servers Num: %d\n", len(servers))
	// set doer of server
	for _, server := range servers {
		server.Context = s
	}

	// ping once
	var wg sync.WaitGroup
	pCtx, fc := context.WithTimeout(context.Background(), time.Second*4)
	dbg.Println("Echo each server...")
	for _, server := range servers {
		wg.Add(1)
		go func(gs *Server) {
			var latency []int64
			if s.config.PingMode == TCP {
				latency, err = gs.TCPPing(pCtx, 1, time.Millisecond, nil)
			} else if s.config.PingMode == ICMP {
				latency, err = gs.ICMPPing(pCtx, 4*time.Second, 1, time.Millisecond, nil)
			} else {
				latency, err = gs.HTTPPing(pCtx, 1, time.Millisecond, nil)
			}
			if err != nil || len(latency) < 1 {
				gs.Latency = PingTimeout
			} else {
				gs.Latency = time.Duration(latency[0]) * time.Nanosecond
			}
			wg.Done()
		}(server)
	}
	wg.Wait()
	fc()

	// Calculate distance
	// If we don't call FetchUserInfo() before FetchServers(),
	// we don't calculate the distance, instead we use the
	// remote computing distance provided by Ookla as default.
	if s.User != nil {
		for _, server := range servers {
			sLat, _ := strconv.ParseFloat(server.Lat, 64)
			sLon, _ := strconv.ParseFloat(server.Lon, 64)
			uLat, _ := strconv.ParseFloat(s.User.Lat, 64)
			uLon, _ := strconv.ParseFloat(s.User.Lon, 64)
			server.Distance = distance(sLat, sLon, uLat, uLon)
		}
	}

	// Sort by distance
	sort.Sort(ByDistance{servers})

	if len(servers) <= 0 {
		return servers, ErrServerNotFound
	}
	return servers, nil
}

// FetchServerListContext retrieves a list of available servers, observing the given context.
func FetchServerListContext(ctx context.Context) (Servers, error) {
	return defaultClient.FetchServerListContext(ctx)
}

func distance(lat1 float64, lon1 float64, lat2 float64, lon2 float64) float64 {
	radius := 6378.137

	phi1 := lat1 * math.Pi / 180.0
	phi2 := lat2 * math.Pi / 180.0

	deltaPhiHalf := (lat1 - lat2) * math.Pi / 360.0
	deltaLambdaHalf := (lon1 - lon2) * math.Pi / 360.0
	sinePhiHalf2 := math.Sin(deltaPhiHalf)*math.Sin(deltaPhiHalf) + math.Cos(phi1)*math.Cos(phi2)*math.Sin(deltaLambdaHalf)*math.Sin(deltaLambdaHalf) // phi half-angle sine ^ 2
	delta := 2 * math.Atan2(math.Sqrt(sinePhiHalf2), math.Sqrt(1-sinePhiHalf2))                                                                       // 2 arc sine
	return radius * delta                                                                                                                             // r * delta
}

// FindServer finds server by serverID in given server list.
// If the id is not found in the given list, return the server with the lowest latency.
func (servers Servers) FindServer(serverID []int) (Servers, error) {
	retServer := Servers{}

	if len(servers) <= 0 {
		return retServer, ErrServerNotFound
	}

	for _, sid := range serverID {
		for _, s := range servers {
			id, _ := strconv.Atoi(s.ID)
			if sid == id {
				retServer = append(retServer, s)
				break
			}
		}
	}

	if len(retServer) == 0 {
		// choose the lowest latency server
		var minLatency int64 = math.MaxInt64
		var minServerIndex int
		for index, server := range servers {
			if server.Latency <= 0 {
				continue
			}
			if minLatency > server.Latency.Milliseconds() {
				minLatency = server.Latency.Milliseconds()
				minServerIndex = index
			}
		}
		retServer = append(retServer, servers[minServerIndex])
	}
	return retServer, nil
}

// String representation of ServerList
func (servers ServerList) String() string {
	slr := ""
	for _, server := range servers.Servers {
		slr += server.String()
	}
	return slr
}

// String representation of Servers
func (servers Servers) String() string {
	slr := ""
	for _, server := range servers {
		slr += server.String()
	}
	return slr
}

// String representation of Server
func (s *Server) String() string {
	if s.Sponsor == "?" {
		return fmt.Sprintf("[%4s] %s", s.ID, s.Name)
	}
	if len(s.Country) == 0 {
		return fmt.Sprintf("[%4s] %.2fkm %s by %s", s.ID, s.Distance, s.Name, s.Sponsor)
	}
	return fmt.Sprintf("[%4s] %.2fkm %s (%s) by %s", s.ID, s.Distance, s.Name, s.Country, s.Sponsor)
}

// CheckResultValid checks that results are logical given UL and DL speeds
func (s *Server) CheckResultValid() bool {
	return !(s.DLSpeed*100 < s.ULSpeed) || !(s.DLSpeed > s.ULSpeed*100)
}

func (s *Server) testDurationTotalCount() {
	total := s.getNotNullValue(s.TestDuration.Ping) +
		s.getNotNullValue(s.TestDuration.Download) +
		s.getNotNullValue(s.TestDuration.Upload)

	s.TestDuration.Total = &total
}

func (s *Server) getNotNullValue(time *time.Duration) time.Duration {
	if time == nil {
		return 0
	}
	return *time
}
