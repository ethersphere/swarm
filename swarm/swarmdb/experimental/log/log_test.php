<?
// php logging
$loggers = array("wolk-debug", "wolk-trace", "wolk-cloud", "wolk-mining", "wolk-netstats", "wolk-tcp", "wolk-http");

foreach ($loggers as $logger) {
    openlog($logger, LOG_PID | LOG_ODELAY, LOG_LOCAL4); // on log06
    $logStr = str_replace("wolk-", "", $logger)." string";
    syslog(LOG_INFO, $logStr);
    closelog();
    echo "LOG: $logger => [$logStr]\n";
}
?>