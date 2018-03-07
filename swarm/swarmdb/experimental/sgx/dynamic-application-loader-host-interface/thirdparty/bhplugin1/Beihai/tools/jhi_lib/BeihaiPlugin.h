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

#ifndef BEIHAIPLUGIN_H
#define BEIHAIPLUGIN_H
#ifdef CPLUSPLUS
extern "C" {
#endif

#ifdef _WIN32
#include <Windows.h>
typedef HANDLE BH_MUTEX;
typedef HANDLE BH_EVENT;
typedef HANDLE BH_THREAD;
#pragma pack(push, 4)

#elif defined(__linux__)
#include <pthread.h>
#include <stdlib.h>
#include <string.h>

struct pevent_t;

typedef pthread_mutex_t *BH_MUTEX;
typedef pevent_t *BH_EVENT;
typedef pthread_t *BH_THREAD;

#pragma pack(push, 4)
#endif

#include "beihai.h"

typedef unsigned int UINT32;
typedef unsigned char UINT8;

typedef long long ADDR;

extern const unsigned char BH_MSG_BEGINNING[4];
extern const unsigned char BH_MSG_FOLLOWING[4];
extern const unsigned char BH_MSG_RESPONSE[4];

#define APPID_LENGTH 16

typedef char JHI_SESSION_ID[APPID_LENGTH];
typedef char APPID[APPID_LENGTH];

#define BH_MAGIC_LENGTH (4)

#define MAGIC_IS_BEGINNING(buf) (memcmp(buf, BH_MSG_BEGINNING, BH_MAGIC_LENGTH) == 0)
#define MAGIC_IS_FOLLOWING(buf) (memcmp(buf, BH_MSG_FOLLOWING, BH_MAGIC_LENGTH) == 0)
#define MAGIC_IS_RESPONSE(buf) (memcmp(buf, BH_MSG_RESPONSE, BH_MAGIC_LENGTH) == 0)

typedef enum {
    HOST_CMD_INIT = 0,
    HOST_CMD_DEINIT = 1,
    HOST_CMD_SENDANDRECV = 2,
    HOST_CMD_DELETE = 3,
    HOST_CMD_DOWNLOAD = 4,
    HOST_CMD_QUERYAPI = 5,
    HOST_CMD_CREATE_SESSION = 6,
    HOST_CMD_CLOSE_SESSION = 7,
    HOST_CMD_RESET = 8,
    HOST_CMD_LIST_PACKAGES = 9,
    HOST_CMD_LIST_SESSIONS = 10,
    HOST_CMD_LIST_PROPERTIES = 11,
    HOST_CMD_FORCE_CLOSE_SESSION = 12,
    HOST_CMD_SENDANDRECV_INTERNAL = 13,
    HOST_CMD_MAX
} host_command_id;

#ifdef _WIN32
#pragma pack(4)
#pragma warning ( disable: 4200 )
#endif

typedef struct {
    APPID appid;
    char buffer[0];
} host_create_session_command;

typedef struct {
    ADDR addr;
} host_destroy_session_command;

typedef struct {
    APPID appid;
    char appblob[0];
} host_download_command;

typedef struct {
    APPID appid;
} host_delete_command;

typedef struct {
    APPID appid;
    char buffer[0];
} host_query_command;

typedef struct {
    ADDR addr;
    int command;
    int outlen;
    char buffer[0];
} host_snr_command;

typedef struct {
    ADDR addr;
    int what;
    int command;
    int outlen;
    char buffer[0];
} host_snr_internal_command;

typedef struct {
    APPID appid;
} host_list_sessions_command;

typedef struct {
    APPID appid;
} host_list_properties_command;

typedef struct {
    int response;		/* this field comes from java BIG endian */
    char buffer[0];
} client_snr_response;

typedef struct {
    int response;		/* this field comes from java BIG endian */
    int request_length;
} client_snr_bof_response;

typedef struct {
    int count;
    ADDR addr[0];
} client_list_sessions_response;

typedef struct {
    int count;
    APPID id[0];
} client_list_packages_response;

typedef struct {
    int response;		/* this field comes from java BIG endian */
    JHI_SESSION_ID session_id;
    char buffer[0];
} spooler_snr_response;

typedef struct {
    int response;		/* this field comes from java BIG endian */
    int request_length;         /* this field comes from java BIG endian */
} spooler_bof_response;

typedef struct {
    unsigned char magic[4];
    unsigned int length;
} jhi_message_header;

typedef struct {
    jhi_message_header h;
    ADDR seq;
    host_command_id id;
    char pad[4];
    char cmd[0];
} bh_message_header;

typedef struct {
    jhi_message_header h;
    ADDR seq;
    ADDR addr;
    BH_ERRNO code;
    char pad[4];
    char data[0];
} bh_response_header;


#pragma pack(pop)

#ifdef CPLUSPLUS
}
#endif

#endif



/* Local Variables: */
/* mode:c           */
/* c-basic-offset: 4 */
/* indent-tabs-mode: nil */
/* End:             */
