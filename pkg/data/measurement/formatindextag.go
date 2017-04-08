package measurement

import (
	"github.com/Sirupsen/logrus"
	"regexp"
	"strconv"
	"strings"
)

// two byte-oriented functions identical except for operator comparing c to 127.
func stripCtlFromBytes(str string) string {
	b := make([]byte, len(str))
	var bl int
	for i := 0; i < len(str); i++ {
		c := str[i]
		if c >= 32 && c != 127 {
			b[bl] = c
			bl++
		}
	}
	return string(b[:bl])
}

func stripCtlAndExtFromBytes(str string) string {
	b := make([]byte, len(str))
	var bl int
	for i := 0; i < len(str); i++ {
		c := str[i]
		if c >= 32 && c < 127 {
			b[bl] = c
			bl++
		}
	}
	return string(b[:bl])
}

func formatDec2ASCII(input string) string {
	sArray := strings.Split(input, ".")
	n := len(sArray)
	bArray := make([]byte, n)
	for i := 0; i < n; i++ {
		num, _ := strconv.Atoi(sArray[i])
		//fmt.Printf("num %d\n",num)
		bArray = append(bArray, byte(num))
	}
	return stripCtlAndExtFromBytes(string(bArray))
}

func formatReGexp(input string, pattern string, replace string) string {
	re, err := regexp.Compile(pattern)
	if err == nil {
		return input
	}
	match := re.FindStringSubmatch(input)

	final := replace
	for i := 1; i < len(match); i++ {
		final = strings.Replace(final, "\\"+strconv.Itoa(i), match[i], -1)
	}
	return final
}

func sectionDotSlice(input string, first int, last int) string {
	sArray := strings.Split(input, ".")
	n := len(sArray)

	if last == -1 {
		last = n - 1
	}
	if last < first {
		return input
	}

	if first > n {
		return input
	}

	if last < n {
		n = last
	}
	output := sArray[first]
	for i := first + 1; i <= n; i++ {
		output = output + "." + sArray[i]
	}
	return output
}

func formatTag(l *logrus.Logger, format string, data map[string]string, def string) string {
	if len(format) == 0 {
		return data[def]
	}

	final := format
	for k, v := range data {
		final = strings.Replace(final, "$"+k, v, -1)
	}
	//check if more varibles defined
	if !strings.Contains(final, "$") {
		return final
	}
	//continue evaluating for each variable
	for k, v := range data {
		//check as many times as repeaded variables it has
		for {
			// ${VARNAME|SECTION|TRANSFORMATION}
			pattern := "\\${" + k + "\\|([^|]*)\\|([^}]*)}"
			re, err := regexp.Compile(pattern)
			if err != nil {
				l.Debugf("FormatTag[%s]: Regex ${VAR|XXX} Error %s with pattern %s", format, err, pattern)
				continue
			}
			match := re.FindStringSubmatch(final)
			if len(match) < 3 {
				l.Debugf("FormatTag[%s]: match length: %d , match  %+v with pattern %s", format, len(match), match, pattern)
				break
				//continue
			}
			//here we hav found a tranformation to do
			sectionmode := match[1]
			transformation := match[2]
			//check defaultvalues
			if len(sectionmode) == 0 {
				sectionmode = "ALL"
			}
			if len(transformation) == 0 {
				transformation = "STRING"
			}
			// Getting Variable Section
			section := v
			switch {
			case sectionmode == "ALL":
				section = v
			case strings.HasPrefix(sectionmode, "DOT["):
				re2, err := regexp.Compile("DOT\\[(\\d*):(\\d*)\\]")
				if err != nil {
					l.Warnf("FormatTag[%s]: DOT - Regex ERROR %s ", format, err)
					break
				}
				match2 := re2.FindStringSubmatch(sectionmode)
				if len(match2) < 3 {
					l.Warnf("FormatTag[%s]: DOT - ERROR on number or paramters %+v  for string %s", format, match2, sectionmode)
					break
				}

				dotInit := match2[1]
				dotEnd := match2[2]

				first, err := strconv.Atoi(dotInit) // if there is an error first will be 0
				if err != nil {
					l.Warnf("FormatTag[%s]: DOT -  error decode first position [%s]   error  %s", format, dotInit, err)
				}
				last, err := strconv.Atoi(dotEnd) // if there is an error last
				if err != nil {
					l.Warnf("FormatTag[%s]: DOT -  error decode last position [%s]   error  %s", format, dotEnd, err)
					last = -1
				}
				section = sectionDotSlice(v, first, last)
				l.Debugf("FormatTag[%s]: final section first/last[%d/%d] from [%s] took [%s] ", format, first, last, v, section)
			default:
				l.Warnf("FormatTag[%s]: Unknown SECTION parameters %s ,  pattern %s", format, sectionmode, pattern)
			}

			//here we have the section we want to decode
			//Doing transfomations over de selected section
			decoded := section
			switch {
			case transformation == "STRING":
				decoded = section
			case strings.HasPrefix(transformation, "DEC2ASCII"):
				decoded = formatDec2ASCII(section)
				l.Debugf("FormatTag[%s]: DEC2ASCII : value %s : Decoded %s", format, v, decoded)
			default:
				l.Warnf("FormatTag[%s]: Unknown TRASNFORMATION parameters %s ,  pattern %s", format, transformation, pattern)
			}
			final = strings.Replace(final, match[0], decoded, -1)
			l.Debugf("Result After Iteration on var Instance [%s]-[%s]", k, final)
		}
		l.Debugf("Result After Iteration on var [%s]-[%s]", k, final)
	}
	l.Debugf("FormatTag[%s]: OK Final result: %s", format, final)
	return stripCtlAndExtFromBytes(final)
}
