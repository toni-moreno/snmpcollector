package measurement

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// MultiIndexFormat defines the required fields to build multiindex result
type MultiIndexFormat struct {
	Label            string
	CurIndexedLabels map[string]string
	TagName          []string
	Index            int
	DepDesc          string
	Dependency       *MultiIndexDependency
}

// MultiIndexDependency defines the dependency pararams required to build mulitindex result
type MultiIndexDependency struct {
	Index    int
	Start    int
	End      int
	Strategy []string
}

// GetDepMultiParams - Parse Dependency description and creates a new MultiIndexDependency object
func (mi *MultiIndexFormat) GetDepMultiParams() error {
	// Dependency syntax follows as:
	// IDX{M}[];DOT[START:END];FILL(XXX)]
	// To simplify its logic, it can be defined also as IDX{M}
	mdep := MultiIndexDependency{}

	// Read dependency params splitted by ';'
	dep := strings.Split(mi.DepDesc, ";")

	if len(dep) < 1 {
		return fmt.Errorf("Invalid dependency description %s", mi.DepDesc)
	}

	// Index dependency - dep[0]
	pattern := "IDX{([0-9]+)}"
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	match := re.FindStringSubmatch(dep[0])
	if len(match) < 1 {
		return fmt.Errorf("Couldn't find any matching dependency on %+v", dep[0])
	}

	// ensure its a valid number and its
	ndep, err := strconv.Atoi(match[1])
	if err != nil {
		return err
	}
	mdep.Index = ndep

	// If only dependency is provided, just load it as defaults and return
	if len(dep) == 1 {
		mdep.Start = -1
		mdep.End = -1
		mdep.Strategy = []string{"", ""}
		mi.Dependency = &mdep
		return nil
	}

	// Check if DOT and Strategy are provided...
	if len(dep) == 3 {

		// dep[1] - Index Position
		ipattern, err := regexp.Compile("DOT\\[(\\d*):(\\d*)\\]")
		if err != nil {
			return err
		}

		imatch := ipattern.FindStringSubmatch(dep[1])
		if len(imatch) < 3 {
			return fmt.Errorf("Couldn't find index matching on %+v", dep[1])
		}
		dotInit := imatch[1]
		dotEnd := imatch[2]

		first, err := strconv.Atoi(dotInit) // if there is an error first will be 0
		if err != nil {
			return err
		}
		last, err := strconv.Atoi(dotEnd) // if there is an error last
		if err != nil {
			return err
		}
		mdep.Start = first
		mdep.End = last

		// ************ PARSE FILL STRATEGY **************************

		fpattern, err := regexp.Compile("(\\w+)(?:\\((.*)\\))?")
		if err != nil {
			return err
		}

		fmatch := fpattern.FindStringSubmatch(dep[2])
		if len(fmatch) < 1 {
			return fmt.Errorf("Error trying to parse dependency strategy %v", fmatch)
		}
		// finally, store the result...
		mdep.Strategy = fmatch

		mi.Dependency = &mdep
		return nil
	}
	return fmt.Errorf("Missing dependency variable %s", mi.DepDesc)
}

// MultiIndexFormatArray - type to implement Sort interface and allow order by dependency
type MultiIndexFormatArray []*MultiIndexFormat

// Len - Sort interface function to return len of array
func (mif MultiIndexFormatArray) Len() int {
	return len(mif)
}

// Less - Sort interface function to add logic on sort function
func (mif MultiIndexFormatArray) Less(i, j int) bool {
	// Retrieve if there are dependencies and retrieve the index of each one...
	idxi := -1
	idxj := -1

	if mif[j].Dependency != nil {
		idxj = mif[j].Dependency.Index
	}
	if mif[i].Dependency != nil {
		idxi = mif[i].Dependency.Index
	}

	return idxi < idxj
}

// Swap - Sort interface function to swap elements between array
func (mif MultiIndexFormatArray) Swap(i, j int) {
	mif[i], mif[j] = mif[j], mif[i]
}

// GetDepIndex - Return the real index of the array based on dependency. Needed as array is sorted
func (mif MultiIndexFormatArray) GetDepIndex(ri int) (int, error) {
	for i, mi := range mif {
		if mi.Index == ri {
			return i, nil
		}
	}
	return -1, fmt.Errorf("Provided index in dependency not found %d", ri)
}

// BuildParseResults - Build result array based on result string
func BuildParseResults(allindex MultiIndexFormatArray, rs []string) (MultiIndexFormatArray, string, error) {
	// combpattern := `^([0-9]+)$|^(IDX\{([0-9])\})$`
	combpattern := `^([0-9]+)$|^(IDX\{([0-9])\}(?:;(DOT\[\d*:\d*\]))?)$`
	re, err := regexp.Compile(combpattern)
	if err != nil {
		return nil, "", err
	}

	// The strategy follows as:
	// find all index splitted by '.' until an IDX{X} is found.
	// all index is stored as index and loaded as index on IDX{X}
	// in case that a prefix is found and it matches with the last dot
	// it is stored on suffix and will be loaded into the last merged tables
	var prefix string
	var suffix string
	var resindex MultiIndexFormatArray

	for i, v := range rs {
		// we need to check if v is a simple number or an expression (using regex?)
		match := re.FindStringSubmatch(v)
		// fmt.Println(match, len(match))
		if len(match) == 0 {
			return nil, "", fmt.Errorf("Invalid value provided on mulitindex result %s", v)
		}
		// Case number is retrieved
		if match[1] != "" {
			if prefix != "" {
				prefix += "."
			}
			prefix += v
			if i == len(rs)-1 {
				suffix += prefix
			}
		}
		// Case IDX{M} is retrieved
		if match[2] != "" {
			ndep, err := strconv.Atoi(match[3])
			if err != nil {
				return nil, "", fmt.Errorf("Invalid provided result index. Error: %s", err)
			}

			ir, err := allindex.GetDepIndex(ndep)
			if err != nil {
				return nil, "", err
			}

			mi := MultiIndexFormat{
				TagName:          allindex[ir].TagName,
				Index:            allindex[ir].Index,
				DepDesc:          allindex[ir].DepDesc,
				Dependency:       allindex[ir].Dependency,
				CurIndexedLabels: make(map[string]string),
			}
			// Deep copy if CurIndexedLabels, index can be reused...
			// Case DOT is provided...
			var first, last int
			var sapply bool
			if match[4] != "" {
				// dep[1] - Index Position
				ipattern, err := regexp.Compile("DOT\\[(\\d*):(\\d*)\\]")
				if err != nil {
					return nil, "", err
				}

				imatch := ipattern.FindStringSubmatch(match[4])
				if len(imatch) < 3 {
					return nil, "", fmt.Errorf("Couldn't find index matching on %+v", match[4])
				}
				dotInit := imatch[1]
				dotEnd := imatch[2]

				first, err = strconv.Atoi(dotInit) // if there is an error first will be 0
				if err != nil {
					return nil, "", err
				}
				last, err = strconv.Atoi(dotEnd) // if there is an error last
				if err != nil {
					return nil, "", err
				}
				sapply = true
			}
			for k, v := range allindex[ir].CurIndexedLabels {
				section := k
				if sapply {
					section, err = sectionDotSlice(k, first, last)
					if err != nil {
						return nil, "", err
					}
				}
				mi.CurIndexedLabels[section] = v
			}
			// mi := allindex[ir]
			kk := make(map[string]string)

			// if there is some prefix, we need to append to the existing map, if not, it will directly be the currend indexes labels
			if len(prefix) > 0 {
				for k, v := range mi.CurIndexedLabels {
					kk[prefix+"."+k] = v
				}
				mi.CurIndexedLabels = kk
			}
			resindex = append(resindex, &mi)
			prefix = ""
		}

	}
	return resindex, suffix, nil
}

// MergeResults - Merge all results
func MergeResults(resindex MultiIndexFormatArray, suffix string) (*MultiIndexFormat, error) {
	if len(resindex) == 0 {
		return &MultiIndexFormat{}, fmt.Errorf("Empty result index after parsed all indexes")
	}

	// Start iteration to over all indexes and it will recursively store results on mi
	mi := resindex[0]
	for i := 1; i < len(resindex); i++ {
		mi = MergeIndex(mi, resindex[i])
	}

	// Finally, check if we must apply some suffix...
	if len(suffix) > 0 {
		mi = AddSuffix(suffix, mi)
	}

	return mi, nil
}

// MergeIndex - Merges two tables by scalar product between indexes
func MergeIndex(tmp, input *MultiIndexFormat) *MultiIndexFormat {
	var pp MultiIndexFormat
	pp.CurIndexedLabels = make(map[string]string)

	for k, v := range tmp.CurIndexedLabels {
		for kk, vv := range input.CurIndexedLabels {
			pp.CurIndexedLabels[k+"."+kk] = v + "|" + vv
		}
	}

	pp.TagName = append(pp.TagName, tmp.TagName...)
	pp.TagName = append(pp.TagName, input.TagName...)
	return &pp
}

// AddSuffix - Add suffix to existing index/label map
func AddSuffix(suffix string, input *MultiIndexFormat) *MultiIndexFormat {
	var pp MultiIndexFormat
	pp.CurIndexedLabels = make(map[string]string)
	for k, v := range input.CurIndexedLabels {
		pp.CurIndexedLabels[k+"."+suffix] = v
	}
	pp.TagName = input.TagName
	return &pp
}
