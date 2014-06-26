package relational

import (
	"errors"
	"strconv"
)

var LatLngUnparsed = errors.New("liquor: no attempt was made to parse latlng")

// ParseLatLng will look for the GET parameters "latitude" and "longitude".
// If either of these parameters exist and are non-empty, they will attempt
// to parse them as latitude and longitude values, returning an error
// if the parameters could no be successfully parsed or were out of range.
func ParseLatLng(rawLat, rawLng string) (lat, lng float64, err error) {
	if rawLat == "" && rawLng == "" {
		err = LatLngUnparsed
		return
	}
	if lat, err = strconv.ParseFloat(rawLat, 64); err != nil {
		return
	}
	lng, err = strconv.ParseFloat(rawLng, 64)
	// TODO Error if out of bounds
	return
}

// ParseIntOrDefault will attempt to parse the given string as an integer and
// if an error is generated, then returned the supplied default instead
func ParseIntOrDefault(s string, d int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return d
	}
	return i
}

// ParseInt64OrDefault will attempt to parse the given string as an integer and
// if an error is generated, then returned the supplied default instead
func ParseInt64OrDefault(s string, d int64) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return d
	}
	return i
}

// ParseFloatOrDefault will attempt to parse the given string as a float and
// if an error is generated, then returned the supplied default instead
func ParseFloatOrDefault(s string, d float64) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return d
	}
	return f
}
