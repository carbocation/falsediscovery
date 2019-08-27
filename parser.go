package falsediscovery

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Value struct {
	ID             string
	pValue         float64
	criticalValue  float64
	adjustedPValue float64
}

func (v *Value) SetAdjustedPValue(in float64) {
	v.adjustedPValue = in
}

func (v *Value) SetCriticalValue(in float64) {
	v.criticalValue = in
}

func (v *Value) AdjustedP() float64 {
	return v.adjustedPValue
}

func (v *Value) P() float64 {
	return v.pValue
}

func (v *Value) Significant() bool {
	if v.pValue < v.criticalValue {
		return true
	}

	return false
}

func (v *Value) CriticalValue() float64 {
	return v.criticalValue
}

// GuessDelimiter identifies the most likely
// field delimiter from input, from the possible
// options of " ,\t|	"
func GuessDelimiter(lines []string) (rune, error) {
	// Check up to this many lines
	tryThisMany := len(lines)
	if tryThisMany > 5 {
		tryThisMany = 5
	}

	input := strings.Join(lines, "\n")

	for _, comma := range " ,\t|	" {
		c := csv.NewReader(strings.NewReader(input))
		c.Comma = comma

		acceptComma := true

		nFields := -1
		for i := 0; i < tryThisMany; i++ {
			fields, err := c.Read()
			if err != nil {
				// If the first two lines agreed, we'll call this the delimiter
				if i >= 2 {
					return comma, nil
				}

				// If not even the first two lines agreed, try another delimiter
				acceptComma = false
				break
			}

			if i == 0 {
				nFields = len(fields)

				// Must be at least a field for p value and for ID (2 fields) or more
				if nFields < 2 {
					acceptComma = false
					break
				}

				continue
			}

			// Note that we have not configured the CSV reader to permit a variable number
			// of fields, so nFields cannot be unequal to len(fields) at this point.
		}

		if acceptComma {
			return comma, nil
		}
	}

	// No comma could be detected
	return ' ', fmt.Errorf("No valid delimiter could be identified")
}

// ParseDelimitedInput takes a string which is composed of
// one record per line, then guesses the delimiter to try
// to identify the P value and the field identifier.
// It then performs FDR, and returns the FDR calculations
// for each detected value.
func ParseDelimitedInput(input string) ([]*Value, error) {
	// Line breaks in browser can be \r\n https://stackoverflow.com/a/14217315/199475

	// So: https://stackoverflow.com/a/54165816/199475
	lines := strings.Split(strings.Replace(input, "\r\n", "\n", -1), "\n")

	delim, err := GuessDelimiter(lines)
	if err != nil && len(lines) > 0 {
		if _, err2 := strconv.ParseFloat(lines[0], 64); err2 != nil {
			// Return the original error if the line isn't a pure number
			return nil, fmt.Errorf("%v => %v", err, err2)
		} else {
			// The line is just a number. Maybe there is no identifier
			delim = ','
			err = nil
		}
	} else if err != nil {
		return nil, err
	}

	c := csv.NewReader(strings.NewReader(input))
	c.Comma = delim

	values := make([]*Value, 0)
	idField, pField := -1, -1
	for {
		rec, err := c.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		// Initialize with detection of the ID and P-value fields
		if idField == -1 && idField == pField {
			idField, pField, err = detectFields(rec)
			if err != nil {
				return nil, fmt.Errorf("Error in line %d of the input: %s", len(values)+1, err)
			}
		}

		// Because the CSV parser requires that all lines have the same field count,
		// we don't need to check that fields pField and idField exist at this point.
		pV, err := strconv.ParseFloat(rec[pField], 64)
		if err != nil {
			return nil, err
		}

		if idField < 0 {
			values = append(values, &Value{pValue: pV, ID: ""})
		} else {
			values = append(values, &Value{pValue: pV, ID: rec[idField]})
		}
	}

	return values, nil
}

// Detects first field which can be parsed as a float as the p Value
//
func detectFields(input []string) (int, int, error) {
	idField, pField := -1, -1

	for k, v := range input {
		// First check for integers, which should be an ID field
		if _, err := strconv.ParseInt(v, 10, 64); err == nil && idField == -1 {
			idField = k
			continue
		}
		if _, err := strconv.ParseFloat(v, 64); err == nil && pField == -1 {
			pField = k
			continue
		}
		if idField == -1 {
			idField = k
			continue
		}
	}

	// Tolerate a missing idField
	// if idField == -1 {
	// 	return idField, pField, fmt.Errorf("Could not detect ID field")
	// }

	if pField == -1 {
		return idField, pField, fmt.Errorf("Could not detect P-value field")
	}

	return idField, pField, nil
}
