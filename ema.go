package main

/* Formula: EMA = Price(t) * k + EMA(y) * (1 - k)
 * t = today, y = yesterday, N = number of days in EMA, k = 2/(2N+1)
 */
func CalcEMA(today float64, yesterday float64, interval float64) float64 {
	var k float64 = 2.0 / (2.0*interval + 1.0)
	return today*k + yesterday*(1-k)
}

func InitAllEMAs(emaArray *map[uint]float64) {
	for _, emaStep := range EMA_STEP {
		(*emaArray)[emaStep] = 0
	}
}

func CalcAllEMAs(emaArray *map[uint]float64, last float64) {
	for _, emaStep := range EMA_STEP {
		if (*emaArray)[emaStep] == 0 {
			(*emaArray)[emaStep] = last
		} else {
			(*emaArray)[emaStep] = CalcEMA(last, (*emaArray)[emaStep], float64(emaStep))
		}
	}
}

func CalcAllMAs(emaArray *map[uint]float64, last float64) {
	for _, emaStep := range EMA_STEP {
		if (*emaArray)[emaStep] == 0 {
			(*emaArray)[emaStep] = last
		} else {
			(*emaArray)[emaStep] = CalcMA(last, (*emaArray)[emaStep], float64(emaStep))
		}
	}
}

func CalcMA(today float64, yesterday float64, interval float64) float64 {
	return today + (yesterday-today)/interval
}

func InitAllMAs(emaArray *map[uint]float64) {
	for _, emaStep := range EMA_STEP {
		(*emaArray)[emaStep] = 0
	}
}
