package main

import (
	"math/rand"
	"net"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	statsd "gopkg.in/alexcesaro/statsd.v2"
)

func main() {
	setupGraphiteReporter("gometricsvsalex")

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
			t := metrics.GetOrRegisterTimer("homepage.response_time", metrics.DefaultRegistry)
			defer t.UpdateSince(time.Now())
			count := metrics.GetOrRegisterCounter("foo.counter", metrics.DefaultRegistry)
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
			t := metrics.GetOrRegisterTimer("homepage.response_time", metrics.DefaultRegistry)
			defer t.UpdateSince(time.Now())
			count := metrics.GetOrRegisterCounter("foo.counter", metrics.DefaultRegistry)
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

func setupGraphiteReporter(namespace string) {
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:2003")

	gc := metrics.GraphiteConfig{
		Addr:          addr,
		Registry:      metrics.DefaultRegistry,
		FlushInterval: 100 * time.Millisecond,
		DurationUnit:  time.Nanosecond,
		Prefix:        namespace,
		Percentiles:   []float64{0.5, 0.75, 0.95, 0.99, 0.999},
	}

	go metrics.GraphiteWithConfig(gc)
}
