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

#ifndef __EVENT_LOG_H
#define __EVENT_LOG_H

#ifdef _WIN32

#include <Windows.h>
#include "EventLogMessages.h" // auto generated form mc file

#define PROVIDER_NAME L"EventLogMessages"

#elif defined (ANDROID)

#include <android/log.h>

#else

#include <syslog.h>

#endif // _WIN32

#include "EventLogMessages.h" // auto generated form mc file

#include <string>

namespace intel_dal
{

#ifdef _WIN32

// logging levels
typedef enum _JHI_EVENT_LOG_TYPE
{
	JHI_EVENT_LOG_ERROR	= EVENTLOG_ERROR_TYPE,
	JHI_EVENT_LOG_WARNING = EVENTLOG_WARNING_TYPE,
	JHI_EVENT_LOG_INFORMATION = EVENTLOG_INFORMATION_TYPE

} JHI_EVENT_LOG_TYPE;

#ifdef SCHANNEL_OVER_SOCKET // emulation mode
#define JHI_EVENT_LOG_SVCNAME TEXT("IntelDalJhi_Emulation")
#else
#define JHI_EVENT_LOG_SVCNAME TEXT("IntelDalJhi")
#endif //SCHANNEL_OVER_SOCKET

void WriteToEventLog(JHI_EVENT_LOG_TYPE EventType, DWORD MessageID);

#else //!_WIN32
// logging levels
#ifdef __ANDROID__
typedef enum _JHI_EVENT_LOG_TYPE
{
        JHI_EVENT_LOG_ERROR       = ANDROID_LOG_ERROR,
        JHI_EVENT_LOG_WARNING     = ANDROID_LOG_WARN,
        JHI_EVENT_LOG_INFORMATION = ANDROID_LOG_INFO
} JHI_EVENT_LOG_TYPE;
#else
typedef enum _JHI_EVENT_LOG_TYPE
{
        JHI_EVENT_LOG_ERROR     = LOG_ERR,
        JHI_EVENT_LOG_WARNING = LOG_WARNING,
        JHI_EVENT_LOG_INFORMATION = LOG_INFO
} JHI_EVENT_LOG_TYPE;
#endif //ANDROID

#ifdef SCHANNEL_OVER_SOCKET // emulation mode
# define JHI_EVENT_LOG_SVCNAME "IntelDalJhi_Emulation"
#else
# define JHI_EVENT_LOG_SVCNAME "IntelDalJhi"
#endif

//void WriteToEventLog(JHI_EVENT_LOG_TYPE EventType, uint32_t MessageID);
// Disable event logs in Linux and Android
#define WriteToEventLog(x,y)

#endif //WIN32

//void WriteToEventLog(JHI_EVENT_LOG_TYPE EventType, const wchar_t* Message);


}

#endif 