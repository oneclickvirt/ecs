package internal

import (
	"fmt"
	"math"
	"time"
)

// Welford Fast standard deviation calculation with moving window
// ref Welford, B. P. (1962). Note on a Method for Calculating Corrected Sums of Squares and Products. Technometrics, 4(3), 419–420. https://doi.org/10.1080/00401706.1962.10490022
type Welford struct {
	n                                    int       // data size
	cap                                  int       // queue capacity
	vector                               []float64 // data set
	mean                                 float64   // mean
	sum                                  float64   // sum
	eraseIndex                           int       // the value will be erased next time
	currentStdDev                        float64
	consecutiveStableIterations          int
	consecutiveStableIterationsThreshold int
	cv                                   float64
	ewmaMean                             float64
	steps                                int
	minSteps                             int
	beta                                 float64
	scale                                float64
	movingVector                         []float64 // data set
	movingAvg                            float64
}

// NewWelford recommended windowSize = cycle / sampling frequency
func NewWelford(cycle, frequency time.Duration) *Welford {
	windowSize := int(cycle / frequency)
	return &Welford{
		vector:                               make([]float64, windowSize),
		movingVector:                         make([]float64, windowSize),
		cap:                                  windowSize,
		consecutiveStableIterationsThreshold: windowSize / 3,                // 33%
		minSteps:                             windowSize * 2,                // set minimum steps with 2x windowSize.
		beta:                                 2 / (float64(windowSize) + 1), // ewma beta ratio
		scale:                                float64(time.Second / frequency),
	}
}

// Update Enter the given value into the measuring system.
// return bool stability evaluation
func (w *Welford) Update(globalAvg, value float64) bool {
	value = value * w.scale
	if w.n == w.cap {
		delta := w.vector[w.eraseIndex] - w.mean
		w.mean -= delta / float64(w.n-1)
		w.sum -= delta * (w.vector[w.eraseIndex] - w.mean)
		// the calc error is approximated to zero
		if w.sum < 0 {
			w.sum = 0
		}
		w.vector[w.eraseIndex] = globalAvg
		w.movingAvg -= w.movingVector[w.eraseIndex]
		w.movingVector[w.eraseIndex] = value
		w.movingAvg += value
		w.eraseIndex++
		if w.eraseIndex == w.cap {
			w.eraseIndex = 0
		}
	} else {
		w.vector[w.n] = globalAvg
		w.movingVector[w.n] = value
		w.movingAvg += value
		w.n++
	}
	delta := globalAvg - w.mean
	w.mean += delta / float64(w.n)
	w.sum += delta * (globalAvg - w.mean)
	w.currentStdDev = math.Sqrt(w.Variance())
	// update C.V
	if w.mean != 0 {
		w.cv = w.currentStdDev / w.mean
	}
	w.ewmaMean = value*w.beta + w.ewmaMean*(1-w.beta)
	// acc consecutiveStableIterations
	if w.n == w.cap && w.cv < 0.03 {
		w.consecutiveStableIterations++
	} else if w.consecutiveStableIterations > 0 {
		w.consecutiveStableIterations--
	}
	w.steps++
	return w.consecutiveStableIterations >= w.consecutiveStableIterationsThreshold && w.steps > w.minSteps
}

func (w *Welford) Mean() float64 {
	return w.mean
}

func (w *Welford) CV() float64 {
	return w.cv
}

func (w *Welford) Variance() float64 {
	if w.n < 2 {
		return 0
	}
	return w.sum / float64(w.n-1)
}

func (w *Welford) StandardDeviation() float64 {
	return w.currentStdDev
}

func (w *Welford) EWMA() float64 {
	return w.ewmaMean*0.5 + w.movingAvg/float64(w.n)*0.5
}

func (w *Welford) String() string {
	return fmt.Sprintf("Mean: %.2f, Standard Deviation: %.2f, C.V: %.2f, EWMA: %.2f", w.Mean(), w.StandardDeviation(), w.CV(), w.EWMA())
}
