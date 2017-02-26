package falsediscovery

import (
	"math/rand"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	rand.Seed(31337)

	os.Exit(m.Run())
}

func TestBenjaminiHochbergBadInput(t *testing.T) {
	_, err := BenjaminiHochberg([]Value{}, 1.2)
	if err == nil {
		t.Error("FDR >= 1.0 should be rejected")
	}
}

func TestBenjaminiHochberg(t *testing.T) {
	N := 249
	FDR := 0.05

	pValues := make([]Value, 0, N)
	for i := 0; i < N; i++ {
		pValues = append(pValues, Value{ID: i, P: rand.Float64()})
	}
	pValues = append(pValues, Value{ID: N, P: 0.0000000000001})

	p, err := BenjaminiHochberg(pValues, FDR)
	if err != nil {
		t.Error(err)
	}

	any := false
	for _, v := range p {
		if v.Significant {
			any = true
			t.Log(v.ID, v.P, v.CriticalValue, v.Significant)
		}
	}

	if any == false {
		t.Log("None of the values were significant after adjustment for FDR", FDR)
	}
}
