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
	"os"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/log"
  statsd "gopkg.in/alexcesaro/statsd.v2"
)

const MetricsEnabledFlag = "metrics"

// MetricsEnabled is the flag specifying if metrics are enable or not.
var MetricsEnabled = false

var MetricsClient *statsd.Client

type MetricsTimer struct {
  timer statsd.Timing
  bucket string
}

// Init enables or disables the metrics system. Since we need this to run before
// any other code gets to create meters and timers, we'll actually do an ugly hack
// and peek into the command line args for the metrics flag.
func init() {
	for _, arg := range os.Args {
		if flag := strings.TrimLeft(arg, "-"); flag == MetricsEnabledFlag {
			log.Info("Enabling metrics collection for Swarm")
			MetricsEnabled = true
		}
	}
}

func SetupMetrics(bzzAccount string) {
  var err error
  MetricsClient, err = statsd.New(
    statsd.TagsFormat(statsd.InfluxDB),
    statsd.Tags("node", bzzAccount),
    statsd.Prefix("swarm.node"),
	)
  if err != nil {
    log.Warn(fmt.Sprintf("Failed to initialize metrics sub-system: %v", err))
  }
}

func TimeEvent(bucket string, value interface{}) {
  if MetricsEnabled {
    MetricsClient.Timing(bucket, value)
  }
}

func Gauge(bucket string, value interface{}) {
  if MetricsEnabled {
    MetricsClient.Gauge(bucket, value)
  }
}

func Increment(bucket string) {
  if MetricsEnabled {
    MetricsClient.Increment(bucket)
  }
}

func Histogram(bucket string, value interface{}) {
  if MetricsEnabled {
    MetricsClient.Histogram(bucket, value)
  }
}

func NewTimer(bucket string) MetricsTimer {
  return MetricsTimer{bucket: bucket, timer: MetricsClient.NewTiming()}
}

func StartTimer(bucket string) MetricsTimer {
  if MetricsEnabled {
    return NewTimer(bucket)
  }
  return MetricsTimer{}
}

func SendTimer(timer MetricsTimer) {
  if MetricsEnabled && timer.bucket != "" {
    timer.timer.Send(timer.bucket)
  }
}
