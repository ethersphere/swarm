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

/*
 *
 * @file  bhp_platform.h
 * @brief This file declares the BHP platform dependent type and interface.
 * @author
 * @version
 *
 */
#ifndef __BHP_PLATFORM_H__
#define __BHP_PLATFORM_H__

#ifdef __cplusplus
extern "C" {
#endif

#ifdef _WIN32
typedef HANDLE bhp_mutex_t;
typedef HANDLE bhp_event_t;
typedef HANDLE bhp_thread_t;
#else
#include <stddef.h>
#include <stdlib.h>
#include <stdarg.h>
#include <string.h>
typedef void* bhp_mutex_t;
typedef void* bhp_event_t;
typedef void* bhp_thread_t;
#endif
#include <stdint.h>
// MUTEX functions
bhp_mutex_t bh_create_mutex(void);
void bh_close_mutex(bhp_mutex_t m);
void mutex_enter(bhp_mutex_t m);
void mutex_exit(bhp_mutex_t m);

//Event functions
bhp_event_t bh_create_event(void);
void bh_close_event(bhp_event_t evt);
void bh_signal_event(bhp_event_t evt);
void bh_wait_event(bhp_event_t evt);
void bh_reset_event(bhp_event_t evt);

//Thread functions
bhp_thread_t bh_thread_create (void* (*func)(void*), void* args);
void bh_thread_cancel (bhp_thread_t thread);
void bh_thread_close (bhp_thread_t thread);
void bh_thread_join(bhp_thread_t thread);

//Atomic operation functions
//unsigned int bh_atomic_inc(volatile unsigned int* pVal);
//unsigned int bh_atomic_dec(volatile unsigned int* pVal);

//debug functions
enum {
    LOG_LEVEL_FATAL = 0,
    //LOG_LEVEL_ERROR = 1,
    LOG_LEVEL_WARN = 2,
    //LOG_LEVEL_INFO = 3,
    LOG_LEVEL_DEBUG = 4,
    //LOG_LEVEL_VERBOSE = 5
};
void bh_debug_print(int level, const char *fmt, ...);
#define BHP_LOG_FATAL(...) bh_debug_print(LOG_LEVEL_FATAL, __VA_ARGS__)
#ifdef DEBUG
#define BHP_LOG_LEVEL LOG_LEVEL_DEBUG
#define BHP_LOG_WARN(...) bh_debug_print(LOG_LEVEL_WARN, __VA_ARGS__)
#define BHP_LOG_DEBUG(...) bh_debug_print(LOG_LEVEL_DEBUG, __VA_ARGS__)
#else
#define BHP_LOG_LEVEL LOG_LEVEL_FATAL
#define BHP_LOG_WARN(...) 
#define BHP_LOG_DEBUG(...)
#endif

//memory trace functions
#ifdef TRACE_MALLOC
void* bhp_trace_malloc (int size, const char* file, int line);
void bhp_trace_free (void* p, const char* file, int line);
#define BHMALLOC(x) bhp_trace_malloc(x, __FILE__, __LINE__)
#define BHFREE(x) bhp_trace_free(x, __FILE__, __LINE__)
#else
#define BHMALLOC malloc
#define BHFREE free
#endif

#ifdef __cplusplus
}
#endif

#endif
