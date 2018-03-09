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

#include <jhi.h>
#include "CommandDispatcher.h"
#include "misc.h"
#include "jhi_service.h"
#include "jhi.h"
#include "EventManager.h"
#include "SessionsManager.h"
#include "AppletsManager.h"
#include "jhi_version.h"
#include "string_s.h"

namespace intel_dal
{
	CommandDispatcher::CommandDispatcher()
	{
	}

	void CommandDispatcher::processCommand(IN const uint8_t* inputData,IN uint32_t inputSize,OUT uint8_t** outputData,OUT uint32_t* outputSize)
	{
		JHI_RESPONSE res_header;
		UINT32 ulRetCode = JHI_SUCCESS;
		bool init_succeeded = false;
		res_header.dataLength = 0;

		do
		{
			if (inputSize<sizeof(JHI_COMMAND) || inputData == NULL)
			{
				TRACE0("recieved invalid input\n");
				ulRetCode = JHI_INTERNAL_ERROR;
				break;
			}

			JHI_COMMAND* cmd_header = (JHI_COMMAND*) inputData;

			if (cmd_header->id >= INVALID_COMMAND_ID)
			{
				TRACE0("invalid command: illegal id in request\n");
				ulRetCode = JHI_INTERNAL_ERROR;
				break;
			}

			if (cmd_header->dataLength != inputSize)
			{
				TRACE0("invalid command: illegal data in request.\n");
				ulRetCode = JHI_INTERNAL_ERROR;
				break;
			}

			if (GlobalsManager::Instance().getJhiState() == JHI_STOPPED)
			{
				GlobalsManager::Instance().initLock.aquireWriterLock();

				// Note that even on suspend/resume or TIMEOUT recovery, init may be called again
				if (GlobalsManager::Instance().getJhiState() == JHI_STOPPED) //Do another init
				{    
					ulRetCode = jhis_init();
					if (ulRetCode != JHI_SUCCESS) 
					{
						GlobalsManager::Instance().initLock.releaseWriterLock();
						break;
					}
				}

				GlobalsManager::Instance().initLock.releaseWriterLock();
			}

			GlobalsManager::Instance().initLock.aquireReaderLock();

			if (GlobalsManager::Instance().getJhiState() != JHI_INITIALIZED)
			{
				ulRetCode = JHI_SERVICE_UNAVAILABLE;
				GlobalsManager::Instance().initLock.releaseReaderLock();
				break;
			}
			
			init_succeeded = true;
			
			// sync all jhi api exept for SendAndRecieve 
			if (cmd_header->id != SEND_AND_RECIEVE)
				_jhiMutex.Lock();

			switch (cmd_header->id)
			{
			case INIT:	InvokeInit(inputData,inputSize,outputData,outputSize);
				break;
			case INSTALL: InvokeInstall(inputData,inputSize,outputData,outputSize);
				break;
			case UNINSTALL:	InvokeUninstall(inputData,inputSize,outputData,outputSize);
				break;
			case SEND_AND_RECIEVE: InvokeSendAndRecieve(inputData,inputSize,outputData,outputSize);
				break;
			case CREATE_SESSION: InvokeCreateSession(inputData,inputSize,outputData,outputSize);
				break;
			case CLOSE_SESSION: InvokeCloseSession(inputData,inputSize,outputData,outputSize);
				break;
			case GET_SESSIONS_COUNT: InvokeGetSessionsCount(inputData,inputSize,outputData,outputSize);
				break;
			case GET_SESSION_INFO: InvokeGetSessionInfo(inputData,inputSize,outputData,outputSize);
				break;
			case SET_SESSION_EVENT_HANDLER:	InvokeSetSessionEventHandler(inputData,inputSize,outputData,outputSize);
				break;
			case GET_EVENT_DATA: InvokeGetSessionEventData(inputData,inputSize,outputData,outputSize);
				break;
			case GET_APPLET_PROPERTY: InvokeGetAppletProperty(inputData,inputSize,outputData,outputSize);
				break;
			case GET_VERSION_INFO: InvokeGetVersionInfo(inputData,inputSize,outputData,outputSize);
				break;
			case LIST_INSTALLED_TAS: InvokeListInstalledTAs(inputData,inputSize,outputData,outputSize);
				break;
			case LIST_INSTALLED_SDS: InvokeListInstalledSDs(inputData, inputSize, outputData, outputSize);
				break;
			case CREATE_SD_SESSION: InvokeOpenSDSession(inputData,inputSize,outputData,outputSize);
				break;
			case CLOSE_SD_SESSION: InvokeCloseSDSession(inputData,inputSize,outputData,outputSize);
				break;
			case SEND_CMD_PKG: InvokeSendCmdPkg(inputData,inputSize,outputData,outputSize);
				break;
			case QUERY_TEE_METADATA: InvokeQueryTeeMetadata(inputData,inputSize,outputData,outputSize);
				break;

#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
			case GET_SESSIONS_DATA_TABLE: InvokeGetSessionDataTable(inputData,inputSize,outputData,outputSize);
				break;
			case GET_LOADED_APPLETS: InvokeGetLoadedApplets(inputData,inputSize,outputData,outputSize);
				break;
#endif
			}

			// sync all jhi api exept for SendAndRecieve 
			if (cmd_header->id != SEND_AND_RECIEVE)
				_jhiMutex.UnLock();

			GlobalsManager::Instance().initLock.releaseReaderLock();
		}
		while(0);

		if (ulRetCode != JHI_SUCCESS)
		{
			if (GlobalsManager::Instance().getJhiState() != JHI_INITIALIZED && init_succeeded)
			{ 
				// JHI is probably being stopped
				res_header.retCode = JHI_SERVICE_UNAVAILABLE;
			}
			else
			{
				res_header.retCode = ulRetCode;
			}
			res_header.dataLength = sizeof(JHI_RESPONSE);

			*outputData = (uint8_t*) JHI_ALLOC(sizeof(JHI_RESPONSE));
			if (*outputData == NULL) {
				TRACE0("malloc of outputData failed .");
				return;
			}

			*((JHI_RESPONSE*)(*outputData)) = res_header;

			*outputSize = sizeof(JHI_RESPONSE);
		}

	}

	bool CommandDispatcher::convertAppIDtoUpperCase(const char *pAppId,UINT8 ucConvertedAppId[LEN_APP_ID+1])
	{

		if (pAppId == NULL) 
			return false;

		if ( ( JhiUtilUUID_Validate(pAppId, ucConvertedAppId) != JHI_SUCCESS))
		{
			TRACE0( "invalid AppId\n");
			return false;
		}

		return true;
	}

	int CommandDispatcher::verifyAppID(char *pAppId)
	{
		int ulRetCode = JHI_SUCCESS;

		do
		{
			if (strlen(pAppId) != LEN_APP_ID)
			{
				TRACE0("illegal applet UUID length\n");
				ulRetCode = JHI_INVALID_APPLET_GUID;
				break;
			}

			//block any command that is involved with the Spooler Applet UUID
			if (strcmp((char*)pAppId, SPOOLER_APPLET_UUID) == 0)
			{
				TRACE0("illegal use of spooler applet UUID\n");
				ulRetCode = JHI_INVALID_APPLET_GUID;
				break;
			}
		} while(0);

		return ulRetCode;
	}

	bool CommandDispatcher::init()
	{
		return true;
	}

	bool CommandDispatcher::deinit()
	{
		return true;
	}

	void CommandDispatcher::InvokeInit(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize)
	{
		// nothig to do since init performed before. send success responce. 
		JHI_RESPONSE res;
		res.retCode = JHI_SUCCESS;
		res.dataLength = sizeof(JHI_RESPONSE);

		if (inputSize != sizeof(JHI_COMMAND))
			res.retCode = JHI_INTERNAL_ERROR;

		*outputData = (uint8_t*) JHI_ALLOC(res.dataLength);
		if (*outputData == NULL) {
			TRACE0("malloc of outputData failed .");
			return;
		}

		*((JHI_RESPONSE*)(*outputData)) = res;

		*outputSize = res.dataLength;
	}

	void CommandDispatcher::InvokeInstall(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize)
	{
		const JHI_COMMAND* cmd = (JHI_COMMAND*) inputData;
		JHI_RESPONSE res;
		JHI_CMD_INSTALL* install = NULL;
		res.dataLength = sizeof(JHI_RESPONSE);

		char* pAppid = NULL;
		FILECHAR* pFile;
		UINT8 ucAppID[LEN_APP_ID+1];

		do
		{
			if (cmd->dataLength != inputSize)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			if (inputSize < sizeof(JHI_COMMAND) + sizeof(JHI_CMD_INSTALL) - 1)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			install = (JHI_CMD_INSTALL*) cmd->data;

			// SrcFile_size cannot be less then empty string and not more than FILENAME_MAX characters
			if (install->SrcFile_size < sizeof(FILECHAR) || install->SrcFile_size > ((FILENAME_MAX + 1) * sizeof(FILECHAR)) ||
				install->SrcFile_size % sizeof(FILECHAR) != 0)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			if (inputSize != sizeof(JHI_COMMAND) + sizeof(JHI_CMD_INSTALL) + install->SrcFile_size -2)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			pAppid = (char*) install->AppId;
			pFile = (FILECHAR*) install->data;

			if (pAppid[LEN_APP_ID] != '\0')
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			if (pFile[(install->SrcFile_size / sizeof(FILECHAR)) - 1] != '\0')
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}


			// convet to uppercase and verify Applet id
			if (!convertAppIDtoUpperCase(pAppid,ucAppID))
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			if (verifyAppID((char*)ucAppID) != JHI_SUCCESS)
			{
				res.retCode = JHI_INVALID_APPLET_GUID;
				break;
			}

			res.retCode = jhis_install((char*)ucAppID,pFile, true, false);

		}
		while(0);

		*outputData = (uint8_t*) JHI_ALLOC(res.dataLength);
		if (*outputData == NULL) {
			TRACE0("malloc of outputData failed .");
			return;
		}

		*((JHI_RESPONSE*)(*outputData)) = res;

		*outputSize = res.dataLength;
	}

	void CommandDispatcher::InvokeUninstall(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize)
	{
		const JHI_COMMAND* cmd = (JHI_COMMAND*) inputData;
		JHI_RESPONSE res;
		JHI_CMD_UNINSTALL* uninstall = NULL;
		res.dataLength = sizeof(JHI_RESPONSE);

		char* pAppid = NULL;
		UINT8 ucAppID[LEN_APP_ID+1];

		do
		{
			if (cmd->dataLength != inputSize)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			uninstall = (JHI_CMD_UNINSTALL*) cmd->data;

			if (inputSize != sizeof(JHI_COMMAND) + sizeof(JHI_CMD_UNINSTALL) - 1)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			pAppid = (char*) uninstall->AppId;

			if (pAppid[LEN_APP_ID] != '\0')
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			// convet to uppercase and verify Applet id
			if (!convertAppIDtoUpperCase(pAppid,ucAppID))
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			if (verifyAppID((char*)ucAppID) != JHI_SUCCESS)
			{
				res.retCode = JHI_INVALID_APPLET_GUID;
				break;
			}

			res.retCode = jhis_uninstall((char*)ucAppID);

		}
		while(0);

		*outputData = (uint8_t*) JHI_ALLOC(res.dataLength);
		if (*outputData == NULL) {
			TRACE0("malloc of outputData failed .");
			return;
		}

		*((JHI_RESPONSE*)(*outputData)) = res;

		*outputSize = res.dataLength;
	}

	void CommandDispatcher::InvokeGetSessionsCount(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize)
	{
		const JHI_COMMAND* cmd = (JHI_COMMAND*) inputData;
		JHI_RESPONSE res;
		JHI_RES_GET_SESSIONS_COUNT res_data;
		JHI_CMD_GET_SESSIONS_COUNT* cmd_data = NULL;
		res.dataLength = 0;

		char* pAppid = NULL;
		UINT8 ucAppID[LEN_APP_ID+1];
		memset(&res_data,0,sizeof(JHI_RES_GET_SESSIONS_COUNT));

		do
		{
			if (cmd->dataLength != inputSize)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			cmd_data = (JHI_CMD_GET_SESSIONS_COUNT*) cmd->data;

			if (inputSize != sizeof(JHI_COMMAND) + sizeof(JHI_CMD_GET_SESSIONS_COUNT) - 1)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			pAppid = (char*) cmd_data->AppId;

			if (pAppid[LEN_APP_ID] != '\0')
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			// convet to uppercase and verify Applet id
			if (!convertAppIDtoUpperCase(pAppid,ucAppID))
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			if (verifyAppID((char*)ucAppID) != JHI_SUCCESS)
			{
				res.retCode = JHI_INVALID_APPLET_GUID;
				break;
			}

			res.retCode = jhis_get_sessions_count((char*)ucAppID,&res_data.SessionCount);


		}
		while(0);

		res.dataLength = sizeof(JHI_RESPONSE) + sizeof(JHI_RES_GET_SESSIONS_COUNT);

		*outputData = (uint8_t*) JHI_ALLOC(res.dataLength);
		if (*outputData == NULL) {
			TRACE0("malloc of outputData failed .");
			return;
		}

		*((JHI_RESPONSE*)(*outputData)) = res;
		*((JHI_RES_GET_SESSIONS_COUNT*)((*((JHI_RESPONSE*)(*outputData))).data)) = res_data;

		*outputSize = res.dataLength;
	}

	void CommandDispatcher::InvokeCreateSession(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize)
	{
		const JHI_COMMAND* cmd = (JHI_COMMAND*) inputData;
		JHI_RESPONSE res;
		JHI_RES_CREATE_SESSION res_data;
		JHI_CMD_CREATE_SESSION* cmd_data = NULL;

		char* pAppid = NULL;
		UINT8 ucAppID[LEN_APP_ID+1];
		memset(&res_data,0,sizeof(JHI_RES_CREATE_SESSION));

		do
		{
			if (cmd->dataLength != inputSize)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			if (inputSize < sizeof(JHI_COMMAND) + sizeof(JHI_CMD_CREATE_SESSION) -1)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			cmd_data = (JHI_CMD_CREATE_SESSION*) cmd->data;

			if (cmd_data->initBuffer_size > JHI_BUFFER_MAX)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			if ((cmd_data->initBuffer_size == 0) && (inputSize != sizeof(JHI_COMMAND) + sizeof(JHI_CMD_CREATE_SESSION) + cmd_data->initBuffer_size -1))
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			if ((cmd_data->initBuffer_size > 0) && (inputSize != sizeof(JHI_COMMAND) + sizeof(JHI_CMD_CREATE_SESSION) + cmd_data->initBuffer_size -2))
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			pAppid = (char*) cmd_data->AppId;

			if (pAppid[LEN_APP_ID] != '\0')
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			// convet to uppercase and verify Applet id
			if (!convertAppIDtoUpperCase(pAppid,ucAppID))
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			if (verifyAppID((char*)ucAppID) != JHI_SUCCESS)
			{
				res.retCode = JHI_INVALID_APPLET_GUID;
				break;
			}

			DATA_BUFFER initData;
			initData.length = cmd_data->initBuffer_size;

			if (initData.length > 0)
			{
				initData.buffer = cmd_data->data;
			}
			else
			{
				initData.buffer = NULL;
			}

			res.retCode = jhis_create_session((char*)ucAppID,&res_data.SessionID,cmd_data->flags,&initData,&(cmd_data->processInfo));


		}
		while(0);

		res.dataLength = sizeof(JHI_RESPONSE) + sizeof(JHI_RES_CREATE_SESSION);

		*outputData = (uint8_t*) JHI_ALLOC(res.dataLength);
		if (*outputData == NULL) {
			TRACE0("malloc of outputData failed .");
			return;
		}

		*((JHI_RESPONSE*)(*outputData)) = res;
		*((JHI_RES_CREATE_SESSION*)((*((JHI_RESPONSE*)(*outputData))).data)) = res_data;

		*outputSize = res.dataLength;

	}

	void CommandDispatcher::InvokeCloseSession(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize)
	{
		const JHI_COMMAND* cmd = (JHI_COMMAND*) inputData;
		JHI_RESPONSE res;
		JHI_CMD_CLOSE_SESSION* cmd_data = NULL;
		res.dataLength = sizeof(JHI_RESPONSE);

		do
		{
			if (cmd->dataLength != inputSize)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			cmd_data = (JHI_CMD_CLOSE_SESSION*) cmd->data;

			if (inputSize != sizeof(JHI_COMMAND) + sizeof(JHI_CMD_CLOSE_SESSION) - 1)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			res.retCode = jhis_close_session(&cmd_data->SessionID,&cmd_data->processInfo, cmd_data->force, true);

		}
		while(0);

		*outputData = (uint8_t*) JHI_ALLOC(res.dataLength);
		if (*outputData == NULL) {
			TRACE0("malloc of outputData failed .");
			return;
		}

		*((JHI_RESPONSE*)(*outputData)) = res;

		*outputSize = res.dataLength;
	}

	void CommandDispatcher::InvokeSetSessionEventHandler(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize)
	{
		const JHI_COMMAND* cmd = (JHI_COMMAND*) inputData;
		JHI_RESPONSE res;
		JHI_CMD_SET_SESSION_EVENT_HANDLER* cmd_data = NULL;
		res.dataLength = sizeof(JHI_RESPONSE);
		char* pHandleName = NULL;
		
		do
		{

			if (cmd->dataLength != inputSize)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			if (inputSize < sizeof(JHI_COMMAND) + sizeof(JHI_CMD_SET_SESSION_EVENT_HANDLER) -1)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			cmd_data = (JHI_CMD_SET_SESSION_EVENT_HANDLER*) cmd->data;

			if ((cmd_data->handleName_size < sizeof(char)) || (cmd_data->handleName_size > JHI_EVENT_HANDLE_SIZE * sizeof(char)))
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			if (inputSize != sizeof(JHI_COMMAND) + sizeof(JHI_CMD_SET_SESSION_EVENT_HANDLER) + cmd_data->handleName_size -2)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			pHandleName = (char*) cmd_data->data;

			if (pHandleName[cmd_data->handleName_size -1] != '\0')
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			res.retCode = EventManager::Instance().SetSessionEventHandler(cmd_data->SessionID, pHandleName);	

		}
		while(0);

		*outputData = (uint8_t*) JHI_ALLOC(res.dataLength);
		if (*outputData == NULL) {
			TRACE0("malloc of outputData failed .");
			return;
		}

		*((JHI_RESPONSE*)(*outputData)) = res;

		*outputSize = res.dataLength;
	}

	void CommandDispatcher::InvokeGetSessionInfo(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize)
	{
		const JHI_COMMAND* cmd = (JHI_COMMAND*) inputData;
		JHI_RESPONSE res;
		JHI_RES_GET_SESSION_INFO res_data;
		JHI_CMD_GET_SESSION_INFO* cmd_data = NULL;
		res.dataLength = 0;

		memset(&res_data,0,sizeof(JHI_RES_GET_SESSION_INFO));

		do
		{
			if (cmd->dataLength != inputSize)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			cmd_data = (JHI_CMD_GET_SESSION_INFO*) cmd->data;

			if (inputSize != (sizeof(JHI_COMMAND) + sizeof(JHI_CMD_GET_SESSION_INFO) - 1))
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			res.retCode = jhis_get_session_info(&cmd_data->SessionID,&res_data.SessionInfo);

		}
		while(0);

		res.dataLength = sizeof(JHI_RESPONSE) + sizeof(JHI_RES_GET_SESSION_INFO);

		*outputData = (uint8_t*) JHI_ALLOC(res.dataLength);
		if (*outputData == NULL) {
			TRACE0("malloc of outputData failed .");
			return;
		}

		*((JHI_RESPONSE*)(*outputData)) = res;
		*((JHI_RES_GET_SESSION_INFO*)((*((JHI_RESPONSE*)(*outputData))).data)) = res_data;

		*outputSize = res.dataLength;
	}

#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
	int getSizeOf_JHI_SESSIONS_DATA_TABLE(JHI_SESSIONS_DATA_TABLE* dataTable)
	{
		int size = 0;
		for (UINT32 session=0; session < (dataTable->sessionsCount); session++)
		{
			size += sizeof(dataTable->dataTable[session]);
			if (dataTable->dataTable[session].ownersListCount > 0)
			{
				size += sizeof(JHI_PROCESS_INFO) * dataTable->dataTable[session].ownersListCount;
			}
		}
		return size;
	}

	void CommandDispatcher::InvokeGetSessionDataTable(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize)
	{
		const JHI_COMMAND* cmd = (JHI_COMMAND*) inputData;
		JHI_RESPONSE res;
		JHI_RES_GET_SESSIONS_DATA_TABLE res_data;
		JHI_SESSION_EXTENDED_INFO* sessionsOffset = NULL;
		JHI_PROCESS_INFORMATION* OwnersListsOffset = NULL;
		res.dataLength = 0;

		if (cmd->dataLength != inputSize)
		{
			res.retCode = JHI_INTERNAL_ERROR;
			return;
		}

		if (inputSize != sizeof(JHI_COMMAND))
		{
			res.retCode = JHI_INTERNAL_ERROR;
			return;
		}

		if (!outputData)
		{
			res.retCode = JHI_INTERNAL_ERROR;
			return;
		}

		res.retCode = jhis_get_sessions_data_table(&res_data.SessionDataTable);

		if (res.retCode != JHI_SUCCESS)
		{
			if (res_data.SessionDataTable.dataTable)
			{
				JHI_DEALLOC_T_ARRAY(res_data.SessionDataTable.dataTable);
				res_data.SessionDataTable.dataTable = NULL;
			}
			*outputSize = sizeof(JHI_RESPONSE);
			*outputData = (uint8_t*) JHI_ALLOC(*outputSize);
			*((JHI_RESPONSE*)(*outputData)) = res;
			return;
		}
		bool failure = false;
		/** arranging all the data in one buffer to return to the dll **/

		//calculating the needed buffer size
		int sessionInfoSize = 0;
		if (res_data.SessionDataTable.sessionsCount > 0)
			sessionInfoSize = sizeof(JHI_SESSION_EXTENDED_INFO) * res_data.SessionDataTable.sessionsCount;
		int ownersListsSize = 0;
		for ( UINT32 i = 0; i < res_data.SessionDataTable.sessionsCount ; ++i )
		{
			ownersListsSize += sizeof(JHI_PROCESS_INFO) * res_data.SessionDataTable.dataTable[i].ownersListCount;
		}
		res.dataLength = sizeof(JHI_RESPONSE) + sizeof(JHI_RES_GET_SESSIONS_DATA_TABLE) + sessionInfoSize + ownersListsSize;

		//allocating the buffer
		*outputData = (uint8_t*) JHI_ALLOC(res.dataLength);
		if (*outputData == NULL)
		{
			failure = true;
			goto error;
		}
		*((JHI_RESPONSE*)(*outputData)) = res;
		*((JHI_RES_GET_SESSIONS_DATA_TABLE*)((*((JHI_RESPONSE*)(*outputData))).data)) = res_data;

		/** pointer calculations **/

		//calculates where the sessions info starts
		sessionsOffset = (JHI_SESSION_EXTENDED_INFO*) ((uint8_t*)(*outputData) + sizeof(JHI_RESPONSE) - sizeof(uint8_t) + sizeof(JHI_RES_GET_SESSIONS_DATA_TABLE));

		//updates the SessionDataTable pointer to where the sessions info starts
		(*((JHI_RES_GET_SESSIONS_DATA_TABLE*)((*((JHI_RESPONSE*)(*outputData))).data))).SessionDataTable.dataTable = sessionsOffset;

		//calculates where the owners lists should start
		OwnersListsOffset = (JHI_PROCESS_INFORMATION*) (((uint8_t*)sessionsOffset) + sessionInfoSize); //(starts at the end of the sessions)


		/** arrange all the data in the new outputbuffer **/

		//the sessions info array:
		for ( UINT32 i = 0; i < res_data.SessionDataTable.sessionsCount ; ++i )
		{
			sessionsOffset[i] = res_data.SessionDataTable.dataTable[i];
			// the owners list
			for ( UINT32 j = 0; j < res_data.SessionDataTable.dataTable[i].ownersListCount; ++j)
			{
				OwnersListsOffset[j] = res_data.SessionDataTable.dataTable[i].ownersList[j];
			}
			// cleaning the copy
			JHI_DEALLOC_T_ARRAY(res_data.SessionDataTable.dataTable[i].ownersList);
			res_data.SessionDataTable.dataTable[i].ownersList = NULL;

			//redirects the pointer
			sessionsOffset[i].ownersList = OwnersListsOffset;

			//updates the offset
			OwnersListsOffset += res_data.SessionDataTable.dataTable[i].ownersListCount;
		}
		JHI_DEALLOC_T_ARRAY(res_data.SessionDataTable.dataTable);
		res_data.SessionDataTable.dataTable = NULL;

		*outputSize = res.dataLength;
error:
		if (failure)
		{
			TRACE0("malloc of outputData failed .");
			if (res_data.SessionDataTable.dataTable)
			{
				JHI_DEALLOC_T_ARRAY(res_data.SessionDataTable.dataTable);
				res_data.SessionDataTable.dataTable = NULL;
			}
			if (*outputData)
			{
				JHI_DEALLOC(*outputData);
				*outputData = NULL;
			}
			*outputSize = sizeof(JHI_RESPONSE);
			*outputData = (uint8_t*) JHI_ALLOC(*outputSize);
			res.retCode = JHI_MEM_ALLOC_FAIL;
			if (*outputData != NULL)
			{
				*((JHI_RESPONSE*)(*outputData)) = res;
			}

			TRACE0("malloc of outputData failed .");
		}
	}

	void CommandDispatcher::InvokeGetLoadedApplets(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize)
	{
		const JHI_COMMAND* cmd = (JHI_COMMAND*) inputData;
		JHI_RESPONSE res;
		JHI_RES_GET_LOADED_APPLETS res_data;
		JHI_RES_GET_LOADED_APPLETS tmp;
		char** outputGUIDs = NULL;
		res.dataLength = 0;
		if (!outputData)
		{
			res.retCode = JHI_INTERNAL_ERROR;
			return;
		}

		memset(&res_data,0,sizeof(JHI_RES_GET_LOADED_APPLETS));

		do
		{
			if (cmd->dataLength != inputSize)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			res.retCode = jhis_get_loaded_applets(&res_data.loadedApplets);

		}
		while(0);

		//arranging the data to return to dll
		if (res.retCode != JHI_SUCCESS)
		{
			freeLoadedAppletsList(&res_data.loadedApplets);
			*outputSize = sizeof(JHI_RESPONSE);
			*outputData = (uint8_t*) JHI_ALLOC(*outputSize);
			*((JHI_RESPONSE*)(*outputData)) = res;
			return;
		}

		bool failure = false;

		//calculating the needed buffer size
		res.dataLength = sizeof(JHI_RESPONSE) + sizeof(JHI_RES_GET_LOADED_APPLETS) + (res_data.loadedApplets.loadedAppletsCount * (LEN_APP_ID + 1));

		//allocating the buffer
		*outputData = (uint8_t*) JHI_ALLOC(res.dataLength);
		if (*outputData == NULL)
		{
			failure = true;
			goto error;
		}
		*((JHI_RESPONSE*)(*outputData)) = res;
		*((JHI_RES_GET_LOADED_APPLETS*)((*((JHI_RESPONSE*)(*outputData))).data)) = res_data;

		// the offset where the GUIDs start
		outputGUIDs = (char**) ((uint8_t*)(*outputData) + sizeof(JHI_RESPONSE) - sizeof(uint8_t) + sizeof(JHI_RES_GET_LOADED_APPLETS));

		tmp = *((JHI_RES_GET_LOADED_APPLETS*)((*((JHI_RESPONSE*)(*outputData))).data));
		tmp.loadedApplets.appsGUIDs = outputGUIDs;

		/** Copying all the data to the outputbuffer **/			
		for (UINT32 i = 0; i < res_data.loadedApplets.loadedAppletsCount; ++i)
		{
			memcpy_s((uint8_t*)outputGUIDs + i*(LEN_APP_ID + 1), LEN_APP_ID + 1, res_data.loadedApplets.appsGUIDs[i], LEN_APP_ID + 1);
		}

		*((JHI_RES_GET_LOADED_APPLETS*)((*((JHI_RESPONSE*)(*outputData))).data)) = tmp;

		//cleanup
		freeLoadedAppletsList(&res_data.loadedApplets);

		*outputSize = res.dataLength;
error:
		if (failure) 
		{
			freeLoadedAppletsList(&res_data.loadedApplets);
			if (*outputData)
			{
				JHI_DEALLOC(*outputData);
				*outputData = NULL;
			}
			*outputSize = sizeof(JHI_RESPONSE);
			*outputData = (uint8_t*) JHI_ALLOC(*outputSize);
			res.retCode = JHI_MEM_ALLOC_FAIL;
			if (*outputData != NULL)
			{
				*((JHI_RESPONSE*)(*outputData)) = res;
			}

			TRACE0("malloc of outputData failed .");
		}
	}

#endif

	void CommandDispatcher::InvokeGetSessionEventData(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize)
	{
		const JHI_COMMAND* cmd = (JHI_COMMAND*) inputData;
		JHI_RESPONSE res;
		JHI_RES_GET_EVENT_DATA res_data;
		JHI_CMD_GET_EVENT_DATA* cmd_data = NULL;
		res.dataLength = 0;
		memset(&res_data,0,sizeof(JHI_RES_GET_EVENT_DATA));

		JHI_EVENT_DATA event_data;
		event_data.data = NULL;
		event_data.datalen = 0;

		do
		{
			if (cmd->dataLength != inputSize)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			cmd_data = (JHI_CMD_GET_EVENT_DATA*) cmd->data;

			if (inputSize != (sizeof(JHI_COMMAND) + sizeof(JHI_CMD_GET_EVENT_DATA) - 1))
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			res.retCode = SessionsManager::Instance().getSessionEventData(cmd_data->SessionID,&event_data);

			res_data.DataBuffer_size = event_data.datalen;
			res_data.DataType = event_data.dataType;

		}
		while(0);

		res.dataLength = sizeof(JHI_RESPONSE) + sizeof(JHI_RES_GET_EVENT_DATA) + event_data.datalen;

		*outputData = (uint8_t*) JHI_ALLOC(res.dataLength);
		if (*outputData == NULL) {
			TRACE0("malloc of outputData failed .");
			// cleanup
			if (event_data.data != NULL)
			{
				JHI_DEALLOC(event_data.data);
				event_data.data = NULL;
			}
			return;
		}

		*((JHI_RESPONSE*)(*outputData)) = res;
		*((JHI_RES_GET_EVENT_DATA*)((*((JHI_RESPONSE*)(*outputData))).data)) = res_data;

		if (event_data.data != NULL)
		{
			uint8_t* pBufferData = (*((JHI_RES_GET_EVENT_DATA*)((*((JHI_RESPONSE*)(*outputData))).data))).data;
			memcpy_s(pBufferData,event_data.datalen,event_data.data,event_data.datalen);

			JHI_DEALLOC(event_data.data);
			event_data.data = NULL;
		}

		*outputSize = res.dataLength;
	}

	void CommandDispatcher::InvokeSendAndRecieve(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize)
	{
		const JHI_COMMAND* cmd = (JHI_COMMAND*) inputData;
		JHI_RESPONSE res;
		JHI_RES_SEND_AND_RECIEVE res_data;
		JHI_CMD_SEND_AND_RECIEVE* cmd_data = NULL;
		res.dataLength = sizeof(JHI_RESPONSE);
		res_data.RecvBuffer_size = 0;
		JVM_COMM_BUFFER IOBuffer;

		IOBuffer.TxBuf->buffer = NULL;
		IOBuffer.TxBuf->length = 0;
		IOBuffer.RxBuf->buffer = NULL;
		IOBuffer.RxBuf->length = 0;

		memset(&res_data,0,sizeof(JHI_RES_SEND_AND_RECIEVE));

		do
		{
			if (cmd->dataLength != inputSize)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			if (inputSize < (sizeof(JHI_COMMAND) + sizeof(JHI_CMD_SEND_AND_RECIEVE) -1))
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			cmd_data = (JHI_CMD_SEND_AND_RECIEVE*) cmd->data;

			if ((cmd_data->SendBuffer_size > JHI_BUFFER_MAX) || (cmd_data->RecvBuffer_size > JHI_BUFFER_MAX))
			{
				res.retCode = JHI_INVALID_BUFFER_SIZE;
				break;
			}

			if ((cmd_data->SendBuffer_size == 0) && (inputSize != (sizeof(JHI_COMMAND) + sizeof(JHI_CMD_SEND_AND_RECIEVE) + cmd_data->SendBuffer_size -1)))
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			if ((cmd_data->SendBuffer_size > 0) && (inputSize != (sizeof(JHI_COMMAND) + sizeof(JHI_CMD_SEND_AND_RECIEVE) + cmd_data->SendBuffer_size -2)))
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			// build JVM_COMM_BUFFER
			IOBuffer.TxBuf->length = cmd_data->SendBuffer_size;
			IOBuffer.TxBuf->buffer = (IOBuffer.TxBuf->length != 0 ? cmd_data->data : NULL);
			IOBuffer.RxBuf->length = cmd_data->RecvBuffer_size;
			IOBuffer.RxBuf->buffer = NULL;

			if (cmd_data->RecvBuffer_size > 0)
			{
				IOBuffer.RxBuf->buffer = (uint8_t*) JHI_ALLOC(cmd_data->RecvBuffer_size);
				if (IOBuffer.RxBuf->buffer == NULL)
				{
					TRACE0("malloc of IOBuffer.RxBuf->buffer failed .");
					return;
				}
				memset(IOBuffer.RxBuf->buffer,0,cmd_data->RecvBuffer_size);
			}

			res.retCode = jhis_txrx_raw(&cmd_data->SessionID, cmd_data->CommandId, &IOBuffer, &res_data.ResponseCode);

			res_data.RecvBuffer_size = IOBuffer.RxBuf->length; // return real response buffer size


			if (res.retCode == JHI_SUCCESS)
			{
				res.dataLength = sizeof(JHI_RESPONSE) + sizeof(JHI_RES_SEND_AND_RECIEVE) + res_data.RecvBuffer_size;
			}
			else
			{
				res.dataLength = sizeof(JHI_RESPONSE) + sizeof(JHI_RES_SEND_AND_RECIEVE);
			}

		}
		while(0);


		*outputData = (uint8_t*) JHI_ALLOC(res.dataLength);
		if (*outputData == NULL) {
			TRACE0("malloc of outputData failed .");
			if (IOBuffer.RxBuf->buffer)
			{
				JHI_DEALLOC(IOBuffer.RxBuf->buffer);
				IOBuffer.RxBuf->buffer = NULL;
			}
			return;
		}
	
		*((JHI_RESPONSE*)(*outputData)) = res;

		if (res.retCode != JHI_INTERNAL_ERROR && res.retCode != JHI_INVALID_BUFFER_SIZE)
		{
			*((JHI_RES_SEND_AND_RECIEVE*)((*((JHI_RESPONSE*)(*outputData))).data)) = res_data;

			// Copy only if returned size can fit the given buffer
			if (res.retCode == JHI_SUCCESS &&
			    res_data.RecvBuffer_size > 0 &&
			    res_data.RecvBuffer_size <= cmd_data->RecvBuffer_size)
			{
				uint8_t* pBufferData = (*((JHI_RES_SEND_AND_RECIEVE*)((*((JHI_RESPONSE*)(*outputData))).data))).data;
				memcpy_s(pBufferData,res_data.RecvBuffer_size,IOBuffer.RxBuf->buffer,res_data.RecvBuffer_size);
			}
		}

		if (IOBuffer.RxBuf->buffer)
		{
			JHI_DEALLOC(IOBuffer.RxBuf->buffer);
			IOBuffer.RxBuf->buffer = NULL;
		}

		*outputSize = res.dataLength;
	}


	void CommandDispatcher::InvokeGetAppletProperty(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize)
	{
		const JHI_COMMAND* cmd = (JHI_COMMAND*) inputData;
		JHI_RESPONSE res;
		JHI_RES_GET_APPLET_PROPERTY res_data;
		JHI_CMD_GET_APPLET_PROPERTY* cmd_data = NULL;
		res.dataLength = sizeof(JHI_RESPONSE);
		res_data.RecvBuffer_size = 0;
		JVM_COMM_BUFFER IOBuffer;

		char* pAppid = NULL;
		UINT8 ucAppID[LEN_APP_ID+1];

		IOBuffer.TxBuf->buffer = NULL;
		IOBuffer.TxBuf->length = 0;
		IOBuffer.RxBuf->buffer = NULL;
		IOBuffer.RxBuf->length = 0;

		memset(&res_data,0,sizeof(JHI_RES_GET_APPLET_PROPERTY));

		do
		{
			if (cmd->dataLength != inputSize)
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			if (inputSize < (sizeof(JHI_COMMAND) + sizeof(JHI_CMD_GET_APPLET_PROPERTY) -1))
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			cmd_data = (JHI_CMD_GET_APPLET_PROPERTY*) cmd->data;

			if (cmd_data->SendBuffer_size > JHI_BUFFER_MAX || cmd_data->RecvBuffer_size > JHI_BUFFER_MAX)
			{
				res.retCode = JHI_INVALID_BUFFER_SIZE;
				break;
			}

			if ((cmd_data->SendBuffer_size == 0) && (inputSize != (sizeof(JHI_COMMAND) + sizeof(JHI_CMD_GET_APPLET_PROPERTY) + cmd_data->SendBuffer_size -1)))
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			if ((cmd_data->SendBuffer_size > 0) && (inputSize != (sizeof(JHI_COMMAND) + sizeof(JHI_CMD_GET_APPLET_PROPERTY) + cmd_data->SendBuffer_size -2)))
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			// build JVM_COMM_BUFFER
			IOBuffer.TxBuf->length = cmd_data->SendBuffer_size;
			IOBuffer.TxBuf->buffer = cmd_data->data;
			IOBuffer.RxBuf->length = cmd_data->RecvBuffer_size;
			IOBuffer.RxBuf->buffer = NULL;

			if (IOBuffer.TxBuf->length > JHI_BUFFER_MAX || IOBuffer.RxBuf->length > JHI_BUFFER_MAX)
			{
				res.retCode = JHI_INVALID_BUFFER_SIZE;
				break;
			}

			if (cmd_data->RecvBuffer_size > 0)
			{
				IOBuffer.RxBuf->buffer = (uint8_t*) JHI_ALLOC(cmd_data->RecvBuffer_size);
				if (IOBuffer.RxBuf->buffer == NULL)
				{
					TRACE0("malloc of IOBuffer.RxBuf->buffer failed .");
					return;
				}
			}

			pAppid = (char*) (cmd_data->AppId);

			if (pAppid[LEN_APP_ID] != '\0')
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			// convet to uppercase and verify Applet id
			if (!convertAppIDtoUpperCase(pAppid,ucAppID))
			{
				res.retCode = JHI_INTERNAL_ERROR;
				break;
			}

			if (verifyAppID((char*)ucAppID) != JHI_SUCCESS)
			{
				res.retCode = JHI_INVALID_APPLET_GUID;
				break;
			}

			res.retCode = jhis_get_applet_property((char*)ucAppID,&IOBuffer);

			res_data.RecvBuffer_size = IOBuffer.RxBuf->length; // return real response buffer size

			if (res.retCode == JHI_SUCCESS)
			{
				res.dataLength = sizeof(JHI_RESPONSE) + sizeof(JHI_RES_GET_APPLET_PROPERTY) + res_data.RecvBuffer_size;
			}
			else
			{
				res.dataLength = sizeof(JHI_RESPONSE) + sizeof(JHI_RES_GET_APPLET_PROPERTY);
			}

		}
		while(0);


		*outputData = (uint8_t*) JHI_ALLOC(res.dataLength);
		if (*outputData == NULL) {
			TRACE0("malloc of outputData failed .");
			if (IOBuffer.RxBuf->buffer)
			{
				JHI_DEALLOC(IOBuffer.RxBuf->buffer);
				IOBuffer.RxBuf->buffer = NULL;
			}
			return;
		}

		*((JHI_RESPONSE*)(*outputData)) = res;

		if (res.retCode == JHI_SUCCESS || res.retCode == JHI_INSUFFICIENT_BUFFER)
		{
			*((JHI_RES_GET_APPLET_PROPERTY*)((*((JHI_RESPONSE*)(*outputData))).data)) = res_data;
			// Copy only if returned size can fit in the given buffer
			if (res.retCode == JHI_SUCCESS &&
			    res_data.RecvBuffer_size > 0 &&
			    res_data.RecvBuffer_size <= cmd_data->RecvBuffer_size)
			{
				uint8_t* pBufferData = (*((JHI_RES_GET_APPLET_PROPERTY*)((*((JHI_RESPONSE*)(*outputData))).data))).data;
				memcpy_s(pBufferData,res_data.RecvBuffer_size,IOBuffer.RxBuf->buffer,res_data.RecvBuffer_size);
			}
		}

		if (IOBuffer.RxBuf->buffer)
		{
			JHI_DEALLOC(IOBuffer.RxBuf->buffer);
			IOBuffer.RxBuf->buffer = NULL;
		}

		*outputSize = res.dataLength;
	}
	
	void CommandDispatcher::InvokeOpenSDSession(const uint8_t* inputData, uint32_t inputSize, uint8_t** outputData,uint32_t* outputSize)
	{
		const JHI_COMMAND* cmd = (JHI_COMMAND*) inputData;
		JHI_RESPONSE res;
		JHI_RES_CREATE_SD_SESSION res_data;
		JHI_CMD_CREATE_SD_SESSION* cmd_data = NULL;
		res.dataLength = sizeof(JHI_RESPONSE);
		char* uuid = NULL;
		JHI_VM_TYPE vmType = GlobalsManager::Instance().getVmType();

		memset(&res_data, 0, sizeof(JHI_RES_CREATE_SD_SESSION));

		do
		{
			if (vmType != JHI_VM_TYPE_BEIHAI_V2)
			{
				res.retCode = TEE_STATUS_UNSUPPORTED_PLATFORM;
				break;
			}

			if (cmd->dataLength != inputSize)
			{
				res.retCode = TEE_STATUS_INTERNAL_ERROR;
				break;
			}

			if (inputSize < (sizeof(JHI_COMMAND) -1 + sizeof(JHI_CMD_CREATE_SD_SESSION)))
			{
				res.retCode = TEE_STATUS_INTERNAL_ERROR;
				break;
			}

			cmd_data = (JHI_CMD_CREATE_SD_SESSION*) cmd->data;

			uuid = (char*) (cmd_data->sdId);

			if (uuid[LEN_APP_ID] != '\0')
			{
				res.retCode = TEE_STATUS_INTERNAL_ERROR;
				break;
			}

			string sdId = string(uuid);

			if (!validateUuidString(sdId))
			{
				res.retCode = TEE_STATUS_INVALID_UUID;
				break;
			}
			
			VM_SESSION_HANDLE sdHandle = NULL;
			VM_Plugin_interface* plugin = NULL;
			if ( (!GlobalsManager::Instance().getPluginTable(&plugin)) || (plugin == NULL) )
			{
				// probably a reset
				res.retCode = TEE_STATUS_NO_FW_CONNECTION;		
			}
			else
			{
				res.retCode = plugin->JHI_Plugin_OpenSDSession(sdId, &sdHandle);
			}

			if (res.retCode == TEE_STATUS_SUCCESS)
			{
				if (sdHandle == NULL)
				{
					res.dataLength = sizeof(JHI_RESPONSE);
					res.retCode = TEE_STATUS_INTERNAL_ERROR;
				}
				res_data.sdHandle = (uint64_t) sdHandle;


				res.dataLength = sizeof(JHI_RESPONSE) -1 + sizeof(JHI_RES_CREATE_SD_SESSION);
			}
			else
			{
				res.dataLength = sizeof(JHI_RESPONSE);
			}
		}
		while(0);

		*outputData = (uint8_t*) JHI_ALLOC(res.dataLength);
		if (*outputData == NULL) {
			TRACE0("malloc of outputData failed .");
			return;
		}

		*((JHI_RESPONSE*)(*outputData)) = res;

		if (res.retCode == TEE_STATUS_SUCCESS)
		{
			*((JHI_RES_CREATE_SD_SESSION*)((*((JHI_RESPONSE*)(*outputData))).data)) = res_data;
		}
		*outputSize = res.dataLength;
	}

	void CommandDispatcher::InvokeCloseSDSession(const uint8_t* inputData, uint32_t inputSize, uint8_t** outputData,uint32_t* outputSize)
	{
		const JHI_COMMAND* cmd = (JHI_COMMAND*) inputData;
		JHI_RESPONSE res;
		JHI_CMD_CLOSE_SD_SESSION* cmd_data = NULL;
		res.dataLength = sizeof(JHI_RESPONSE);
		JHI_VM_TYPE vmType = GlobalsManager::Instance().getVmType();

		do
		{
			if (vmType != JHI_VM_TYPE_BEIHAI_V2)
			{
				res.retCode = TEE_STATUS_UNSUPPORTED_PLATFORM;
				break;
			}

			if (cmd->dataLength != inputSize)
			{
				res.retCode = TEE_STATUS_INTERNAL_ERROR;
				break;
			}

			if (inputSize < (sizeof(JHI_COMMAND) + sizeof(JHI_CMD_CLOSE_SD_SESSION) -1))
			{
				res.retCode = TEE_STATUS_INTERNAL_ERROR;
				break;
			}

			cmd_data = (JHI_CMD_CLOSE_SD_SESSION*) cmd->data;

			if (cmd_data->sdHandle == 0)
			{
				res.retCode = TEE_STATUS_INTERNAL_ERROR;
				break;
			}

			VM_Plugin_interface* plugin = NULL;
			if ( (!GlobalsManager::Instance().getPluginTable(&plugin)) || (plugin == NULL) )
			{
				// probably a reset
				res.retCode = TEE_STATUS_NO_FW_CONNECTION;		
			}
			else
			{
				res.retCode = plugin->JHI_Plugin_CloseSDSession((VM_SESSION_HANDLE*)&cmd_data->sdHandle);
			}
		}
		while(0);

		*outputData = (uint8_t*) JHI_ALLOC(res.dataLength);
		if (*outputData == NULL) {
			TRACE0("malloc of outputData failed .");
			return;
		}

		*((JHI_RESPONSE*)(*outputData)) = res;

		*outputSize = res.dataLength;
	}

	void CommandDispatcher::InvokeSendCmdPkg(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize)
	{
		const JHI_COMMAND* cmd = (JHI_COMMAND*) inputData;
		JHI_RESPONSE res;
		JHI_CMD_SEND_CMD_PKG* cmdPkg = NULL;
		res.dataLength = sizeof(JHI_RESPONSE);
		JHI_VM_TYPE vmType = GlobalsManager::Instance().getVmType();

		do
		{
			if (vmType != JHI_VM_TYPE_BEIHAI_V2)
			{
				res.retCode = TEE_STATUS_UNSUPPORTED_PLATFORM;
				break;
			}

			if (cmd->dataLength != inputSize)
			{
				res.retCode = TEE_STATUS_INTERNAL_ERROR;
				break;
			}

			cmdPkg = (JHI_CMD_SEND_CMD_PKG*) cmd->data;

			if ( (cmdPkg->blobSize == 0) || (cmdPkg->sdHandle == 0) )
			{
				res.retCode = TEE_STATUS_INTERNAL_ERROR;
				break;
			}

			if (inputSize != sizeof(JHI_COMMAND) -1 + sizeof(JHI_CMD_SEND_CMD_PKG) -1 + cmdPkg->blobSize)
			{
				res.retCode = TEE_STATUS_INTERNAL_ERROR;
				break;
			}

			//construct the vector to pass down.
			vector<uint8_t> blob(cmdPkg->blobSize);
			std::copy(&cmdPkg->blob[0], &cmdPkg->blob[0] + cmdPkg->blobSize, blob.begin());

			res.retCode = jhis_send_cmd_pkg((VM_SESSION_HANDLE)cmdPkg->sdHandle, blob);
		}
		while(0);

		*outputData = (uint8_t*) JHI_ALLOC(res.dataLength);
		if (*outputData == NULL) {
			TRACE0("malloc of outputData failed .");
			return;
		}

		*((JHI_RESPONSE*)(*outputData)) = res;

		*outputSize = res.dataLength;
	}


	void CommandDispatcher::InvokeListInstalledTAs(const uint8_t* inputData, uint32_t inputSize, uint8_t** outputData, uint32_t* outputSize)
	{
		const JHI_COMMAND* cmd = (JHI_COMMAND*) inputData;
		JHI_RESPONSE res;
		JHI_RES_LIST_INSTALLED_TAS res_data;		
		JHI_CMD_LIST_INSTALLED_TAS* cmd_data = NULL;
		res.dataLength = sizeof(JHI_RESPONSE);
		vector<string> uuids;
		JHI_VM_TYPE vmType = GlobalsManager::Instance().getVmType();

		memset(&res_data,0,sizeof(JHI_RES_LIST_INSTALLED_TAS));

		do
		{
			if (vmType != JHI_VM_TYPE_BEIHAI_V2)
			{
				res.retCode = TEE_STATUS_UNSUPPORTED_PLATFORM;
				break;
			}

			if (cmd->dataLength != inputSize)
			{
				res.retCode = TEE_STATUS_INTERNAL_ERROR;
				break;
			}

			if (inputSize < (sizeof(JHI_COMMAND) -1 + sizeof(JHI_CMD_LIST_INSTALLED_TAS)))
			{
				res.retCode = TEE_STATUS_INTERNAL_ERROR;
				break;
			}

			cmd_data = (JHI_CMD_LIST_INSTALLED_TAS*) cmd->data;

			if (cmd_data->sdHandle == 0)
			{
				res.retCode = TEE_STATUS_INTERNAL_ERROR;
				break;
			}

			VM_Plugin_interface* plugin = NULL;
			if ( (!GlobalsManager::Instance().getPluginTable(&plugin)) || (plugin == NULL) )
			{
				// probably a reset
				res.retCode = TEE_STATUS_NO_FW_CONNECTION;		
			}
			else
			{
				res.retCode = plugin->JHI_Plugin_ListInstalledTAs((VM_SESSION_HANDLE)cmd_data->sdHandle, uuids);
			}

			if (res.retCode == TEE_STATUS_SUCCESS)
			{
				res_data.count = (uint32_t)uuids.size();
				res.dataLength = (sizeof(JHI_RESPONSE) -1) + (sizeof(JHI_RES_LIST_INSTALLED_TAS) -1) + res_data.count * (LEN_APP_ID +1) + 1;
			}
			else
			{
				res.dataLength = sizeof(JHI_RESPONSE);
			}
		}
		while(0);

		*outputData = (uint8_t*) JHI_ALLOC(res.dataLength);
		if (*outputData == NULL) {
			TRACE0("malloc of outputData failed .");
			return;
		}

		JHI_RESPONSE* jhiOutputData = (JHI_RESPONSE*)(*outputData);
		*jhiOutputData = res;

		if (res.retCode == TEE_STATUS_SUCCESS)
		{
			// Build the output buffer.
			// The buffer containg all the UUIDs (including their null termination) concatenated one after the other.
			memcpy_s(jhiOutputData->data, sizeof(res_data), &res_data, sizeof(res_data)); // copy internal struct.

			if (res.retCode == TEE_STATUS_SUCCESS && res_data.count > 0)
			{
				char* pBufferData = (char*)((JHI_RES_LIST_INSTALLED_TAS*)(jhiOutputData->data))->data; //set pointer to the internal data in the struct.
				for (uint32_t i = 0; i < res_data.count; ++i)
				{
					strcpy_s(pBufferData, LEN_APP_ID + 1, uuids.at(i).c_str()); //stuff the uuids one after the other w/o anything between them.
					pBufferData += LEN_APP_ID + 1; 
				}
				*pBufferData = '\0'; // null at the end.
			}
		}
		*outputSize = res.dataLength;
	}

	void CommandDispatcher::InvokeListInstalledSDs(const uint8_t* inputData, uint32_t inputSize, uint8_t** outputData, uint32_t* outputSize)
	{
		const JHI_COMMAND* cmd = (JHI_COMMAND*)inputData;
		JHI_RESPONSE res;
		JHI_RES_LIST_INSTALLED_SDS res_data;
		JHI_CMD_LIST_INSTALLED_SDS* cmd_data = NULL;
		res.dataLength = sizeof(JHI_RESPONSE);
		vector<string> uuids;
		JHI_VM_TYPE vmType = GlobalsManager::Instance().getVmType();

		memset(&res_data, 0, sizeof(JHI_RES_LIST_INSTALLED_SDS));

		do
		{
			if (vmType != JHI_VM_TYPE_BEIHAI_V2)
			{
				res.retCode = TEE_STATUS_UNSUPPORTED_PLATFORM;
				break;
			}

			if (cmd->dataLength != inputSize)
			{
				res.retCode = TEE_STATUS_INTERNAL_ERROR;
				break;
			}

			if (inputSize < (sizeof(JHI_COMMAND) - 1 + sizeof(JHI_CMD_LIST_INSTALLED_SDS)))
			{
				res.retCode = TEE_STATUS_INTERNAL_ERROR;
				break;
			}

			cmd_data = (JHI_CMD_LIST_INSTALLED_SDS*)cmd->data;

			if (cmd_data->sdHandle == 0)
			{
				res.retCode = TEE_STATUS_INTERNAL_ERROR;
				break;
			}

			VM_Plugin_interface* plugin = NULL;
			if ((!GlobalsManager::Instance().getPluginTable(&plugin)) || (plugin == NULL))
			{
				// probably a reset
				res.retCode = TEE_STATUS_NO_FW_CONNECTION;
			}
			else
			{
				res.retCode = plugin->JHI_Plugin_ListInstalledSDs((VM_SESSION_HANDLE)cmd_data->sdHandle, uuids);
			}

			if (res.retCode == TEE_STATUS_SUCCESS)
			{
				res_data.count = (uint32_t)uuids.size();
				res.dataLength = (sizeof(JHI_RESPONSE) - 1) + (sizeof(JHI_RES_LIST_INSTALLED_SDS) - 1) + res_data.count * (LEN_APP_ID + 1) + 1;
			}
			else
			{
				res.dataLength = sizeof(JHI_RESPONSE);
			}
		} while (0);

		*outputData = (uint8_t*)JHI_ALLOC(res.dataLength);
		if (*outputData == NULL) {
			TRACE0("malloc of outputData failed .");
			return;
		}

		JHI_RESPONSE* jhiOutputData = (JHI_RESPONSE*)(*outputData);
		*jhiOutputData = res;

		if (res.retCode == TEE_STATUS_SUCCESS)
		{
			// Build the output buffer.
			// The buffer containg all the UUIDs (including their null termination) concatenated one after the other.
			memcpy_s(jhiOutputData->data, sizeof(res_data), &res_data, sizeof(res_data)); // copy internal struct.

			if (res.retCode == TEE_STATUS_SUCCESS && res_data.count > 0)
			{
				char* pBufferData = (char*)((JHI_RES_LIST_INSTALLED_SDS*)(jhiOutputData->data))->data; //set pointer to the internal data in the struct.
				for (uint32_t i = 0; i < res_data.count; ++i)
				{
					strcpy_s(pBufferData, LEN_APP_ID + 1, uuids.at(i).c_str()); //stuff the uuids one after the other w/o anything between them.
					pBufferData += LEN_APP_ID + 1;
				}
				*pBufferData = '\0'; // null at the end.
			}
		}
		*outputSize = res.dataLength;
	}

	void CommandDispatcher::InvokeGetVersionInfo(const uint8_t* inputData, uint32_t inputSize, uint8_t** outputData, uint32_t* outputSize)
	{
		JHI_VERSION_INFO info;
		JHI_VERSION_INFO* infoPtr = NULL;
		JHI_RESPONSE res;
		res.dataLength = sizeof(JHI_RESPONSE) + sizeof(JHI_VERSION_INFO);

		if (inputSize != sizeof(JHI_COMMAND))
		{
			res.retCode = JHI_INTERNAL_ERROR;
		}
		else
		{
			GlobalsManager::Instance().getFwVersionString(info.fw_version);
			strcpy_s(info.jhi_version,VERSION_BUFFER_SIZE, VER_PRODUCTVERSION_STR);

			TEE_TRANSPORT_TYPE transport = GlobalsManager::Instance().getTransportType();
			if (transport != TEE_TRANSPORT_TYPE_SOCKET)
			{
				info.comm_type = JHI_HECI;
			}
			else
			{
				info.comm_type = JHI_SOCKETS;
			}

			info.platform_id = GlobalsManager::Instance().getPlatformId();

			info.vm_type = GlobalsManager::Instance().getVmType();

			res.retCode = JHI_SUCCESS;
		}

		*outputData = (uint8_t*) JHI_ALLOC(res.dataLength);
		if (*outputData == NULL) {
			TRACE0("malloc of outputData failed .");
			return;
		}

		*((JHI_RESPONSE*)(*outputData)) = res;

		if (res.retCode == JHI_SUCCESS)
		{
			infoPtr = (JHI_VERSION_INFO*) (*((JHI_RESPONSE*)(*outputData))).data;
			memcpy_s(infoPtr,sizeof(JHI_VERSION_INFO),&info,sizeof(JHI_VERSION_INFO));
		}

		*outputSize = res.dataLength;
	}

	void CommandDispatcher::InvokeQueryTeeMetadata(const uint8_t* inputData, uint32_t inputSize, uint8_t** outputData, uint32_t* outputSize)
	{
		unsigned char* metadata = NULL;
		unsigned int length = 0;
		const JHI_COMMAND* cmd = (JHI_COMMAND*)inputData;
		JHI_RESPONSE res = {0};
		JHI_RES_QUERY_TEE_METADATA res_data = {0};
		res.dataLength = sizeof(JHI_RESPONSE);
		JHI_VM_TYPE vmType = GlobalsManager::Instance().getVmType();

		do
		{
			if (vmType != JHI_VM_TYPE_BEIHAI_V2)
			{
				res.retCode = TEE_STATUS_UNSUPPORTED_PLATFORM;
				break;
			}

			if (outputData == NULL || outputSize == NULL)
			{
				TRACE0("InvokeQueryTeeMetadata ERROR: Invalid params. OutputData or OutputSize are NULL");
				return;
			}

			if (cmd->dataLength != inputSize)
			{
				res.retCode = TEE_STATUS_INTERNAL_ERROR;
				break;
			}

			if (inputSize < sizeof(JHI_COMMAND))
			{
				res.retCode = TEE_STATUS_INTERNAL_ERROR;
				break;
			}

			VM_Plugin_interface* plugin = NULL;
			if ((!GlobalsManager::Instance().getPluginTable(&plugin)) || (plugin == NULL))
			{
				// probably a reset
				res.retCode = TEE_STATUS_NO_FW_CONNECTION;
			}
			else
			{
				res.retCode = plugin->JHI_Plugin_QueryTeeMetadata(&metadata, &length);
			}

			if (res.retCode == TEE_STATUS_SUCCESS)
			{
				res_data.length = length;
				res.dataLength = (sizeof(JHI_RESPONSE) - 1) + (sizeof(JHI_RES_QUERY_TEE_METADATA) - 1) + res_data.length;
			}
			else
			{
				res.dataLength = sizeof(JHI_RESPONSE);
			}
		} while (0);

		// Output
		*outputData = (uint8_t*)JHI_ALLOC(res.dataLength);
		if (*outputData == NULL) {
			TRACE0("malloc of outputData failed.");
			return;
		}

		JHI_RESPONSE* jhiOutputData = (JHI_RESPONSE*)(*outputData);
		*jhiOutputData = res;

		if (res.retCode == TEE_STATUS_SUCCESS)
		{
			memcpy_s(jhiOutputData->data, sizeof(res_data), &res_data, sizeof(res_data)); // copy internal struct.
			char* pBufferData = (char*)((JHI_RES_QUERY_TEE_METADATA*)(jhiOutputData->data))->metadata; //set pointer to the internal data in the struct.
			memcpy_s(pBufferData, res_data.length, metadata, res_data.length);
			JHI_DEALLOC(metadata);
		}
		*outputSize = res.dataLength;
	}
}