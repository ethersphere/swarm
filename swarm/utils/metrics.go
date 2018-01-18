// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

//NOTE: The metrics package for swarm should align with the general go-ethereum metrics,
//which uses the `rcrowley/go-metrics` package. The latter has limitations with
//aggregating and resetting metrics, which sparked some research on the swarm team
//regarding metrics. Therefore, this swarm metrics package introduces some indirection,
//which abstracts a metrics API, allowing to use a different metrics library if needed.
//If ultimately the go-metrics issues are solved,
//it may be omitted in favor of directly using the `metrics/metrics` package in go-ethereum

package utils

import (
	"fmt"
	"reflect"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"

	gometrics "github.com/nonsense/go-metrics"
	influxdb "github.com/nonsense/go-metrics-influxdb"
)

var (
	MetricsEnabled = metrics.Enabled
	bzzAccount     string
	SwarmRegistry  gometrics.Registry
)

type MetricsTimer struct {
	timer  gometrics.Timer
	bucket string
}

func SetupMetrics(bzzAccount string) {
	bzzAccount = bzzAccount[:12]
	SwarmRegistry = gometrics.NewRegistry()

	//go influxdb.InfluxDBWithTags(SwarmRegistry, 5*time.Second, "http://10.0.1.245:8086", "metrics", "admin", "admin", "swarm.", map[string]string{
	//"host": bzzAccount,
	//})

	go influxdb.InfluxDBWithTags(SwarmRegistry, 5*time.Second, "http://localhost:8086", "metrics", "admin", "admin", "swarm.", map[string]string{
		"host": bzzAccount,
	})
}

func Gauge(bucket string, value interface{}) {
	if metrics.Enabled {
		g := gometrics.GetOrRegisterGauge(bucket, SwarmRegistry)
		if val, ok := value.(int); ok {
			g.Update(int64(val))
			return
		}
		if val, ok := value.(uint64); ok {
			g.Update(int64(val))
			return
		}
		if val, ok := value.(int64); ok {
			g.Update(val)
			return
		}
		log.Warn(fmt.Sprintf("Invalid value type for Gauge %v", reflect.TypeOf(value)))
	}
}

func Increment(bucket string) {
	if metrics.Enabled {
		c := gometrics.GetOrRegisterCounter(bucket, SwarmRegistry)
		c.Inc(1)
	}
}

func Histogram(bucket string, value interface{}) {
	if metrics.Enabled {
		h := gometrics.GetOrRegisterHistogram(bucket, SwarmRegistry, gometrics.NewUniformSample(100))
		h.Update(int64(value.(int)))
	}
}

func NewTimer(bucket string) MetricsTimer {
	return MetricsTimer{bucket: bucket, timer: gometrics.GetOrRegisterTimer(bucket, SwarmRegistry)}
}

func StartTimer(bucket string) MetricsTimer {
	if metrics.Enabled {
		return NewTimer(bucket)
	}
	return MetricsTimer{}
}

func SendTimer(timer MetricsTimer) {
	if metrics.Enabled && timer.bucket != "" {
		timer.timer.UpdateSince(time.Now())
	}
}
