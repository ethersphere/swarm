/* Copyright 2014 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
#ifndef __HELPERS_H
#define __HELPERS_H

#ifdef _WIN32
#include <windows.h>
#include <stdio.h>
#include <stdarg.h>
#include "libteewin.h"

	#if _DEBUG
		#define PRINTS_ENABLE
	#endif

	#define MALLOC(X)   HeapAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, X)
	#define FREE(X)     {if(X) { HeapFree(GetProcessHeap(), 0, X); X = NULL ; } }

	#define DEBUG_MSG_LEN 1024

	static void DebugPrint(const char* args, ...)
	{
		char msg[DEBUG_MSG_LEN];
		va_list varl;
		va_start(varl, args);
		vsprintf_s(msg, DEBUG_MSG_LEN, args, varl);
		va_end(varl);

		OutputDebugStringA(msg);
	}

	#define ErrorPrint(fmt, ...) DebugPrint(fmt, __VA_ARGS__)
	#define IS_HANDLE_INVALID(h) (NULL == h || 0 == h->handle || INVALID_HANDLE_VALUE == h->handle)
	#define INIT_STATUS TEE_INTERNAL_ERROR
#else
	#ifdef ANDROID
		// For debugging
		//#define LOG_NDEBUG 0
		#define LOG_TAG "libtee"
		#include <cutils/log.h>
		#define DebugPrint(fmt, ...) ALOGV_IF(true, fmt, ##__VA_ARGS__)
		#define ErrorPrint(fmt, ...) ALOGE_IF(true, fmt, ##__VA_ARGS__)
		#if LOG_NDEBUG
			#define PRINTS_ENABLE
		#endif
	#else /* LINUX */
		#include <stdlib.h>
		#ifdef DEBUG
			#define PRINTS_ENABLE
		#endif
		#define DebugPrint(fmt, ...) fprintf(stderr, fmt, ##__VA_ARGS__)
		#define ErrorPrint(fmt, ...) DebugPrint(fmt, ##__VA_ARGS__)
	#endif /* ANDROID */

	#define MALLOC(X)   malloc(X)
	#define FREE(X)     { if(X) { free(X); X = NULL ; } }

	#define IS_HANDLE_INVALID(h) (NULL == h || 0 == h->handle || -1 == h->handle)
	#define INIT_STATUS -EPERM
#endif /* _WIN32 */


#ifdef PRINTS_ENABLE
#define DBGPRINT(_x_, ...) \
	DebugPrint("TEELIB: (%s:%s():%d) ",__FILE__,__FUNCTION__,__LINE__); \
	DebugPrint(_x_, ##__VA_ARGS__)
#else
	#define DBGPRINT(_x_, ...)
#endif /* PRINTS_ENABLE */

#ifdef PRINTS_ENABLE
#define ERRPRINT(_x_, ...) \
	ErrorPrint("TEELIB: (%s:%s():%d) ",__FILE__,__FUNCTION__,__LINE__); \
	ErrorPrint(_x_, ##__VA_ARGS__)
#else
	#define ERRPRINT(_x_, ...)
#endif

#define FUNC_ENTRY()         ERRPRINT("Entry\n")
#define FUNC_EXIT(status)    ERRPRINT("Exit with status: %d\n", status)

#endif /* __HELPERS_H */
