package falsediscovery

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"testing"
)

const TestEpsilon = 0.0000001

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

func TestAdjustedP(t *testing.T) {
	inputs := []struct {
		name     string
		input    string
		expected map[string]float64
	}{
		{
			`insig`,
			`A 0.01
B 0.03
C 0.05
D 0.12
E 0.2
F 0.3`,
			map[string]float64{"A": 0.06, "B": 0.09, "C": 0.10, "D": 0.18, "E": 0.24, "F": 0.3},
		},
		{
			`insig`,
			`A 0.001052588
B 0.002887613
C 0.004249
D 0.2431061
E 0.4909836
F 0.589325`,
			map[string]float64{
				"A": 0.006315528,
				"B": 0.008498000,
				"C": 0.008498000,
				"D": 0.364659150,
				"E": 0.589180320,
				"F": 0.589325000},
		},
		{
			`somesig`,
			`A 0.006920873
B 0.0074964
C 0.190987137
D 0.204375554
E 0.376956048
F 0.742802872
G 0.743635244
H 0.883128442`,
			map[string]float64{
				"A": 0.0299856,
				"B": 0.0299856,
				"C": 0.4087511,
				"D": 0.4087511,
				"E": 0.6031297,
				"F": 0.8498689,
				"G": 0.8498689,
				"H": 0.8831284,
			},
		},
	}

	for _, v := range inputs {
		t.Run(v.name, func(t *testing.T) {
			FDR := 0.05
			values, err := ParseDelimitedInput(v.input)
			if err != nil {
				t.Error(err)
			}

			tStats := ValuesToTestStatistics(values)
			if err := BenjaminiHochberg(FDR, tStats...); err != nil {
				t.Error(err)
			}

			for id, stat := range tStats {
				value := stat.(*Value)
				if math.Abs(value.AdjustedP()-v.expected[value.ID]) > TestEpsilon {
					t.Error("Entry", id, value.AdjustedP(), "is not equal to", v.expected[value.ID])
				}
			}
		})
	}
}

func TestChangedDelimiter(t *testing.T) {
	input := `A 0.2
B 0.3
C,0.25`
	_, err := ParseDelimitedInput(input)
	if err == nil {
		t.Error("Should have detected an issue with a change in the delimiter")
	}
}
func TestNoPValueInFirstLine(t *testing.T) {
	input := `A X
B 0.3
C`
	_, err := ParseDelimitedInput(input)
	if err == nil {
		t.Error("Should have detected the lack of P value in the first line")
	}
}

func TestDetectFieldsMissingID(t *testing.T) {
	input := []string{`0.025`}
	_, _, err := detectFields(input)
	if err != nil {
		t.Error("Should have tolerated the lack of ID field")
	}
}

func TestNoPValueInSubsequentLine(t *testing.T) {
	input := `A 0.02
B 0.2
C X`
	_, err := ParseDelimitedInput(input)
	if err == nil {
		t.Error("Should have detected the lack of P value in the third line")
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
			t.Log(v.ID, v.P(), v.CriticalValue(), v.Significant())
		}
	}

	if any == false {
		t.Log("None of the values were significant after adjustment for FDR", FDR)
	}
}
