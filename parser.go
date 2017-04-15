package falsediscovery

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Value struct {
	ID            string
	pValue        float64
	criticalValue float64
}

func (v *Value) SetCriticalValue(in float64) {
	v.criticalValue = in
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

// By way of https://play.golang.org/p/UwR9nF1kv_
func GuessDelimiter(lines []string) rune {
	if len(lines) < 2 {
		return ' '
	}
	for _, c := range " ,\t|" {
		followStd := true
		std := Parse([]rune(lines[0]), c)
		if std < 2 {
			// Must have at least 2 fields for our purposes
			continue
		}
		//fmt.Println(std)
		for i := 1; i < len(lines); i++ {
			if std != Parse([]rune(lines[i]), c) {
				//fmt.Println(Parse(lines[i], c))
				followStd = false
				break
			}
			//fmt.Println(lines[i], i)
		}
		if followStd {
			return c
		}
	}
	return ' '
}

// By way of https://play.golang.org/p/UwR9nF1kv_
func Parse(line []rune, d rune) int {
	field := 1
	inQuote := false
	for i := 0; i < len(line); i++ {
		if line[i] == '"' {
			if i == len(line)-1 || line[i+1] != '"' {
				inQuote = !inQuote
				continue
			} else {
				i++
				continue
			}
		}
		if !inQuote && line[i] == d {
			field++
		}
	}
	return field
}

func ParseDelimitedInput(input string) ([]*Value, error) {
	// TODO: Use a better detector of delimiters. For example, should
	// be able to detect spaces as the delimiter even if there is a
	// variable number of spaces between fields.

	lines := strings.Split(input, "\n")
	delim := GuessDelimiter(lines)

	c := csv.NewReader(strings.NewReader(input))
	c.Comma = delim

	values := make([]*Value, 0)
	idField, pField := -1, -1
	for {
		rec, err := c.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if idField == pField {
			idField, pField = detectFields(rec)
			if idField == pField {
				return nil, fmt.Errorf("Could not detect ID and p-Value fields")
			}
		}

		pV, err := strconv.ParseFloat(rec[pField], 64)
		if err != nil {
			return nil, err
		}

		values = append(values, &Value{pValue: pV, ID: rec[idField]})
	}

	return values, nil
}

// Detects first field with . as p Value, first field without . as ID
func detectFields(input []string) (idField, pField int) {
	for k, v := range input {
		if _, err := strconv.ParseFloat(v, 64); err == nil && pField == 0 {
			pField = k
		} else if idField == 0 {
			idField = k
		}
	}

	return
}
