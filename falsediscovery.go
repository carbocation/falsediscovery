package falsediscovery

import (
	"fmt"
	"math"
	"sort"
)

type TestStatistic interface {
	P() float64
	AdjustedP() float64
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
		kRank := 1 + k
		pValues[k].SetCriticalValue(float64(kRank) / float64(nTests) * FDR)

		// Choosing the min of 1 and the Q value is similar to other
		// implementations, e.g.,
		// https://github.com/StoreyLab/qvalue/blob/master/R/qvalue.R#L118
		pValues[k].SetAdjustedPValue(math.Min(1.0, float64(nTests)/float64(kRank)*pValues[k].P()))
	}

	// Enforce monotonicity. See https://stats.stackexchange.com/a/402217/8853
	for i := 0; ; i++ {
		lastP := 0.0
		backtracked := false
		for k := range pValues {
			x := pValues[k].AdjustedP()
			if x < lastP {
				// Nonmonotonicity detected. Set the prior P value to the
				// current (better) P value.
				pValues[k-1].SetAdjustedPValue(x)
				backtracked = true
			}

			lastP = x

		}
		if !backtracked {
			break
		}
		if i > 200 {
			return fmt.Errorf("nonmonotonic P values. Could not resolve after 200 iterations")
		}
	}

	return nil
}
