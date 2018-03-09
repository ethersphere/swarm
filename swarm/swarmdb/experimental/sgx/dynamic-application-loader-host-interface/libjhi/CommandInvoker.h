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
**    @file CommandInvoker.h
**
**    @brief  Contains API for jhi commands
**
**    @author Elad Dabool
**
********************************************************************************
*/
#ifndef _COMMAND_INVOKER_H_
#define _COMMAND_INVOKER_H_

#include "CommandsClientFactory.h"
#include "jhi.h"
#include "jhi_i.h"
#include <string_s.h>
#include "teemanagement.h"
#include "dal_tee_metadata.h"

#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
#include "jhi_sdk.h"
#endif

namespace intel_dal
{

	class CommandInvoker
	{
	private:
		ICommandsClient* client;
		bool InvokeCommand(const uint8_t* inputBuffer,uint32_t inputBufferSize,uint8_t** outputBuffer,uint32_t* outputBufferSize);

		// disabling copy constructor and assignment operator by declaring them as private
		CommandInvoker&  operator = (const CommandInvoker& other) { return *this; }
		CommandInvoker(const CommandInvoker& other) { }
	public:
		CommandInvoker();
		~CommandInvoker();

		JHI_RET JhisInit();
		JHI_RET JhisInstall(char* AppId, const FILECHAR* pSrcFile);
		JHI_RET JhisUninstall(char* AppId);
		JHI_RET JhisGetSessionsCount(char* AppId, uint32_t* pSessionCount);
		JHI_RET JhisCreateSession(char* AppId, JHI_SESSION_ID* pSessionID, uint32_t flags,DATA_BUFFER* initBuffer, JHI_PROCESS_INFO* processInfo);
		JHI_RET JhisCloseSession(JHI_SESSION_ID* SessionID,JHI_PROCESS_INFO* processInfo, bool force);
		JHI_RET JhisGetSessionInfo(JHI_SESSION_ID* SessionID,JHI_SESSION_INFO* pSessionInfo);
		JHI_RET JhisSetSessionEventHandler(JHI_SESSION_ID* SessionID, const char* handleName);
		JHI_RET JhisGetEventData(JHI_SESSION_ID* SessionID,uint32_t* DataBufferSize,uint8_t** pDataBuffer,uint8_t* pDataType);
		JHI_RET JhisSendAndRecv(JHI_SESSION_ID* SessionID,int32_t CommandId,const uint8_t* SendBuffer,uint32_t SendBufferSize,uint8_t* RecvBuffer,uint32_t* RecvBufferSize,int32_t* responseCode);
		JHI_RET JhisGetAppletProperty(char* AppId,const uint8_t* SendBuffer,uint32_t SendBufferSize,uint8_t* RecvBuffer,uint32_t* RecvBufferSize);
		JHI_RET JhisGetVersionInfo(JHI_VERSION_INFO* pVersionInfo);
		// Management API
		TEE_STATUS JhisOpenSDSession(IN const string& sdId, OUT SD_SESSION_HANDLE*	sdHandle);
		TEE_STATUS JhisCloseSDSession(IN OUT SD_SESSION_HANDLE* sdHandle);
		TEE_STATUS JhisSendAdminCmdPkg(IN const SD_SESSION_HANDLE sdHandle, IN const uint8_t* package, IN uint32_t packageSize);
		TEE_STATUS JhisListInstalledTAs(IN SD_SESSION_HANDLE sdHandle, OUT	UUID_LIST* uuidList);
		TEE_STATUS JhisListInstalledSDs(IN SD_SESSION_HANDLE sdHandle, OUT	UUID_LIST* uuidList);
		TEE_STATUS JhisQueryTEEMetadata(OUT dal_tee_metadata* metadata, size_t max_length);

#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
		JHI_RET JhisGetSessionTable(JHI_SESSIONS_DATA_TABLE** SessionDataTable);
		JHI_RET JhisGetLoadedAppletsList(JHI_LOADED_APPLET_GUIDS** appGUIDs);
#endif
	};

}
#endif // _COMMAND_INVOKER_H_
