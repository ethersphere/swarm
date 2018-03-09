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
 * @file  bhp_impl.h
 * @brief This file declares the BHP internal data type definition and interface.
 * @author
 * @version
 *
 */
#ifndef __BHP_IMPL_H__
#define __BHP_IMPL_H__

#include "bh_shared_conf.h"
#include "bh_shared_errcode.h"
#include "bh_shared_types.h"
#include "bhp_platform.h"
#include <map>

typedef struct {
    bhp_event_t wait_event; //the event which sender thread waits on
    BH_RET code; //the response code from firmware
    unsigned int length; //length of the response buffer
    void *buffer; //the response buffer
    BH_U64 addr; //remote address in firmware
    int is_session; //whether this record relates with session
    int killed; //whether this session is killed or not, valid only for is_session is 1
    unsigned int count; //the count of users who are using this session, valid only for is_session is 1
    bhp_mutex_t session_lock; //for exclusive operation on this session, valid only for is_session is 1
} bh_response_record;
//bh_connection_item using c++ map class, and should not be inside extern "C"
typedef struct {
    bhp_mutex_t lock; //for exclusive access of this item
    volatile uintptr_t handle;  //physical connection handle
    bhp_mutex_t bhm_send;  //exclusive pkg sending on this connection
    bhp_mutex_t bhm_rrmap; //exclusive access of rrmap
    std::map<BH_U64, bh_response_record*> rrmap;
    bhp_thread_t recv_thread; //recv_thread handle
    volatile unsigned int conn_count; //VM connection counter, only valid for VM
    BH_SDID sdid; //the sd id it serves, only valid for VM
} bh_connection_item;

#ifdef __cplusplus
extern "C" {
#endif
//maximum concurrent activities on one session
#define MAX_SESSION_LIMIT 20

//heci command header buffer size in bytes
#define CMDBUF_SIZE 100

typedef enum {
    CONN_IDX_START = 1,
    CONN_IDX_LAUNCHER = 1,
    CONN_IDX_SDM = 2,
    CONN_IDX_IVM = 3,
    CONN_IDX_SVM = 4,
    MAX_CONNECTIONS = 5
} BHP_CONN_IDX_T;

typedef enum {
    DEINITED = 0,
    INITED = 1,
} BHP_STATE_T;

//whether BHP is inited or not
int is_bhp_inited(void);

//Add a rr to rrmap and return a new seq number.
BH_U64 rrmap_add(int conn_idx, bh_response_record* rr);

//session enter with session handle seq
bh_response_record* session_enter(int conn_idx, BH_U64 seq, int lock_session);

//session exit
void session_exit(int conn_idx, bh_response_record* session, BH_U64 seq, int unlock_session);

//session close
void session_close(int conn_idx, bh_response_record* session, BH_U64 seq, int unlock_session);

//send one message through heci
BH_RET bh_send_message (int conn_idx, void* cmd, unsigned int clen, const void* data, unsigned int dlen, BH_U64 seq);

enum {
    BHP_OPEN_VM_QUERY_MODE = 0,
    BHP_OPEN_VM_NORMAL_MODE = 1
};

//open vm connection for sdid and increase vm connection counter by 1
BH_RET bh_do_openVM (BH_SDID sdid, int* conn_idx, int mode);

//decrease vm connection counter by 1
BH_RET bh_do_closeVM(int conn_idx);

#ifdef __cplusplus
}
#endif

#endif
