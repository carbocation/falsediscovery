package falsediscovery

import (
	"fmt"
	"sort"
)

type Value struct {
	ID            int
	P             float64
	CriticalValue float64
	Significant   bool
}

func BenjaminiHochberg(pValues []Value, FDR float64) ([]Value, error) {
	if FDR >= 1.0 || FDR <= 0.0 {
		return nil, fmt.Errorf("BenjaminiHochberg: FDR must be on the range (0, 1)")
	}

	sort.Slice(pValues, func(i, j int) bool { return pValues[i].P < pValues[j].P })

	nTests := len(pValues)
	for k, v := range pValues {
		pValues[k].CriticalValue = float64(1+k) / float64(nTests) * FDR
		if v.P < pValues[k].CriticalValue {
			pValues[k].Significant = true
		}
	}

	return pValues, nil
}
