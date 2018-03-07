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
**    @file dbg-android.c
**
**    @brief  Debug functions writing to logcat
**
**    @author Niveditha Sundaram
**    @author Venky Gokulranga
**    @author Alexander Usyskin
**
********************************************************************************
*/
#include <stdarg.h>
#include <android/log.h>
#include "dbg.h"

#define LOG_TAG "jhi"

JHI_LOG_LEVEL g_jhiLogLevel = JHI_LOG_LEVEL_RELEASE;

#ifdef PRINT_TID
void _print(const char *format, va_list& ap)
{
	const int buflen = 8192;
	char Buffer[buflen];

	vsnprintf(Buffer, buflen, format, ap);
	__android_log_print(ANDROID_LOG_DEBUG, LOG_TAG, "[%ld] %s", GetCurrentThreadId(), Buffer);
}
#else
inline void _print(const char *format, va_list& ap)
{
	__android_log_vprint(ANDROID_LOG_DEBUG, LOG_TAG, format, ap);
}
#endif//PRINT_TID

UINT32 JHI_Log(const char *format, ...)
{
	va_list ap;

	va_start(ap, format);
	_print(format, ap);
	va_end(ap);

	/* to comply with the API */
	return 1;
}

UINT32 JHI_Trace(const char *format, ...)
{
#ifdef DEBUG
	va_list ap;

	va_start(ap, format);
	_print(format, ap);
	va_end(ap);
#endif
	/* to comply with the API */
	return 1;
}
