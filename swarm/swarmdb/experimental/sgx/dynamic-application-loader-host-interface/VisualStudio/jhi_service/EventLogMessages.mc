MessageIdTypedef=uint32_t
SeverityNames=(Informational=0x1:STATUS_SEVERITY_INFORMATIONAL
               Warning=0x2:STATUS_SEVERITY_WARNING
               Error=0x3:STATUS_SEVERITY_ERROR
			   )
			   
LanguageNames=(All=0x000:MSG00001)

;//*********** JHI Event Definitions**************
MessageId=0
Severity=Informational
SymbolicName=MSG_SERVICE_START
Language=All
Intel(R) Dynamic Application Loader Host Interface Service started.
.
MessageId=1
Severity=Informational
SymbolicName=MSG_SERVICE_STOP
Language=All
Intel(R) Dynamic Application Loader Host Interface Service stopped.
.
MessageId=2
Severity=Informational
SymbolicName=MSG_SERVICE_RESET
Language=All
Intel(R) Dynamic Application Loader Host Interface Service has been reset.
.
MessageId=3
Severity=Error
SymbolicName=MSG_SPOOLER_NOT_FOUND
Language=All
Intel(R) Dynamic Application Loader Host Interface Service initialization failure - the spooler applet wasn't found.
.
MessageId=4
Severity=Error
SymbolicName=MSG_INVALID_SPOOLER
Language=All
Intel(R) Dynamic Application Loader Host Interface Service initialization failure - the spooler applet is invalid.
.
MessageId=5
Severity=Error
SymbolicName=MSG_FW_COMMUNICATION_ERROR
Language=All
Intel(R) Dynamic Application Loader Host Interface Service initialization failure - there is no communication with FW.
.
MessageId=6
Severity=Error
SymbolicName=MSG_REGISTRY_READ_ERROR
Language=All
Intel(R) Dynamic Application Loader Host Interface Service has failed to read from registry.
.
MessageId=7
Severity=Error
SymbolicName=MSG_REGISTRY_WRITE_ERROR
Language=All
Intel(R) Dynamic Application Loader Host Interface Service has failed to write to registry.
.
MessageId=8
Severity=Error
SymbolicName=MSG_REPOSITORY_NOT_FOUND
Language=All
Intel(R) Dynamic Application Loader Host Interface Service cannot find applet repository.
.
MessageId=9
Severity=Warning
SymbolicName=MSG_INSTALL_FAILURE
Language=All
Intel(R) Dynamic Application Loader Host Interface Service has failed to install applet.
.
MessageId=10
Severity=Warning
SymbolicName=MSG_CREATE_SESSION_FAILURE
Language=All
Intel(R) Dynamic Application Loader Host Interface Service has failed to create an applet session.
.
MessageId=11
Severity=Error
SymbolicName=MSG_CONNECT_FAILURE
Language=All
Intel(R) Dynamic Application Loader Host Interface Service has encountered an internal connection problem.
.




