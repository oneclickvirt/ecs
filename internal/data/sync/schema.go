package datasync

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type speedtestServerSchema struct {
	CC              string `json:"cc"`
	City            string `json:"city"`
	CityZH          string `json:"cityzh"`
	Code            *int   `json:"code"`
	Country         string `json:"country"`
	Distance        *int   `json:"distance"`
	ForcePingSelect *int   `json:"force_ping_select"`
	Host            string `json:"host"`
	HTTPSFunctional *int   `json:"https_functional"`
	ID              string `json:"id"`
	Latitude        string `json:"lat"`
	Longitude       string `json:"lon"`
	Name            string `json:"name"`
	Preferred       *int   `json:"preferred"`
	Provider        string `json:"provider"`
	ProviderZH      string `json:"providerzh"`
	Sponsor         string `json:"sponsor"`
	Status          string `json:"status"`
	URL             string `json:"url"`
}

type dnsblZoneSchema struct {
	Zone string `json:"zone"`
	IPv4 *bool  `json:"ipv4"`
	IPv6 *bool  `json:"ipv6"`
}

func strictJSONArray[T any](data []byte) ([]T, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var records []T
	if err := decoder.Decode(&records); err != nil {
		return nil, err
	}
	if records == nil {
		return nil, errors.New("records must be a non-null JSON array")
	}
	if len(records) == 0 {
		return nil, errors.New("records must not be empty")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, errors.New("trailing JSON value")
		}
		return nil, err
	}
	return records, nil
}

func validateTCPTargetSchema(data []byte) error {
	records, err := strictJSONArray[tcpTarget](data)
	if err != nil {
		return err
	}
	seenEndpoints := make(map[string]struct{}, len(records))
	keys := make([]string, 0, len(records))
	for index, record := range records {
		record.ID = strings.TrimSpace(record.ID)
		record.Name = strings.TrimSpace(record.Name)
		record.Host = strings.ToLower(strings.TrimSpace(record.Host))
		record.Category = strings.TrimSpace(record.Category)
		if record.ID == "" || record.Name == "" || record.Category == "" || !validSchemaHost(record.Host) || record.Port < 1 || record.Port > 65535 {
			return fmt.Errorf("record %d has empty or invalid required fields", index)
		}
		endpoint := net.JoinHostPort(record.Host, strconv.Itoa(record.Port))
		if _, exists := seenEndpoints[endpoint]; exists {
			return fmt.Errorf("record %d duplicates endpoint %q", index, endpoint)
		}
		seenEndpoints[endpoint] = struct{}{}
		keys = append(keys, record.ID)
	}
	return validateStableUniqueKeys(keys)
}

func validateProvinceRouteSchema(data []byte) error {
	records, err := strictJSONArray[provinceRoute](data)
	if err != nil {
		return err
	}
	seenCodes := make(map[string]struct{}, len(records))
	seenNames := make(map[string]struct{}, len(records))
	seenNumbers := make(map[int]struct{}, len(records))
	previousProvince := 0
	for index, record := range records {
		if !validUpperCountryCode(record.Code) || strings.TrimSpace(record.Name) == "" || strings.TrimSpace(record.Short) == "" || record.Province <= 0 || len(record.Targets) != 3 {
			return fmt.Errorf("record %d has empty or invalid province fields", index)
		}
		if index > 0 && record.Province <= previousProvince {
			return errors.New("province records are duplicated or not stably sorted")
		}
		previousProvince = record.Province
		if _, exists := seenCodes[record.Code]; exists {
			return fmt.Errorf("duplicate province code %q", record.Code)
		}
		if _, exists := seenNames[record.Name]; exists {
			return fmt.Errorf("duplicate province name %q", record.Name)
		}
		if _, exists := seenNumbers[record.Province]; exists {
			return fmt.Errorf("duplicate province number %d", record.Province)
		}
		seenCodes[record.Code], seenNames[record.Name], seenNumbers[record.Province] = struct{}{}, struct{}{}, struct{}{}
		for targetIndex, target := range record.Targets {
			if target.Carrier != []string{"ct", "cu", "cm"}[targetIndex] || !validSchemaHost(target.IPv4) || !validSchemaHost(target.IPv6) {
				return fmt.Errorf("record %d carrier %d has invalid fields", index, targetIndex)
			}
		}
	}
	return nil
}

func validateSpeedtestServerSchema(data []byte) error {
	records, err := strictJSONArray[speedtestServerSchema](data)
	if err != nil {
		return err
	}
	keys := make([]string, 0, len(records))
	for index, record := range records {
		if anyEmpty(record.CC, record.City, record.CityZH, record.Country, record.Host, record.ID, record.Latitude, record.Longitude, record.Name, record.Provider, record.ProviderZH, record.Sponsor, record.Status, record.URL) || !validUpperCountryCode(record.CC) {
			return fmt.Errorf("record %d has an empty required string", index)
		}
		if record.Code == nil || *record.Code <= 0 || record.Distance == nil || *record.Distance < 0 || !binarySchemaFlag(record.ForcePingSelect) || !binarySchemaFlag(record.HTTPSFunctional) || !binarySchemaFlag(record.Preferred) {
			return fmt.Errorf("record %d has missing or invalid numeric fields", index)
		}
		latitude, latitudeErr := strconv.ParseFloat(record.Latitude, 64)
		longitude, longitudeErr := strconv.ParseFloat(record.Longitude, 64)
		if latitudeErr != nil || longitudeErr != nil || latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 {
			return fmt.Errorf("record %d has invalid coordinates", index)
		}
		if !validSchemaHostPort(record.Host) || !validHTTPURL(record.URL) || !validAvailabilityStatus(record.Status) {
			return fmt.Errorf("record %d has invalid host, URL, or status", index)
		}
		keys = append(keys, record.ID)
	}
	return validateStableUniqueKeys(keys)
}

func validateTransferTargetSchema(data []byte) error {
	records, err := strictJSONArray[transferTarget](data)
	if err != nil {
		return err
	}
	keys := make([]string, 0, len(records))
	for index, record := range records {
		if anyEmpty(record.ID, record.Host, record.Provider, record.Country, record.City, record.Status) || !validUpperCountryCode(record.Country) || !validSchemaHost(record.Host) || record.PortFrom < 1 || record.PortTo < record.PortFrom || record.PortTo > 65535 || !validAvailabilityStatus(record.Status) {
			return fmt.Errorf("record %d has empty or invalid transfer target fields", index)
		}
		keys = append(keys, record.ID)
	}
	return validateStableUniqueKeys(keys)
}

func validateDNSBLZoneSchema(data []byte) error {
	records, err := strictJSONArray[dnsblZoneSchema](data)
	if err != nil {
		return err
	}
	keys := make([]string, 0, len(records))
	for index, record := range records {
		zone := strings.TrimSpace(record.Zone)
		if zone == "" || zone != strings.ToLower(zone) || !validDNSName(zone) || record.IPv4 == nil || record.IPv6 == nil || (!*record.IPv4 && !*record.IPv6) {
			return fmt.Errorf("record %d has empty or invalid DNSBL fields", index)
		}
		keys = append(keys, zone)
	}
	return validateStableUniqueKeys(keys)
}

func validateASNMapSchema(data []byte) error {
	records, err := strictJSONArray[asnName](data)
	if err != nil {
		return err
	}
	previous := uint32(0)
	for index, record := range records {
		if record.ASN == 0 || strings.TrimSpace(record.Name) == "" {
			return fmt.Errorf("record %d has empty or invalid ASN fields", index)
		}
		if index > 0 && record.ASN <= previous {
			return errors.New("ASN records are duplicated or not stably sorted")
		}
		previous = record.ASN
	}
	return nil
}

func validateMediaProviderSchema(data []byte) error {
	records, err := strictJSONArray[providerName](data)
	if err != nil {
		return err
	}
	keys := make([]string, 0, len(records))
	for index, record := range records {
		if strings.TrimSpace(record.ID) == "" || strings.TrimSpace(record.Name) == "" || record.ID != slug(record.Name) || len(record.Groups) == 0 {
			return fmt.Errorf("record %d has an empty provider ID or name", index)
		}
		for groupIndex, group := range record.Groups {
			if strings.TrimSpace(group) == "" || groupIndex > 0 && group <= record.Groups[groupIndex-1] {
				return fmt.Errorf("record %d has invalid or unsorted groups", index)
			}
		}
		keys = append(keys, record.ID)
	}
	return validateStableUniqueKeys(keys)
}

func validateStableUniqueKeys(keys []string) error {
	for index, key := range keys {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("record %d has an empty sort key", index)
		}
		if index > 0 && key <= keys[index-1] {
			return errors.New("records are duplicated or not stably sorted")
		}
	}
	return nil
}

func validSchemaHost(value string) bool {
	value = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
	return net.ParseIP(value) != nil || validDNSName(value)
}

func validSchemaHostPort(value string) bool {
	host, port, err := net.SplitHostPort(strings.TrimSpace(value))
	if err != nil || !validSchemaHost(strings.Trim(host, "[]")) {
		return false
	}
	number, err := strconv.Atoi(port)
	return err == nil && number >= 1 && number <= 65535
}

func validHTTPURL(value string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Hostname() != ""
}

func validAvailabilityStatus(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "available", "unavailable":
		return true
	default:
		return false
	}
}

func binarySchemaFlag(value *int) bool {
	return value != nil && (*value == 0 || *value == 1)
}

func validUpperCountryCode(value string) bool {
	return len(value) == 2 && value[0] >= 'A' && value[0] <= 'Z' && value[1] >= 'A' && value[1] <= 'Z'
}

func anyEmpty(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return true
		}
	}
	return false
}
