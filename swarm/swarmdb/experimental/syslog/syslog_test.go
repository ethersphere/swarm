package syslog

import (
	"fmt"
	"testing"
)

func TestSyslog(t *testing.T) {
	 
	 syslogInit() 
	 Syslog_ethereum_debug("test syslog Syslog_ethereum_debug");
	 fmt.Printf("test syslog Syslog_ethereum_debug")
	 

}