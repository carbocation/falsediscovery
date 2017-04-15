package falsediscovery

func ValuesToTestStatistics(values []*Value) []TestStatistic {
	tStats := make([]TestStatistic, len(values), len(values))
	for k, v := range values {
		tStats[k] = TestStatistic(v)
	}

	return tStats
}
