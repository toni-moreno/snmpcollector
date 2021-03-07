package utils

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"encoding/base64"

	"github.com/sirupsen/logrus"
)

// WaitAlignForNextCycle waiths untile a next cycle begins aligned with second 00 of each minute
func WaitAlignForNextCycle(SecPeriod int, l *logrus.Logger) {
	i := int64(time.Duration(SecPeriod) * time.Second)
	remain := i - (time.Now().UnixNano() % i)
	l.Infof("Waiting %s to round until nearest interval... (Cycle = %d seconds)", time.Duration(remain).String(), SecPeriod)
	time.Sleep(time.Duration(remain))
}

// RemoveDuplicatesUnordered removes duplicated elements in the array string
func RemoveDuplicatesUnordered(elements []string) []string {
	encountered := map[string]bool{}

	// Create a map of all unique elements.
	for v := range elements {
		encountered[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []string{}
	for key := range encountered {
		result = append(result, key)
	}
	return result
}

// DiffSlice return de Difference between two Slices
func DiffSlice(X, Y []string) []string {

	diff := []string{}
	vals := map[string]struct{}{}

	for _, x := range X {
		vals[x] = struct{}{}
	}

	for _, x := range Y {
		if _, ok := vals[x]; !ok {
			diff = append(diff, x)
		}
	}

	return diff
}

func diffKeysInMap(X, Y map[string]string) map[string]string {

	diff := map[string]string{}

	for k, vK := range X {
		if _, ok := Y[k]; !ok {
			diff[k] = vK
		}
	}

	return diff
}

// DiffKeyValuesInMap does a diff key and values from 2 strings maps
func DiffKeyValuesInMap(X, Y map[string]string) map[string]string {

	diff := map[string]string{}

	for kX, vX := range X {
		if vY, ok := Y[kX]; !ok {
			//not exist
			diff[kX] = vX
		} else {
			//exist
			if vX != vY {
				//but value is different
				diff[kX] = vX
			}
		}
	}
	return diff
}

// CSV2IntArray CSV intenger array conversion
func CSV2IntArray(csv string) ([]int64, error) {
	var iarray []int64
	result := Splitter(csv, ",;|")
	for i, v := range result {
		vc, err := strconv.Atoi(v)
		if err != nil {
			return iarray, fmt.Errorf("Bad Format in CSV array item %d | value %s | Error %s", i, v, err)
		}
		iarray = append(iarray, int64(vc))
	}
	return iarray, nil
}

// Splitter multiple value split
func Splitter(s string, splits string) []string {
	m := make(map[rune]int)
	for _, r := range splits {
		m[r] = 1
	}

	splitter := func(r rune) bool {
		return m[r] == 1
	}

	return strings.FieldsFunc(s, splitter)
}

type NetworkAddress struct {
	Host string
	Port string
}

func SplitHostPortDefault(input, defaultHost, defaultPort string) (NetworkAddress, error) {
	addr := NetworkAddress{
		Host: defaultHost,
		Port: defaultPort,
	}
	if len(input) == 0 {
		return addr, nil
	}

	start := 0
	// Determine if IPv6 address, in which case IP address will be enclosed in square brackets
	if strings.Index(input, "[") == 0 {
		addrEnd := strings.LastIndex(input, "]")
		if addrEnd < 0 {
			// Malformed address
			return addr, fmt.Errorf("Malformed IPv6 address: '%s'", input)
		}

		start = addrEnd
	}
	if strings.LastIndex(input[start:], ":") < 0 {
		// There's no port section of the input
		// It's still useful to call net.SplitHostPort though, since it removes IPv6
		// square brackets from the address
		input = fmt.Sprintf("%s:%s", input, defaultPort)
	}

	host, port, err := net.SplitHostPort(input)
	if err != nil {
		return addr, fmt.Errorf("net.SplitHostPort failed for '%s' Error: %s,", input, err)
	}

	if len(host) > 0 {
		addr.Host = host
	}
	if len(port) > 0 {
		addr.Port = port
	}

	return addr, nil
}

func DecodeBasicAuthHeader(header string) (string, string, error) {
	var code string
	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 2 && parts[0] == "Basic" {
		code = parts[1]
	}

	decoded, err := base64.StdEncoding.DecodeString(code)
	if err != nil {
		return "", "", err
	}

	userAndPass := strings.SplitN(string(decoded), ":", 2)
	if len(userAndPass) != 2 {
		return "", "", errors.New("Invalid basic auth header")
	}

	return userAndPass[0], userAndPass[1], nil
}
