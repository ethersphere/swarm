package syslog

import (
	"fmt"
	"log/syslog"
)

var Syslog_Writer_ethereum_debug    *syslog.Writer
var Syslog_Writer_ethereum_trace    *syslog.Writer
var Syslog_Writer_ethereum_swarmdb  *syslog.Writer
var Syslog_Writer_ethereum_cloud    *syslog.Writer
var Syslog_Writer_ethereum_netstat  *syslog.Writer
var Syslog_Writer_ethereum_mining   *syslog.Writer
var Syslog_Writer_ethereum_client   *syslog.Writer
var Syslog_Writer_ethereum_tcp      *syslog.Writer
var Syslog_Writer_ethereum_http     *syslog.Writer


func Syslog_ethereum_debug(info string) {
	err := Syslog_Writer_ethereum_debug.Info(info)
	if err != nil {
		fmt.Println("Syslog_Writer_ethereum_debug err: +v%", err)
	}
}

func Syslog_ethereum_trace(info string) {
	err := Syslog_Writer_ethereum_trace.Info(info)
	if err != nil {
		fmt.Println("Syslog_Writer_ethereum_trace err: +v%", err)
	}
}

func Syslog_ethereum_swarmdb(info string) {
	err := Syslog_Writer_ethereum_swarmdb.Info(info)
	if err != nil {
		fmt.Println("Syslog_Writer_ethereum_swarmdb err: +v%", err)
	}
}

func Syslog_ethereum_cloud(info string) {
	err := Syslog_Writer_ethereum_cloud.Info(info)
	if err != nil {
		fmt.Println("Syslog_Writer_ethereum_cloud err: +v%", err)
	}
}

func Syslog_ethereum_netstat(info string) {
	err := Syslog_Writer_ethereum_netstat.Info(info)
	if err != nil {
		fmt.Println("Syslog_Writer_ethereum_netstat err: +v%", err)
	}
}

func Syslog_ethereum_mining(info string) {
	err := Syslog_Writer_ethereum_mining.Info(info)
	if err != nil {
		fmt.Println("Syslog_Writer_ethereum_mining err: +v%", err)
	}
}

func Syslog_ethereum_client(info string) {
	err := Syslog_Writer_ethereum_client.Info(info)
	if err != nil {
		fmt.Println("Syslog_Writer_ethereum_client err: +v%", err)
	}
}

func Syslog_ethereum_tcp(info string) {
	err := Syslog_Writer_ethereum_tcp.Info(info)
	if err != nil {
		fmt.Println("Syslog_Writer_ethereum_tcp err: +v%", err)
	}
}

func Syslog_ethereum_http(info string) {
	err := Syslog_Writer_ethereum_http.Info(info)
	if err != nil {
		fmt.Println("Syslog_Writer_ethereum_http err: +v%", err)
	}
}

func syslogInit() {
	var err error

	Syslog_Writer_ethereum_debug, err = syslog.Dial("tcp", "127.0.0.1:5000", syslog.LOG_ERR, "ethereum-debug")
	defer Syslog_Writer_ethereum_debug.Close()
	if err != nil {
		fmt.Println("syslog Dial Syslog_Writer_ethereum_debug err: +v%", err)
	}

	Syslog_Writer_ethereum_trace, err = syslog.Dial("tcp", "127.0.0.1:5000", syslog.LOG_ERR, "ethereum-trace")
	defer Syslog_Writer_ethereum_trace.Close()
	if err != nil {
		fmt.Println("syslog Dial Syslog_Writer_ethereum_trace err: +v%", err)
	}

	Syslog_Writer_ethereum_swarmdb, err = syslog.Dial("tcp", "127.0.0.1:5000", syslog.LOG_ERR, "ethereum-swarmdb")
	defer Syslog_Writer_ethereum_swarmdb.Close()
	if err != nil {
		fmt.Println("syslog Dial Syslog_Writer_ethereum_swarmdb err: +v%", err)
	}

	Syslog_Writer_ethereum_cloud, err = syslog.Dial("tcp", "127.0.0.1:5000", syslog.LOG_ERR, "ethereum-cloud")
	defer Syslog_Writer_ethereum_cloud.Close()
	if err != nil {
		fmt.Println("syslog Dial Syslog_Writer_ethereum_cloud err: +v%", err)
	}

	Syslog_Writer_ethereum_netstat, err = syslog.Dial("tcp", "127.0.0.1:5000", syslog.LOG_ERR, "ethereum-netstat")
	defer Syslog_Writer_ethereum_netstat.Close()
	if err != nil {
		fmt.Println("syslog Dial Syslog_Writer_ethereum_netstat err: +v%", err)
	}

	Syslog_Writer_ethereum_mining, err = syslog.Dial("tcp", "127.0.0.1:5000", syslog.LOG_ERR, "ethereum-mining")
	defer Syslog_Writer_ethereum_mining.Close()
	if err != nil {
		fmt.Println("syslog Dial Syslog_Writer_ethereum_mining err: +v%", err)
	}

	Syslog_Writer_ethereum_client, err = syslog.Dial("tcp", "127.0.0.1:5000", syslog.LOG_ERR, "ethereum-client")
	defer Syslog_Writer_ethereum_client.Close()
	if err != nil {
		fmt.Println("syslog Dial Syslog_Writer_ethereum_client err: +v%", err)
	}

	Syslog_Writer_ethereum_tcp, err = syslog.Dial("tcp", "127.0.0.1:5000", syslog.LOG_ERR, "ethereum-tcp")
	defer Syslog_Writer_ethereum_tcp.Close()
	if err != nil {
		fmt.Println("syslog Dial Syslog_Writer_ethereum_debug err: +v%", err)
	}

	Syslog_Writer_ethereum_http, err = syslog.Dial("tcp", "127.0.0.1:5000", syslog.LOG_ERR, "ethereum-http")
	defer Syslog_Writer_ethereum_http.Close()
	if err != nil {
		fmt.Println("syslog Dial Syslog_Writer_ethereum_http err: +v%", err)
	}

}






