package speedtest

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Location struct {
	Name string
	CC   string
	Lat  float64
	Lon  float64
}

// Locations TODO more location need to added
var Locations = map[string]*Location{
	"brasilia":     {"brasilia", "br", -15.793876, -47.8835327},
	"hongkong":     {"hongkong", "hk", 22.3106806, 114.1700546},
	"tokyo":        {"tokyo", "jp", 35.680938, 139.7674114},
	"london":       {"london", "uk", 51.5072493, -0.1288861},
	"moscow":       {"moscow", "ru", 55.7497248, 37.615989},
	"beijing":      {"beijing", "cn", 39.8721243, 116.4077473},
	"paris":        {"paris", "fr", 48.8626026, 2.3477229},
	"sanfrancisco": {"sanfrancisco", "us", 37.7540028, -122.4429967},
	"newyork":      {"newyork", "us", 40.7200876, -74.0220945},
	"yishun":       {"yishun", "sg", 1.4230218, 103.8404728},
	"delhi":        {"delhi", "in", 28.6251287, 77.1960896},
	"monterrey":    {"monterrey", "mx", 25.6881435, -100.3073485},
	"berlin":       {"berlin", "de", 52.5168128, 13.4009469},
	"maputo":       {"maputo", "mz", -25.9579267, 32.5760444},
	"honolulu":     {"honolulu", "us", 20.8247065, -156.918706},
	"seoul":        {"seoul", "kr", 37.6086268, 126.7179721},
	"osaka":        {"osaka", "jp", 34.6952743, 135.5006967},
	"shanghai":     {"shanghai", "cn", 31.2292105, 121.4661666},
	"urumqi":       {"urumqi", "cn", 43.8256624, 87.6058564},
	"ottawa":       {"ottawa", "ca", 45.4161836, -75.7035467},
	"capetown":     {"capetown", "za", -33.9391993, 18.4316716},
	"sydney":       {"sydney", "au", -33.8966622, 151.1731861},
	"perth":        {"perth", "au", -31.9551812, 115.8591904},
	"warsaw":       {"warsaw", "pl", 52.2396659, 21.0129345},
	"kampala":      {"kampala", "ug", 0.3070027, 32.5675581},
	"bangkok":      {"bangkok", "th", 13.7248936, 100.493026},
}

func PrintCityList() {
	fmt.Println("Available city labels (case insensitive): ")
	fmt.Println(" CC\t\tCityLabel\tLocation")
	for k, v := range Locations {
		fmt.Printf("(%v)\t%20s\t[%v, %v]\n", v.CC, k, v.Lat, v.Lon)
	}
}

func GetLocation(locationName string) (*Location, error) {
	loc, ok := Locations[strings.ToLower(locationName)]
	if ok {
		return loc, nil
	}
	return nil, errors.New("not found location")
}

// NewLocation new a Location
func NewLocation(locationName string, latitude float64, longitude float64) *Location {
	var loc Location
	loc.Lat = latitude
	loc.Lon = longitude
	loc.Name = locationName
	Locations[locationName] = &loc
	return &loc
}

// ParseLocation parse latitude and longitude string
func ParseLocation(locationName string, coordinateStr string) (*Location, error) {
	ll := strings.Split(coordinateStr, ",")
	if len(ll) == 2 {
		// parameters check
		lat, err := betweenRange(ll[0], 90)
		if err != nil {
			return nil, err
		}
		lon, err := betweenRange(ll[1], 180)
		if err != nil {
			return nil, err
		}
		name := "Custom-%s"
		if len(locationName) == 0 {
			name = "Custom-Default"
		}
		return NewLocation(fmt.Sprintf(name, locationName), lat, lon), nil
	}
	return nil, fmt.Errorf("invalid location input: %s", coordinateStr)
}

func (l *Location) String() string {
	return fmt.Sprintf("(%s) [%v, %v]", l.Name, l.Lat, l.Lon)
}

// betweenRange latitude and longitude range check
func betweenRange(inputStrValue string, interval float64) (float64, error) {
	value, err := strconv.ParseFloat(inputStrValue, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid input: %v", inputStrValue)
	}
	if value < -interval || interval < value {
		return 0, fmt.Errorf("invalid input. got: %v, expected between -%v and %v", inputStrValue, interval, interval)
	}
	return value, nil
}
