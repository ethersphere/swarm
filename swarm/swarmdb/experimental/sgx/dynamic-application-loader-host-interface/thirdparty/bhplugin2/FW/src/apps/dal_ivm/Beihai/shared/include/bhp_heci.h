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
 * @file  bhp_heci.h
 * @brief This file defines heci command and response format for Beihai Host Proxy (BHP) module.
 * @author
 * @version
 *
 */

#ifndef __BHP_HECI_H
#define __BHP_HECI_H
#ifdef __cplusplus
extern "C" {
#endif

#include "bh_shared_types.h"
#include "bh_shared_errcode.h"

typedef BH_I8 JHI_SESSION_ID[BH_GUID_LENGTH];

#define BHP_MSG_MAGIC_LENGTH (4)
#define BHP_MSG_CMD_MAGIC "\xff\xa3\xaa\x55"
#define BHP_MSG_RESPONSE_MAGIC "\xff\xa5\xaa\x55"

typedef enum {
    BHP_CMD_INIT = 0,
    BHP_CMD_DEINIT,
    BHP_CMD_VERIFY_JAVATA,
    BHP_CMD_DOWNLOAD_JAVATA,
    BHP_CMD_OPEN_JTASESSION,
    BHP_CMD_CLOSE_JTASESSION,
    BHP_CMD_FORCECLOSE_JTASESSION,
    BHP_CMD_SENDANDRECV,
    BHP_CMD_SENDANDRECV_INTERNAL,
    BHP_CMD_RUN_NATIVETA,
    BHP_CMD_STOP_NATIVETA,
    BHP_CMD_OPEN_SDSESSION,
    BHP_CMD_CLOSE_SDSESSION,
    BHP_CMD_INSTALL_SD,
    BHP_CMD_UNINSTALL_SD,
    BHP_CMD_INSTALL_JAVATA,
    BHP_CMD_UNINSTALL_JAVATA,
    BHP_CMD_INSTALL_NATIVETA,
    BHP_CMD_UNINSTALL_NATIVETA,
    BHP_CMD_LIST_SD,
    BHP_CMD_LIST_TA,
    BHP_CMD_RESET,
    BHP_CMD_LIST_TA_PROPERTIES,
    BHP_CMD_QUERY_TA_PROPERTY,
    BHP_CMD_LIST_JTA_SESSIONS,
    BHP_CMD_LIST_TA_PACKAGES,
    BHP_CMD_GET_ISD,
    BHP_CMD_GET_SD_BY_TA,
    BHP_CMD_LAUNCH_VM,
    BHP_CMD_CLOSE_VM,
    BHP_CMD_QUERY_NATIVETA_STATUS,
    BHP_CMD_QUERY_SD_STATUS,
    BHP_CMD_LIST_DOWNLOADED_NTA,
    BHP_CMD_UPDATE_SVL,
    BHP_CMD_CHECK_SVL_TA_BLOCKED_STATE,
    BHP_CMD_QUERY_TEE_METADATA,
    BHP_CMD_MAX
} bhp_command_id;

#ifdef _WIN32
#pragma warning (disable:4200)
#pragma pack(push, 4) //some structs below need 4-byte align for correct size in win32
#else
#pragma pack(4)
#endif

typedef struct {
    BH_U8 magic[BHP_MSG_MAGIC_LENGTH];
    BH_U32 length;
} transport_msg_header;

typedef struct {
    transport_msg_header h;
    BH_U64 seq;
    bhp_command_id id;
    BH_U8 pad[4];
    BH_I8 cmd[0];
} bhp_command_header;

typedef struct {
    transport_msg_header h;
    BH_U64 seq;
    BH_U64 addr;
    BH_RET code;
    BH_U8 pad[4];
    BH_I8 data[0];
} bhp_response_header;

typedef struct {
    BH_TAID appid;
    BH_I8 appblob[0];
} bhp_verify_javata_cmd;

typedef struct {
    BH_TAID appid;
    BH_I8 appblob[0];
} bhp_download_javata_cmd;

typedef struct {
    BH_TAID appid;
    BH_I8 buffer[0];
} bhp_open_jtasession_cmd;

typedef struct {
    BH_U64 ta_session_id;
} bhp_close_jtasession_cmd;

typedef struct {
    BH_U64 ta_session_id;
} bhp_forceclose_jtasession_cmd;

typedef struct {
    BH_U64 ta_session_id;
    BH_I32 command;
    BH_U32 outlen;
    BH_I8 buffer[0];
} bhp_snr_cmd;

typedef struct {
    BH_U64 ta_session_id;
    BH_I32 what;
    BH_I32 command;
    BH_U32 outlen;
    BH_I8 buffer[0];
} bhp_snr_internal_cmd;

typedef struct {
    BH_TAID appid;
    BH_I8 appblob[0];
} bhp_run_nativeta_cmd;

typedef struct {
    BH_TAID appid;
} bhp_stop_nativeta_cmd;

typedef struct {
    BH_SDID sdid;
} bhp_open_sdsession_cmd;

typedef struct {
    BH_U64 sd_session_id;
} bhp_close_sdsession_cmd;

typedef struct {
    BH_U64 sd_session_id;
    BH_I8 sd_install_pkg[0];
} bhp_install_sd_cmd;

typedef struct {
    BH_U64 sd_session_id;
    BH_I8 sd_uninstall_pkg[0];
} bhp_uninstall_sd_cmd;

typedef struct {
    BH_U64 sd_session_id;
    BH_I8 update_svl_pkg[0];
} bhp_update_svl_cmd;

typedef struct {
    BH_TAID taid;
} bhp_check_svl_ta_blocked_state_cmd;

typedef struct {
    BH_U64 sd_session_id;
    BH_I8 ta_install_pkg[0];
} bhp_install_javata_cmd;

typedef struct {
    BH_U64 sd_session_id;
    BH_I8 ta_uninstall_pkg[0];
} bhp_uninstall_javata_cmd;

typedef struct {
    BH_U64 sd_session_id;
    BH_I8 ta_install_pkg[0];
} bhp_install_nativeta_cmd;

typedef struct {
    BH_U64 sd_session_id;
    BH_I8 ta_uninstall_pkg[0];
} bhp_uninstall_nativeta_cmd;

typedef struct {
    BH_U64 sd_session_id;
} bhp_list_sd_cmd;

typedef struct {
    BH_U64 sd_session_id;
    BH_SDID sdid;
} bhp_list_ta_cmd;

typedef struct {
    BH_TAID appid;
} bhp_list_ta_properties_cmd;

typedef struct {
    BH_TAID appid;
    BH_I8 buffer[0];
} bhp_query_ta_property_cmd;

typedef struct {
    BH_TAID appid;
} bhp_list_ta_sessions_cmd;

typedef struct {
    BH_SDID sdid;
} bhp_launch_vm_cmd;

typedef struct {
    BH_SDID sdid;
} bhp_close_vm_cmd;

typedef struct {
    BH_TAID taid;
} bhp_query_nativeta_status_cmd;

typedef struct {
    BH_SDID sdid;
} bhp_query_sd_status_cmd;

typedef struct {
    BH_SDID sdid;
} bhp_list_downloaded_nta_cmd;

typedef struct {
    BH_U32 count;
    BH_TAID nta_ids[0];
} bhp_list_downloaded_nta_response;

typedef struct {
    BH_I32 heci_port;
} bhp_launch_vm_response;

typedef struct {
    BH_U32 count; //count of svm heci ports
    BH_I32 vm_heci_port_list[0];
} bhp_reset_launcher_response;

typedef struct {
    BH_TAID taid;
} bhp_get_sd_by_ta_cmd;

typedef struct {
    BH_SDID sdid;
} bhp_get_sd_by_ta_response;

typedef struct {
    BH_SDID sdid;
} bhp_get_isd_response;

typedef struct {
    // field response comes from java BIG endian
    BH_I32 response;
    BH_I8 buffer[0];
} bhp_snr_response;

typedef struct {
    // field response comes from java BIG endian
    BH_I32 response;
    BH_U32 request_length;
} bhp_snr_bof_response;

typedef struct {
    BH_U32 count;
    BH_U64 addr[0];
} bhp_list_ta_sessions_response;

typedef struct {
    BH_U32 count;
    BH_TAID appIds[0];
} bhp_list_ta_packages_response;

typedef struct {
    BH_U32 count;
    BH_SDID sd_ids[0];
} bhp_list_sd_response;

typedef struct {
    BH_U32 count;
    BH_TAID ta_ids[0];
} bhp_list_ta_response;

typedef struct {
    // field response comes from java BIG endian
    BH_I32 response;
    JHI_SESSION_ID session_id;
    BH_I8 buffer[0];
} bhp_spooler_snr_response;

typedef struct {
    // field response comes from java BIG endian
    BH_I32 response;
    // field request_length comes from java BIG endian
    BH_U32 request_length;
} bhp_spooler_bof_response;

#ifdef _WIN32
#pragma pack(pop)  //restore original packing
#else
#pragma pack()
#endif

// HECI port number list. It must match with the values defined in BeihaiHAL.h.
enum {
  BH_LAUNCHER_HECI_PORT = 10000,
  BH_SDM_HECI_PORT = 10001,
  BH_IVM_HECI_PORT = 10002,
  BH_SVM_HECI_PORT = 10003
};

#ifdef __cplusplus
}
#endif

#endif
