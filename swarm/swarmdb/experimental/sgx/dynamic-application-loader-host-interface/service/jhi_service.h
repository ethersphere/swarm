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
**    @file jhi_service.h
**
**    @brief  Defines service intenal interface and typedefs
**
**    @author Elad Dabool
**
********************************************************************************
*/

#ifndef __JHI_SERVICE_H__
#define __JHI_SERVICE_H__

#include "jhi_i.h"
#include "jhi_plugin_loader.h"


#define JHI_EVENT_DATA_BUFFER_SIZE 1024
#define JHI_EVENT_HANDLE_SIZE 88

// internal errors
#define	JHI_SM_TIMEOUT			 702
#define JHI_MEM_ALLOC_FAIL		 703
#define JHI_ITEM_NOT_EXISTS		 706

// internal Install errors
#define JHI_APPLET_AUTHENTICATION_FAILURE		JHI_FILE_ERROR_AUTH     // FW rejected the applet while trying to install it
#define JHI_BAD_APPLET_FORMAT					0x2001						

// Applet related
#define MAX_DAL_APPLETS 6
#define SPOOLER_COMMAND_GET_EVENT 1

JHI_RET_I jhis_init();

JHI_RET_I
jhis_txrx_raw( 
			  JHI_SESSION_ID* pSessionID,
			  INT32  nCommandId,
			  JVM_COMM_BUFFER* pIOBuffer,
			  INT32* pResponseCode
			  );

//int deinit
JHI_RET_I
jhis_uninstall(const char* pAppId, const VM_SESSION_HANDLE handle = NULL, vector<uint8_t>* blob = NULL);

JHI_RET_I 
jhis_unload(const char* pAppId, const VM_SESSION_HANDLE handle = NULL, vector<uint8_t>* blob = NULL);

JHI_RET_I
jhis_create_session(
	const char* pAppId,
	JHI_SESSION_ID* pSessionID,
	UINT32 flags,
	DATA_BUFFER* initBuffer,
	JHI_PROCESS_INFO* processInfo
);

JHI_RET_I
jhis_close_session(
	JHI_SESSION_ID *pSessionID,
	JHI_PROCESS_INFO *processInfo,
	bool force,
	bool removeFromVM
);

JHI_RET_I
jhis_get_sessions_count(
	const char* pAppId,
	UINT32* pSessionsCount
);

JHI_RET_I
jhis_get_session_info(
	JHI_SESSION_ID*		 pSessionID,
	JHI_SESSION_INFO*    pSessionInfo
);

#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
JHI_RET_I
jhis_get_sessions_data_table(
	JHI_SESSIONS_DATA_TABLE*	 pSessionsDataTable
);

JHI_RET_I
	jhis_get_loaded_applets(JHI_LOADED_APPLET_GUIDS* loadedAppletsList);
#endif

JHI_RET_I
	jhis_install(
	const char* pAppId,
	const FILECHAR* pFile,
	bool visibleApp,
	bool isAcp
	);

JHI_RET_I
	jhis_send_cmd_pkg(
	const VM_SESSION_HANDLE handle, 
	vector<uint8_t>& blob
	);

JHI_RET_I
	jhis_get_applet_property (
	const char* pAppId,
	JVM_COMM_BUFFER* pCommBuffer
);

bool TryUnloadUnusedApplet();

void JhiReset();

#endif
