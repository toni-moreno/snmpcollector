package utils

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
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
