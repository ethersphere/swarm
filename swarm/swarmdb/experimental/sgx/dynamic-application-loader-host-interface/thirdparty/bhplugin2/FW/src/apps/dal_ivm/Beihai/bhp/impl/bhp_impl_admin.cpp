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
 * @file  bhp_impl_admin.cpp
 * @brief This file implements Beihai Host Proxy (BHP) module TA management API.
 * @author
 * @version
 *
 */
#include <dbg.h>
#include "bhp_exp.h"
#include "bhp_heci.h"
#include "bh_acp_exp.h"
#include "bh_acp_util.h"
#include "bhp_platform.h"
#include "bhp_impl.h"

BH_RET BHP_OpenSDSession(const char* SD_ID, SD_SESSION_HANDLE* pSession)
{
    char cmdbuf[CMDBUF_SIZE] = {0};
    bhp_command_header* h = (bhp_command_header*) cmdbuf;
    bhp_open_sdsession_cmd* cmd = (bhp_open_sdsession_cmd*) h->cmd;
    bh_response_record* rr = NULL;
    BH_U64 seq;
    BH_RET ret = BH_SUCCESS;

    if (!is_bhp_inited()) return BPE_NOT_INIT;
    if (SD_ID == NULL || pSession == NULL) return BPE_INVALID_PARAMS;

    if (!string_to_uuid(SD_ID,(char*)&cmd->sdid)) return BPE_INVALID_PARAMS;

    rr = (bh_response_record*) BHMALLOC (sizeof (bh_response_record));
    if (!rr) {
        return BPE_OUT_OF_MEMORY;
    }
    memset (rr, 0, sizeof(bh_response_record));
    
    rr->session_lock = bh_create_mutex();
    if (!rr->session_lock) {
        if (rr) BHFREE(rr);
        return BPE_OUT_OF_RESOURCE;
    }
    rr->count = 1;
    rr->is_session = 1;
    seq = rrmap_add(CONN_IDX_SDM, rr);

    h->id = BHP_CMD_OPEN_SDSESSION;
    
    BHP_LOG_DEBUG("Beihai BHP_OpenSDSession %x\n", rr);

    ret = bh_send_message(CONN_IDX_SDM, (char*)h, sizeof (*h)+sizeof(*cmd), NULL, 0, seq);
    if( ret == BH_SUCCESS) ret = rr->code;

    BHP_LOG_DEBUG("Beihai BHP_OpenSDSession %x ret %x\n", rr, rr->code);

    if (rr->buffer) {
        BHFREE(rr->buffer);
        rr->buffer = NULL;
    }
    if (ret == BH_SUCCESS) {
        *pSession = (SD_SESSION_HANDLE) (uintptr_t)seq;
        session_exit(CONN_IDX_SDM, rr, seq, 0);
    } else {
        session_close(CONN_IDX_SDM, rr, seq, 0);
    }

    return ret;
}


BH_RET BHP_CloseSDSession(const SD_SESSION_HANDLE handle)
{
    char cmdbuf[CMDBUF_SIZE] = {0};
    bhp_command_header* h = (bhp_command_header*) cmdbuf;
    bhp_close_sdsession_cmd* cmd = (bhp_close_sdsession_cmd*) h->cmd;
    bh_response_record* rr = NULL;
    BH_U64 seq = (BH_U64)(uintptr_t)handle;
    BH_RET ret = BH_SUCCESS;
    int conn_idx = CONN_IDX_SDM;

    if (!is_bhp_inited()) return BPE_NOT_INIT;

    rr = session_enter(conn_idx, seq, 1);
    if(!rr) {
        return BPE_INVALID_PARAMS;
    }

    rr->buffer = NULL;

    h->id = BHP_CMD_CLOSE_SDSESSION;
    cmd->sd_session_id = rr->addr;

    BHP_LOG_DEBUG("Beihai CloseSDSession %x\n", rr);

    ret = bh_send_message(conn_idx, (char*)h, sizeof(*h) + sizeof (*cmd), NULL, 0, seq);
    if (ret == BH_SUCCESS)	ret = rr->code;

    BHP_LOG_DEBUG ("Beihai CloseSDSession %x ret %x\n", rr, rr->code);
    if (rr->killed) ret = BHE_UNCAUGHT_EXCEPTION;

    session_close(conn_idx, rr, seq, 1);

    return ret;
}

static BH_RET bh_get_cmdtype_by_cmd_pkg(const char* cmd_pkg, unsigned int pkg_len, int* cmd_type){
    if (cmd_pkg == NULL || pkg_len == 0) return BPE_INVALID_PARAMS;
    if (cmd_type == NULL) return BPE_INVALID_PARAMS;
    return ACP_get_cmd_id(cmd_pkg, pkg_len, cmd_type);
}

static BH_RET bh_get_tainfo_by_cmd_pkg_installjta(const char* cmd_pkg, unsigned int pkg_len, BH_TAID* ta_id, unsigned int* ta_pkg_offset){
    BH_RET ret = BPE_INVALID_PARAMS;
    ACInsJTAPackExt pack;

    if (cmd_pkg == NULL || pkg_len == 0) return BPE_INVALID_PARAMS;
    if (ta_pkg_offset == NULL) return BPE_INVALID_PARAMS;
    ret = ACP_pload_ins_jta(cmd_pkg, pkg_len, &pack);
    if (ret == BH_SUCCESS) {
        *ta_id = pack.cmd_pack.head->ta_id;
        *ta_pkg_offset = (unsigned int)(pack.ta_pack - cmd_pkg);
    }
    return ret;
}

static BH_RET bh_get_tainfo_by_cmd_pkg_uninstalljta(const char* cmd_pkg, unsigned int pkg_len, BH_TAID* ta_id)
{
    BH_RET ret = BPE_INVALID_PARAMS;
    ACUnsTAPackExt pack = {0};

    if (cmd_pkg == NULL || pkg_len == 0) return BPE_INVALID_PARAMS;
    ret = ACP_pload_uns_jta(cmd_pkg, pkg_len, &pack);
    if (ret == BH_SUCCESS) {
        *ta_id = *pack.cmd_pack.p_taid;
    }
    return ret;
}

static BH_RET bh_do_uninstall_jta(const SD_SESSION_HANDLE handle, const char* cmd_pkg, unsigned int pkg_len) 
{
    char cmdbuf[CMDBUF_SIZE] = {0};
    bhp_command_header* h = (bhp_command_header*) cmdbuf;
    bhp_uninstall_javata_cmd* cmd = (bhp_uninstall_javata_cmd*) h->cmd;
    bh_response_record* rr = NULL;
    BH_U64 seq = (BH_U64)(uintptr_t)handle;
    BH_RET ret = BH_SUCCESS;
    BH_TAID ta_id = {0};
   
    if (cmd_pkg == NULL || pkg_len ==0) return BPE_INVALID_PARAMS;

    ret = bh_get_tainfo_by_cmd_pkg_uninstalljta(cmd_pkg, pkg_len, &ta_id);
    if (ret != BH_SUCCESS) return ret;

    {
    //Check with VM whether the TA has live session or not
    char ta_id_string[BH_GUID_LENGTH *2 +1] = {0};
    unsigned int session_count = 0;
    JAVATA_SESSION_HANDLE* handles = NULL;

    uuid_to_string((char*)&ta_id, ta_id_string);
    ret = BHP_ListTASessions(ta_id_string, &session_count, &handles);
    if (handles) BHP_Free(handles);
    if (ret == BH_SUCCESS && session_count > 0) return BHE_EXIST_LIVE_SESSION;
    }

    //send uninstall cmd to SDM
    rr = session_enter(CONN_IDX_SDM, seq, 1);
    if (!rr) {
        return BPE_INVALID_PARAMS;
    }

    rr->buffer = NULL;
    h->id = BHP_CMD_UNINSTALL_JAVATA;

    cmd->sd_session_id = rr->addr;

    BHP_LOG_DEBUG("Beihai bh_do_uninstall_jta %x\n", rr);

    ret = bh_send_message(CONN_IDX_SDM, (char*)h, sizeof (*h) + sizeof (*cmd), cmd_pkg, pkg_len, seq);
    if (ret == BH_SUCCESS) ret = rr->code;

    BHP_LOG_DEBUG("Beihai bh_do_uninstall_jta %x ret %x\n", rr, rr->code);

    if (rr->killed) {
        ret = BHE_UNCAUGHT_EXCEPTION;
    }

    if (rr->buffer) {
        BHFREE (rr->buffer);
        rr->buffer = NULL;
    }

    session_exit(CONN_IDX_SDM, rr, seq, 1);

    return ret;
}

/*
 * 
 * This function is inside sd-session-lock
 */
static BH_RET bh_proxy_installjavata(SD_SESSION_HANDLE handle, bh_response_record* rr, const char* cmd_pkg, unsigned int cmd_pkg_len)
{
    char cmdbuf[CMDBUF_SIZE] = {0};
    bhp_command_header* h = (bhp_command_header*) cmdbuf;
    bhp_install_javata_cmd* cmd = (bhp_install_javata_cmd*) h->cmd;
    BH_RET ret = BH_SUCCESS;
    BH_U64 seq = (BH_U64)(uintptr_t)handle;

    if (cmd_pkg == NULL || cmd_pkg_len == 0) return BPE_INVALID_PARAMS;
    if (rr == NULL) return BPE_INVALID_PARAMS;

    rr->buffer = NULL;
    h->id = BHP_CMD_INSTALL_JAVATA;

    cmd->sd_session_id = rr->addr;

    BHP_LOG_DEBUG("Beihai bh_proxy_installjavata %x\n", rr);

    ret = bh_send_message(CONN_IDX_SDM, (char*)h, sizeof (*h) + sizeof (*cmd), cmd_pkg, cmd_pkg_len, seq);
    if (ret == BH_SUCCESS) ret = rr->code;

    BHP_LOG_DEBUG("Beihai bh_proxy_installjavata %x ret %x\n", rr, rr->code);

    if (rr->killed) {
        ret = BHE_UNCAUGHT_EXCEPTION;
    }
    if (rr->buffer) {
        BHFREE(rr->buffer);
        rr->buffer = NULL;
    }
    return ret;
}

/*
 * 
 * This function is inside sd-session-lock
 */
static BH_RET bh_proxy_verifyjavata(int conn_idx, BH_TAID ta_id, const char* ta_pkg, unsigned int ta_pkg_len)
{
    char cmdbuf[CMDBUF_SIZE] = {0};
    bhp_command_header* h = (bhp_command_header*) cmdbuf;
    bhp_verify_javata_cmd *cmd = (bhp_verify_javata_cmd*)h->cmd;
    bh_response_record rr = {0};
    BH_RET ret = BH_SUCCESS;

    if (ta_pkg == NULL || ta_pkg_len == 0) return BPE_INVALID_PARAMS;

    h->id = BHP_CMD_VERIFY_JAVATA;
    cmd->appid = ta_id;

    BHP_LOG_DEBUG ("Beihai bh_proxy_verifyjavata %x\n", &rr);

    ret = bh_send_message(conn_idx, (char*)h, sizeof(*h) + sizeof (*cmd), ta_pkg, ta_pkg_len, rrmap_add(conn_idx, &rr));
    if (ret == BH_SUCCESS) ret = rr.code;

    BHP_LOG_DEBUG ("Beihai bh_proxy_verifyjavata %x ret %x\n", &rr, rr.code);

    if (rr.buffer) BHFREE(rr.buffer);

    return ret;
}

static BH_RET bh_do_install_jta(const SD_SESSION_HANDLE handle, const char* cmd_pkg, unsigned int pkg_len) 
{
    bh_response_record* rr = NULL;
    BH_RET ret = BH_SUCCESS;
    unsigned int ta_pkg_offset = 0;
    const char* ta_pkg = NULL;
    BH_TAID ta_id = {0};
    BH_U64 seq = (BH_U64)(uintptr_t)handle;

    if (cmd_pkg == NULL || pkg_len == 0) return BPE_INVALID_PARAMS;

    if (bh_get_tainfo_by_cmd_pkg_installjta(cmd_pkg,pkg_len,&ta_id,&ta_pkg_offset) != BH_SUCCESS) return BPE_INVALID_PARAMS;
    ta_pkg = (cmd_pkg + ta_pkg_offset);

    rr = session_enter(CONN_IDX_SDM, seq, 1);
    if (!rr) return BPE_INVALID_PARAMS;

    //first step: send installjta cmd to sdm
    ret = bh_proxy_installjavata(handle, rr, cmd_pkg, ta_pkg_offset);
    if (ret != BH_SUCCESS) {
        goto cleanup;
    }

    //second step: verifyjavata cmd to IVM
    ret = bh_proxy_verifyjavata(CONN_IDX_IVM, ta_id, ta_pkg, pkg_len - ta_pkg_offset);

cleanup:
    session_exit(CONN_IDX_SDM, rr, seq, 1);

    return ret;
}

#if (BEIHAI_ENABLE_SVM || BEIHAI_ENABLE_OEM_SIGNING_IOTG)
static BH_RET bh_do_install_sd(const SD_SESSION_HANDLE handle, const char* cmd_pkg, unsigned int pkg_len) 
{
    char cmdbuf[CMDBUF_SIZE] = {0};
    bhp_command_header* h = (bhp_command_header*) cmdbuf;
    bhp_install_sd_cmd *cmd = (bhp_install_sd_cmd*)h->cmd;
    bh_response_record* rr = NULL;
    BH_RET ret = BH_SUCCESS;
    BH_U64 seq = (BH_U64)(uintptr_t)handle;

    if (cmd_pkg == NULL || pkg_len == 0) return BPE_INVALID_PARAMS;

    rr = session_enter(CONN_IDX_SDM, seq, 1);
    if (!rr) return BPE_INVALID_PARAMS;

    //send installsd cmd to SDM
    h->id = BHP_CMD_INSTALL_SD;
    cmd->sd_session_id = seq;
    rr->buffer = NULL;

    BHP_LOG_DEBUG ("Beihai bh_proxy_installsd %x\n", &rr);

    ret = bh_send_message(CONN_IDX_SDM, (char*)h, sizeof(*h) + sizeof (*cmd), cmd_pkg, pkg_len, seq);
    if (ret == BH_SUCCESS) ret = rr->code;

    BHP_LOG_DEBUG ("Beihai bh_proxy_installsd %x ret %x\n", rr, rr->code);

    if (rr->buffer) {
        BHFREE(rr->buffer);
        rr->buffer = NULL;
    }

    session_exit(CONN_IDX_SDM, rr, seq, 1);

    return ret;
}

static BH_RET bh_get_sdinfo_by_cmd_pkg_uninstallsd(const char* cmd_pkg, unsigned int pkg_len, BH_SDID* sd_id)
{
    BH_RET ret = BPE_INVALID_PARAMS;
    ACUnsSDPackExt pack = {0};

    if (cmd_pkg == NULL || pkg_len == 0) return BPE_INVALID_PARAMS;
    ret = ACP_pload_uns_sd(cmd_pkg, pkg_len, &pack);
    if (ret == BH_SUCCESS) {
        *sd_id = *pack.cmd_pack.p_sdid;
    }
    return ret;
}

static BH_RET bh_do_uninstall_sd(const SD_SESSION_HANDLE handle, const char* cmd_pkg, unsigned int pkg_len)
{
    char cmdbuf[CMDBUF_SIZE] = {0};
    bhp_command_header* h = (bhp_command_header*) cmdbuf;
    bhp_uninstall_sd_cmd *cmd = (bhp_uninstall_sd_cmd*)h->cmd;
    bh_response_record* rr = NULL;
    BH_RET ret = BH_SUCCESS;
    BH_U64 seq = (BH_U64)(uintptr_t)handle;
    BH_SDID sd_id = {0};

    if (cmd_pkg == NULL || pkg_len == 0) return BPE_INVALID_PARAMS;
    if (bh_get_sdinfo_by_cmd_pkg_uninstallsd(cmd_pkg,pkg_len,&sd_id) != BH_SUCCESS) return BPE_INVALID_PARAMS;

#if BEIHAI_ENABLE_SVM
    // Step 1: ask Launcher to query sd running status
    if (bh_proxy_query_sd_status(sd_id) == BH_SUCCESS) {
        //the sd's svm or nta is running, so uninstalling fails.
        return BHE_EXIST_LIVE_SESSION;
    }
#endif

    // Step 2: send UninstallSD cmd to SDM
    rr = session_enter(CONN_IDX_SDM, seq, 1);
    if (!rr) return BPE_INVALID_PARAMS;

    h->id = BHP_CMD_UNINSTALL_SD;
    cmd->sd_session_id = seq;
    rr->buffer = NULL;

    BHP_LOG_DEBUG ("Beihai bh_proxy_uninstallsd %x\n", rr);

    ret = bh_send_message(CONN_IDX_SDM, (char*)h, sizeof(*h) + sizeof (*cmd), cmd_pkg, pkg_len, seq);
    if (ret == BH_SUCCESS) ret = rr->code;

    BHP_LOG_DEBUG ("Beihai bh_proxy_uninstallsd %x ret %x\n", rr, rr->code);

    if (rr->buffer) {
        BHFREE(rr->buffer);
        rr->buffer = NULL;
    }

    session_exit(CONN_IDX_SDM, rr, seq, 1);

    return ret;
}
#endif

#if BEIHAI_ENABLE_NATIVETA
static BH_RET bh_get_tainfo_by_cmd_pkg_installnta(const char* cmd_pkg, unsigned int pkg_len, BH_TAID* ta_id, unsigned int* ta_pkg_offset){
    BH_RET ret = BPE_INVALID_PARAMS;
    ACInsNTAPackExt pack = {0};

    if (cmd_pkg == NULL || pkg_len == 0) return BPE_INVALID_PARAMS;
    if (ta_pkg_offset == NULL) return BPE_INVALID_PARAMS;
    ret = ACP_pload_ins_nta(cmd_pkg, pkg_len, &pack);
    if (ret == BH_SUCCESS) {
        *ta_id = pack.cmd_pack.head->ta_id;
        *ta_pkg_offset = (unsigned int)pack.ta_pack - (unsigned int)cmd_pkg;
    }
    return ret;
}

static BH_RET bh_do_install_nta(const SD_SESSION_HANDLE handle, const char* cmd_pkg, unsigned int pkg_len)
{
    char cmdbuf[CMDBUF_SIZE] = {0};
    bhp_command_header* h = (bhp_command_header*) cmdbuf;
    bhp_install_nativeta_cmd* cmd = (bhp_install_nativeta_cmd*) h->cmd;
    bh_response_record* rr = NULL;
    BH_RET ret = BH_SUCCESS;
    BH_U64 seq = (BH_U64)(BH_U32)handle;
    unsigned int ta_pkg_offset = 0;
    BH_TAID ta_id = {0};

    if (cmd_pkg == NULL || pkg_len == 0) return BPE_INVALID_PARAMS;
    if (bh_get_tainfo_by_cmd_pkg_installnta(cmd_pkg,pkg_len,&ta_id,&ta_pkg_offset) != BH_SUCCESS) return BPE_INVALID_PARAMS;

    rr = session_enter(CONN_IDX_SDM, seq, 1);
    if (!rr) return BPE_INVALID_PARAMS;

    rr->buffer = NULL;
    h->id = BHP_CMD_INSTALL_NATIVETA;

    cmd->sd_session_id = seq;

    BHP_LOG_DEBUG("Beihai bh_proxy_install_nativeta %x\n", rr);
    //excluding NativeTA package at install time to save lots of RAM requirement for SDM process
    ret = bh_send_message(CONN_IDX_SDM, (char*)h, sizeof (*h) + sizeof (*cmd), cmd_pkg, ta_pkg_offset, seq);
    if (ret == BH_SUCCESS) ret = rr->code;

    BHP_LOG_DEBUG("Beihai bh_proxy_install_nativeta %x ret %x\n", rr, rr->code);

    if (rr->buffer) {
        BHFREE(rr->buffer);
        rr->buffer = NULL;
    }

    session_exit(CONN_IDX_SDM, rr, seq, 1);

    return ret;
}

static BH_RET bh_get_tainfo_by_cmd_pkg_uninstallnta(const char* cmd_pkg, unsigned int pkg_len, BH_TAID* ta_id)
{
    BH_RET ret = BPE_INVALID_PARAMS;
    ACUnsTAPackExt pack = {0};

    if (cmd_pkg == NULL || pkg_len == 0) return BPE_INVALID_PARAMS;
    ret = ACP_pload_uns_nta(cmd_pkg, pkg_len, &pack);
    if (ret == BH_SUCCESS) {
        *ta_id = *pack.cmd_pack.p_taid;
    }
    return ret;
}

static BH_RET bh_proxy_query_nta_status(BH_TAID ta_id)
{
    char cmdbuf[CMDBUF_SIZE] = {0};
    bhp_command_header* h = (bhp_command_header*) cmdbuf;
    bhp_query_nativeta_status_cmd *cmd = (bhp_query_nativeta_status_cmd*)h->cmd;
    bh_response_record rr = {0};
    BH_RET ret = BH_SUCCESS;

    h->id = BHP_CMD_QUERY_NATIVETA_STATUS;
    cmd->taid = ta_id;

    BHP_LOG_DEBUG ("Beihai bh_proxy_query_nta_status 0x%x\n", &rr);

    ret = bh_send_message(CONN_IDX_LAUNCHER, (char*)h, sizeof(*h) + sizeof (*cmd), NULL, 0, rrmap_add(CONN_IDX_LAUNCHER, &rr));
    if (ret == BH_SUCCESS) ret = rr.code;

    BHP_LOG_DEBUG ("Beihai bh_proxy_query_nta_status 0x%x ret %x\n", &rr, rr.code);

    if (rr.buffer) BHFREE(rr.buffer);

    return ret;
}

static BH_RET bh_do_uninstall_nta(const SD_SESSION_HANDLE handle, const char* cmd_pkg, unsigned int pkg_len)
{
    char cmdbuf[CMDBUF_SIZE] = {0};
    bhp_command_header* h = (bhp_command_header*) cmdbuf;
    bhp_uninstall_nativeta_cmd* cmd = (bhp_uninstall_nativeta_cmd*) h->cmd;
    bh_response_record* rr = NULL;
    BH_RET ret = BH_SUCCESS;
    BH_TAID ta_id = {0};
    BH_U64 seq = (BH_U64)(BH_U32)handle;

    if (cmd_pkg == NULL || pkg_len == 0) return BPE_INVALID_PARAMS;

    if (bh_get_tainfo_by_cmd_pkg_uninstallnta(cmd_pkg,pkg_len,&ta_id) != BH_SUCCESS) return BPE_INVALID_PARAMS;

    //step1: ask Launcher to query nativeta running status
    if (bh_proxy_query_nta_status(ta_id) == BH_SUCCESS) {
        //the nta is running, so uninstalling fails.
        return BHE_EXIST_LIVE_SESSION;
    }

    //step2: send UninstallNTA cmd to SDM
    rr = session_enter(CONN_IDX_SDM, seq, 1);
    if (!rr) return BPE_INVALID_PARAMS;

    rr->buffer = NULL;
    h->id = BHP_CMD_UNINSTALL_NATIVETA;

    cmd->sd_session_id = seq;

    BHP_LOG_DEBUG("Beihai bh_proxy_uninstall_nativeta %x\n", rr);

    ret = bh_send_message(CONN_IDX_SDM, (char*)h, sizeof (*h) + sizeof (*cmd), cmd_pkg, pkg_len, seq);
    if (ret == BH_SUCCESS) ret = rr->code;

    BHP_LOG_DEBUG("Beihai bh_proxy_uninstall_nativeta %x ret %x\n", rr, rr->code);

    if (rr->buffer) {
        BHFREE(rr->buffer);
        rr->buffer = NULL;
    }

    session_exit(CONN_IDX_SDM, rr, seq, 1);

    return ret;
}
#endif

static BH_RET bh_do_update_svl(const SD_SESSION_HANDLE handle, const char* cmd_pkg, unsigned int pkg_len)
{
    char cmdbuf[CMDBUF_SIZE] = {0};
    bhp_command_header* h = (bhp_command_header*) cmdbuf;
    bhp_update_svl_cmd *cmd = (bhp_update_svl_cmd*)h->cmd;
    bh_response_record* rr = NULL;
    BH_RET ret = BH_SUCCESS;
    BH_U64 seq = (BH_U64)(uintptr_t)handle;

    if (cmd_pkg == NULL || pkg_len == 0) return BPE_INVALID_PARAMS;

    rr = session_enter(CONN_IDX_SDM, seq, 1);
    if (!rr) return BPE_INVALID_PARAMS;

    //send updatesvl cmd to SDM
    h->id = BHP_CMD_UPDATE_SVL;
    cmd->sd_session_id = seq;
    rr->buffer = NULL;

    BHP_LOG_DEBUG ("Beihai bh_do_update_svl %x\n", &rr);

    ret = bh_send_message(CONN_IDX_SDM, (char*)h, sizeof(*h) + sizeof (*cmd), cmd_pkg, pkg_len, seq);
    if (ret == BH_SUCCESS) ret = rr->code;

    BHP_LOG_DEBUG ("Beihai bh_do_update_svl %x ret %x\n", rr, rr->code);

    if (rr->buffer) {
        BHFREE(rr->buffer);
        rr->buffer = NULL;
    }

    session_exit(CONN_IDX_SDM, rr, seq, 1);

    return ret;
}

BH_RET BHP_SendAdminCmdPkg(const SD_SESSION_HANDLE handle, const char* cmd_pkg, unsigned int pkg_len)
{
    BH_RET ret = BPE_INVALID_PARAMS;
    int cmd_type = 0;

    if (!is_bhp_inited()) return BPE_NOT_INIT;
    if (cmd_pkg == NULL || pkg_len == 0) return BPE_INVALID_PARAMS;

    if (bh_get_cmdtype_by_cmd_pkg(cmd_pkg, pkg_len, &cmd_type) != BH_SUCCESS) return BPE_INVALID_PARAMS;

    switch (cmd_type){
#if (BEIHAI_ENABLE_SVM || BEIHAI_ENABLE_OEM_SIGNING_IOTG)
        case AC_INSTALL_SD:
			TRACE0("The command is AC_INSTALL_SD");
            ret = bh_do_install_sd(handle, cmd_pkg, pkg_len);
            break;
        case AC_UNINSTALL_SD:
			TRACE0("The command is AC_UNINSTALL_SD");
            ret = bh_do_uninstall_sd(handle, cmd_pkg, pkg_len);
            break;
#endif
#if BEIHAI_ENABLE_NATIVETA
        case AC_INSTALL_NTA:
        	TRACE0("The command is AC_INSTALL_NTA");
            ret = bh_do_install_nta(handle, cmd_pkg, pkg_len);
            break;
        case AC_UNINSTALL_NTA:
        	TRACE0("The command is AC_UNINSTALL_NTA");
            ret = bh_do_uninstall_nta(handle, cmd_pkg, pkg_len);
            break;
#endif
        case AC_INSTALL_JTA:
			TRACE0("The command is AC_INSTALL_JTA");
            ret = bh_do_install_jta(handle, cmd_pkg, pkg_len);
            break;
        case AC_UNINSTALL_JTA:
			TRACE0("The command is AC_UNINSTALL_JTA");
            ret = bh_do_uninstall_jta(handle, cmd_pkg, pkg_len);
            break;
        case AC_UPDATE_SVL:
			TRACE0("The command is AC_UPDATE_SVL");
            ret = bh_do_update_svl(handle, cmd_pkg, pkg_len);
            break;
        default:
            ret = BPE_INVALID_PARAMS;
            break;
    }

    return ret;
}

BH_RET BHP_ListInstalledSDs (const SD_SESSION_HANDLE handle, unsigned int* count, char*** sdIdStrs)
{
    char cmdbuf[CMDBUF_SIZE] = {0};
    bhp_command_header* h = (bhp_command_header*) cmdbuf;
    bhp_list_sd_cmd *cmd = (bhp_list_sd_cmd*)h->cmd;
    bh_response_record* rr = NULL;
    BH_RET ret = BH_SUCCESS;
    BH_U64 seq = (BH_U64)(uintptr_t)handle;

    if (!is_bhp_inited()) return BPE_NOT_INIT;
    if (count == NULL || sdIdStrs == NULL) return BPE_INVALID_PARAMS;

    rr = session_enter(CONN_IDX_SDM, seq, 1);
    if (!rr) return BPE_INVALID_PARAMS;

    //send listSD cmd to SDM
    h->id = BHP_CMD_LIST_SD;
    cmd->sd_session_id = seq;
    rr->buffer = NULL;

    BHP_LOG_DEBUG ("Beihai List SD %x\n", rr);

    ret = bh_send_message(CONN_IDX_SDM, (char*)h, sizeof(*h) + sizeof (*cmd), NULL, 0, seq);
    if (ret == BH_SUCCESS) ret = rr->code;

    BHP_LOG_DEBUG ("Beihai List SD %x ret %x\n", rr, rr->code);

    *count = 0;
    *sdIdStrs = NULL;
    char** outbuf = NULL;
    int total_count = 0;
    do {
        if (ret != BH_SUCCESS)  break;
        if (rr->buffer == NULL) {
            ret = BPE_MESSAGE_ILLEGAL;
            break;
        }

        bhp_list_sd_response* resp = (bhp_list_sd_response*) rr->buffer;
        total_count = resp->count;
        if (total_count == 0)  break;

        if (rr->length != sizeof(BH_SDID) * total_count + sizeof(bhp_list_sd_response)) {
            ret = BPE_MESSAGE_ILLEGAL;
            break;
        }
        outbuf = (char**) BHMALLOC(sizeof(char*) * (total_count+1));
        if (!outbuf) {
            ret = BPE_OUT_OF_MEMORY;
            break;
        }
        memset (outbuf, 0, sizeof(char*) * (total_count+1));

        for (int i=0; i< total_count; i++) {
            outbuf[i] = (char*) BHMALLOC(BH_GUID_LENGTH *2 +1);
            if (outbuf[i] == NULL) {
                ret = BPE_OUT_OF_MEMORY;
                break;
            }
            uuid_to_string((char*)&resp->sd_ids[i], outbuf[i]);
        }
        if (ret != BH_SUCCESS) break;

        *count  = total_count;
        *sdIdStrs = outbuf;
    } while(0);

    if (ret != BH_SUCCESS) {
        for (int i=0; i<total_count; i++) {
            if (outbuf && outbuf[i]) BHFREE(outbuf[i]);
        }
        if (outbuf) BHFREE(outbuf);
    }
    if (rr->buffer) {
        BHFREE(rr->buffer);
        rr->buffer = NULL;
    }

    session_exit(CONN_IDX_SDM, rr, seq, 1);

    return ret;
}

BH_RET BHP_ListInstalledTAs (const SD_SESSION_HANDLE handle, const char* SD_ID, unsigned int * count, char*** appIdStrs)
{
    char cmdbuf[CMDBUF_SIZE] = {0};
    bhp_command_header* h = (bhp_command_header*) cmdbuf;
    bhp_list_ta_cmd *cmd = (bhp_list_ta_cmd*)h->cmd;
    bh_response_record* rr = NULL;
    BH_RET ret = BH_SUCCESS;
    BH_U64 seq = (BH_U64)(uintptr_t)handle;
    BH_SDID sdid = {0};

    if (!is_bhp_inited()) return BPE_NOT_INIT;
    if (SD_ID == NULL || count == NULL || appIdStrs == NULL) return BPE_INVALID_PARAMS;
    if (!string_to_uuid(SD_ID, (char*)&sdid)) return BPE_INVALID_PARAMS;

    rr = session_enter(CONN_IDX_SDM, seq, 1);
    if (!rr) return BPE_INVALID_PARAMS;

    //send listTA cmd to SDM
    h->id = BHP_CMD_LIST_TA;
    cmd->sd_session_id = seq;
    cmd->sdid = sdid;
    rr->buffer = NULL;

    BHP_LOG_DEBUG ("Beihai List TA %x\n", rr);

    ret = bh_send_message(CONN_IDX_SDM, (char*)h, sizeof(*h) + sizeof (*cmd), NULL, 0, seq);
    if (ret == BH_SUCCESS) ret = rr->code;

    BHP_LOG_DEBUG ("Beihai List TA %x ret %x\n", rr, rr->code);

    *count = 0;
    *appIdStrs = NULL;
    char** outbuf = NULL;
    int total_count = 0;
    do {
        if (ret != BH_SUCCESS)  break;
        if (rr->buffer == NULL) {
            ret = BPE_MESSAGE_ILLEGAL;
            break;
        }

        bhp_list_ta_response* resp = (bhp_list_ta_response*) rr->buffer;
        total_count = resp->count;
        if (total_count == 0)  break;

        if (rr->length != sizeof(BH_TAID) * total_count + sizeof(bhp_list_ta_response)) {
            ret = BPE_MESSAGE_ILLEGAL;
            break;
        }
        outbuf = (char**) BHMALLOC(sizeof(char*) * (total_count+1));
        if (!outbuf) {
            ret = BPE_OUT_OF_MEMORY;
            break;
        }
        memset (outbuf, 0, sizeof(char*) * (total_count+1));

        for (int i=0; i< total_count; i++) {
            outbuf[i] = (char*) BHMALLOC(BH_GUID_LENGTH *2 +1);
            if (outbuf[i] == NULL) {
                ret = BPE_OUT_OF_MEMORY;
                break;
            }
            uuid_to_string((char*)&resp->ta_ids[i], outbuf[i]);
        }
        if (ret != BH_SUCCESS) break;

        *count  = total_count;
        *appIdStrs = outbuf;
    } while(0);

    if (ret != BH_SUCCESS) {
        for (int i=0; i<total_count; i++) {
            if (outbuf && outbuf[i]) BHFREE(outbuf[i]);
        }
        if (outbuf) BHFREE(outbuf);
    }
    if (rr->buffer) {
        BHFREE(rr->buffer);
        rr->buffer = NULL;
    }

    session_exit(CONN_IDX_SDM, rr, seq, 1);

    return ret;
}
