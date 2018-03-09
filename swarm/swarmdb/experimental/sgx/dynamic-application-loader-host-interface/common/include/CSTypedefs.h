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

/**                                                                            
********************************************************************************
**
**    @file CSTypedefs.h
**
**    @brief  Contains common type declarations used throughout the client server communcation
**
**    @author Elad Dabool
**
********************************************************************************
*/
#ifndef _CSTYPEDEFS_H_
#define _CSTYPEDEFS_H_

#include "typedefs.h"
#include "jhi_i.h"

#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
#include "jhi_sdk.h"
#endif

// connection definitions

#define JHI_MAX_CLIENTS_CONNECTIONS 10
#define JHI_MAX_TRANSPORT_DATA_SIZE 5242880 // we limit the data size recieved from and to the server to 5 Megabyte


// command id
typedef enum _JHI_COMMAND_ID
{
	INIT						= 0,
	INSTALL,
	UNINSTALL,
	SEND_AND_RECIEVE,
	CREATE_SESSION,
	CLOSE_SESSION,
	GET_SESSIONS_COUNT,
	GET_SESSION_INFO,
	SET_SESSION_EVENT_HANDLER,
	GET_EVENT_DATA,
	GET_APPLET_PROPERTY,
	GET_VERSION_INFO,
	SEND_CMD_PKG,
	CREATE_SD_SESSION,
	CLOSE_SD_SESSION,
	LIST_INSTALLED_TAS,
	QUERY_TEE_METADATA,
	LIST_INSTALLED_SDS,

#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
	GET_SESSIONS_DATA_TABLE,
	GET_LOADED_APPLETS,
#endif
	INVALID_COMMAND_ID			  // mark last valid command id
}JHI_COMMAND_ID;

// command and response structs
#pragma pack(1)

// JHI Command Header //
typedef struct {
	uint8_t id;
	uint32_t dataLength;
	uint8_t data[1];
} JHI_COMMAND;

// JHI Response Header //
typedef struct {
	uint32_t retCode;
	uint32_t dataLength;
	uint8_t data[1];
} JHI_RESPONSE;

// JHI Install //
typedef struct {
	uint8_t AppId[LEN_APP_ID+1];
	uint32_t SrcFile_size;
	uint8_t data[1];
} JHI_CMD_INSTALL;

// JHI Uninstall //
typedef struct {
	uint8_t AppId[LEN_APP_ID+1];
} JHI_CMD_UNINSTALL;

// JHI Get Session Count //
typedef struct {
	uint8_t AppId[LEN_APP_ID+1];
} JHI_CMD_GET_SESSIONS_COUNT;

typedef struct {
	uint32_t SessionCount;
} JHI_RES_GET_SESSIONS_COUNT;

// JHI Create JTA Session //
typedef struct {
	uint8_t AppId[LEN_APP_ID+1];
	uint32_t initBuffer_size;
	uint32_t flags;
	JHI_PROCESS_INFO processInfo;
	uint8_t data[1];
} JHI_CMD_CREATE_SESSION;

typedef struct {
	JHI_SESSION_ID SessionID; 
} JHI_RES_CREATE_SESSION;

// JHI Close Session //
typedef struct {
	JHI_SESSION_ID SessionID;
	JHI_PROCESS_INFO processInfo;
	bool force;
} JHI_CMD_CLOSE_SESSION;

// JHI Set Session Event Handler
typedef struct {
	JHI_SESSION_ID SessionID;
	uint32_t handleName_size;
	uint8_t data[1];
} JHI_CMD_SET_SESSION_EVENT_HANDLER;

// JHI Get Session Information
typedef struct {
	JHI_SESSION_ID SessionID;
} JHI_CMD_GET_SESSION_INFO;

typedef struct {
	JHI_SESSION_INFO SessionInfo;
} JHI_RES_GET_SESSION_INFO;

// JHI Create SD Session //
typedef struct {
	uint8_t sdId[LEN_APP_ID + 1];
} JHI_CMD_CREATE_SD_SESSION;

typedef struct {
	uint64_t sdHandle;
} JHI_RES_CREATE_SD_SESSION;

// JHI Close SD Session //
typedef struct {
	uint64_t sdHandle;
} JHI_CMD_CLOSE_SD_SESSION;

// JHI Send command package //
typedef struct {
	uint64_t sdHandle;
	uint32_t blobSize;
	uint8_t blob[1];
} JHI_CMD_SEND_CMD_PKG;

// JHI List Installed TAs //
typedef struct {
	uint64_t sdHandle;
} JHI_CMD_LIST_INSTALLED_TAS;

typedef struct {
	uint32_t count; // Number of UUIDs recieved.
	uint8_t data[1]; // The buffer containing all the UUIDs concatenated one after the other with no spaces between them. + null termination at the end.
} JHI_RES_LIST_INSTALLED_TAS;

// JHI List Installed SDs //
typedef struct {
	uint64_t sdHandle;
} JHI_CMD_LIST_INSTALLED_SDS;

typedef struct {
	uint32_t count; // Number of UUIDs recieved.
	uint8_t data[1]; // The buffer containing all the UUIDs concatenated one after the other with no spaces between them. + null termination at the end.
} JHI_RES_LIST_INSTALLED_SDS;

typedef struct {
	uint32_t length; // length of metadata recieved
	uint8_t metadata[1];
} JHI_RES_QUERY_TEE_METADATA;

#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
// JHI Get Session Data Table
typedef struct {
	JHI_SESSIONS_DATA_TABLE SessionDataTable;
} JHI_RES_GET_SESSIONS_DATA_TABLE;

typedef struct {
	JHI_LOADED_APPLET_GUIDS loadedApplets;
} JHI_RES_GET_LOADED_APPLETS;
#endif

// JHI Get Event Data
typedef struct {
	JHI_SESSION_ID SessionID;
} JHI_CMD_GET_EVENT_DATA;

typedef struct {
	uint32_t DataBuffer_size;
	uint8_t DataType;
	uint8_t data[1];
} JHI_RES_GET_EVENT_DATA;

// JHI Send And Recieve
typedef struct {
	JHI_SESSION_ID SessionID;
	int32_t CommandId;
	uint32_t SendBuffer_size;
	uint32_t RecvBuffer_size;
	uint8_t data[1];
} JHI_CMD_SEND_AND_RECIEVE;

typedef struct {
	int32_t ResponseCode;
	uint32_t RecvBuffer_size;
	uint8_t data[1];
} JHI_RES_SEND_AND_RECIEVE;

// JHI Get Applet property
typedef struct {
	uint8_t AppId[LEN_APP_ID+1];
	uint32_t SendBuffer_size;
	uint32_t RecvBuffer_size;
	uint8_t data[1];
} JHI_CMD_GET_APPLET_PROPERTY;

typedef struct {
	uint32_t RecvBuffer_size;
	uint8_t data[1];
} JHI_RES_GET_APPLET_PROPERTY;

#pragma pack()

#endif