//*********** JHI Event Definitions**************
//
//  Values are 32 bit values laid out as follows:
//
//   3 3 2 2 2 2 2 2 2 2 2 2 1 1 1 1 1 1 1 1 1 1
//   1 0 9 8 7 6 5 4 3 2 1 0 9 8 7 6 5 4 3 2 1 0 9 8 7 6 5 4 3 2 1 0
//  +---+-+-+-----------------------+-------------------------------+
//  |Sev|C|R|     Facility          |               Code            |
//  +---+-+-+-----------------------+-------------------------------+
//
//  where
//
//      Sev - is the severity code
//
//          00 - Success
//          01 - Informational
//          10 - Warning
//          11 - Error
//
//      C - is the Customer code flag
//
//      R - is a reserved bit
//
//      Facility - is the facility code
//
//      Code - is the facility's status code
//
//
// Define the facility codes
//


//
// Define the severity codes
//
#define STATUS_SEVERITY_WARNING          0x2
#define STATUS_SEVERITY_INFORMATIONAL    0x1
#define STATUS_SEVERITY_ERROR            0x3


//
// MessageId: MSG_SERVICE_START
//
// MessageText:
//
// Intel(R) Dynamic Application Loader Host Interface Service started.
//
#define MSG_SERVICE_START                ((uint32_t)0x40000000L)

//
// MessageId: MSG_SERVICE_STOP
//
// MessageText:
//
// Intel(R) Dynamic Application Loader Host Interface Service stopped.
//
#define MSG_SERVICE_STOP                 ((uint32_t)0x40000001L)

//
// MessageId: MSG_SERVICE_RESET
//
// MessageText:
//
// Intel(R) Dynamic Application Loader Host Interface Service has been reset.
//
#define MSG_SERVICE_RESET                ((uint32_t)0x40000002L)

//
// MessageId: MSG_SPOOLER_NOT_FOUND
//
// MessageText:
//
// Intel(R) Dynamic Application Loader Host Interface Service initialization failure - the spooler applet wasn't found.
//
#define MSG_SPOOLER_NOT_FOUND            ((uint32_t)0xC0000003L)

//
// MessageId: MSG_INVALID_SPOOLER
//
// MessageText:
//
// Intel(R) Dynamic Application Loader Host Interface Service initialization failure - the spooler applet is invalid.
//
#define MSG_INVALID_SPOOLER              ((uint32_t)0xC0000004L)

//
// MessageId: MSG_FW_COMMUNICATION_ERROR
//
// MessageText:
//
// Intel(R) Dynamic Application Loader Host Interface Service initialization failure - there is no communication with FW.
//
#define MSG_FW_COMMUNICATION_ERROR       ((uint32_t)0xC0000005L)

//
// MessageId: MSG_REGISTRY_READ_ERROR
//
// MessageText:
//
// Intel(R) Dynamic Application Loader Host Interface Service has failed to read from registry.
//
#define MSG_REGISTRY_READ_ERROR          ((uint32_t)0xC0000006L)

//
// MessageId: MSG_REGISTRY_WRITE_ERROR
//
// MessageText:
//
// Intel(R) Dynamic Application Loader Host Interface Service has failed to write to registry.
//
#define MSG_REGISTRY_WRITE_ERROR         ((uint32_t)0xC0000007L)

//
// MessageId: MSG_REPOSITORY_NOT_FOUND
//
// MessageText:
//
// Intel(R) Dynamic Application Loader Host Interface Service cannot find applet repository.
//
#define MSG_REPOSITORY_NOT_FOUND         ((uint32_t)0xC0000008L)

//
// MessageId: MSG_INSTALL_FAILURE
//
// MessageText:
//
// Intel(R) Dynamic Application Loader Host Interface Service has failed to install applet.
//
#define MSG_INSTALL_FAILURE              ((uint32_t)0x80000009L)

//
// MessageId: MSG_CREATE_SESSION_FAILURE
//
// MessageText:
//
// Intel(R) Dynamic Application Loader Host Interface Service has failed to create an applet session.
//
#define MSG_CREATE_SESSION_FAILURE       ((uint32_t)0x8000000AL)

//
// MessageId: MSG_CONNECT_FAILURE
//
// MessageText:
//
// Intel(R) Dynamic Application Loader Host Interface Service has encountered an internal connection problem.
//
#define MSG_CONNECT_FAILURE              ((uint32_t)0xC000000BL)

