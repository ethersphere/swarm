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

/**
********************************************************************************
**
**    @file dbg.c
**
**    @brief  Debug functions
**
**    @author Niveditha Sundaram
**    @author Venky Gokulranga
**
********************************************************************************
*/
#include <stdio.h>
#include <stdlib.h>
#include <stdarg.h>
#include <syslog.h>
#include "dbg.h"

#define LOG_APP (LOG_CONS | LOG_PID | LOG_NDELAY)

JHI_LOG_LEVEL g_jhiLogLevel = JHI_LOG_LEVEL_RELEASE;

#ifdef PRINT_TID
void _print(const char *format, va_list& ap)
{
	const int buflen = 8192;
	char Buffer[buflen];

	openlog ("jhi", LOG_APP, LOG_LOCAL1);
	vsnprintf(Buffer, buflen, format, ap);
	syslog(LOG_DEBUG, "[%ld] %s", GetCurrentThreadId(), Buffer);
	closelog();
}
#else

inline void _print(const char *format, va_list& ap)
{
	openlog ("jhi", LOG_APP, LOG_LOCAL1);
	vsyslog(LOG_DEBUG, format, ap);
	closelog();
}
#endif//PRINT_TID

UINT32 JHI_Log(const char *format, ...)
{
	if(g_jhiLogLevel >= JHI_LOG_LEVEL_RELEASE)
	{
		va_list ap;
		va_start(ap, format);
		_print(format, ap);
		va_end(ap);
	}
	/* to comply with the API */
	return 1;
}

UINT32 JHI_Trace(const char *format, ...)
{
	if(g_jhiLogLevel >= JHI_LOG_LEVEL_DEBUG)
	{
		va_list ap;
		va_start(ap, format);
		_print(format, ap);
		va_end(ap);
	}
	/* to comply with the API */
	return 1;
}

UINT32
JHI_T_Trace(const wchar_t *format, ...)
{
	char msg[2024];
	wchar_t wmsg[1024];
	va_list ap;

	va_start(ap, format);	
	vswprintf(wmsg, 1024, format, ap);
	va_end(ap);

	wcstombs(msg, wmsg, 2024);
	openlog("jhi", LOG_APP, LOG_LOCAL1);
	syslog(LOG_INFO, "%s", msg);
	closelog();

	/* to comply with the API */
	return 1;
}
