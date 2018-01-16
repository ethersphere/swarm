package metrics

import (
	"fmt"
	"log"
	"math"
	uurl "net/url"
	"sort"
	"time"

	"github.com/influxdata/influxdb/client"
)

type reporter struct {
	reg      Registry
	interval time.Duration

	url       uurl.URL
	database  string
	username  string
	password  string
	namespace string
	tags      map[string]string

	client *client.Client

	cache map[string]int64
}

// InfluxDB starts a InfluxDB reporter which will post the from the given registry at each d interval.
func InfluxDB(r Registry, d time.Duration, url, database, username, password, namespace string) {
	InfluxDBWithTags(r, d, url, database, username, password, namespace, nil)
}

// InfluxDBWithTags starts a InfluxDB reporter which will post the from the given registry at each d interval with the specified tags
func InfluxDBWithTags(r Registry, d time.Duration, url, database, username, password, namespace string, tags map[string]string) {
	u, err := uurl.Parse(url)
	if err != nil {
		log.Printf("unable to parse InfluxDB url %s. err=%v", url, err)
		return
	}

	rep := &reporter{
		reg:       r,
		interval:  d,
		url:       *u,
		database:  database,
		username:  username,
		password:  password,
		namespace: namespace,
		tags:      tags,
		cache:     make(map[string]int64),
	}
	if err := rep.makeClient(); err != nil {
		log.Printf("unable to make InfluxDB client. err=%v", err)
		return
	}

	rep.run()
}

func (r *reporter) makeClient() (err error) {
	r.client, err = client.NewClient(client.Config{
		URL:      r.url,
		Username: r.username,
		Password: r.password,
	})

	return
}

func (r *reporter) run() {
	intervalTicker := time.Tick(r.interval)
	pingTicker := time.Tick(time.Second * 5)

	for {
		select {
		case <-intervalTicker:
			if err := r.send(); err != nil {
				log.Printf("unable to send to InfluxDB. err=%v", err)
			}
		case <-pingTicker:
			_, _, err := r.client.Ping()
			if err != nil {
				log.Printf("got error while sending a ping to InfluxDB, trying to recreate client. err=%v", err)

				if err = r.makeClient(); err != nil {
					log.Printf("unable to make InfluxDB client. err=%v", err)
				}
			}
		}
	}
}

func (r *reporter) send() error {
	var pts []client.Point

	r.reg.Each(func(name string, i interface{}) {
		now := time.Now()
		namespace := r.namespace

		switch metric := i.(type) {
		case Counter:
			v := metric.Count()
			l := r.cache[name]
			pts = append(pts, client.Point{
				Measurement: fmt.Sprintf("%s%s.count", namespace, name),
				Tags:        r.tags,
				Fields: map[string]interface{}{
					"value": v - l,
				},
				Time: now,
			})
			r.cache[name] = v
		case Gauge:
			ms := metric.Snapshot()
			pts = append(pts, client.Point{
				Measurement: fmt.Sprintf("%s%s.gauge", namespace, name),
				Tags:        r.tags,
				Fields: map[string]interface{}{
					"value": ms.Value(),
				},
				Time: now,
			})
		case GaugeFloat64:
			ms := metric.Snapshot()
			pts = append(pts, client.Point{
				Measurement: fmt.Sprintf("%s%s.gauge", namespace, name),
				Tags:        r.tags,
				Fields: map[string]interface{}{
					"value": ms.Value(),
				},
				Time: now,
			})
		case Histogram:
			ms := metric.Snapshot()
			ps := ms.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999, 0.9999})
			pts = append(pts, client.Point{
				Measurement: fmt.Sprintf("%s%s.histogram", namespace, name),
				Tags:        r.tags,
				Fields: map[string]interface{}{
					"count":    ms.Count(),
					"max":      ms.Max(),
					"mean":     ms.Mean(),
					"min":      ms.Min(),
					"stddev":   ms.StdDev(),
					"variance": ms.Variance(),
					"p50":      ps[0],
					"p75":      ps[1],
					"p95":      ps[2],
					"p99":      ps[3],
					"p999":     ps[4],
					"p9999":    ps[5],
				},
				Time: now,
			})
		case Meter:
			ms := metric.Snapshot()
			pts = append(pts, client.Point{
				Measurement: fmt.Sprintf("%s%s.meter", namespace, name),
				Tags:        r.tags,
				Fields: map[string]interface{}{
					"count": ms.Count(),
					"m1":    ms.Rate1(),
					"m5":    ms.Rate5(),
					"m15":   ms.Rate15(),
					"mean":  ms.RateMean(),
				},
				Time: now,
			})
		case Timer:
			ms := metric.Snapshot()
			ps := ms.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999, 0.9999})
			pts = append(pts, client.Point{
				Measurement: fmt.Sprintf("%s%s.timer", namespace, name),
				Tags:        r.tags,
				Fields: map[string]interface{}{
					"count":    ms.Count(),
					"max":      ms.Max(),
					"mean":     ms.Mean(),
					"min":      ms.Min(),
					"stddev":   ms.StdDev(),
					"variance": ms.Variance(),
					"p50":      ps[0],
					"p75":      ps[1],
					"p95":      ps[2],
					"p99":      ps[3],
					"p999":     ps[4],
					"p9999":    ps[5],
					"m1":       ms.Rate1(),
					"m5":       ms.Rate5(),
					"m15":      ms.Rate15(),
					"meanrate": ms.RateMean(),
				},
				Time: now,
			})
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

				ps := []int64{}

				thresholdBoundary := max

				for _, pct := range percentiles {
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
						ps = append(ps, thresholdBoundary)
						//fmt.Fprintf(w, "%s.%s.upper_%s %d %d\n", c.Prefix, name, k, thresholdBoundary, now)
					} else {
						ps = append(ps, thresholdBoundary)
						//fmt.Fprintf(w, "%s.%s.lower_%s %d %d\n", c.Prefix, name, k, thresholdBoundary, now)
					}
				}

				sum := cumulativeValues[count-1]
				mean := float64(sum) / float64(count)

				pts = append(pts, client.Point{
					Measurement: fmt.Sprintf("%s%s.span", namespace, name),
					Tags:        r.tags,
					Fields: map[string]interface{}{
						"count": count,
						"max":   max,
						"mean":  mean,
						"min":   min,
						"p50":   ps[0],
						"p95":   ps[1],
						"p99":   ps[2],
					},
					Time: now,
				})
			}

		}
	})

	bps := client.BatchPoints{
		Points:   pts,
		Database: r.database,
	}

	_, err := r.client.Write(bps)
	return err
}
