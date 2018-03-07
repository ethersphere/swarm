/*
   Copyright 2010-2016 Intel Corporation

   This software is licensed to you in accordance
   with the agreement between you and Intel Corporation.

   Alternatively, you can use this file in compliance
   with the Apache license, Version 2.


   Apache License, Version 2.0

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

#include "EventLog.h"


namespace intel_dal
{

#ifdef _WIN32

void WriteToEventLog(JHI_EVENT_LOG_TYPE EventType, DWORD MessageID)
{
	HANDLE  hEventSource;
	//LPCWSTR msg = LPCWSTR(Message);

	hEventSource = RegisterEventSource(NULL, JHI_EVENT_LOG_SVCNAME);
	
	if (hEventSource != NULL)
	{
		/* Write to event log. */	

		ReportEvent(hEventSource,        // event log handle
					EventType,			 // event type
					0,                   // event category
					MessageID,			 // event identifier
					NULL,                // no security identifier
					0,                   // num of strings in the message array
					0,                   // no binary data
					NULL,				 // the message
					NULL);               // no binary data


		DeregisterEventSource(hEventSource);
	}

}
#elif defined (ANDROID)
void WriteToEventLog(JHI_EVENT_LOG_TYPE EventType, uint32_t MessageID)
{
	__android_log_print(EventType, JHI_EVENT_LOG_SVCNAME, "%d", MessageID);
}
#else //Linux
void WriteToEventLog(JHI_EVENT_LOG_TYPE EventType, uint32_t MessageID)
{
	openlog(JHI_EVENT_LOG_SVCNAME, LOG_CONS | LOG_PID | LOG_NDELAY, LOG_USER);
	syslog(EventType, "%d", MessageID);
	closelog();
}
#endif // _WIN32

}