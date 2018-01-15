package metrics

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"net"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Int64Slice []int64

func (s Int64Slice) Len() int {
	return len(s)
}
func (s Int64Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s Int64Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

// GraphiteReporter
type GraphiteReporter struct {
	gc *GraphiteConfig

	cache map[string]int64
}

func NewGraphiteReporter(gc *GraphiteConfig) *GraphiteReporter {
	return &GraphiteReporter{
		gc:    gc,
		cache: make(map[string]int64),
	}
}

// GraphiteConfig provides a container with configuration parameters for
// the Graphite exporter
type GraphiteConfig struct {
	Addr          *net.TCPAddr  // Network address to connect to
	Registry      Registry      // Registry to be exported
	FlushInterval time.Duration // Flush interval
	DurationUnit  time.Duration // Time conversion unit for durations
	Prefix        string        // Prefix to be prepended to metric names
	Percentiles   []float64     // Percentiles to export from timers and histograms
}

// Flush is a blocking exporter function which reports in the registry
// to the statsd client, flushing every d duration
func (r *GraphiteReporter) Flush() {
	defer func() {
		if rec := recover(); rec != nil {
			handlePanic(rec)
		}
	}()

	for range time.Tick(r.gc.FlushInterval) {
		if err := r.FlushOnce(); err != nil {
			log.Println(err)
		}
	}
}

func handlePanic(rec interface{}) {
	callers := ""
	for i := 2; true; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		callers = callers + fmt.Sprintf("%v:%v\n", file, line)
	}
	log.Printf("Recovered from panic: %#v \n%v", rec, callers)
}

// FlushOnce performs a single submission to Graphite, returning a
// non-nil error on failed connections. This can be used in a loop
// similar to GraphiteWithConfig for custom error handling.
func (r *GraphiteReporter) FlushOnce() error {
	c := r.gc

	now := time.Now().Unix()
	du := float64(c.DurationUnit)
	conn, err := net.DialTCP("tcp", nil, c.Addr)
	if nil != err {
		return err
	}
	defer conn.Close()
	w := bufio.NewWriter(conn)
	c.Registry.Each(func(name string, i interface{}) {
		switch metric := i.(type) {
		case Counter:
			v := metric.Count()
			l := r.cache[name]
			fmt.Println(v - l)
			fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, v-l, now)
			r.cache[name] = v
		case Gauge:
			fmt.Fprintf(w, "%s.%s.value %d %d\n", c.Prefix, name, metric.Value(), now)
		case GaugeFloat64:
			fmt.Fprintf(w, "%s.%s.value %f %d\n", c.Prefix, name, metric.Value(), now)
		case Histogram:
			h := metric.Snapshot()
			ps := h.Percentiles(c.Percentiles)
			fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, h.Count(), now)
			fmt.Fprintf(w, "%s.%s.min %d %d\n", c.Prefix, name, h.Min(), now)
			fmt.Fprintf(w, "%s.%s.max %d %d\n", c.Prefix, name, h.Max(), now)
			fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", c.Prefix, name, h.Mean(), now)
			fmt.Fprintf(w, "%s.%s.std-dev %.2f %d\n", c.Prefix, name, h.StdDev(), now)
			for psIdx, psKey := range c.Percentiles {
				key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
				fmt.Fprintf(w, "%s.%s.%s-percentile %.2f %d\n", c.Prefix, name, key, ps[psIdx], now)
			}
		case Meter:
			m := metric.Snapshot()
			fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, m.Count(), now)
			fmt.Fprintf(w, "%s.%s.one-minute %.2f %d\n", c.Prefix, name, m.Rate1(), now)
			fmt.Fprintf(w, "%s.%s.five-minute %.2f %d\n", c.Prefix, name, m.Rate5(), now)
			fmt.Fprintf(w, "%s.%s.fifteen-minute %.2f %d\n", c.Prefix, name, m.Rate15(), now)
			fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", c.Prefix, name, m.RateMean(), now)
		case Timer:
			t := metric.Snapshot()
			ps := t.Percentiles(c.Percentiles)
			fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, t.Count(), now)
			fmt.Fprintf(w, "%s.%s.min %d %d\n", c.Prefix, name, t.Min()/int64(du), now)
			fmt.Fprintf(w, "%s.%s.max %d %d\n", c.Prefix, name, t.Max()/int64(du), now)
			fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", c.Prefix, name, t.Mean()/du, now)
			fmt.Fprintf(w, "%s.%s.std-dev %.2f %d\n", c.Prefix, name, t.StdDev()/du, now)
			for psIdx, psKey := range c.Percentiles {
				key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
				fmt.Fprintf(w, "%s.%s.%s-percentile %.2f %d\n", c.Prefix, name, key, ps[psIdx], now)
			}
			fmt.Fprintf(w, "%s.%s.one-minute %.2f %d\n", c.Prefix, name, t.Rate1(), now)
			fmt.Fprintf(w, "%s.%s.five-minute %.2f %d\n", c.Prefix, name, t.Rate5(), now)
			fmt.Fprintf(w, "%s.%s.fifteen-minute %.2f %d\n", c.Prefix, name, t.Rate15(), now)
			fmt.Fprintf(w, "%s.%s.mean-rate %.2f %d\n", c.Prefix, name, t.RateMean(), now)
		case ResettingTimer:
			t := metric.Snapshot()
			sort.Sort(Int64Slice(t.Values()))

			val := t.Values()
			count := len(val)
			if count > 0 {
				min := val[0]
				max := val[count-1]

				cumulativeValues := make([]int64, count)
				cumulativeValues[0] = min
				for i := 1; i < count; i++ {
					cumulativeValues[i] = val[i] + cumulativeValues[i-1]
				}

				percentiles := map[string]float64{
					"50": 50,
					"95": 95,
					"99": 99,
				}

				thresholdBoundary := max

				for k, pct := range percentiles {
					if count > 1 {
						var abs float64
						if pct >= 0 {
							abs = pct
						} else {
							abs = 100 + pct
						}
						// poor man's math.Round(x):
						// math.Floor(x + 0.5)
						indexOfPerc := int(math.Floor(((abs / 100.0) * float64(count)) + 0.5))
						if pct >= 0 {
							indexOfPerc -= 1 // index offset=0
						}
						thresholdBoundary = val[indexOfPerc]
					}

					if pct > 0 {
						fmt.Fprintf(w, "%s.%s.upper_%s %d %d\n", c.Prefix, name, k, thresholdBoundary, now)
					} else {
						fmt.Fprintf(w, "%s.%s.lower_%s %d %d\n", c.Prefix, name, k, thresholdBoundary, now)
					}
				}

				sum := cumulativeValues[count-1]
				mean := float64(sum) / float64(count)

				fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, count, now)
				fmt.Fprintf(w, "%s.%s.min %d %d\n", c.Prefix, name, min, now)
				fmt.Fprintf(w, "%s.%s.max %d %d\n", c.Prefix, name, max, now)
				fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", c.Prefix, name, mean, now)
			}
		}
		w.Flush()
	})
	return nil
}
