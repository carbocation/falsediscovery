package falsediscovery

import (
	"fmt"
	"sort"
)

type TestStatistic interface {
	P() float64
	SetCriticalValue(in float64)
	SetAdjustedPValue(in float64)
}

/*
BenjaminiHochberg implements the Benjamini-Hochberg procedure to control for the false
discovery rate. A "discovery" is a significant result; false discoveries are results
where the null hypothesis is incorrectly rejected. The Benjamini-Hochberg procedure
aims to set the expected proportion of false positives that you are willing to accept.

After providing a slice of values that conform to TestStatistic and supplying your
desired FDR, the slice will be updated with the critical value for each entry. If
an entry's P-value is less than the critical value, the result is considered
significant by this procedure.
*/
func BenjaminiHochberg(FDR float64, pValues ...TestStatistic) error {
	if FDR >= 1.0 || FDR <= 0.0 {
		return fmt.Errorf("BenjaminiHochberg: FDR must be on the range (0, 1)")
	}

	sort.Slice(pValues, func(i, j int) bool { return pValues[i].P() < pValues[j].P() })

	nTests := len(pValues)
	for k := range pValues {
		pValues[k].SetCriticalValue(float64(1+k) / float64(nTests) * FDR)
		pValues[k].SetAdjustedPValue(float64(nTests) / float64(1+k) * pValues[k].P())
	}

	return nil
}
