package swarmdblog

import (
	"log/syslog"
)

type Logger struct {
	Debug_logger *syslog.Writer  // unstructured raw stuff, summary
	Trace_logger *syslog.Writer  // unstructured raw stuff, detailed
	Swarmdb_logger *syslog.Writer    // server put/get/insert operations, success/fail (structured)
	Client_logger *syslog.Writer     // client JSON input messages, structured (?)
	Cloud_logger *syslog.Writer      // communications between SWARMDB nodes, structured (?)
	Miner_logger *syslog.Writer      // claim/validation messages with WLK, structured (?)
	Netstats_logger *syslog.Writer   // Swarmlog summary,  structured
	TCP_logger *syslog.Writer        // TCP Server operations only 
	HTTP_logger *syslog.Writer       // HTTP Server operations only
}

func (l *Logger) Debug(info string) (err error){
	l.Debug_logger, err = syslog.Dial("tcp", "127.0.0.1:5000", syslog.LOG_ERR, "wolk-debug")
	if err != nil {
		return err
	}

	err2 := l.Debug_logger.Info(info)
	if err2 != nil {
		return err2
	}
	return nil
}

func (l *Logger) Trace(info string) (err error){
	l.Trace_logger, err = syslog.Dial("tcp", "127.0.0.1:5000", syslog.LOG_ERR, "wolk-trace")
	if err != nil {
		return err
	}

	err2 := l.Trace_logger.Info(info)
	if err2 != nil {
		return err2
	}
	return nil
}

func (l *Logger) Cloud(info string) (err error){
	l.Cloud_logger, err = syslog.Dial("tcp", "127.0.0.1:5000", syslog.LOG_ERR, "wolk-cloud")
	if err != nil {
		return err
	}

	err2 := l.Cloud_logger.Info(info)
	if err2 != nil {
		return err2
	}
	return nil
}

func (l *Logger) Mining(info string) (err error){
	l.Miner_logger, err = syslog.Dial("tcp", "127.0.0.1:5000", syslog.LOG_ERR, "wolk-mining")
	if err != nil {
		return err
	}

	err2 := l.Miner_logger.Info(info)
	if err2 != nil {
		return err2
	}
	return nil
}


func (l *Logger) Netstats(info string) (err error){
	l.Netstats_logger, err = syslog.Dial("tcp", "127.0.0.1:5000", syslog.LOG_ERR, "wolk-netstats")
	if err != nil {
		return err
	}

	err2 := l.Netstats_logger.Info(info)
	if err2 != nil {
		return err2
	}
	return nil
}

func (l *Logger) TCP(info string) (err error){
	l.TCP_logger, err = syslog.Dial("tcp", "127.0.0.1:5000", syslog.LOG_ERR, "wolk-tcp")
	if err != nil {
		return err
	}

	err2 := l.TCP_logger.Info(info)
	if err2 != nil {
		return err2
	}
	return nil
}

func (l *Logger) HTTP(info string) (err error){
	l.HTTP_logger, err = syslog.Dial("tcp", "127.0.0.1:5000", syslog.LOG_ERR, "wolk-http")
	if err != nil {
		return err
	}

	err2 := l.HTTP_logger.Info(info)
	if err2 != nil {
		return err2
	}
	return nil
}

func NewLogger() (l *Logger) {
	l = new(Logger)
	return l
}

