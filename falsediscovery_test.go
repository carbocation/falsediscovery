package falsediscovery

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
)

func TestMain(m *testing.M) {
	rand.Seed(31337)

	os.Exit(m.Run())
}

func TestBenjaminiHochbergBadInput(t *testing.T) {
	if err := BenjaminiHochberg(1.2, &Value{}); err == nil {
		t.Error("FDR >= 1.0 should be rejected")
	}
}

func TestBenjaminiHochberg(t *testing.T) {
	N := 249
	FDR := 0.05

	pValues := make([]*Value, 0, N)
	for i := 0; i < N; i++ {
		pValues = append(pValues, &Value{ID: strconv.Itoa(i), pValue: rand.Float64()})
	}
	pValues = append(pValues, &Value{ID: strconv.Itoa(N), pValue: 0.0000000000001})

	tStats := ValuesToTestStatistics(pValues)

	if err := BenjaminiHochberg(FDR, tStats...); err != nil {
		t.Error(err)
	}

	for k, v := range tStats {
		pValues[k] = v.(*Value)
	}

	any := false
	for _, v := range pValues {
		if v.Significant() {
			any = true
			t.Log(v.ID, v.P(), v.criticalValue, v.Significant())
		}
	}

	if any == false {
		t.Log("None of the values were significant after adjustment for FDR", FDR)
	}
}

func TestDelimiterDetector(t *testing.T) {
	FDR := 0.05
	inputs := []struct {
		input   string
		divider string
	}{
		{input: `3	0.005
4	0.34
Six	0.11`, divider: "tab"},
		{input: `3 0.005
4 0.34
Six 0.11`, divider: "space"},
		{input: `3 0.005
4  0.34
Six 0.11`, divider: "mixed-space"},
		{input: `3,0.005
4,0.34
Six,0.11`, divider: "comma"},
	}

	for _, input := range inputs {
		t.Run(input.divider, func(t *testing.T) {
			values, err := ParseDelimitedInput(input.input)
			if input.divider == "mixed-space" {
				if err == nil {
					t.Error(fmt.Errorf("mixed-space should not be parsed correctly"))
				}
			} else {
				if err != nil {
					t.Error(err)
				}
			}

			significanceHelper(t, FDR, values)
		})
	}
}

func significanceHelper(t *testing.T, FDR float64, values []*Value) {
	tStats := ValuesToTestStatistics(values)
	if err := BenjaminiHochberg(FDR, tStats...); err != nil {
		t.Error(err)
	}

	for k, v := range tStats {
		values[k] = v.(*Value)
	}

	any := false
	for _, v := range values {
		if v.Significant() {
			any = true
			t.Log(v.ID, v.P(), v.criticalValue, v.Significant())
		}
	}

	if any == false {
		t.Log("None of the values were significant after adjustment for FDR", FDR)
	}
}
