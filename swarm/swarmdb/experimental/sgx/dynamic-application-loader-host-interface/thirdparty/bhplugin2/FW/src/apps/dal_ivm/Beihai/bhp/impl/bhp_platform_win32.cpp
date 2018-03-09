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
 * @file  bhp_platform_win32.cpp
 * @brief This file implements BHP platform dependent part on Win32 platform.
 * @author
 * @version
 *
 */
#include <Windows.h>
#include <stdio.h>
#include "bhp_platform.h"

#ifdef TRACE_MALLOC
struct {
    void* (*pm) (int size);
	void* (*pml) (int size, const char* file, int line); //malloc with filename and lineno
	void (*pf) (void* p);
	void (*pfl) (void* p, const char* file, int line); //free with filename and lineno
} plat_mem_procs = {0};

void* bhp_trace_malloc (int size, const char* file, int line)
{
    if (plat_mem_procs.pml) return plat_mem_procs.pml(size, file, line);
    return plat_mem_procs.pm(size);
}

void bhp_trace_free (void* p, const char* file, int line)
{
    if (plat_mem_procs.pfl) return plat_mem_procs.pfl(p, file, line);
    return  plat_mem_procs.pf(p);
}

__declspec(dllexport) void BHP_SetupAllocate(void* alloc_f, void* alloc_location_f, void* free_f, void* free_location_f)
{
    plat_mem_procs.pm =  (void* (*) (int)) alloc_f;
	plat_mem_procs.pml = (void* (*) (int, const char*, int)) alloc_location_f;
	plat_mem_procs.pf = (void (*) (void*)) free_f;
	plat_mem_procs.pfl = (void (*) (void*, const char*, int)) free_location_f;
}
#endif

bhp_mutex_t bh_create_mutex(void) 
{
    return ::CreateMutex(NULL, FALSE, NULL);
}

void bh_close_mutex(bhp_mutex_t m) 
{
    ::CloseHandle(m);
}

void mutex_enter(bhp_mutex_t m) 
{
    ::WaitForSingleObject(m, INFINITE);
}

void mutex_exit(bhp_mutex_t m) 
{
    ::ReleaseMutex(m);
}

bhp_event_t bh_create_event(void)
{
    return ::CreateEvent(NULL, FALSE, FALSE, NULL);
}

void bh_close_event(bhp_event_t evt)
{
    ::CloseHandle(evt);
}

void bh_signal_event(bhp_event_t evt)
{
    ::SetEvent(evt);
}

void bh_reset_event(bhp_event_t evt)
{
    ::ResetEvent(evt);
}

void bh_wait_event(bhp_event_t evt)
{
    ::WaitForSingleObject(evt, INFINITE);
}

bhp_thread_t bh_thread_create (void* (*func)(void*), void* args)
{
    return (bhp_thread_t)::CreateThread (NULL, 0, (LPTHREAD_START_ROUTINE) func, args, 0, NULL);
}

void bh_thread_close (bhp_thread_t thread)
{
    ::CloseHandle(thread);
}

void bh_thread_join (bhp_thread_t thread) {
    if (thread == NULL) return;
    //daoming: don't kill recv_thread, as the thread will exit automatically when heci disconnected.
    //TerminateThread((HANDLE) thread, 0);
    ::WaitForSingleObject(thread, INFINITE);
}

void bh_thread_cancel (bhp_thread_t thread)
{
	// In Windows nothing needs to be done here.
}

/*
unsigned int bh_atomic_inc(volatile unsigned int* pVal)
{
    return InterlockedIncrement(pVal);
}

unsigned int bh_atomic_dec(volatile unsigned int* pVal)
{
    return InterlockedDecrement(pVal);
}
*/
/*
static unsigned int bh_atomic_comp_and_swap(volatile unsigned int* dest, unsigned int exchange, unsigned int comparand)
{
    return InterlockedCompareExchange(dest, exchange, comparand);
}
*/

void bh_debug_print(int level, const char *format, ... )
{
#define DEBUG_BUF_LEN 1024
    if (level <= BHP_LOG_LEVEL) {
        char     buffer[DEBUG_BUF_LEN];
        va_list  args;

        va_start(args, format);
        vsprintf_s(buffer, DEBUG_BUF_LEN, format, args);
        va_end(args);

        ::OutputDebugStringA(buffer);
        //fprintf(stderr, "%s", buffer); //for linux
        //fflush(stderr);
        //only fatal msg shown in the console
        if (level == LOG_LEVEL_FATAL) printf(buffer);
    }
}


