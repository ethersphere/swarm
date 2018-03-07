/*
 * INTEL CONFIDENTIAL
 * Copyright 2010-2013 Intel Corporation All Rights Reserved.
 * The source code contained or described herein and all documents related to
 * the source code ("Material") are owned by Intel Corporation or its suppliers
 * or licensors. Title to the Material remains with Intel Corporation or its
 * suppliers and licensors. The Material contains trade secrets and proprietary
 * and confidential information of Intel or its suppliers and licensors. The
 * Material is protected by worldwide copyright and trade secret laws and treaty
 * provisions. No part of the Material may be used, copied, reproduced, modified,
 * published, uploaded, posted, transmitted, distributed, or disclosed in any way
 * without Intel's prior express written permission.
 *
 * No license under any patent, copyright, trade secret or other intellectual
 * property right is granted to or conferred upon you by disclosure or delivery
 * of the Materials, either expressly, by implication, inducement, estoppel or
 * otherwise. Any license under such intellectual property rights must be express
 * and approved by Intel in writing.
 *
 *
 * @file  bhp_platform_linux.cpp
 * @brief This file implements BHP platform dependent part on Linux 32bit platform.
 * @author
 * @version
 *
 */
#include <pthread.h>
#include <sys/types.h>
#include <stdio.h>
#include <unistd.h>
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

void BHP_SetupAllocate(void* alloc_f, void* alloc_location_f, void* free_f, void* free_location_f)
{
    plat_mem_procs.pm =  (void* (*) (int)) alloc_f;
    plat_mem_procs.pml = (void* (*) (int, const char*, int)) alloc_location_f;
    plat_mem_procs.pf = (void (*) (void*)) free_f;
    plat_mem_procs.pfl = (void (*) (void*, const char*, int)) free_location_f;
}
#endif

bhp_mutex_t bh_create_mutex(void)
{
    pthread_mutex_t* m;
    pthread_mutexattr_t attr;
    int ret;

    m = (pthread_mutex_t*) BHMALLOC(sizeof(pthread_mutex_t));
    if (m == NULL) return NULL;

    pthread_mutexattr_init(&attr);
    pthread_mutexattr_settype(&attr, PTHREAD_MUTEX_RECURSIVE);
    ret = pthread_mutex_init(m, &attr);
    pthread_mutexattr_destroy(&attr);
    if (ret != 0) {
        BHFREE(m);
        return NULL;
    }
    return (bhp_mutex_t)m;
}

void bh_close_mutex(bhp_mutex_t mt)
{
    pthread_mutex_t* m = (pthread_mutex_t*)mt;
    if (m == NULL) return;

    pthread_mutex_destroy(m);
    BHFREE(m);
}

void mutex_enter(bhp_mutex_t mt)
{
    pthread_mutex_t* m = (pthread_mutex_t*)mt;
    if (m == NULL) return;

    pthread_mutex_lock(m);
}

void mutex_exit(bhp_mutex_t mt)
{
    pthread_mutex_t* m = (pthread_mutex_t*)mt;
    if (m == NULL) return;

    pthread_mutex_unlock(m);
}

struct pevent_t {
    pthread_mutex_t mutex;
    pthread_cond_t cond;
    int triggered;
};

bhp_event_t bh_create_event(void)
{
   pevent_t* evt;

   evt = (pevent_t*) BHMALLOC(sizeof(pevent_t));
   if(!evt) return NULL;
   
   pthread_mutex_init(&evt->mutex, 0);
   pthread_cond_init(&evt->cond, 0);
   evt->triggered = 0;

   return (bhp_event_t)evt;
}

void bh_close_event(bhp_event_t evt)
{
    pevent_t* ev = (pevent_t*)evt;
    if (ev == NULL) return;

    pthread_mutex_destroy(&ev->mutex);
    pthread_cond_destroy(&ev->cond);
    BHFREE(ev);
}

void bh_signal_event(bhp_event_t evt)
{
    pevent_t* ev = (pevent_t*)evt;
    if (ev == NULL) return;

    pthread_mutex_lock(&ev->mutex);
    ev->triggered = 1;
    pthread_cond_signal(&ev->cond);
    pthread_mutex_unlock(&ev->mutex);
}

void bh_reset_event(bhp_event_t evt)
{
    pevent_t* ev = (pevent_t*)evt;
    if (ev == NULL) return;

    pthread_mutex_lock(&ev->mutex);
    ev->triggered = 0;
    pthread_mutex_unlock(&ev->mutex);
}

void bh_wait_event(bhp_event_t evt)
{
    pevent_t* ev = (pevent_t*)evt;
    if (ev == NULL) return;

    pthread_mutex_lock(&ev->mutex);
    while (!ev->triggered) {
        pthread_cond_wait(&ev->cond, &ev->mutex);
    }
    pthread_mutex_unlock(&ev->mutex);
}

bhp_thread_t bh_thread_create (void* (*func)(void*), void* args) {
    pthread_t* t;
    int ret;
    
    t = (pthread_t*) BHMALLOC(sizeof(pthread_t));
    if (t == NULL) return NULL;

    ret = pthread_create(t, NULL, func, args);
    if (ret != 0) {
        BHFREE(t);
        return NULL;
    }
    return (bhp_thread_t)t;
}

void bh_thread_close (bhp_thread_t th)
{
    pthread_t* t = (pthread_t*)th;
    if (t == NULL) return;

    BHFREE(t);
}

void bh_thread_join (bhp_thread_t th) {
    pthread_t* t = (pthread_t*)th;
    if (t == NULL) return;

    pthread_join(*t, NULL);
}

void bh_thread_cancel (bhp_thread_t th)
{
// TODO: Do we need an alternative for Android?
// Bionic doesn't implement it.
#ifndef __ANDROID__
    pthread_t* t = (pthread_t*)th;
    if (t == NULL) return;

    pthread_cancel(*t);
#endif
}

void bh_debug_print(int level, const char *format, ...)
{
#define DEBUG_BUF_LEN 1024
    if (level <= BHP_LOG_LEVEL) {
        char     buffer[DEBUG_BUF_LEN];
        va_list  args;

        va_start(args, format);
        vsnprintf(buffer, DEBUG_BUF_LEN, format, args);
        va_end(args);

        fprintf(stderr, "%s", buffer);
        fflush(stderr);
    }
}

