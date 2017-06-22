package utils

import (
	"github.com/Sirupsen/logrus"
	"time"
)

func WaitAlignForNextCicle(SecPeriod int, l *logrus.Logger) {
	i := int64(time.Duration(SecPeriod) * time.Second)
	remain := i - (time.Now().UnixNano() % i)
	l.Infof("Waiting %s to round until nearest interval... (Cicle = %d seconds)", time.Duration(remain).String(), SecPeriod)
	time.Sleep(time.Duration(remain))
}

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

func diffSlice(X, Y []string) []string {

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
