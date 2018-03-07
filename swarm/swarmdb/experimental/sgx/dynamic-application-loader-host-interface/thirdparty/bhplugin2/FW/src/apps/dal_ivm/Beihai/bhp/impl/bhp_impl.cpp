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
 * @file  bhp_impl.cpp
 * @brief This file implements Beihai Host Proxy (BHP) module core functionality.
 * @author
 * @version
 *
 */
#include "bhp_exp.h"
#include "bhp_heci.h"
#include "bh_acp_util.h"
#include "bhp_platform.h"
#include "bhp_impl.h"

#ifndef max
#define max(a,b)            (((a) > (b)) ? (a) : (b))
#endif
#ifndef min
#define min(a,b)            (((a) < (b)) ? (a) : (b))
#endif

static bh_connection_item connections[MAX_CONNECTIONS]; //slot 0 is reserved

static const int heci_port_list[MAX_CONNECTIONS] = { //should be same order as connections[]
    0, //placeholder
    BH_LAUNCHER_HECI_PORT,
    BH_SDM_HECI_PORT,
    BH_IVM_HECI_PORT,
    BH_SVM_HECI_PORT
};

static volatile unsigned int init_state = DEINITED;

static bhp_mutex_t bhm_gInit = NULL; //global init/deinit mutex

static BHP_TRANSPORT bhp_tx_itf= {0}; //transport func list, set during init

static BH_SDID g_isd_uuid = {0}; //the isd-id in the firmware, got during init

#define MAX_TXRX_LENGTH 4096
static unsigned char skip_buffer[MAX_TXRX_LENGTH] = {0};

static bhp_mutex_t bhm_seqno = NULL; //seq no operation mutex

int is_bhp_inited(void) {
    return (init_state == INITED);
}

static bh_response_record* addr2record(int conn_idx, BH_U64 seq)
{
    bh_response_record* rr = NULL;
    mutex_enter(connections[conn_idx].bhm_rrmap);
    if (connections[conn_idx].rrmap.find(seq) != connections[conn_idx].rrmap.end())
        rr = (bh_response_record*) connections[conn_idx].rrmap[seq];
    mutex_exit(connections[conn_idx].bhm_rrmap);
    return rr;
}

static void destroy_session(bh_response_record* session)
{
    BHP_LOG_DEBUG("destroy_session %x\n", session);
    if (session->session_lock != NULL) {
        bh_close_mutex(session->session_lock);
        session->session_lock = NULL;
    }
    if (session->buffer != NULL ) {
        BHFREE(session->buffer);
        session->buffer = NULL;
    }
    BHFREE (session);
}

bh_response_record* session_enter(int conn_idx, BH_U64 seq, int lock_session)
{
    bh_response_record* session = NULL;

    mutex_enter(connections[conn_idx].bhm_rrmap);
    if (connections[conn_idx].rrmap.find (seq) != connections[conn_idx].rrmap.end() 
        && connections[conn_idx].rrmap[seq]->is_session
        && !connections[conn_idx].rrmap[seq]->killed) {
        session = connections[conn_idx].rrmap[seq];
        if (session->count < MAX_SESSION_LIMIT) {
            session->count++;
        } else {
            session = NULL;
        }
    }
    mutex_exit(connections[conn_idx].bhm_rrmap);

    if (session && lock_session) {
        mutex_enter(session->session_lock);
        //check whether session has been killed before session operation
        if (session->killed) {
            session_exit(conn_idx,session, seq, 1);
            session = NULL;
        }
    }
    return session;
}

void session_exit(int conn_idx, bh_response_record* session, BH_U64 seq, int unlock_session)
{
    int closeVMConn = 0;

    mutex_enter(connections[conn_idx].bhm_rrmap);
    session->count--;
    if (session->count == 0 && session->killed) {
        connections[conn_idx].rrmap.erase(seq);
        if (unlock_session) mutex_exit(session->session_lock);
        destroy_session(session);
        if (conn_idx > CONN_IDX_IVM) closeVMConn = 1;
    } else {
        if (unlock_session) mutex_exit(session->session_lock);
    }
    mutex_exit(connections[conn_idx].bhm_rrmap);

    if (closeVMConn) {
        //remove the VM conn counter of this session:only for connected SVM
        bh_do_closeVM(conn_idx);
    }
}

void session_close(int conn_idx, bh_response_record* session, BH_U64 seq, int unlock_session)
{
    int closeVMConn = 0;

    mutex_enter(connections[conn_idx].bhm_rrmap);
    session->count--;
    if (session->count == 0) {
        connections[conn_idx].rrmap.erase(seq);
        if (unlock_session) mutex_exit(session->session_lock);
        destroy_session(session);
        if (conn_idx > CONN_IDX_IVM) closeVMConn = 1;
    } else {
        session->killed = 1;
        if (unlock_session) mutex_exit(session->session_lock);
    }
    mutex_exit(connections[conn_idx].bhm_rrmap);

    if (closeVMConn) {
        //remove the VM conn counter of this session:only for connected SVM
        bh_do_closeVM(conn_idx);
    }
}

static void* bh_close_svm_thread_func (void* args) {
    bh_do_closeVM((int)(uintptr_t)args);
    return NULL;
}

static void session_kill(uintptr_t conn_idx, bh_response_record* session, BH_U64 seq, int callerIsSVMRecvThread)
{
    int closeVMConn = 0;

    mutex_enter(connections[conn_idx].bhm_rrmap);
    session->killed = 1;
    if (session->count == 0) {
        connections[conn_idx].rrmap.erase(seq);
        destroy_session(session);
        if (conn_idx > CONN_IDX_IVM) closeVMConn = 1;
    }
    mutex_exit(connections[conn_idx].bhm_rrmap);

    if (closeVMConn) {
        //decrease the VM conn counter of this session:only for connected SVM
        //Note: callerIsSVMRecvThread is always 1 in current impl, as caller of this func is only bh_recv_message().
        if (!callerIsSVMRecvThread) {
            bh_do_closeVM((int)conn_idx);
        } else {
            mutex_enter(connections[conn_idx].lock);
            if (connections[conn_idx].conn_count == 1) {
                //this is the last vm connection to be closed, so startup new thread to close svm, 
                //    otherwise the recv_thread will deadlock.
                bhp_thread_t closeSvmThread = bh_thread_create(bh_close_svm_thread_func, (void*)conn_idx);
                if (closeSvmThread != NULL) bh_thread_close(closeSvmThread);
            } else {
                connections[conn_idx].conn_count --;
            }
            mutex_exit(connections[conn_idx].lock);
        }
    }
}

/*
 * function inc_seqno():
 *   increase the shared variable g_seqno by 1 and wrap around if needed.
 * note: g_seqno is shared resource among all connections/threads.
 * As the JAVATA_SESSION_HANDLE/SD_SESSION_HANDLE is (void*) type,
 * it could only store 32-bit value in 32-bit machine.
 * so we define g_seqno as BH_U32, and it should be enough for usage.
 */
static BH_U64 inc_seqno() {
    static BH_U32 g_seqno = 0;
    BH_U32 ret = 0;

    if (bhm_seqno == NULL) BHP_LOG_FATAL("[BHP] FATAL: Out of resource, bhm_seqno is NULL\n");
    mutex_enter(bhm_seqno);
    g_seqno++;
    //wrap around. g_seqno must not be 0, as required by Firmware VM.
    if (g_seqno == 0) g_seqno = 1;
    ret = g_seqno;
    mutex_exit(bhm_seqno);

    return (BH_U64)ret;
}

BH_U64 rrmap_add(int conn_idx, bh_response_record* rr)
{
    BH_U64 seq= inc_seqno();

    mutex_enter(connections[conn_idx].bhm_rrmap);
    connections[conn_idx].rrmap[seq] = rr;
    BHP_LOG_DEBUG("rrmap_add idx-%d %I64x %x\n", conn_idx, seq, rr);
    mutex_exit(connections[conn_idx].bhm_rrmap);

    return seq;
}

static bh_response_record* rrmap_remove(int conn_idx, BH_U64 seq)
{
    bh_response_record* rr = NULL;

    mutex_enter(connections[conn_idx].bhm_rrmap);
    if (connections[conn_idx].rrmap.find(seq) != connections[conn_idx].rrmap.end()) {
        rr = connections[conn_idx].rrmap[seq];
        if (!rr->is_session) {
            connections[conn_idx].rrmap.erase(seq);
            BHP_LOG_DEBUG("rrmap_erase idx-%d %I64x %x\n", conn_idx, seq, rr);
        }
    }
    mutex_exit(connections[conn_idx].bhm_rrmap);

    return rr;
}

static BH_RET bh_transport_init(const BHP_TRANSPORT* context)
{
    BH_RET ret = BH_SUCCESS;

    memcpy(&bhp_tx_itf, context, sizeof(BHP_TRANSPORT));

    if (bhp_tx_itf.pfnConnect == NULL
        || bhp_tx_itf.pfnClose == NULL
        || bhp_tx_itf.pfnSend == NULL
        || bhp_tx_itf.pfnRecv == NULL) {
        BHP_LOG_DEBUG("FATAL error: Invalid transport function.\n");
        ret = BPE_INVALID_PARAMS;
    }

    return ret;
}


static BH_RET bh_transport_recv (uintptr_t handle, void* buffer, uint32_t size)
{
    unsigned int got = 0;
    unsigned int count = 0;
    int status = 0;

    if (!handle) return BPE_COMMS_ERROR;

    while (size - count > 0) {
        if (buffer) {
            got = min (size - count, MAX_TXRX_LENGTH);
            status = bhp_tx_itf.pfnRecv (handle, (unsigned char*) buffer + count,  &got);
        } else {
            got = min (MAX_TXRX_LENGTH, size - count);
            status = bhp_tx_itf.pfnRecv (handle, skip_buffer, &got);
        }

        if (status != 0) return BPE_COMMS_ERROR;

        count += got;
    }

    return BH_SUCCESS;
}

static BH_RET bh_transport_send (uintptr_t handle, const void* buffer, unsigned int size)
{
    int status = 0;
	if (!handle) return BPE_COMMS_ERROR;

	status = bhp_tx_itf.pfnSend(handle, (unsigned char*)buffer, size);
	if (status != 0) {
		return BPE_COMMS_ERROR;
	}
    return BH_SUCCESS;
}

static BH_RET bh_recv_message(int conn_idx)
{
    bhp_response_header headbuf = {0};
    bhp_response_header *head = &headbuf;
    char* data = NULL;
    unsigned int dlen = 0;
    BH_RET ret = BH_SUCCESS;
    bh_response_record* rr = NULL;

    ret = bh_transport_recv(connections[conn_idx].handle, (char*) head, sizeof (bhp_response_header));
    if (ret != BH_SUCCESS) return ret;
 
    /* check magic */
    if (memcmp(BHP_MSG_RESPONSE_MAGIC, head->h.magic, BHP_MSG_MAGIC_LENGTH) != 0) return BPE_MESSAGE_ILLEGAL;

    // verify rr
    rr = rrmap_remove(conn_idx, head->seq);
    if (!rr) {
        BHP_LOG_WARN ("Beihai RECV invalid rr idx-%d 0x%I64x\n", conn_idx, head->seq);
    }

    BHP_LOG_DEBUG("enter bh_recv_message 0x%x 0x%I64x %d\n", rr, head->seq, head->code);

    if (head->h.length > sizeof(bhp_response_header)) {
        dlen = head->h.length - sizeof(bhp_response_header);
        data = (char*) BHMALLOC(dlen);
        ret = bh_transport_recv(connections[conn_idx].handle, data, dlen);
        if (ret == BH_SUCCESS && data == NULL) ret = BPE_OUT_OF_MEMORY;
    }

    BHP_LOG_DEBUG("exit bh_recv_message %x %I64x %d\n", rr, head->seq, ret);

    if (rr) {
        rr->buffer = data;
        rr->length = dlen;

        if (ret == BH_SUCCESS) rr->code = (BH_RET)head->code;
        else rr->code = ret;

        if (head->addr) rr->addr = head->addr;

        int sessionKilled = (rr->is_session &&
            (rr->code == BHE_WD_TIMEOUT
            || rr->code == BHE_UNCAUGHT_EXCEPTION
            || rr->code == BHE_APPLET_CRASHED));
        if (sessionKilled) rr->killed = 1; //set killed flag before wake up send_wait thread.

        if (rr->wait_event) {
            bh_signal_event (rr->wait_event);
        } else if (sessionKilled) {
            //VM instance abnormal exit, and no corresponding send_wait thread.
            session_kill(conn_idx, rr, head->seq, 1);
        }
    } else {
        if (data) BHFREE(data);
    }

    return ret;
}

static BH_RET _send_message (int conn_idx, void* cmd, unsigned int clen, const void* data, unsigned int dlen, bh_response_record *rr, BH_U64 seq)
{
    BH_RET ret = BH_SUCCESS;
    bhp_command_header *h = NULL;

    if (clen < sizeof(bhp_command_header) || !cmd || !rr) return BPE_INVALID_PARAMS;

    rr->buffer = NULL;
    rr->length = 0;
    rr->wait_event = bh_create_event();
    if (rr->wait_event == NULL) {
        ret = BPE_OUT_OF_RESOURCE;
        goto cleanup;
    }

    memcpy (cmd, BHP_MSG_CMD_MAGIC, BHP_MSG_MAGIC_LENGTH);

    h = (bhp_command_header*)cmd;
    h->h.length = clen + dlen;
    h->seq = seq;

    ret = bh_transport_send(connections[conn_idx].handle, cmd, clen);
    if (ret == BH_SUCCESS && dlen>0) {
        ret = bh_transport_send(connections[conn_idx].handle, data, dlen);
    }

cleanup:
    if (ret != BH_SUCCESS)	{
        if (rr->wait_event) bh_close_event(rr->wait_event);
        rrmap_remove(conn_idx, seq);
    }

    return ret;
}

BH_RET bh_send_message (int conn_idx, void* cmd, unsigned int clen, const void* data, unsigned int dlen, BH_U64 seq)
{
    BH_RET ret = BH_SUCCESS;
    bh_response_record *rr = addr2record(conn_idx, seq);

    if (!rr) { //should not happen
        BHP_LOG_FATAL("[BHP] FATAL: rr record NULL with seq=%d.\n", (BH_U32)seq);
        return BPE_INTERNAL_ERROR;
    }
    mutex_enter(connections[conn_idx].bhm_send);
    BHP_LOG_DEBUG("enter bh_send_message %x %d\n", rr, clen+dlen);
    ret = _send_message (conn_idx, cmd, clen, data, dlen, rr, seq);
    BHP_LOG_DEBUG ("done bh_send_message %x %d\n", rr, clen+dlen);

    if (ret == BH_SUCCESS && rr && rr->wait_event) {
        mutex_exit (connections[conn_idx].bhm_send);
        bh_wait_event(rr->wait_event);
        bh_close_event(rr->wait_event);
        rr->wait_event = NULL;
    } else {
        mutex_exit (connections[conn_idx].bhm_send);
    }

    return ret;
}

static void unblock_threads (int conn_idx, BH_RET code) {
    std::map<BH_U64, bh_response_record*>::iterator it, it2;

    mutex_enter(connections[conn_idx].bhm_rrmap);
    for (it = connections[conn_idx].rrmap.begin(); it != connections[conn_idx].rrmap.end(); ) {
        bh_response_record* rr = it->second;
        it2 = it;    it++;
        if (rr) {
            rr->code = code;
            if (rr->wait_event) {
                //set killed flag before wakeup, so the session obj would be released.
                if (rr->is_session) rr->killed = 1;
                bh_signal_event(rr->wait_event);
                if (!rr->is_session) connections[conn_idx].rrmap.erase(it2);
            } else {
                if (rr->is_session && rr->count == 0) {
                    //rr is not used in any sender thread, but cached in user app.
                    destroy_session(rr);
                    //no need to decrease this session's conn counter, as the connection disconnected.
                    connections[conn_idx].rrmap.erase(it2);
                }
                //Don't call rrmap.erase() to let session-rr or non-session-rr continues its work and call erase by itself thread
            }
        } else {
            //invalid record
            connections[conn_idx].rrmap.erase(it2);
        }
    }

    mutex_exit(connections[conn_idx].bhm_rrmap);

    BHP_LOG_DEBUG("unblock_threads conn_idx=%d, rrmap.empty()=%d\n", conn_idx, connections[conn_idx].rrmap.empty());

    //Note: Recv thread doesn't need to wait rrmap empty before exiting,
    //because JHI service programming and reset-svm processing will ask Launcher for svm status.
    mutex_enter(connections[conn_idx].bhm_rrmap);
    connections[conn_idx].rrmap.clear();
    mutex_exit(connections[conn_idx].bhm_rrmap);
}

static void* bh_recv_thread_func (void* args) {
    int conn_idx = (int)(uintptr_t)args;
    BH_RET ret = BH_SUCCESS;
    int i =0;

    while(1) {
        ret = bh_recv_message(conn_idx);
        if (ret != BH_SUCCESS) { //heci connection disconnected?
            bhp_tx_itf.pfnClose(connections[conn_idx].handle);
            connections[conn_idx].handle = 0; //reset handle
            if (conn_idx < CONN_IDX_SVM) { // fatal error:IBL process disconnected
                if (conn_idx == CONN_IDX_START) {
                    for (i=CONN_IDX_START+1; i< MAX_CONNECTIONS; i++) {
                        if (connections[i].handle != 0) {
                            //notify other connections to disconnect
                            bhp_tx_itf.pfnClose(connections[i].handle);
                        }
                    }
                } else if (connections[CONN_IDX_START].handle != 0) {
                    bhp_tx_itf.pfnClose(connections[CONN_IDX_START].handle);
                }
            }
            break;
        }
    }
    //cleanup
    unblock_threads(conn_idx, BPE_COMMS_ERROR); //connection with firmware error
    BHP_LOG_DEBUG("bh_recv_thread exit, conn_idx=%d.\n", conn_idx);
    return NULL;
}

static BH_RET bh_do_connect(uintptr_t conn_idx, int heci_port)
{
    BH_RET ret = BH_SUCCESS;
    int temp_ret =0;
    uintptr_t handle = 0;

    connections[conn_idx].handle = 0;
    connections[conn_idx].recv_thread = NULL;
    connections[conn_idx].conn_count = 0;
    connections[conn_idx].rrmap.clear();
    memset(&connections[conn_idx].sdid,0,sizeof(BH_SDID));

    temp_ret = bhp_tx_itf.pfnConnect(heci_port, &handle);
    if (temp_ret != 0) {
        ret = BPE_CONNECT_FAILED;
        BHP_LOG_WARN("bh_do_connect() failed: idx=%d, port=%d.\n", conn_idx, heci_port);
        goto cleanup;
    }
    connections[conn_idx].handle = handle;

    connections[conn_idx].recv_thread = bh_thread_create(bh_recv_thread_func, (void*)conn_idx);
    if (connections[conn_idx].recv_thread == NULL) {
        ret = BPE_OUT_OF_RESOURCE;
        goto cleanup;
    }

cleanup:
    if (ret != BH_SUCCESS) {
        if (connections[conn_idx].handle != 0) {
            bhp_tx_itf.pfnClose(connections[conn_idx].handle);
            connections[conn_idx].handle = 0;
            //here: recv_thread creation must be failed, and don't need to terminate it.
        }
    }

    return ret;
}

static BH_RET bh_do_disconnect(int conn_idx) 
{
    BH_RET ret = BH_SUCCESS;

    if (connections[conn_idx].handle != 0) {
        bhp_tx_itf.pfnClose(connections[conn_idx].handle);
        //connections[conn_idx].handle will be reset to 0 when recv_thread exits
        //wait for the recv thread exit
        
        /*pfnClose() function above doesn't unblock recv thread in linux due to 
        linux heci driver issue, so we call bh_thread_cancel here for workaround*/
        bh_thread_cancel(connections[conn_idx].recv_thread);

        bh_thread_join(connections[conn_idx].recv_thread);
        bh_thread_close(connections[conn_idx].recv_thread);
    }
    connections[conn_idx].conn_count = 0;
    connections[conn_idx].handle = 0;
    connections[conn_idx].recv_thread = NULL;
    connections[conn_idx].rrmap.clear();
    memset(&connections[conn_idx].sdid,0,sizeof(BH_SDID));

    return ret;
}

static BH_RET bh_connections_init(void)
{
    BH_RET ret = BH_SUCCESS;
    int i = 0;

    for (i=CONN_IDX_START;i<MAX_CONNECTIONS;i++){
        connections[i].conn_count = 0;
        connections[i].handle = 0;
        connections[i].recv_thread = NULL;
        connections[i].rrmap.clear();
        connections[i].lock = bh_create_mutex();
        if (connections[i].lock == NULL) {
            ret = BPE_OUT_OF_RESOURCE;
            goto cleanup;
        }
        connections[i].bhm_send = bh_create_mutex();
        if (connections[i].bhm_send == NULL) {
            ret = BPE_OUT_OF_RESOURCE;
            goto cleanup;
        }
        connections[i].bhm_rrmap = bh_create_mutex();
        if (connections[i].bhm_rrmap == NULL) {
            ret = BPE_OUT_OF_RESOURCE;
            goto cleanup;
        }
    }

    for (i=CONN_IDX_START; i<CONN_IDX_SVM; i++) {
        //connect to predefined heci ports, except SVM
        ret = bh_do_connect(i, heci_port_list[i]);
        if (ret != BH_SUCCESS) break;
    }

cleanup:
    if (ret != BH_SUCCESS) {
        for (i=CONN_IDX_START; i<CONN_IDX_SVM; i++) {
            if (connections[i].handle != 0) {
                bhp_tx_itf.pfnClose(connections[i].handle);
                bh_thread_join(connections[i].recv_thread);
                bh_thread_close(connections[i].recv_thread);
                //when recv_thread exits, it will reset the handle to 0.
                connections[i].recv_thread = NULL;
            }
            if (connections[i].bhm_rrmap) {
                bh_close_mutex(connections[i].bhm_rrmap);
                connections[i].bhm_rrmap = NULL;
            }
            if (connections[i].bhm_send) {
                bh_close_mutex(connections[i].bhm_send);
                connections[i].bhm_send = NULL;
            }
            if (connections[i].lock) {
                bh_close_mutex(connections[i].lock);
                connections[i].lock = NULL;
            }
        }
    }

    return ret;
}

static void bh_connections_deinit () {
    int i=0;

    BHP_LOG_DEBUG("BHP bh_connections_deinit \n");

    for (i=CONN_IDX_START; i<MAX_CONNECTIONS; i++) {
        bh_do_disconnect(i);

        if (connections[i].bhm_send) {
            bh_close_mutex(connections[i].bhm_send);
            connections[i].bhm_send = NULL;
        }
        if (connections[i].bhm_rrmap) {
            bh_close_mutex(connections[i].bhm_rrmap);
            connections[i].bhm_rrmap = NULL;
        }
        if (connections[i].lock) {
            bh_close_mutex(connections[i].lock);
            connections[i].lock = NULL;
        }
    }
}

static BH_RET bh_proxy_reset(int conn_idx)
{
    char cmdbuf[CMDBUF_SIZE] = {0};
    bhp_command_header* h = (bhp_command_header*) cmdbuf;
    bh_response_record rr = {0};
    BH_RET ret = BH_SUCCESS;

    h->id = BHP_CMD_RESET;

    ret = bh_send_message(conn_idx,(char*)h, sizeof(*h), NULL, 0, rrmap_add(conn_idx, &rr));
    if (ret == BH_SUCCESS) ret = rr.code;

    if (rr.buffer) BHFREE(rr.buffer);

    return ret;
}

static BH_RET bh_proxy_reset_launcher(unsigned int* count, int** ports)
{
    char cmdbuf[CMDBUF_SIZE] = {0};
    bhp_command_header* h = (bhp_command_header*) cmdbuf;
    bh_response_record rr = {0};
    BH_RET ret = BH_SUCCESS;

    if (count == NULL || ports == NULL) return BPE_INVALID_PARAMS;

    h->id = BHP_CMD_RESET;

    ret = bh_send_message(CONN_IDX_LAUNCHER,(char*)h, sizeof(*h), NULL, 0, rrmap_add(CONN_IDX_LAUNCHER, &rr));
    if (ret == BH_SUCCESS) ret = rr.code;

    *ports = NULL;
    *count = 0;
    do {
        if (ret != BH_SUCCESS) break;
        if (rr.buffer == NULL
            || rr.length < sizeof (bhp_reset_launcher_response)) {
            ret = BPE_MESSAGE_ILLEGAL;
            break;
        }
        bhp_reset_launcher_response* resp = (bhp_reset_launcher_response*) rr.buffer;
        if (resp->count == 0) break;

        if (rr.length != sizeof (int) * resp->count + sizeof (bhp_reset_launcher_response)) {
            ret = BPE_MESSAGE_ILLEGAL;
            break;
        }
        *ports = (int*) BHMALLOC(sizeof(int) * resp->count);
        if (*ports == NULL) {
            ret = BPE_OUT_OF_MEMORY;
            break;
        }
        memcpy((void*)*ports, resp->vm_heci_port_list, resp->count * sizeof(int));
        *count = resp->count;
    } while(0);

    if (rr.buffer) BHFREE(rr.buffer);

    return ret;
}

static BH_RET bh_proxy_close_vm(BH_SDID sdid)
{
    char cmdbuf[CMDBUF_SIZE] = {0};
    bhp_command_header* h = (bhp_command_header*) cmdbuf;
    bhp_close_vm_cmd* cmd = (bhp_close_vm_cmd*) h->cmd;
    bh_response_record rr = {0};
    BH_RET ret = BH_SUCCESS;

    h->id = BHP_CMD_CLOSE_VM;
    cmd->sdid = sdid;

    ret = bh_send_message(CONN_IDX_LAUNCHER, (char*)h, sizeof(*h) + sizeof(*cmd), NULL, 0, rrmap_add(CONN_IDX_LAUNCHER,&rr));
    if (ret == BH_SUCCESS) ret = rr.code;

    if (rr.buffer) BHFREE(rr.buffer);

    return ret;
}

static BH_RET bh_proxy_reset_svm(int conn_idx)
{
    BH_RET ret = BPE_INVALID_PARAMS;
    BH_SDID sdid = {0};

    if (conn_idx <= CONN_IDX_IVM || connections[conn_idx].handle == 0) return ret;
    sdid = connections[conn_idx].sdid;
    //send RESET cmd to VM
    ret = bh_proxy_reset(conn_idx);
    if (ret == BH_SUCCESS) {
        //wait for the SVM recv thread exit.
        bh_thread_join(connections[conn_idx].recv_thread);
        bh_thread_close(connections[conn_idx].recv_thread);
        connections[conn_idx].recv_thread = NULL;
        //send closeVM to Launcher, which will wait for SVM process exit
        ret = bh_proxy_close_vm(sdid);
    }

    return ret;
}

/*
 * function: bh_proxy_get_isd()
 *   get isd uuid from SDM in Firmware side.
 */
static BH_RET bh_proxy_get_isd(void) {
    char cmdbuf[CMDBUF_SIZE] = {0};
    bhp_command_header* h = (bhp_command_header*) cmdbuf;
    bh_response_record rr = {0};
    BH_RET ret = BH_SUCCESS;

    BHP_LOG_DEBUG ("Beihai get_isd 0x%x\n", &rr);
    h->id = BHP_CMD_GET_ISD;

    ret = bh_send_message(CONN_IDX_SDM, (char*)h, sizeof(*h), NULL, 0, rrmap_add(CONN_IDX_SDM, &rr));
    if (ret == BH_SUCCESS)  ret = rr.code;

    BHP_LOG_DEBUG ("Beihai get_isd 0x%x ret 0x%x\n", &rr, rr.code);

    if (ret == BH_SUCCESS) {
        if (rr.buffer && rr.length == sizeof(bhp_get_isd_response)) {
            bhp_get_isd_response* resp = (bhp_get_isd_response*)rr.buffer;
            g_isd_uuid = resp->sdid;
        } else {
            ret = BPE_MESSAGE_ILLEGAL;
        }
    }

    if (rr.buffer) BHFREE(rr.buffer);
    return ret;
}

BH_RET BHP_Init(const BHP_TRANSPORT* transport, int do_vm_reset)
{
    BH_RET ret = BH_SUCCESS;
    unsigned int count_svm = 0;
    int* ports_svm = NULL;

    BHP_LOG_DEBUG("BHP INIT \n");

    if (transport == NULL) return BPE_INVALID_PARAMS;
    if (bhm_gInit == NULL) {
        bhm_gInit = bh_create_mutex();
        if (bhm_gInit == NULL) return BPE_OUT_OF_RESOURCE;
    }

    mutex_enter(bhm_gInit);

    bhm_seqno = bh_create_mutex();
    if (bhm_seqno == NULL) {
        ret = BPE_OUT_OF_RESOURCE;
        goto cleanup;
    }

    if (init_state == INITED) {
        ret = BPE_INITIALIZED_ALREADY;
        goto cleanup;
    }

    //step 1: init connections to each process
    ret = bh_transport_init(transport);
    if (ret == BH_SUCCESS) ret = bh_connections_init();
    if (ret != BH_SUCCESS) {
        goto cleanup;
    }

    //step 2: send reset cmd to each process in correct order - do vm reset only if needed.
    if (do_vm_reset) {
        ret = bh_proxy_reset(CONN_IDX_SDM);
        if (ret == BH_SUCCESS) ret = bh_proxy_reset_launcher(&count_svm, &ports_svm);
        if (ret == BH_SUCCESS && count_svm > 0) {
            //we have at most 1 svm
            int port = ports_svm[0];
            ret = bh_do_connect(CONN_IDX_SVM, port);
            if (ret == BH_SUCCESS) {
                ret = bh_proxy_reset_svm(CONN_IDX_SVM);
            }
        }
        if (ports_svm) BHFREE(ports_svm);
        if (ret == BH_SUCCESS) ret = bh_proxy_reset(CONN_IDX_IVM);
    }

    //step 3: get isd-uuid from SDM
    if (ret == BH_SUCCESS) {
        ret = bh_proxy_get_isd();
    }

    if (ret != BH_SUCCESS) {
        bh_connections_deinit();
    } else {
        //this assignment is atomic operation
        init_state = INITED;
    }

cleanup:
    mutex_exit(bhm_gInit);

    return ret;
}

BH_RET BHP_Deinit(int do_vm_reset)
{
    BH_RET ret = BH_SUCCESS;

    if (!is_bhp_inited()) return BPE_NOT_INIT;
    mutex_enter(bhm_gInit);

    if (init_state == INITED) {
        //do vm reset only if needed
        if (do_vm_reset) {
            BHP_Reset(); //reset fw and let SVM(if any) exit
        }

        bh_connections_deinit();
        init_state = DEINITED;

        bh_close_mutex(bhm_seqno);
        bhm_seqno = NULL;
    } else {
        ret = BPE_NOT_INIT;
    }

    mutex_exit(bhm_gInit);

    return ret;
}

BH_RET BHP_Reset(void)
{
    BH_RET ret = BH_SUCCESS;
    BH_RET ret_tmp = BH_SUCCESS;
    unsigned int count_svm = 0;
    int* ports_svm = NULL;

    if (!is_bhp_inited()) return BPE_NOT_INIT;
    mutex_enter(bhm_gInit);

    //disconnect svm and unblock all user threads to avoid the recursive reset_svm() below
    bh_do_disconnect(CONN_IDX_SVM);
    //send reset cmd to each process in correct order
    ret_tmp = bh_proxy_reset(CONN_IDX_SDM);
    if (ret_tmp != BH_SUCCESS) ret = ret_tmp;

    ret_tmp = bh_proxy_reset_launcher(&count_svm, &ports_svm);
    if (ret_tmp == BH_SUCCESS && count_svm > 0) {
        //we have at most 1 svm
        int port = ports_svm[0];
        ret_tmp = bh_do_connect(CONN_IDX_SVM, port);
        if (ret_tmp == BH_SUCCESS) {
            ret_tmp = bh_proxy_reset_svm(CONN_IDX_SVM);
        }
    }
    if (ports_svm) BHFREE(ports_svm);
    if (ret_tmp != BH_SUCCESS) ret = ret_tmp;

    ret_tmp = bh_proxy_reset(CONN_IDX_IVM);
    if (ret_tmp != BH_SUCCESS) ret = ret_tmp;

    mutex_exit(bhm_gInit);

    return ret;
}

BH_RET bh_do_openVM (BH_SDID sdid, int* conn_idx, int mode)
{
#if BEIHAI_ENABLE_OEM_SIGNING_IOTG
    if (conn_idx == NULL) return BPE_INVALID_PARAMS;
    *conn_idx = CONN_IDX_IVM;
    return BH_SUCCESS;
#else
    BH_RET ret = BPE_SERVICE_UNAVAILABLE;

    if (conn_idx == NULL) return BPE_INVALID_PARAMS;
    if (memcmp(&sdid, &g_isd_uuid,sizeof(BH_SDID)) == 0) {
        *conn_idx = CONN_IDX_IVM;
        return BH_SUCCESS;
    }
#if (!BEIHAI_ENABLE_SVM)
    return BPE_INVALID_PARAMS;
#else
    mutex_enter(connections[CONN_IDX_SVM].lock);
    if (connections[CONN_IDX_SVM].handle > 0
        && memcmp(&connections[CONN_IDX_SVM].sdid, &sdid, sizeof(BH_SDID)) == 0) {
        unsigned int val = (++connections[CONN_IDX_SVM].conn_count);
        BHP_LOG_DEBUG("svm conn_count inc = %d\n", val);
        ret = BH_SUCCESS;
    }
    if (mode == BHP_OPEN_VM_QUERY_MODE || ret == BH_SUCCESS) {
        //simply query vm conn status or sd-id match
        goto cleanup;
    }
    do {
        //sd-id has checked and not match
        if (connections[CONN_IDX_SVM].handle > 0) {
            if (connections[CONN_IDX_SVM].conn_count > 0) {
                ret = BPE_OUT_OF_RESOURCE;
                break;
            }
            //should not happen
            ret = BPE_INTERNAL_ERROR;
            break;
        }
        //need launch vm and connect
        //1: launch vm
        int heci_port = 0;
        ret = bh_proxy_launch_vm(sdid, &heci_port);
        if (ret != BH_SUCCESS) {
            BHP_LOG_FATAL("BHP-open-vm lauchVM failed, ret=0x%x.\n", ret);
            ret = BPE_OUT_OF_RESOURCE;
            break;
        }
        //2: connect to the heci-port
        ::Sleep(3000); //Daoming: wait some time for SVM HECI ready
        ret = bh_do_connect(CONN_IDX_SVM, heci_port);
        if (ret != BH_SUCCESS) {
            //NOTE: this should not happen. If it happens, host record will
            //be inconsistent with fw status.
            //TODO: Should we ask Launcher to terminate SVM?
            BHP_LOG_FATAL("BHP-open-vm connectSVM failed, ret=0x%x, heci_port=%d.\n", ret, heci_port);
            break;
        }
        //3: update bhp record
        connections[CONN_IDX_SVM].sdid = sdid;
        connections[CONN_IDX_SVM].conn_count = 1;
        ret = BH_SUCCESS;
    } while (0);

cleanup:
    mutex_exit(connections[CONN_IDX_SVM].lock);
    if (ret == BH_SUCCESS) *conn_idx = CONN_IDX_SVM;

    return ret;
#endif
#endif
}

BH_RET bh_do_closeVM(int conn_idx) {
    BH_RET ret = BH_SUCCESS;
#if (!BEIHAI_ENABLE_OEM_SIGNING_IOTG)
    unsigned int count = 0;

    //only close connected SVM
    if (conn_idx <= CONN_IDX_IVM || connections[conn_idx].handle == 0) return ret;

    mutex_enter(connections[conn_idx].lock);
    if (connections[conn_idx].conn_count == 0) {
        BHP_LOG_FATAL("[BHP]FATAL: svm conn_idx %d, closeVM called when conn_count is already 0 \n", conn_idx);
    }
    count = (--connections[conn_idx].conn_count);
    BHP_LOG_DEBUG("svm conn_idx %d, conn_count dec = %d \n", conn_idx, count);
    if (count == 0) {
        ret = bh_proxy_reset_svm(conn_idx);
    }
    mutex_exit(connections[conn_idx].lock);
#endif
    return ret;
}

BH_RET BHP_QueryTEEMetadata(unsigned char** metadata, unsigned int* length)
{
    char cmdbuf[CMDBUF_SIZE] = {0};
    bhp_command_header* h = (bhp_command_header*) cmdbuf;
    bh_response_record rr = {0};
    BH_RET ret = BH_SUCCESS;

    if (!is_bhp_inited())  return BPE_NOT_INIT;

    if (metadata == NULL || length == NULL) return BPE_INVALID_PARAMS;

    BHP_LOG_DEBUG ("Beihai QueryTEEMetadata 0x%x\n", &rr);
    h->id = BHP_CMD_QUERY_TEE_METADATA;

    ret = bh_send_message(CONN_IDX_IVM, (char*)h, sizeof(*h), NULL, 0, rrmap_add(CONN_IDX_IVM, &rr));
    if (ret == BH_SUCCESS)  ret = rr.code;

    BHP_LOG_DEBUG ("Beihai QueryTEEMetadata 0x%x ret 0x%x\n", &rr, rr.code);

    if (ret == BH_SUCCESS) {
        if (rr.buffer) {
            *length = rr.length;
            *metadata = (unsigned char*)BHMALLOC(rr.length);
            if (*metadata == NULL) {
                ret = BPE_OUT_OF_MEMORY;
            } else {
                memcpy(*metadata, rr.buffer, rr.length);
            }
        } else {
            ret = BPE_MESSAGE_ILLEGAL;
        }
    }

    if (rr.buffer) BHFREE(rr.buffer);
    return ret;
}

void BHP_Free(void * p)
{
    BHFREE(p);
}
