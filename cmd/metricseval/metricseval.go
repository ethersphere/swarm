package main

import (
	"math/rand"
	"net"
	"time"

	metrics "github.com/nonsense/go-metrics"
	statsd "gopkg.in/alexcesaro/statsd.v2"
)

var (
	host1reg = metrics.NewRegistry()
	host2reg = metrics.NewRegistry()
)

func main() {
	//setupGraphiteReporter("gometrics.host1", host1reg)
	//setupGraphiteReporter("gometrics.host2", host2reg)

	go metrics.InfluxDBWithTags(host1reg, 5*time.Second, "http://localhost:8086", "metrics", "admin", "admin", "infl.", map[string]string{
		"host": "host-1.lvh.me",
	})

	go metrics.InfluxDBWithTags(host2reg, 5*time.Second, "http://localhost:8086", "metrics", "admin", "admin", "infl.", map[string]string{
		"host": "host-2.lvh.me",
	})

	go percentiles()

	for j := 0; j < 5; j++ {
		go host2()
		host1()

		time.Sleep(15 * time.Second)
	}
}

func ping(host string, delay int) {
	sl := delay + rand.Intn(delay)

	time.Sleep(time.Duration(sl) * time.Millisecond)
}

func host1() {
	c, err := statsd.New(
		statsd.TagsFormat(statsd.InfluxDB),
		statsd.Tags("host", "host-1.lvh.me"),
		statsd.Prefix("alex_statsd"),
	)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 20; i++ {
		pingHomepage := func(ind int) {
			// statsd
			defer c.NewTiming().Send("homepage.response_time")
			c.Increment("foo.counter")

			// go-metrics
			t := metrics.GetOrRegisterResettingTimer("homepage.response_time", host1reg)
			defer t.UpdateSince(time.Now())
			count := metrics.GetOrRegisterCounter("foo.counter", host1reg)
			count.Inc(1)

			if ind == 10 {
				ping("http://swarm-gateways.net/", 5000)
			} else {
				ping("http://swarm-gateways.net/", 500)
			}
		}
		pingHomepage(i)

		time.Sleep(5 * time.Second)
	}
}

func host2() {
	c, err := statsd.New(
		statsd.TagsFormat(statsd.InfluxDB),
		statsd.Tags("host", "host-2.lvh.me"),
		statsd.Prefix("alex_statsd"),
	)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 20; i++ {

		pingHomepage := func(ind int) {
			// statsd
			defer c.NewTiming().Send("homepage.response_time")
			c.Increment("foo.counter")

			// go-metrics
			t := metrics.GetOrRegisterResettingTimer("homepage.response_time", host2reg)
			defer t.UpdateSince(time.Now())
			count := metrics.GetOrRegisterCounter("foo.counter", host2reg)
			count.Inc(1)

			if ind == 10 {
				ping("http://swarm-gateways.net/", 1000)
			} else {
				ping("http://swarm-gateways.net/", 100)
			}
		}
		pingHomepage(i)

		time.Sleep(5 * time.Second)
	}
}

func setupGraphiteReporter(namespace string, reg metrics.Registry) {
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:2003")

	reporter := metrics.NewGraphiteReporter(&metrics.GraphiteConfig{
		Addr:          addr,
		Registry:      reg,
		FlushInterval: 5000 * time.Millisecond,
		DurationUnit:  time.Nanosecond,
		Prefix:        namespace,
		Percentiles:   []float64{0.5, 0.75, 0.95, 0.99, 0.999},
	})

	go reporter.Flush()
}

func percentiles() {
	t := metrics.GetOrRegisterResettingTimer("rand_300ms", host2reg)
	for {
		t.Time(func() { time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond) })
	}
}
