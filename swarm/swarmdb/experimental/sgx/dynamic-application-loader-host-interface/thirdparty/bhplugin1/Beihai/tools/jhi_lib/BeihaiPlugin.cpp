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

#include "BeihaiPlugin.h"
#include <set>
#include <map>

#define CMDBUF_SIZE 100

const unsigned char BH_MSG_BEGINNING[] = { 0xff, 0xa3, 0xaa, 0x55 };
const unsigned char BH_MSG_FOLLOWING[] = { 0xff, 0xa4, 0xaa, 0x55 };
const unsigned char BH_MSG_RESPONSE[] = { 0xff, 0xa5, 0xaa, 0x55 };

static BH_MUTEX bhm_state = NULL;
static BH_MUTEX bhm_send = NULL;
static BH_MUTEX bhm_rrmap = NULL;

static BH_EVENT start_event = NULL;

static BH_THREAD recv_thread = NULL;

typedef struct {
	BH_EVENT event;
	int count;
	BH_ERRNO code;
	UINT32 length;
	ADDR addr;
	void *buffer;
	BH_MUTEX session_lock;
	int is_session;
	int killed;
} bh_response_record;

typedef void* (*malloc_t)(int size);
typedef void* (*malloc_location_t)(int size, const char* file, int line);
typedef void (*free_t)(void* p);
typedef void (*free_location_t)(void* p, const char* file, int line);

#ifdef TRACE_MALLOC
malloc_t bhmalloc = (malloc_t) malloc;
malloc_location_t bhmalloc_location = NULL;
free_t bhfree = (free_t) free;
free_location_t bhfree_location = NULL;

#define BHMALLOC(x) _bhmalloc(x, __FILE__, __LINE__)
#define BHFREE(x) _bhfree(x, __FILE__, __LINE__)

void *_bhmalloc (int size, const char* file, int line)
{
	if (bhmalloc_location)
		return bhmalloc_location(size, file, line);
	return bhmalloc(size);
}

void _bhfree (void*p, const char* file, int line)
{
	if (bhfree_location)
		return bhfree_location(p, file, line);
	return
		bhfree(p);
}

DLL_EXPORT void BH_SetupAllocate(void* alloc, void* alloc_location, void* free, void* free_location)
{
	bhmalloc = (malloc_t) alloc;
	bhmalloc_location = (malloc_location_t) alloc_location;
	bhfree = (free_t) free;
	bhfree_location = (free_location_t) free_location;
}
#else
#define BHMALLOC malloc
#define BHFREE free
#endif

#define TRACE0                         trace
#define TRACE1(fmt,p1)                 trace(fmt,p1)
#define TRACE2(fmt,p1,p2)              trace(fmt,p1,p2)
#define TRACE3(fmt,p1,p2,p3)           trace(fmt,p1,p2,p3)
#define TRACE4(fmt,p1,p2,p3,p4)        trace(fmt,p1,p2,p3,p4)
#define TRACE5(fmt,p1,p2,p3,p4,p5)     trace(fmt,p1,p2,p3,p4,p5)
#define TRACE6(fmt,p1,p2,p3,p4,p5,p6)  trace(fmt,p1,p2,p3,p4,p5,p6)

#ifdef DEBUG
#include <stdarg.h>
#include <stdio.h>
#endif

static int trace(
	 const char*  Format,
	 ... )
{
	UINT32       dwChars=0;
#ifdef DEBUG
	char     Buffer [1024] ;
	int buflen = sizeof(Buffer);
	va_list  args ;

	va_start ( args, Format ) ;
	
#ifdef __linux__
		int tracerNameLen = 0;
		tracerNameLen = strlen("BH_Plugin: ");
		strcpy(Buffer, "BH_Plugin: ") ;
		dwChars = vsnprintf (Buffer + tracerNameLen, buflen - tracerNameLen, Format, args);
#else
	dwChars = vsnprintf_s ( Buffer, buflen, Format, args ) ;
#endif
	
	va_end (args) ;

#ifdef _WIN32
	OutputDebugStringA ( Buffer) ;
#else
	fprintf (stderr, "BH_Plugin: %s", Buffer ) ;
	fflush(stderr);
#endif
#endif

	return dwChars ;
}

static inline void swapchar(char* i, char* j) {
	char t;
	t = *i; *i = *j; *j = t;
}

void byte_order_swapi (void* i) {
	char* c = (char*)i;
	swapchar(&c[3], &c[0]);
	swapchar(&c[2], &c[1]);
}

void byte_order_swaps (void* i) {
	char* c = (char*)i;
	swapchar(&c[1], &c[0]);
}

#ifdef _WIN32

void bh_init_state() {
	if (!bhm_state)
		bhm_state = ::CreateMutex(NULL, NULL, NULL);
}

void bh_init_mutex() {
	if (!bhm_send)
		bhm_send = ::CreateMutex(NULL, NULL, NULL);
	if (!bhm_rrmap)
		bhm_rrmap = ::CreateMutex(NULL, NULL, NULL);
	if (!start_event)
		start_event = ::CreateEvent(NULL, NULL, FALSE, NULL);
}

#define bh_create_mutex() ::CreateMutex(NULL, FALSE, NULL)
#define bh_close_mutex(mutex) do {		\
		if (mutex)			\
			::CloseHandle((mutex));	\
	} while(0)

#define mutex_enter(mutex) ::WaitForSingleObject((mutex), INFINITE)
#define mutex_exit(mutex) ::ReleaseMutex((mutex))

#define bh_create_event() ::CreateEvent(NULL, NULL, FALSE, NULL)
#define bh_close_event(event) do {		\
		if (event)			\
			::CloseHandle((event));	\
	} while(0)

#define bh_event_reset(event) ::ResetEvent((event))
#define bh_event_wait(event) ::WaitForSingleObject((HANDLE)(event), INFINITE)
#define bh_event_signal(event) ::SetEvent(event)
#define bh_signal_and_wait(event1, event2) ::SignalObjectAndWait((HANDLE)event1, (HANDLE)event2, INFINITE, FALSE)

BH_THREAD bh_thread_create (void* (*func)(void*)) {
	return CreateThread (NULL, 0, (LPTHREAD_START_ROUTINE) func, NULL, 0, NULL);
}

void bh_thread_cancel(BH_THREAD thread) {
	// Not needed on windows
}

void bh_thread_join (BH_THREAD thread) {
    if (thread == NULL) 
		return;

    ::WaitForSingleObject(thread, INFINITE);
}

void bh_thread_close (BH_THREAD thread)
{
    ::CloseHandle(thread);
}

#elif __linux__

#include <ctype.h>

static pthread_mutex_t bhm_state_s = PTHREAD_MUTEX_INITIALIZER;
static pthread_mutex_t bhm_send_s = PTHREAD_MUTEX_INITIALIZER;
static pthread_mutex_t bhm_rrmap_s = PTHREAD_MUTEX_INITIALIZER;

BH_MUTEX bh_create_mutex()
{
	BH_MUTEX lock;
	int ret;

	lock = (BH_MUTEX) BHMALLOC(sizeof(pthread_mutex_t));
	if (!lock)
		return lock;

	ret = pthread_mutex_init(lock, NULL);
	if (ret == 0)
		return lock;
	else {
		BHFREE(lock);
		return NULL;
	}
}

void bh_close_mutex(BH_MUTEX lock)
{
	pthread_mutex_destroy(lock);
	BHFREE(lock);
}

void mutex_enter(BH_MUTEX lock)
{
	int ret = pthread_mutex_lock(lock);
	if (ret != 0)
		abort();
}

void mutex_exit(BH_MUTEX lock)
{
	int ret = pthread_mutex_unlock(lock);
	if (ret != 0)
		abort();
}

struct pevent_t {
    pthread_mutex_t mutex;
    pthread_cond_t cond;
    bool triggered;
};

BH_EVENT bh_create_event() {
	BH_EVENT ev = (BH_EVENT) BHMALLOC(sizeof(pevent_t));
	if(!ev)
		return NULL;

	pthread_mutex_init(&ev->mutex, 0);
	pthread_cond_init(&ev->cond, 0);
	ev->triggered = false;
	return ev;
}

void bh_close_event(BH_EVENT ev) {
	pthread_mutex_destroy(&ev->mutex);
	pthread_cond_destroy(&ev->cond);
	BHFREE(ev);
}

void bh_event_signal(BH_EVENT ev) {
	pthread_mutex_lock(&ev->mutex);
	ev->triggered = true;
	pthread_cond_signal(&ev->cond);
	pthread_mutex_unlock(&ev->mutex);
}

void bh_event_reset(BH_EVENT ev) {
	pthread_mutex_lock(&ev->mutex);
	ev->triggered = false;
	pthread_mutex_unlock(&ev->mutex);
}

void bh_event_wait(BH_EVENT ev) {
	pthread_mutex_lock(&ev->mutex);
	while (!ev->triggered) {
		pthread_cond_wait(&ev->cond, &ev->mutex);
	}
	pthread_mutex_unlock(&ev->mutex);
}

void bh_signal_and_wait(BH_MUTEX mutex, BH_EVENT ev) {
	pthread_mutex_unlock(mutex);
	bh_event_wait(ev);
}

void bh_init_state() {
	if (!bhm_state) {
		pthread_mutex_init(&bhm_state_s, NULL);
		bhm_state = &bhm_state_s;
	} 
}

static pevent_t start_event_data;

void bh_init_mutex() {
	if (!bhm_send) {
		pthread_mutex_init(&bhm_send_s, NULL);
		bhm_send = &bhm_send_s;
	} 
	if (!bhm_rrmap) {
		pthread_mutex_init(&bhm_rrmap_s, NULL);
		bhm_rrmap = &bhm_rrmap_s;
	} 
	if (!start_event) {
		start_event = &start_event_data;
		pthread_mutex_init(&start_event->mutex, 0);
		pthread_cond_init(&start_event->cond, 0);
		start_event->triggered = false;
	}
}

pthread_t recv_thread_struct;

BH_THREAD bh_thread_create (void* (*func)(void*)) {
	BH_THREAD t = &recv_thread_struct;
	int ret = pthread_create(t, NULL, func, NULL);
	if (ret == 0)
		return t;
	else {
		return NULL;
	}
}

void bh_thread_cancel(BH_THREAD thread) {
#ifndef __ANDROID__
	pthread_cancel(*thread);
#endif
// pthread_cancel is not implemented for android
}

void bh_thread_join (BH_THREAD thread) {
	pthread_join(*thread, NULL);
}

void bh_thread_close (BH_THREAD thread)
{
    (void)thread;
}

UINT32 min(UINT32 x, UINT32 y)
{
	return x>y? y : x;
}

#endif

#define MAX_SESSION_LIMIT 20

using namespace std;

ADDR seqno = 1000;
map<ADDR, bh_response_record*> rrmap;

static void session_exit(bh_response_record* session, ADDR seq);

static bh_response_record* addr2record(ADDR seq)
{
	bh_response_record* rr = NULL;
	mutex_enter(bhm_rrmap);
	if (rrmap.find(seq) != rrmap.end())
		rr = (bh_response_record*) rrmap[seq];
	mutex_exit(bhm_rrmap);
	return rr;
}

static void destroy_session(bh_response_record* session)
{
	TRACE1("destroy_session %x\n", session);
	bh_close_mutex(session->session_lock);
	BHFREE (session->buffer);
	BHFREE (session);
}

static bh_response_record* session_enter_nolock(ADDR seq)
{
	bh_response_record* session = NULL;

	mutex_enter(bhm_rrmap);
	if (rrmap.find (seq) != rrmap.end() && rrmap[seq]->is_session && !rrmap[seq]->killed) {
		session = rrmap[seq];
		if (session->count < MAX_SESSION_LIMIT) {
			session->count++;
		} else
			session = NULL;
	}
	mutex_exit(bhm_rrmap);
	return session;
}

static bh_response_record* session_enter(ADDR seq)
{
	bh_response_record* session = session_enter_nolock(seq);

	if (session) {
		mutex_enter(session->session_lock);
		if (session->killed) {
			session_exit(session, seq);
			return NULL;
		}
	}

	return session;
}

static void session_exit(bh_response_record* session, ADDR seq)
{
	mutex_enter(bhm_rrmap);
	session->count --;
	if (session->count == 0 && session->killed) {
		mutex_exit(session->session_lock);
		rrmap.erase(seq);
		destroy_session(session);
	} else {
		mutex_exit(session->session_lock);
	}
	mutex_exit(bhm_rrmap);
}

static void session_close_nolock(bh_response_record* session, ADDR seq)
{
	mutex_enter(bhm_rrmap);
	session->count --;
	if (session->count == 0) {
		rrmap.erase(seq);
		destroy_session(session);
	} else {
		session->killed = 1;
	}
	mutex_exit(bhm_rrmap);
}

static void session_close(bh_response_record* session, ADDR seq)
{
	mutex_enter(bhm_rrmap);
	session->count --;
	if (session->count == 0) {
		mutex_exit(session->session_lock);
		rrmap.erase(seq);
		destroy_session(session);
	} else {
		session->killed = 1;
		mutex_exit(session->session_lock);
	}
	mutex_exit(bhm_rrmap);
}

static ADDR rrmap_add(bh_response_record* rr)
{
	ADDR seq = seqno ++;

	mutex_enter(bhm_rrmap);
	rrmap[seq] = rr;
	TRACE2("rrmap_add %llx %x\n", seq, rr);
	mutex_exit(bhm_rrmap);

	return seq;
}

static bh_response_record* rrmap_remove(ADDR seq)
{
	bh_response_record* rr = NULL;

	mutex_enter(bhm_rrmap);
	if (rrmap.find(seq) != rrmap.end()) {
		rr = rrmap[seq];
		if (!rr->is_session) {
			rrmap.erase(seq);
			TRACE2("rrmap_erase %llx %x\n", seq, rr);
		}
	}
	mutex_exit(bhm_rrmap);
	return rr;
}

static int char2hex(char c)
{
	if (isdigit(c)) 
		return (c - '0');
	else 
		return (toupper(c) - 'A' + 0xA);
}

static int string_check1_uuid(const char* str)
{
	int i;

	if (strlen (str) != APPID_LENGTH*2)
		return false;

	for(i=0; i<APPID_LENGTH*2; i++, str++)
		if(! ((*str >= '0' && *str <= '9') || 
		      (*str >= 'a' && *str <= 'f') || 
		      (*str >= 'A' && *str <= 'F')))
			return false;

	return true;
}

static int string_check2_uuid(const char* str)
{
	int i;

	if (strlen (str) != APPID_LENGTH*2 + 4)
		return false;

	for(i=0; i<APPID_LENGTH*2; i++, str++) {
		if (*str == '-' && (i==8 || i==12 || i==16 || i== 20))
			str++;
		if (! ((*str >= '0' && *str <= '9') || 
		       (*str >= 'a' && *str <= 'f') || 
		       (*str >= 'A' && *str <= 'F')))
			return false;
	}

	return true;
}

int string_to_uuid(const char* str, char* uuid)
{
	int i;

	if (!string_check1_uuid(str) && !string_check2_uuid(str))
		return false;

	for(i=0; i<APPID_LENGTH; i++, uuid++) {
		if(*str == '-')
			str++;

		*uuid = char2hex(*str++);
		*uuid <<= 4;
		*uuid += char2hex(*str++);
	}

	return true;
}

static int hex2asc (char c)
{
	if (c < 10)
		return '0' + c;
	else
		return 'a' + c - 10;
}

void uuid_to_string(char* uuid, char* str)
{
	int i;
	str[APPID_LENGTH * 2] = 0;
	for (i=0; i<APPID_LENGTH; i++, uuid++) {
		*str++ = hex2asc((*uuid & 0xf0) >> 4); 
		*str++ = hex2asc(*uuid & 0xf);
	}
}

static UINT32 tdesc = 0;
static UINT32 handle = 0;
static PFN_BH_TRANSPORT_SEND heci_send = NULL;
static PFN_BH_TRANSPORT_RECEIVE heci_recv = NULL;
static PFN_BH_TRANSPORT_CLOSE heci_close = NULL;

#define MAX_TXRX_LENGTH 4096
static unsigned char skip_buffer[MAX_TXRX_LENGTH];

void bh_unblock_recv_thread()
{
	// Close the transport handle - in such case the recv() in recv_thread _main() will
	// return with an error and the thread will be terminated.

    heci_close(handle); 

	// On Linux, because of a HECI limitation (?),  the recv thread will not be notified
	// about the closure and needs to be cancelled.
	bh_thread_cancel(recv_thread);

	// Wait for the recv_thread_main thread to exit.
	bh_thread_join(recv_thread);

	// Release thread handle.
    bh_thread_close(recv_thread);
}

int bh_transport_init(void* context)
{
	BH_PLUGIN_TRANSPORT* p = (BH_PLUGIN_TRANSPORT*)context;
	heci_send = p->pfnSend;
	heci_recv = p->pfnRecv;
	heci_close = p->pfnClose;
	tdesc = p->handle;
	handle = p->handle;
	return BH_SUCCESS;
}

int bh_transport_deinit()
{
	tdesc = 0;
	return BH_SUCCESS;
}

int bh_transport_connect ()
{
	return BH_SUCCESS;
}

BH_ERRNO bh_transport_recv (const void* buffer, UINT32 size)
{
	UINT32 got;
	UINT32 count = 0;
	int status;

	if (!tdesc)
		return BPE_COMMS_ERROR;

	while (size - count > 0) {
		if (buffer) {
			got = min (size - count, MAX_TXRX_LENGTH);
			status = heci_recv (tdesc, (unsigned char*) buffer + count,  &got);
		} else {
			got = min (MAX_TXRX_LENGTH, size - count);
			status = heci_recv (tdesc, skip_buffer, &got);
		}

		if (status != BH_SUCCESS)
			return BPE_COMMS_ERROR;

		count += got;
	}

	return BH_SUCCESS;
}

BH_ERRNO bh_transport_send (const void* buffer, UINT32 size)
{
	int status;

	if (!tdesc)
		return BPE_COMMS_ERROR;

	status = heci_send (tdesc, (UINT8*) buffer,  size);
	if (status != BH_SUCCESS) {
		return BPE_COMMS_ERROR;
	}

	return BH_SUCCESS;
}

BH_ERRNO bh_recv_message ()
{
	bh_response_header headbuf;
	bh_response_header *head = &headbuf;
	char* data = NULL;
	BH_ERRNO ret;
	bh_response_record* rr;

	ret = bh_transport_recv((char*) head, sizeof (bh_response_header));
	if (ret != BH_SUCCESS)
		return ret;
 
	/* check magic */
	if (memcmp(BH_MSG_RESPONSE, head->h.magic, sizeof(BH_MSG_RESPONSE)) != 0)
		return BPE_COMMS_ERROR;

	// verify rr
	rr = rrmap_remove(head->seq);
	if (!rr) {
		TRACE1 ("Beihai RECV invalid rr %llx\n", head->seq);
	}

	TRACE3 ("enter bh_recv_message %x %llx %d\n", rr, head->seq, head->code);

	if (head->h.length != sizeof(bh_response_header)) {
		data = (char*) BHMALLOC(head->h.length - sizeof(bh_response_header));
		ret = bh_transport_recv(data, head->h.length - sizeof(bh_response_header));
 		if (ret == BH_SUCCESS && data == NULL)
			ret = BPE_OUT_OF_MEMORY;			
	}
	
	TRACE3 ("exit bh_recv_message %x %llx %d\n", rr, head->seq, ret);

	if (rr) {
		rr->buffer = data;
		rr->length = head->h.length - sizeof(bh_response_header);

		if (ret == BH_SUCCESS)
			rr->code = head->code;
		else
			rr->code = ret;

		if (head->addr)
			rr->addr = head->addr;

		if (rr->event)
			bh_event_signal (rr->event);
	} else
		BHFREE(data);

	return ret;
}

static BH_ERRNO _send_message (void* cmd, UINT32 clen, const void* data, UINT32 dlen, bh_response_record *rr, ADDR seq)
{
	BH_ERRNO ret;
	bh_message_header *h;

	if (clen < sizeof(bh_message_header) || !cmd)
		return BPE_INVALID_PARAMS;

	rr->buffer = NULL;
	rr->length = 0;
	rr->event = bh_create_event();
	if (!rr->event) {
		ret = BPE_OUT_OF_RESOURCE;
		goto failure;
	}

	memcpy (cmd, BH_MSG_BEGINNING, sizeof(BH_MSG_BEGINNING));

	h = (bh_message_header*)cmd;
	h->h.length = clen + dlen;
	h->seq = seq;

	ret = bh_transport_send (cmd, clen);
	if (ret != BH_SUCCESS)
		goto failure;

	if(dlen > 0) 
		ret = bh_transport_send (data, dlen);

	if (ret != BH_SUCCESS)
		goto failure;

	return ret;

failure:
	if(rr->event)
		bh_close_event(rr->event);
	(void) rrmap_remove(seq);
	return ret;
}


BH_ERRNO bh_send_message (void* cmd, UINT32 clen, const void* data, UINT32 dlen, ADDR seq)
{
	BH_ERRNO ret;
	bh_response_record *rr = addr2record(seq);

	if (!rr)
		abort();

	mutex_enter(bhm_send);
	TRACE3 ("enter bh_send_message %x %x %d\n", rr, seq, clen+dlen);
	ret = _send_message (cmd, clen, data, dlen, rr, seq);
	TRACE3 ("done bh_send_message %x %x %d\n", rr, seq, clen+dlen);
	if (ret == BH_SUCCESS && rr && rr->event) {
		bh_signal_and_wait(bhm_send, rr->event);
		bh_close_event(rr->event);
		rr->event = NULL;
	} else 
		mutex_exit (bhm_send);

	return ret;
}


void unblock_threads (BH_ERRNO code) {
	map<ADDR, bh_response_record*>::iterator it;

	mutex_enter(bhm_rrmap);
	for(map<ADDR, bh_response_record*>::iterator it = rrmap.begin(); it != rrmap.end(); it++) {
		bh_response_record* rr = it->second;
		if(rr) {
			rr->code = code;
			if (rr->is_session)
				rr->killed = 1;

			if (rr->event)
				bh_event_signal (rr->event);

			if (rr->is_session && rr->count == 0)
				destroy_session(rr);
		}
	}

	rrmap.clear();
	TRACE0("rrmap_clear\n");
	mutex_exit(bhm_rrmap);
}

void teardown () {
	TRACE0("PLUGIN TEARDOWN \n");
	bh_event_reset(start_event);
	bh_transport_deinit();
	unblock_threads(BPE_SERVICE_UNAVAILABLE);
}

typedef enum {
	DEINITED = 0,
	INITED = 1,
	OUT_OF_SERVICE = 2,
} init_state_t;

static init_state_t init_state = DEINITED;

static void enter_state ()
{
	if (!bhm_state) {
		bh_init_state();
	}
	mutex_enter(bhm_state);
}

static void exit_state ()
{
	mutex_exit(bhm_state);
}

void* recv_thread_main (void*) 
{
	while(1) {
		BH_ERRNO ret;

		ret = bh_recv_message ();

		if ( ret != BH_SUCCESS ) {
			if (init_state == INITED) {
				init_state = OUT_OF_SERVICE;
				unblock_threads(ret);
			}

			break;
		}
	}
	return NULL;
}

static int is_init() {
	int ret;
	enter_state ();
	ret = (init_state == INITED);
	exit_state ();
	return ret;
}

static BH_ERRNO init(void* context)
{
	TRACE0("PLUGIN INIT \n");
	bh_init_mutex();
	mutex_enter(bhm_rrmap);
	rrmap.clear();
	TRACE0("rrmap_clear\n");
	mutex_exit(bhm_rrmap);

	if (bh_transport_init(context) != BH_SUCCESS) {
		teardown();
		return BPE_NO_CONNECTION_TO_FIRMWARE;
	}

	if (bh_transport_connect() != BH_SUCCESS) {
		teardown();
		return BPE_COMMS_ERROR;
	}

	if (recv_thread == NULL)
		recv_thread =  bh_thread_create(recv_thread_main);
	else
		bh_event_signal(start_event);

	return BH_SUCCESS;
}

static BH_ERRNO reset ();

BH_ERRNO BH_PluginInit (BH_PLUGIN_TRANSPORT* transport, int do_vm_reset)
{
	BH_ERRNO ret = BH_SUCCESS;

	enter_state();
	if (init_state == DEINITED) {
		ret = init(transport);

		if (ret == BH_SUCCESS) {
			// Avoid dead lock with recv thread
			exit_state();
			if (do_vm_reset)
				ret = reset();
			enter_state();

			if (ret == BH_SUCCESS)
				init_state = INITED;
			else {
				teardown();
				init_state = DEINITED;
			}
		} else {
			teardown();
			init_state = DEINITED;
		}

	} else if (init_state == INITED) {
		ret = BPE_INITIALIZED_ALREADY;
	} else if (init_state == OUT_OF_SERVICE) {
		ret = BPE_SERVICE_UNAVAILABLE;
	}
	exit_state();

	return ret;
}

BH_ERRNO BH_PluginDeinit ()
{
	enter_state();
	if (init_state != DEINITED) {
		teardown ();
		bh_unblock_recv_thread();
		init_state = DEINITED;
	}
	exit_state();
	return BH_SUCCESS;
}

static BH_ERRNO reset ()
{
	char cmdbuf[CMDBUF_SIZE];
	bh_message_header* h = (bh_message_header*) cmdbuf;
	bh_response_record rr = {0};
	BH_ERRNO ret;

	h->id = HOST_CMD_RESET;

	ret = bh_send_message((char*)h, sizeof(*h), NULL, 0, rrmap_add(&rr));
 	if ( ret == BH_SUCCESS )
		ret = rr.code;

	return ret;
}

BH_ERRNO BH_PluginReset ()
{
	if (!is_init())
		return BPE_NOT_INIT;

	return reset();
}

BH_ERRNO BH_PluginSendAndRecv ( SHANDLE pSession, int nCommandId, const void* input, UINT32 length, void** output, UINT32* output_length, int* pResponseCode)
{
	char cmdbuf[CMDBUF_SIZE];
	bh_message_header* h = (bh_message_header*) cmdbuf;
	host_snr_command* cmd = (host_snr_command*) h->cmd;
	ADDR seq = (ADDR) pSession;
	bh_response_record* rr;
	client_snr_response *resp;
	BH_ERRNO ret;

	if (!is_init())
		return BPE_NOT_INIT;

	if (!input && length != 0)
		return BPE_INVALID_PARAMS;

	if (!pSession || !output_length)
		return BPE_INVALID_PARAMS;

	if (output)
		*output = NULL;

	rr = session_enter(seq);
	if(!rr) {
		return BPE_INVALID_PARAMS;
	}

	rr->buffer = NULL;
	h->id = HOST_CMD_SENDANDRECV;

	cmd->addr = rr->addr;
	cmd->command = nCommandId;
	cmd->outlen = *output_length;

	TRACE1 ("Beihai SendAndReceive %x\n", rr);

	ret = bh_send_message((char*)h, sizeof (*h) + sizeof (*cmd), (char*)input, length, seq);
	if ( ret == BH_SUCCESS )
		ret = rr->code;

	TRACE2 ("Beihai SendAndReceive %x ret %x\n", rr, rr->code);

	if (ret == BH_SUCCESS ) {
		if(rr->buffer && rr->length >= sizeof(client_snr_response)){
			resp = (client_snr_response *) rr->buffer;
			if (pResponseCode) {
				*pResponseCode = resp->response;
				byte_order_swapi(pResponseCode);
			}

			UINT32 len = rr->length - sizeof (client_snr_response);

			if (len>0) {
				if (output && *output_length >= len) {
					*output = (char*) BHMALLOC(len);
					if (*output) {
						memcpy (*output, resp->buffer, len);
					} else 
						ret = BPE_OUT_OF_MEMORY;
				} else 
					ret = BHE_APPLET_SMALL_BUFFER;
			}

			*output_length = len;
		}else
			ret = BPE_MESSAGE_TOO_SHORT;
	} else if (ret == BHE_APPLET_SMALL_BUFFER && rr->buffer && rr->length == sizeof (client_snr_bof_response)) {
		client_snr_bof_response* resp = (client_snr_bof_response *) rr->buffer;
		if (pResponseCode) {
			*pResponseCode = resp->response;
			byte_order_swapi(pResponseCode);
		}

		*output_length = resp->request_length;
		byte_order_swapi(output_length);
	} else if (ret == BHE_UNCAUGHT_EXCEPTION || ret == BHE_WD_TIMEOUT || ret == BHE_APPLET_CRASHED /* or whatever is received when the session dies */) {
		rr->killed = 1;
	}

	BHFREE (rr->buffer);
	rr->buffer = NULL;
	session_exit(rr, seq);
	return ret;
}

BH_ERRNO BH_PluginSendAndRecvInternal ( SHANDLE pSession, int what, int nCommandId, const void* input, UINT32 length, void** output, UINT32* output_length, int* pResponseCode)
{
	char cmdbuf[CMDBUF_SIZE];
	bh_message_header* h = (bh_message_header*) cmdbuf;
	host_snr_internal_command* cmd = (host_snr_internal_command*) h->cmd;
	ADDR seq = (ADDR) pSession;
	bh_response_record* rr;
	client_snr_response *resp;
	BH_ERRNO ret;

	if (!is_init())
		return BPE_NOT_INIT;

	if (!input && length != 0)
		return BPE_INVALID_PARAMS;

	if (!pSession || !output_length)
		return BPE_INVALID_PARAMS;

	if (output)
		*output = NULL;

	rr = session_enter(seq);
	if(!rr) {
		return BPE_INVALID_PARAMS;
	}

	rr->buffer = NULL;
	h->id = HOST_CMD_SENDANDRECV_INTERNAL;

	cmd->addr = rr->addr;
	cmd->what = what;
	cmd->command = nCommandId;
	cmd->outlen = *output_length;
	
	TRACE1 ("Beihai SendAndReceive %x\n", rr);

	ret = bh_send_message((char*)h, sizeof (*h) + sizeof (*cmd), (char*)input, length, seq);
	if ( ret == BH_SUCCESS )
		ret = rr->code;

	TRACE2 ("Beihai SendAndReceive %x ret %x\n", rr, rr->code);

	if (ret == BH_SUCCESS ) {
		if(rr->buffer && rr->length >= sizeof(client_snr_response)){
			UINT32 length;
			resp = (client_snr_response *) rr->buffer;
			if (pResponseCode) {
				*pResponseCode = resp->response;
				byte_order_swapi(pResponseCode);
			}

			length = rr->length - sizeof (client_snr_response);

			if (length>0) {
				if (output && *output_length >= length) {
					*output = (char*) BHMALLOC(length);
					if (*output) {
						memcpy (*output, resp->buffer, length);
					} else 
						ret = BPE_OUT_OF_MEMORY;
				} else 
					ret = BHE_APPLET_SMALL_BUFFER;
			}

			*output_length = length;
		}else
			ret = BPE_MESSAGE_TOO_SHORT;
	} else if (ret == BHE_APPLET_SMALL_BUFFER && rr->buffer && rr->length == sizeof (client_snr_bof_response)) {
		client_snr_bof_response* resp = (client_snr_bof_response *) rr->buffer;
		if (pResponseCode) {
			*pResponseCode = resp->response;
			byte_order_swapi(pResponseCode);
		}
		*output_length = resp->request_length;
		byte_order_swapi(output_length);
	} else if (ret == BHE_UNCAUGHT_EXCEPTION || ret == BHE_WD_TIMEOUT || ret == BHE_APPLET_CRASHED /* or whatever is received when the session dies */) {
		rr->killed = 1;
	}

	BHFREE (rr->buffer);
	rr->buffer = NULL;
	session_exit(rr, seq);
	return ret;
}

BH_ERRNO BH_PluginUnload ( const char *AppId )
{
	char cmdbuf[CMDBUF_SIZE];
	bh_message_header* h = (bh_message_header*) cmdbuf;
	host_delete_command* cmd = (host_delete_command*) h->cmd;
	bh_response_record rr = {0};
	BH_ERRNO ret;

	if (!is_init())
		return BPE_NOT_INIT;

	if ( !AppId )
		return BPE_INVALID_PARAMS;

	if (!string_to_uuid(AppId, (char*)&(cmd->appid)))
		return BPE_INVALID_PARAMS;

	h->id = HOST_CMD_DELETE;

	TRACE0 ("Beihai Delete\n");

	ret = bh_send_message((char*)h, sizeof(*h) + sizeof (*cmd), NULL, 0, rrmap_add(&rr));
	if ( ret == BH_SUCCESS )
		ret = rr.code;

	TRACE1 ("Beihai Delete ret %x\n", rr.code);

	BHFREE (rr.buffer);
	return ret;
}


BH_ERRNO BH_PluginDownload ( const char *pAppId, const void* pAppBlob, UINT32 AppSize)
{
	char cmdbuf[CMDBUF_SIZE];
	bh_message_header* h = (bh_message_header*) cmdbuf;
	host_download_command* cmd = (host_download_command*) h->cmd;
	bh_response_record rr = {0};
	BH_ERRNO ret;

	if (!is_init())
		return BPE_NOT_INIT;

	if ( !pAppId || !pAppBlob)
		return BPE_INVALID_PARAMS;

	if (!string_to_uuid(pAppId, (char*)&(cmd->appid)))
		return BPE_INVALID_PARAMS;

	// if (AppSize + sizeof(bh_message_header) > 0x10000)
	// 	return BH_ILLEGAL_VALUE;
	h->id = HOST_CMD_DOWNLOAD;
	
	TRACE1 ("Beihai Download %x\n", &rr);

	ret = bh_send_message((char*)h, sizeof(*h) + sizeof (*cmd), pAppBlob, AppSize, rrmap_add(&rr));
	if ( ret == BH_SUCCESS )
		ret = rr.code;

	TRACE2 ("Beihai Download %x ret %x\n", &rr, rr.code);

	BHFREE (rr.buffer);
	return ret;
}


BH_ERRNO BH_PluginQueryAPI ( const char *AppId, const void* input, UINT32 length, char** output)
{
	char cmdbuf[CMDBUF_SIZE];
	bh_message_header* h = (bh_message_header*) cmdbuf;
	host_query_command* cmd = (host_query_command*) h->cmd;
	bh_response_record rr = {0};
	BH_ERRNO ret;

	if (!is_init())
		return BPE_NOT_INIT;

	if (!AppId || !input || !length || !output)
		return BPE_INVALID_PARAMS;

	if (!string_to_uuid(AppId, (char*)&(cmd->appid)))
		return BPE_INVALID_PARAMS;

	h->id = HOST_CMD_QUERYAPI;

	TRACE1 ("Beihai Query %x\n", &rr);

	ret = bh_send_message((char*)h, sizeof(*h) + sizeof (*cmd), input, length, rrmap_add(&rr));
	if ( ret == BH_SUCCESS )
		ret = rr.code;

	TRACE2 ("Beihai Query %x ret %x\n", &rr, rr.code);

	if ( ret == BH_SUCCESS ) {
		if (rr.length > 0 && rr.buffer) {
			int len = rr.length;
			if (output) {
				*output = (char*) BHMALLOC (len + 1);
				if (*output) {
					memcpy (*output, rr.buffer, len);
					((char*) *output) [len] = '\0';
				} else {
					ret = BPE_OUT_OF_MEMORY;
				}
			}
		} else if (rr.length == 0) {
			*output = NULL;
		} else
			ret = BPE_MESSAGE_TOO_SHORT;
	}

	BHFREE (rr.buffer);
	return ret;
}


BH_ERRNO BH_PluginCreateSession ( const char *pAppId, SHANDLE* psession, const void* init_buffer, UINT32 init_len)
{
	char cmdbuf[CMDBUF_SIZE];
	bh_message_header* h = (bh_message_header*) cmdbuf;
	host_create_session_command* cmd = (host_create_session_command*)h->cmd;
	bh_response_record* session;
	ADDR seq;
	BH_ERRNO ret;

	if (!is_init())
		return BPE_NOT_INIT;

	if (!pAppId || !psession)
		return BPE_INVALID_PARAMS;

	if (init_buffer == NULL && init_len != 0)
		return BPE_INVALID_PARAMS;

	if (!string_to_uuid(pAppId, (char*)&(cmd->appid)))
		return BPE_INVALID_PARAMS;

	session = (bh_response_record*) BHMALLOC (sizeof (bh_response_record));
	if (!session) {
		return BPE_OUT_OF_MEMORY;
	}

	session->session_lock = bh_create_mutex();
	if (!session->session_lock) {
		BHFREE(session);
		return BPE_OUT_OF_RESOURCE;
	}

	session->count = 1;
	session->is_session = 1;
	session->killed = 0;
	seq = rrmap_add(session);

	h->id = HOST_CMD_CREATE_SESSION;

	TRACE2 ("Beihai CreateSession %x %llx\n", session, seq);

	ret = bh_send_message((char*)h, sizeof (*h) + sizeof (*cmd), (char*)init_buffer, init_len, seq);
	if( ret == BH_SUCCESS)
		ret = session->code;

	TRACE2 ("Beihai CreateSession %x ret %x\n", session, session->code);

	BHFREE(session->buffer);
	session->buffer = NULL;

	if (ret == BH_SUCCESS) {
		*psession = (SHANDLE) seq;
		session_exit(session, seq);
	} else {
		session_close(session, seq);
		*psession = NULL;
	}

	return ret;
}


BH_ERRNO BH_PluginCloseSession (SHANDLE pSession)
{
	char cmdbuf[CMDBUF_SIZE];
	bh_message_header* h = (bh_message_header*) cmdbuf;
	host_destroy_session_command* cmd = (host_destroy_session_command*) h->cmd;
	ADDR seq = (ADDR) pSession;
	bh_response_record* rr;
	BH_ERRNO ret;

	if (!is_init())
		return BPE_NOT_INIT;

	if (!pSession)
		return BPE_INVALID_PARAMS;

	rr = session_enter(seq);
	if(!rr) {
		return BPE_INVALID_PARAMS;
	}


	h->id = HOST_CMD_CLOSE_SESSION;
	cmd->addr = rr->addr;

	TRACE1 ("Beihai CloseSession %x\n", rr);

	ret = bh_send_message((char*)h, sizeof(*h) + sizeof (*cmd), NULL, 0, seq);
	if ( ret == BH_SUCCESS)
		ret = rr->code;

	TRACE2 ("Beihai CloseSession %x ret %x\n", rr, rr->code);

	session_close(rr, seq);
	return ret;

}

BH_ERRNO BH_PluginForceCloseSession (SHANDLE pSession)
{
	char cmdbuf[CMDBUF_SIZE];
	bh_message_header* h = (bh_message_header*) cmdbuf;
	host_destroy_session_command* cmd = (host_destroy_session_command*) h->cmd;
	ADDR seq = (ADDR) pSession;
	bh_response_record rr = {0};
	bh_response_record* session_rr;
	BH_ERRNO ret;

	if (!is_init())
		return BPE_NOT_INIT;

	if (!pSession)
		return BPE_INVALID_PARAMS;

	session_rr = session_enter_nolock(seq);
	if(!session_rr) {
		return BPE_INVALID_PARAMS;
	}

	h->id = HOST_CMD_FORCE_CLOSE_SESSION;
	cmd->addr = session_rr->addr;

	TRACE1 ("Beihai ForceCloseSession %x\n", &rr);

	ret = bh_send_message((char*)h, sizeof(*h) + sizeof (*cmd), NULL, 0, rrmap_add(&rr));
	if ( ret == BH_SUCCESS)
		ret = rr.code;

	TRACE2 ("Beihai ForceCloseSession %x ret %x\n", &rr, rr.code);

	BHFREE(rr.buffer);
	session_close_nolock(session_rr, seq);
	return ret;
}


BH_ERRNO BH_PluginListProperties ( const char* AppId, int *number, char*** array)
{
	char cmdbuf[CMDBUF_SIZE];
	bh_message_header* h = (bh_message_header*) cmdbuf;
	host_list_properties_command* cmd = (host_list_properties_command*) h->cmd;
	bh_response_record rr = {0};
	BH_ERRNO ret;

	if (!is_init())
		return BPE_NOT_INIT;

	if (!AppId || !array || !number)
		return BPE_INVALID_PARAMS;

	if (!string_to_uuid(AppId, (char*)&(cmd->appid)))
		return BPE_INVALID_PARAMS;		

	h->id = HOST_CMD_LIST_PROPERTIES;

	TRACE1 ("Beihai List Sessions %x\n", &rr);

	ret = bh_send_message((char*)h, sizeof(*h) + sizeof (*cmd), NULL, 0, rrmap_add(&rr));
	if ( ret == BH_SUCCESS )
		ret = rr.code;

	TRACE2 ("Beihai List Sessions %x ret %x\n", &rr, rr.code);

	*array = NULL;
	*number = 0;

	do {
		if ( ret != BH_SUCCESS)
			break;
		if ( rr.buffer == NULL ) {
			ret = BPE_MESSAGE_ILLEGAL;
			break;
		}

		char* buf  = (char*) rr.buffer;
		if( buf[rr.length - 1] != '\0') {
			ret = BPE_MESSAGE_ILLEGAL;
			break;
		}

		int count = 0;
		char* pos = (char*) rr.buffer;

		while(pos < (char*) rr.buffer + rr.length) {
			pos += strlen(pos) + 1;
			count ++;
		}

		if (count == 0)
			break;

		char** output = (char**) BHMALLOC ((count+1) * sizeof (char*));
		if (!output) {
			ret = BPE_OUT_OF_MEMORY;
			break;
		}

		memset(output, 0, (count+1) * sizeof (char*));
		pos = (char*) rr.buffer;
		for (int i = 0; i< count; i++) {
			int pos_len = (int)strlen(pos) + 1;
			output[i] = (char*) BHMALLOC(pos_len);
			if (output[i] == NULL) {
				ret = BPE_OUT_OF_MEMORY;
				break;
			}

#ifdef _WIN32
			strncpy_s (output[i], pos_len, pos, pos_len);
#else
            strncpy (output[i], pos, pos_len);
#endif
			pos += pos_len;
		}

		if (ret == BPE_OUT_OF_MEMORY) {
			for (int i = 0; i< count; i++)
				BHFREE(output[i]);
			BHFREE(output);
			break;
		}

		*array = output;
		*number = count;
	} while(0);

	BHFREE (rr.buffer);
	return ret;
}

BH_ERRNO BH_PluginListSessions ( const char* AppId, int* count, SHANDLE** array)
{
	char cmdbuf[CMDBUF_SIZE];
	bh_message_header* h = (bh_message_header*) cmdbuf;
	host_list_sessions_command* cmd = (host_list_sessions_command*) h->cmd;
	bh_response_record rr = {0};
	BH_ERRNO ret;

	if (!is_init())
		return BPE_NOT_INIT;

	if (!AppId || !count || !array)
		return BPE_INVALID_PARAMS;

	if (!string_to_uuid(AppId, (char*)&(cmd->appid)))
		return BPE_INVALID_PARAMS;

	h->id = HOST_CMD_LIST_SESSIONS;

	TRACE1 ("Beihai List Sessions %x\n", &rr);

	ret = bh_send_message((char*)h, sizeof(*h) + sizeof (*cmd), NULL, 0, rrmap_add(&rr));
	if ( ret == BH_SUCCESS )
		ret = rr.code;

	TRACE2 ("Beihai List Sessions %x ret %x\n", &rr, rr.code);

	*array = NULL;
	*count = 0;

	do {
		if ( ret != BH_SUCCESS)
			break;
		if ( rr.buffer == NULL) {
			ret = BPE_MESSAGE_ILLEGAL;
			break;
		}
		
		client_list_sessions_response* resp = (client_list_sessions_response*) rr.buffer;
		if (resp->count == 0)
			break;

		if (rr.length != sizeof (ADDR) * resp->count + sizeof (client_list_sessions_response)) {
			ret = BPE_MESSAGE_ILLEGAL;
			break;
		}

		SHANDLE* outbuf = (SHANDLE*) BHMALLOC (sizeof (void*) * resp->count);
		if (!outbuf) {
			ret = BPE_OUT_OF_MEMORY;
			break;
		}

		memset (outbuf, 0, sizeof (SHANDLE) * resp->count);

		for (int i=0; i< resp->count; i++) {
			outbuf[i] = (SHANDLE) resp->addr[i];
		}

		*array = outbuf;
		*count = resp->count;
	} while(0);

	BHFREE (rr.buffer);
	return ret;
}


BH_ERRNO BH_PluginListPackages (int *number, char*** array)
{
	char cmdbuf[CMDBUF_SIZE];
	bh_message_header* h = (bh_message_header*) cmdbuf;
	bh_response_record rr = {0};
	BH_ERRNO ret;

	if (!is_init())
		return BPE_NOT_INIT;

	if (!array || !number)
		return BPE_INVALID_PARAMS;

	h->id = HOST_CMD_LIST_PACKAGES;

	TRACE1 ("Beihai List Packages %x\n", &rr);

	ret = bh_send_message((char*)h, sizeof(*h), NULL, 0, rrmap_add(&rr));
	if (  ret == BH_SUCCESS )
		ret = rr.code;

	TRACE2 ("Beihai List Packages %x ret %x\n", &rr, rr.code);

	*array = NULL;
	*number = 0;

	do {
		if ( ret != BH_SUCCESS)
			break;
		if ( rr.buffer == NULL) {
			ret = BPE_MESSAGE_ILLEGAL;
			break;
		}

		client_list_packages_response* resp = (client_list_packages_response*) rr.buffer;
		if (resp->count == 0)
			break;
			
		if (rr.length != sizeof (APPID) * resp->count + sizeof (client_list_packages_response)) {
			ret = BPE_MESSAGE_ILLEGAL;
			break;
		}
			
		char ** outbuf = (char**) BHMALLOC (sizeof (char*) * (resp->count + 1));
		if (!outbuf) {
			ret = BPE_OUT_OF_MEMORY;
			break;
		}

		outbuf[resp->count] = 0;
		for (int i = 0; i< resp->count; i++) {
			outbuf[i] = (char*) BHMALLOC(APPID_LENGTH * 2 + 1);
			if (outbuf[i] == NULL) {
				ret = BPE_OUT_OF_MEMORY;
				break;
			}
		}

		if (ret == BPE_OUT_OF_MEMORY) {
			for (int i = 0; i< resp->count; i++)
				BHFREE (outbuf[i]);
			BHFREE (outbuf);
			break;
		} 

		for (int i=0; i<resp->count; i++)
			uuid_to_string((char*)&(resp->id[i]), outbuf[i]);

		*array = outbuf;
		*number = resp->count;
	}while(0);

	BHFREE (rr.buffer);
	return ret;
}

void BH_FREE(void * p)
{
	BHFREE(p);
}

