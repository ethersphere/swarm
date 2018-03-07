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

#include "CommandInvoker.h"

using std::string;
namespace intel_dal
{
	CommandInvoker::CommandInvoker()
	{
		CommandsClientFactory ccfactory;
		client = ccfactory.createInstance();
	}

	CommandInvoker::~CommandInvoker()
	{
		if (client != NULL)
		{
			// is there a problem with inheritance?
			JHI_DEALLOC_T(client);
			client = NULL;
		}
	}

	bool CommandInvoker::InvokeCommand(const uint8_t* inputBuffer,uint32_t inputBufferSize,uint8_t** outputBuffer,uint32_t* outputBufferSize)
	{
		if (!client->Connect())
		{
			TRACE0("CommandInvoker: Failed connect to JHI service");
			return false;
		}

		if (!client->Invoke(inputBuffer,inputBufferSize,outputBuffer,outputBufferSize))
		{
			TRACE0("CommandInvoker: Send Command failed\n");
			client->Disconnect();
			return false;
		}

		if (!client->Disconnect())
		{
			TRACE0("CommandInvoker: Failed disconnect from server");
			return false;
		}

		return true;
	}

	JHI_RET CommandInvoker::JhisInit()
	{
		JHI_RET ret = JHI_INTERNAL_ERROR;
		JHI_RESPONSE* res;
		JHI_COMMAND cmd;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize;

		cmd.id = INIT;
		cmd.dataLength = sizeof(JHI_COMMAND);

		do
		{
			if (!InvokeCommand((const uint8_t*) &cmd,cmd.dataLength,&outputBuffer,&outputBufferSize))
			{
				ret = JHI_SERVICE_UNAVAILABLE;
				break;
			}

			if (outputBufferSize != sizeof(JHI_RESPONSE))
				break;

			res = (JHI_RESPONSE*) outputBuffer;

			if (outputBufferSize != res->dataLength)
				break;

			ret = res->retCode;
		}
		while (0);

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}

		return ret;
	}

	JHI_RET CommandInvoker::JhisInstall(char* AppId,const FILECHAR* pSrcFile)
	{
		//*****************Command Buffer*******************//
		// JHI_COMMAND | JHI_CMD_INSTALL | pSrcFile         //

		JHI_RET ret = JHI_INTERNAL_ERROR;
		JHI_RESPONSE* res;
		JHI_COMMAND cmd;
		uint8_t* inputBuffer = NULL;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize;

		JHI_CMD_INSTALL install_data;
		install_data.SrcFile_size = (uint32_t)((FILECHARLEN(pSrcFile) + 1) * sizeof(FILECHAR));

		cmd.id = INSTALL;
		cmd.dataLength = sizeof(JHI_COMMAND) + sizeof(JHI_CMD_INSTALL) + install_data.SrcFile_size - 2;

		do
		{
			// build the command buffer
			inputBuffer = (uint8_t*)JHI_ALLOC(cmd.dataLength);
			if (inputBuffer == NULL)
			{
				TRACE0("CommandInvoker: failed to allocate inputBuffer memory.");
				ret = JHI_INTERNAL_ERROR;
				break;
			}
			//fill the buffer
			*((JHI_COMMAND*) inputBuffer) = cmd;

			JHI_CMD_INSTALL* install = (JHI_CMD_INSTALL*)(((JHI_COMMAND*) inputBuffer)->data);
			*install = install_data;

			memcpy_s(install->AppId,LEN_APP_ID+1,AppId,LEN_APP_ID+1);

			uint8_t* srcfile = install->data;
			memcpy_s(srcfile,install_data.SrcFile_size,pSrcFile,install_data.SrcFile_size);

			// send the command buffer
			if (!InvokeCommand(inputBuffer,cmd.dataLength,&outputBuffer,&outputBufferSize))
			{
				ret = JHI_SERVICE_UNAVAILABLE;
				break;
			}

			if (outputBufferSize != sizeof(JHI_RESPONSE))
				break;

			res = (JHI_RESPONSE*) outputBuffer;

			if (outputBufferSize != res->dataLength)
				break;

			ret = res->retCode;
		}
		while (0);

		// cleanup:
		if (inputBuffer)
		{
			JHI_DEALLOC(inputBuffer);
			inputBuffer = NULL;
		}

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}

		return ret;
	}

	JHI_RET CommandInvoker::JhisUninstall(char* AppId)
	{
		//*****************Command Buffer*******************//
		// JHI_COMMAND | JHI_CMD_UNINSTALL                  //

		JHI_RET ret = JHI_INTERNAL_ERROR;
		JHI_RESPONSE* res;
		JHI_COMMAND cmd;
		uint8_t* inputBuffer = NULL;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize;

		JHI_CMD_UNINSTALL uninstall_data;
		memcpy_s(uninstall_data.AppId,LEN_APP_ID+1,AppId,LEN_APP_ID+1);

		cmd.id = UNINSTALL;
		cmd.dataLength = sizeof(JHI_COMMAND)+sizeof(JHI_CMD_UNINSTALL) -1;

		do
		{
			// build the command buffer
			inputBuffer = (uint8_t*)JHI_ALLOC(cmd.dataLength);
			if (inputBuffer == NULL)
			{
				TRACE0("CommandInvoker: failed to allocate inputBuffer memory.");
				ret = JHI_INTERNAL_ERROR;
				break;
			}
			
			//fill the buffer
			*((JHI_COMMAND*) inputBuffer) = cmd;

			JHI_CMD_UNINSTALL* uninstall = (JHI_CMD_UNINSTALL*)(((JHI_COMMAND*) inputBuffer)->data);
			memcpy_s(uninstall,sizeof(JHI_CMD_UNINSTALL),&uninstall_data,sizeof(JHI_CMD_UNINSTALL));

			// send the command buffer
			if (!InvokeCommand(inputBuffer,cmd.dataLength,&outputBuffer,&outputBufferSize))
			{
				ret = JHI_SERVICE_UNAVAILABLE;
				break;
			}

			if (outputBufferSize != sizeof(JHI_RESPONSE))
				break;

			res = (JHI_RESPONSE*) outputBuffer;

			if (outputBufferSize != res->dataLength)
				break;

			ret = res->retCode;
		}
		while (0);

		// cleanup:
		if (inputBuffer)
		{
			JHI_DEALLOC(inputBuffer);
			inputBuffer = NULL;
		}

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}

		return ret;
	}

	JHI_RET CommandInvoker::JhisGetSessionsCount(char* AppId, uint32_t* pSessionCount)
	{
		//*******************Command Buffer*********************//
		// JHI_COMMAND | JHI_CMD_GET_SESSIONS_COUNT             //

		//******************Response Buffer*********************//
		// JHI_RESPONSE | JHI_RES_GET_SESSIONS_COUNT            //

		JHI_RET ret = JHI_INTERNAL_ERROR;
		JHI_RESPONSE* res;
		JHI_COMMAND cmd;
		uint8_t* inputBuffer = NULL;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize;

		JHI_CMD_GET_SESSIONS_COUNT command_data;

		memset(&command_data,0,sizeof(JHI_CMD_GET_SESSIONS_COUNT));
		memcpy_s(command_data.AppId,LEN_APP_ID+1,AppId,LEN_APP_ID+1);

		cmd.id = GET_SESSIONS_COUNT;
		cmd.dataLength = sizeof(JHI_COMMAND)+sizeof(JHI_CMD_GET_SESSIONS_COUNT) -1;

		do
		{
			// build the command buffer
			inputBuffer = (uint8_t*)JHI_ALLOC(cmd.dataLength);
			if (inputBuffer == NULL)
			{
				TRACE0("CommandInvoker: failed to allocate inputBuffer memory.");
				ret = JHI_INTERNAL_ERROR;
				break;
			}
			
			//fill the buffer
			*((JHI_COMMAND*) inputBuffer) = cmd;

			JHI_CMD_GET_SESSIONS_COUNT* cmd_data = (JHI_CMD_GET_SESSIONS_COUNT*)(((JHI_COMMAND*) inputBuffer)->data);
			memcpy_s(cmd_data,sizeof(JHI_CMD_GET_SESSIONS_COUNT),&command_data,sizeof(JHI_CMD_GET_SESSIONS_COUNT));

			// send the command buffer
			if (!InvokeCommand(inputBuffer,cmd.dataLength,&outputBuffer,&outputBufferSize))
			{
				ret = JHI_SERVICE_UNAVAILABLE;
				break;
			}

			if (outputBufferSize < sizeof(JHI_RESPONSE))
				break;

			res = (JHI_RESPONSE*) outputBuffer;

			if (outputBufferSize != res->dataLength)
				break;

			if (outputBufferSize == (sizeof(JHI_RESPONSE) + sizeof(JHI_RES_GET_SESSIONS_COUNT)))
			{	
				// update OUT parameters
				JHI_RES_GET_SESSIONS_COUNT* res_data = (JHI_RES_GET_SESSIONS_COUNT*) res->data;
				*pSessionCount = res_data->SessionCount;
			}

			ret = res->retCode;
		}
		while (0);

		// cleanup:
		if (inputBuffer)
		{
			JHI_DEALLOC(inputBuffer);
			inputBuffer = NULL;
		}

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}
		
		return ret;
	}

	JHI_RET CommandInvoker::JhisCreateSession(char* AppId, JHI_SESSION_ID* pSessionID,uint32_t flags,DATA_BUFFER* initBuffer,JHI_PROCESS_INFO* processInfo)
	{
		//**********************Command Buffer***********************//
		// JHI_COMMAND | JHI_CMD_CREATE_SESSION | initBuffer         //

		JHI_RET ret = JHI_INTERNAL_ERROR;
		JHI_RESPONSE* res;
		JHI_COMMAND cmd;
		uint8_t* inputBuffer = NULL;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize;

		JHI_CMD_CREATE_SESSION cmd_data;
		cmd_data.flags = flags;
		cmd_data.initBuffer_size = initBuffer->length;
		cmd_data.processInfo = *processInfo;

		cmd.id = CREATE_SESSION;

		if (cmd_data.initBuffer_size == 0)
			cmd.dataLength = sizeof(JHI_COMMAND)+sizeof(JHI_CMD_CREATE_SESSION) + cmd_data.initBuffer_size -1;
		else
			cmd.dataLength = sizeof(JHI_COMMAND)+sizeof(JHI_CMD_CREATE_SESSION) + cmd_data.initBuffer_size -2;

		do
		{
			// build the command buffer
			inputBuffer = (uint8_t*)JHI_ALLOC(cmd.dataLength);
			if (inputBuffer == NULL)
			{
				TRACE0("CommandInvoker: failed to allocate inputBuffer memory.");
				ret = JHI_INTERNAL_ERROR;
				break;
			}
			
			//fill the buffer
			*((JHI_COMMAND*) inputBuffer) = cmd;

			JHI_CMD_CREATE_SESSION* cSession = (JHI_CMD_CREATE_SESSION*)(((JHI_COMMAND*) inputBuffer)->data);
			*cSession = cmd_data;

			memcpy_s(cSession->AppId,LEN_APP_ID+1,AppId,LEN_APP_ID+1);

			if (initBuffer->length > 0)
			{
				uint8_t* pBuffer = cSession->data;
				memcpy_s(pBuffer,initBuffer->length,initBuffer->buffer,initBuffer->length);
			}

			// send the command buffer
			if (!InvokeCommand(inputBuffer,cmd.dataLength,&outputBuffer,&outputBufferSize))
			{
				ret = JHI_SERVICE_UNAVAILABLE;
				break;
			}

			if (outputBufferSize < sizeof(JHI_RESPONSE))
				break;

			res = (JHI_RESPONSE*) outputBuffer;

			if (outputBufferSize != res->dataLength)
				break;

			if (outputBufferSize == sizeof(JHI_RESPONSE) + sizeof(JHI_RES_CREATE_SESSION))
			{
				// update OUT parameters
				JHI_RES_CREATE_SESSION* res_data = (JHI_RES_CREATE_SESSION*) res->data;
				memcpy_s(pSessionID,sizeof(JHI_SESSION_ID),&(res_data->SessionID),sizeof(JHI_SESSION_ID));
			}

			ret = res->retCode;
		}
		while (0);

		// cleanup:
		if (inputBuffer)
		{
			JHI_DEALLOC(inputBuffer);
			inputBuffer = NULL;
		}

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}

		return ret;
	}

	JHI_RET CommandInvoker::JhisCloseSession(JHI_SESSION_ID* SessionID,JHI_PROCESS_INFO* processInfo, bool force)
	{
		//*****************Command Buffer*******************//
		// JHI_COMMAND | JHI_CMD_CLOSE_SESSION              //

		JHI_RET ret = JHI_INTERNAL_ERROR;
		JHI_RESPONSE* res;
		JHI_COMMAND cmd;
		uint8_t* inputBuffer = NULL;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize;

		JHI_CMD_CLOSE_SESSION cmd_data;
		memcpy_s(&cmd_data.SessionID,sizeof(JHI_SESSION_ID),SessionID,sizeof(JHI_SESSION_ID));
		memcpy_s(&cmd_data.processInfo,sizeof(JHI_PROCESS_INFO),processInfo,sizeof(JHI_PROCESS_INFO));
		cmd_data.force = force;

		cmd.id = CLOSE_SESSION;
		cmd.dataLength = sizeof(JHI_COMMAND) + sizeof(JHI_CMD_CLOSE_SESSION) -1;

		do
		{
			// build the command buffer
			inputBuffer = (uint8_t*)JHI_ALLOC(cmd.dataLength);
			if (inputBuffer == NULL)
			{
				TRACE0("CommandInvoker: failed to allocate inputBuffer memory.");
				ret = JHI_INTERNAL_ERROR;
				break;
			}
			
			//fill the buffer
			*((JHI_COMMAND*) inputBuffer) = cmd;

			JHI_CMD_CLOSE_SESSION* cSession = (JHI_CMD_CLOSE_SESSION*)(((JHI_COMMAND*) inputBuffer)->data);
			*cSession = cmd_data;

			// send the command buffer
			if (!InvokeCommand(inputBuffer,cmd.dataLength,&outputBuffer,&outputBufferSize))
			{
				ret = JHI_SERVICE_UNAVAILABLE;
				break;
			}

			if (outputBufferSize != sizeof(JHI_RESPONSE))
				break;

			res = (JHI_RESPONSE*) outputBuffer;

			if (outputBufferSize != res->dataLength)
				break;

			ret = res->retCode;
		}
		while (0);

		// cleanup:
		if (inputBuffer)
		{
			JHI_DEALLOC(inputBuffer);
			inputBuffer = NULL;
		}

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}
		
		return ret;
	}

	JHI_RET CommandInvoker::JhisGetSessionInfo(JHI_SESSION_ID* SessionID, JHI_SESSION_INFO* pSessionInfo)
	{
		//*******************Command Buffer*******************//
		// JHI_COMMAND | JHI_CMD_GET_SESSION_INFO             //

		//******************Response Buffer*******************//
		// JHI_RESPONSE | JHI_RES_GET_SESSION_INFO            //

		JHI_RET ret = JHI_INTERNAL_ERROR;
		JHI_RESPONSE* res;
		JHI_COMMAND cmd;
		uint8_t* inputBuffer = NULL;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize;

		JHI_CMD_GET_SESSION_INFO command_data;
		memcpy_s(&command_data.SessionID, sizeof(JHI_SESSION_ID), SessionID, sizeof(JHI_SESSION_ID));

		cmd.id = GET_SESSION_INFO;
		cmd.dataLength = sizeof(JHI_COMMAND) + sizeof(JHI_CMD_GET_SESSION_INFO) - 1;

		do
		{
			// build the command buffer
			inputBuffer = (uint8_t*)JHI_ALLOC(cmd.dataLength);
			if (inputBuffer == NULL)
			{
				TRACE0("CommandInvoker: failed to allocate inputBuffer memory.");
				ret = JHI_INTERNAL_ERROR;
				break;
			}

			//fill the buffer
			*((JHI_COMMAND*)inputBuffer) = cmd;

			JHI_CMD_GET_SESSION_INFO* cmd_data = (JHI_CMD_GET_SESSION_INFO*)(((JHI_COMMAND*)inputBuffer)->data);
			*cmd_data = command_data;

			// send the command buffer
			if (!InvokeCommand(inputBuffer, cmd.dataLength, &outputBuffer, &outputBufferSize))
			{
				ret = JHI_SERVICE_UNAVAILABLE;
				break;
			}

			if (outputBufferSize < sizeof(JHI_RESPONSE))
				break;

			res = (JHI_RESPONSE*)outputBuffer;

			if (outputBufferSize != res->dataLength)
				break;

			// update OUT parameters

			if (outputBufferSize == (sizeof(JHI_RESPONSE) + sizeof(JHI_RES_GET_SESSION_INFO)))
			{
				JHI_RES_GET_SESSION_INFO* res_data = (JHI_RES_GET_SESSION_INFO*)res->data;
				*pSessionInfo = res_data->SessionInfo;
			}

			ret = res->retCode;
		} while (0);

		// cleanup:
		if (inputBuffer)
		{
			JHI_DEALLOC(inputBuffer);
			inputBuffer = NULL;
		}

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}

		return ret;
	}

	JHI_RET CommandInvoker::JhisSetSessionEventHandler(JHI_SESSION_ID* SessionID, const char* handleName)
	{
		//***********************Command Buffer**************************//
		// JHI_COMMAND | JHI_CMD_SET_SESSION_EVENT_HANDLER | handleName  //

		JHI_RET ret = JHI_INTERNAL_ERROR;
		JHI_RESPONSE* res;
		JHI_COMMAND cmd;
		uint8_t* inputBuffer = NULL;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize;

		JHI_CMD_SET_SESSION_EVENT_HANDLER cmd_data;
		cmd_data.handleName_size = (uint32_t)(strlen(handleName) + 1);
		memcpy_s(&cmd_data.SessionID,sizeof(JHI_SESSION_ID),SessionID,sizeof(JHI_SESSION_ID));

		cmd.id = SET_SESSION_EVENT_HANDLER;
		cmd.dataLength = sizeof(JHI_COMMAND)+sizeof(JHI_CMD_SET_SESSION_EVENT_HANDLER)+cmd_data.handleName_size -2;

		do
		{
			// build the command buffer
			inputBuffer = (uint8_t*)JHI_ALLOC(cmd.dataLength);
			if (inputBuffer == NULL)
			{
				TRACE0("CommandInvoker: failed to allocate inputBuffer memory.");
				ret = JHI_INTERNAL_ERROR;
				break;
			}
			
			//fill the buffer
			*((JHI_COMMAND*) inputBuffer) = cmd;

			JHI_CMD_SET_SESSION_EVENT_HANDLER* pCmd = (JHI_CMD_SET_SESSION_EVENT_HANDLER*)(((JHI_COMMAND*) inputBuffer)->data);
			*pCmd = cmd_data;

			char* pHandleName = (char*) pCmd->data;
			memcpy_s(pHandleName,cmd_data.handleName_size,handleName,cmd_data.handleName_size);

			// send the command buffer
			if (!InvokeCommand(inputBuffer,cmd.dataLength,&outputBuffer,&outputBufferSize))
			{
				ret = JHI_SERVICE_UNAVAILABLE;
				break;
			}

			if (outputBufferSize != sizeof(JHI_RESPONSE))
				break;

			res = (JHI_RESPONSE*) outputBuffer;

			if (outputBufferSize != res->dataLength)
				break;

			ret = res->retCode;
		}
		while (0);

		// cleanup:
		if (inputBuffer)
		{
			JHI_DEALLOC(inputBuffer);
			inputBuffer = NULL;
		}

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}

		return ret;
	}

	JHI_RET CommandInvoker::JhisGetEventData(JHI_SESSION_ID* SessionID, uint32_t* DataBufferSize, uint8_t** pDataBuffer, uint8_t* pDataType)
	{
		//*******************Command Buffer*********************//
		// JHI_COMMAND | JHI_CMD_GET_EVENT_DATA                 //

		//******************Response Buffer*********************//
		// JHI_RESPONSE | JHI_RES_GET_EVENT_DATA | DataBuffer   //

		JHI_RET ret = JHI_INTERNAL_ERROR;
		JHI_RESPONSE* res;
		JHI_COMMAND cmd;
		uint8_t* inputBuffer = NULL;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize;

		JHI_CMD_GET_EVENT_DATA command_data;
		memcpy_s(&command_data.SessionID, sizeof(JHI_SESSION_ID), SessionID, sizeof(JHI_SESSION_ID));

		cmd.id = GET_EVENT_DATA;
		cmd.dataLength = sizeof(JHI_COMMAND) + sizeof(JHI_CMD_GET_EVENT_DATA) - 1;

		do
		{
			// build the command buffer
			inputBuffer = (uint8_t*)JHI_ALLOC(cmd.dataLength);
			if (inputBuffer == NULL)
			{
				TRACE0("CommandInvoker: failed to allocate inputBuffer memory.");
				ret = JHI_INTERNAL_ERROR;
				break;
			}

			//fill the buffer
			*((JHI_COMMAND*)inputBuffer) = cmd;

			JHI_CMD_GET_EVENT_DATA* cmd_data = (JHI_CMD_GET_EVENT_DATA*)(((JHI_COMMAND*)inputBuffer)->data);
			*cmd_data = command_data;

			// send the command buffer
			if (!InvokeCommand(inputBuffer, cmd.dataLength, &outputBuffer, &outputBufferSize))
			{
				ret = JHI_SERVICE_UNAVAILABLE;
				break;
			}

			if (outputBufferSize < sizeof(JHI_RESPONSE))
				break;

			res = (JHI_RESPONSE*)outputBuffer;

			if (outputBufferSize != res->dataLength)
				break;

			if (outputBufferSize < (sizeof(JHI_RESPONSE) + sizeof(JHI_RES_GET_EVENT_DATA)))
			{
				if (outputBufferSize == sizeof(JHI_RESPONSE))
					ret = res->retCode;

				break;
			}

			// update OUT parameters	
			JHI_RES_GET_EVENT_DATA* res_data = (JHI_RES_GET_EVENT_DATA*)res->data;
			*pDataType = res_data->DataType;
			*DataBufferSize = res_data->DataBuffer_size;

			if (outputBufferSize != (sizeof(JHI_RESPONSE) + sizeof(JHI_RES_GET_EVENT_DATA) + *DataBufferSize))
				break;

			ret = res->retCode;

			if (res_data->DataBuffer_size > 0)
			{
				*pDataBuffer = (uint8_t*)JHI_ALLOC(res_data->DataBuffer_size);
				if (*pDataBuffer == NULL)
				{
					TRACE0("CommandInvoker: failed to allocate pDataBuffer memory.");
					ret = JHI_INTERNAL_ERROR;
					break;
				}

				memcpy_s(*pDataBuffer, res_data->DataBuffer_size, res_data->data, res_data->DataBuffer_size);
			}
		} while (0);

		// cleanup:
		if (inputBuffer)
		{
			JHI_DEALLOC(inputBuffer);
			inputBuffer = NULL;
		}

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}
		return ret;
	}

	JHI_RET CommandInvoker::JhisSendAndRecv(JHI_SESSION_ID* SessionID, int32_t CommandId, const uint8_t* SendBuffer, uint32_t SendBufferSize, uint8_t* RecvBuffer, uint32_t* RecvBufferSize, int32_t* responseCode)
	{
		//*******************Command Buffer**********************//
		// JHI_COMMAND  | JHI_CMD_SEND_AND_RECIEVE | SendBuffer  //

		//******************Response Buffer**********************//
		// JHI_RESPONSE | JHI_RES_SEND_AND_RECIEVE | RecvBuffer  //

		JHI_RET ret = JHI_INTERNAL_ERROR;
		JHI_RESPONSE* res;
		JHI_COMMAND cmd;
		uint8_t* inputBuffer = NULL;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize;

		JHI_CMD_SEND_AND_RECIEVE command_data;
		command_data.CommandId = CommandId;
		command_data.SendBuffer_size = SendBufferSize;
		command_data.RecvBuffer_size = *RecvBufferSize;

		memcpy_s(&command_data.SessionID, sizeof(JHI_SESSION_ID), SessionID, sizeof(JHI_SESSION_ID));

		cmd.id = SEND_AND_RECIEVE;

		if (command_data.SendBuffer_size == 0)
			cmd.dataLength = sizeof(JHI_COMMAND) + sizeof(JHI_CMD_SEND_AND_RECIEVE) + command_data.SendBuffer_size - 1;
		else
			cmd.dataLength = sizeof(JHI_COMMAND) + sizeof(JHI_CMD_SEND_AND_RECIEVE) + command_data.SendBuffer_size - 2;

		do
		{
			// build the command buffer
			inputBuffer = (uint8_t*)JHI_ALLOC(cmd.dataLength);
			if (inputBuffer == NULL)
			{
				TRACE0("CommandInvoker: failed to allocate inputBuffer memory.");
				ret = JHI_INTERNAL_ERROR;
				break;
			}

			//fill the buffer
			*((JHI_COMMAND*)inputBuffer) = cmd;

			JHI_CMD_SEND_AND_RECIEVE* cmd_data = (JHI_CMD_SEND_AND_RECIEVE*)(((JHI_COMMAND*)inputBuffer)->data);
			*cmd_data = command_data;

			memcpy_s(cmd_data->data, SendBufferSize, SendBuffer, SendBufferSize);

			// send the command buffer
			if (!InvokeCommand(inputBuffer, cmd.dataLength, &outputBuffer, &outputBufferSize))
			{
				ret = JHI_SERVICE_UNAVAILABLE;
				break;
			}

			if (outputBufferSize < sizeof(JHI_RESPONSE))
				break;

			res = (JHI_RESPONSE*)outputBuffer;

			if (outputBufferSize != res->dataLength)
				break;

			// update OUT parameters
			ret = res->retCode;

			if (outputBufferSize == sizeof(JHI_RESPONSE))
				break;

			if ((ret != JHI_INTERNAL_ERROR) && (ret != JHI_INVALID_BUFFER_SIZE))
			{

				if (outputBufferSize < (sizeof(JHI_RESPONSE) + sizeof(JHI_RES_SEND_AND_RECIEVE)))
				{
					ret = JHI_INTERNAL_ERROR;
					break;
				}

				JHI_RES_SEND_AND_RECIEVE* res_data = (JHI_RES_SEND_AND_RECIEVE*)res->data;

				if ((ret == JHI_SUCCESS) && (outputBufferSize != (sizeof(JHI_RESPONSE) + sizeof(JHI_RES_SEND_AND_RECIEVE) + res_data->RecvBuffer_size)))
				{
					ret = JHI_INTERNAL_ERROR;
					break;
				}
				else if ((ret != JHI_SUCCESS) && (outputBufferSize != (sizeof(JHI_RESPONSE) + sizeof(JHI_RES_SEND_AND_RECIEVE))))
				{
					ret = JHI_INTERNAL_ERROR;
					break;
				}

				if ((ret == JHI_SUCCESS) && (res_data->RecvBuffer_size > 0) && (*RecvBufferSize >= res_data->RecvBuffer_size))
				{
					memcpy_s(RecvBuffer, *RecvBufferSize, res_data->data, res_data->RecvBuffer_size);
				}

				if (ret == JHI_SUCCESS || ret == JHI_INSUFFICIENT_BUFFER)
				{
					*RecvBufferSize = res_data->RecvBuffer_size;
				}

				if (responseCode != NULL)
					*responseCode = res_data->ResponseCode;
			}

		} while (0);

		// cleanup:
		if (inputBuffer)
		{
			JHI_DEALLOC(inputBuffer);
			inputBuffer = NULL;
		}

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}

		return ret;
	}

	JHI_RET CommandInvoker::JhisGetAppletProperty(char* AppId, const uint8_t* SendBuffer, uint32_t SendBufferSize, uint8_t* RecvBuffer, uint32_t* RecvBufferSize)
	{
		//***************************Command Buffer**************************//
		// JHI_COMMAND  | JHI_CMD_GET_APPLET_PROPERTY | SendBuffer           //

		//**************************Response Buffer**************************//
		// JHI_RESPONSE | JHI_RES_GET_APPLET_PROPERTY | RecvBuffer           //

		JHI_RET ret = JHI_INTERNAL_ERROR;
		JHI_RESPONSE* res;
		JHI_COMMAND cmd;
		uint8_t* inputBuffer = NULL;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize;

		JHI_CMD_GET_APPLET_PROPERTY command_data;
		command_data.SendBuffer_size = SendBufferSize + 1;  // convert the sizes from character length to buffer length
		command_data.RecvBuffer_size = *RecvBufferSize + 1; // convert the sizes from character length to buffer length

		cmd.id = GET_APPLET_PROPERTY;

		if (SendBufferSize == 0)
			cmd.dataLength = sizeof(JHI_COMMAND) + sizeof(JHI_CMD_GET_APPLET_PROPERTY) + command_data.SendBuffer_size - 1;
		else
			cmd.dataLength = sizeof(JHI_COMMAND) + sizeof(JHI_CMD_GET_APPLET_PROPERTY) + command_data.SendBuffer_size - 2;
		do
		{
			// build the command buffer
			inputBuffer = (uint8_t*)JHI_ALLOC(cmd.dataLength);
			if (inputBuffer == NULL)
			{
				TRACE0("CommandInvoker: failed to allocate inputBuffer memory.");
				return JHI_INTERNAL_ERROR;
			}

			//fill the buffer
			*((JHI_COMMAND*)inputBuffer) = cmd;

			JHI_CMD_GET_APPLET_PROPERTY* cmd_data = (JHI_CMD_GET_APPLET_PROPERTY*)(((JHI_COMMAND*)inputBuffer)->data);
			*cmd_data = command_data;

			if (SendBufferSize != 0)
			{
				// copy send buffer
				memcpy_s(cmd_data->data, command_data.SendBuffer_size, SendBuffer, command_data.SendBuffer_size);
			}

			// copy appid
			memcpy_s(cmd_data->AppId, LEN_APP_ID + 1, AppId, LEN_APP_ID + 1);

			// send the command buffer
			if (!InvokeCommand(inputBuffer, cmd.dataLength, &outputBuffer, &outputBufferSize))
			{
				ret = JHI_SERVICE_UNAVAILABLE;
				break;
			}

			if (outputBufferSize < sizeof(JHI_RESPONSE))
				break;

			res = (JHI_RESPONSE*)outputBuffer;

			if (outputBufferSize != res->dataLength)
				break;

			// update OUT parameters
			ret = res->retCode;

			if (outputBufferSize == sizeof(JHI_RESPONSE))
				break;

			if ((ret != JHI_INTERNAL_ERROR) && (ret != JHI_INVALID_BUFFER_SIZE))
			{
				if (outputBufferSize < (sizeof(JHI_RESPONSE) + sizeof(JHI_RES_GET_APPLET_PROPERTY)))
				{
					ret = JHI_INTERNAL_ERROR;
					break;
				}

				JHI_RES_GET_APPLET_PROPERTY* res_data = (JHI_RES_GET_APPLET_PROPERTY*)res->data;

				if ((ret == JHI_SUCCESS) && (outputBufferSize != (sizeof(JHI_RESPONSE) + sizeof(JHI_RES_GET_APPLET_PROPERTY) + res_data->RecvBuffer_size)))
				{
					ret = JHI_INTERNAL_ERROR;
					break;
				}
				if ((ret != JHI_SUCCESS) && (outputBufferSize != (sizeof(JHI_RESPONSE) + sizeof(JHI_RES_GET_APPLET_PROPERTY))))
				{
					ret = JHI_INTERNAL_ERROR;
					break;
				}

				if ((ret == JHI_SUCCESS) && (res_data->RecvBuffer_size > 0) && (command_data.RecvBuffer_size >= res_data->RecvBuffer_size) && RecvBuffer != NULL)
				{
					memcpy_s(RecvBuffer, command_data.RecvBuffer_size, res_data->data, res_data->RecvBuffer_size);
				}

				if (ret == JHI_SUCCESS || ret == JHI_INSUFFICIENT_BUFFER)
				{
					*RecvBufferSize = res_data->RecvBuffer_size - 1; // return the length as characters
				}
			}
		} while (0);

		// cleanup:

		if (inputBuffer)
		{
			JHI_DEALLOC(inputBuffer);
			inputBuffer = NULL;
		}

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}
		return ret;
	}

	JHI_RET CommandInvoker::JhisGetVersionInfo(JHI_VERSION_INFO* pVersionInfo)
	{
		//*******************Command Buffer*********************//
		// JHI_COMMAND                                          //

		//******************Response Buffer*********************//
		// JHI_RESPONSE | JHI_VERSION_INFO                      //

		JHI_RET ret = JHI_INTERNAL_ERROR;
		JHI_RESPONSE* res;
		JHI_COMMAND cmd;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize;

		cmd.id = GET_VERSION_INFO;
		cmd.dataLength = sizeof(JHI_COMMAND);

		do
		{
			// send the command buffer
			if (!InvokeCommand((const uint8_t*)&cmd, cmd.dataLength, &outputBuffer, &outputBufferSize))
			{
				ret = JHI_SERVICE_UNAVAILABLE;
				break;
			}


			if (outputBufferSize < sizeof(JHI_RESPONSE))
				break;

			res = (JHI_RESPONSE*)outputBuffer;

			if ((res->retCode == JHI_SUCCESS) && (outputBufferSize != sizeof(JHI_RESPONSE) + sizeof(JHI_VERSION_INFO)))
			{
				// invalid response size for this error code
				break;
			}
			else if ((res->retCode != JHI_SUCCESS) && (outputBufferSize != sizeof(JHI_RESPONSE)))
			{
				// invalid response size for this error code
				break;
			}

			if (outputBufferSize != res->dataLength)
				break;

			ret = res->retCode;

			if (ret == JHI_SUCCESS)
			{
				memcpy_s(pVersionInfo, sizeof(JHI_VERSION_INFO), res->data, sizeof(JHI_VERSION_INFO));
			}

		} while (0);

		// cleanup:

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}

		return ret;
	}
	
	///////////////////////////
	// TeeManagement Methods //
	///////////////////////////

	TEE_STATUS CommandInvoker::JhisOpenSDSession(IN const string& sdId, OUT SD_SESSION_HANDLE*	sdHandle)
	{
		//**********************Command Buffer***********************//
		// JHI_COMMAND | JHI_CMD_CREATE_SD_SESSION | initBuffer         //

		TEE_STATUS ret = TEE_STATUS_INTERNAL_ERROR;
		JHI_RESPONSE* res;
		JHI_COMMAND cmd;
		uint8_t* inputBuffer = NULL;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize;

		cmd.id = CREATE_SD_SESSION;
		cmd.dataLength = sizeof(JHI_COMMAND) -1 + sizeof(JHI_CMD_CREATE_SD_SESSION);

		do
		{
			if (sdHandle == NULL)
			{
				ret = TEE_STATUS_INVALID_PARAMS;
				break;
			}

			// build the command buffer
			inputBuffer = (uint8_t*)JHI_ALLOC(cmd.dataLength);
			if (inputBuffer == NULL)
			{
				TRACE0("CommandInvoker: failed to allocate inputBuffer memory.");
				ret = TEE_STATUS_INTERNAL_ERROR;
				break;
			}
			
			//fill the buffer
			*((JHI_COMMAND*) inputBuffer) = cmd;

			JHI_CMD_CREATE_SD_SESSION* cSession = (JHI_CMD_CREATE_SD_SESSION*)(((JHI_COMMAND*) inputBuffer)->data);

			memcpy_s(cSession->sdId, LEN_APP_ID, sdId.c_str(), LEN_APP_ID);

			cSession->sdId[LEN_APP_ID] = '\0';

			// send the command buffer
			if (!InvokeCommand(inputBuffer, cmd.dataLength, &outputBuffer, &outputBufferSize))
			{
				ret = TEE_STATUS_SERVICE_UNAVAILABLE;
				break;
			}

			if ( (outputBufferSize < sizeof(JHI_RESPONSE)) || (outputBuffer == NULL) )
			{
				ret = TEE_STATUS_INTERNAL_ERROR;
				break;
			}

			res = (JHI_RESPONSE*) outputBuffer;
			if (outputBufferSize != res->dataLength)
			{
				ret = TEE_STATUS_INTERNAL_ERROR;
				break;
			}

			if (outputBufferSize == sizeof(JHI_RESPONSE) -1 + sizeof(JHI_RES_CREATE_SD_SESSION))
			{
				// update OUT parameters
				JHI_RES_CREATE_SD_SESSION* res_data = (JHI_RES_CREATE_SD_SESSION*) res->data;
				*sdHandle = (SD_SESSION_HANDLE*) res_data->sdHandle;
			}

			ret = jhiErrorToTeeError(res->retCode);
		}
		while (0);

		// cleanup:
		if (inputBuffer)
		{
			JHI_DEALLOC(inputBuffer);
			inputBuffer = NULL;
		}

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}

		return ret;
	}

	TEE_STATUS CommandInvoker::JhisCloseSDSession(IN OUT SD_SESSION_HANDLE* sdHandle)
	{
		//**********************Command Buffer***********************//
		// JHI_COMMAND | JHI_CMD_CLOSE_SD_SESSION   //

		TEE_STATUS ret = TEE_STATUS_INTERNAL_ERROR;
		JHI_RESPONSE* res;
		JHI_COMMAND cmd;
		uint8_t* inputBuffer = NULL;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize;


		cmd.id = CLOSE_SD_SESSION;

		cmd.dataLength = sizeof(JHI_COMMAND) -1 + sizeof(JHI_CMD_CLOSE_SD_SESSION);

		do
		{
			if (sdHandle == NULL)
			{
				ret = TEE_STATUS_INVALID_PARAMS;
				break;
			}

			// build the command buffer
			inputBuffer = (uint8_t*)JHI_ALLOC(cmd.dataLength);
			if (inputBuffer == NULL)
			{
				TRACE0("CommandInvoker: failed to allocate inputBuffer memory.");
				ret = TEE_STATUS_INTERNAL_ERROR;
				break;
			}
			
			//fill the buffer
			*((JHI_COMMAND*) inputBuffer) = cmd;

			JHI_CMD_CLOSE_SD_SESSION* cmd_data = (JHI_CMD_CLOSE_SD_SESSION*)(((JHI_COMMAND*) inputBuffer)->data);
			cmd_data->sdHandle = (uint64_t)*sdHandle;

			// send the command buffer
			if (!InvokeCommand(inputBuffer,cmd.dataLength,&outputBuffer,&outputBufferSize))
			{
				ret = TEE_STATUS_SERVICE_UNAVAILABLE;
				break;
			}

			if (outputBufferSize != sizeof(JHI_RESPONSE))
				break;

			res = (JHI_RESPONSE*) outputBuffer;

			if (outputBufferSize != res->dataLength)
				break;

			ret = jhiErrorToTeeError(res->retCode);

			if (ret == TEE_STATUS_SUCCESS)
			{
				*sdHandle = NULL;
			}
		}
		while (0);

		// cleanup:
		if (inputBuffer)
		{
			JHI_DEALLOC(inputBuffer);
			inputBuffer = NULL;
		}

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}

		return ret;
	}

	TEE_STATUS CommandInvoker::JhisSendAdminCmdPkg(IN const SD_SESSION_HANDLE sdHandle, IN const uint8_t* package, IN uint32_t packageSize)
	{
		//*****************Command Buffer*******************//
		// JHI_COMMAND | JHI_CMD_SEND_CMD_PKG | pSrcFile //

		TEE_STATUS ret = TEE_STATUS_INTERNAL_ERROR;
		JHI_RESPONSE* res;
		JHI_COMMAND cmd;
		uint8_t* inputBuffer = NULL;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize;

		JHI_CMD_SEND_CMD_PKG send_cmd_pkg_data;
		send_cmd_pkg_data.blobSize = packageSize;

		cmd.id = SEND_CMD_PKG;
		cmd.dataLength = sizeof(JHI_COMMAND) -1 + sizeof(JHI_CMD_SEND_CMD_PKG) -1 + send_cmd_pkg_data.blobSize;

		do
		{
			// build the command buffer
			inputBuffer = (uint8_t*)JHI_ALLOC(cmd.dataLength);
			if (inputBuffer == NULL)
			{
				TRACE0("CommandInvoker: failed to allocate inputBuffer memory.");
				ret = TEE_STATUS_INTERNAL_ERROR;
				break;
			}
			//fill the buffer
			*((JHI_COMMAND*) inputBuffer) = cmd;

			JHI_CMD_SEND_CMD_PKG* cmdPkg = (JHI_CMD_SEND_CMD_PKG*)(((JHI_COMMAND*) inputBuffer)->data);
			cmdPkg->blobSize = send_cmd_pkg_data.blobSize;

			cmdPkg->sdHandle = (uint64_t)sdHandle;

			uint8_t* blob = cmdPkg->blob;
			memcpy_s(blob, send_cmd_pkg_data.blobSize, package, send_cmd_pkg_data.blobSize);			

			// send the command buffer
			if (!InvokeCommand(inputBuffer, cmd.dataLength, &outputBuffer, &outputBufferSize))
			{
				ret = TEE_STATUS_SERVICE_UNAVAILABLE;
				break;
			}

			if (outputBufferSize != sizeof(JHI_RESPONSE))
				break;

			res = (JHI_RESPONSE*) outputBuffer;

			if (outputBufferSize != res->dataLength)
				break;

			ret = jhiErrorToTeeError(res->retCode);
		}
		while (0);

		// cleanup:
		if (inputBuffer)
		{
			JHI_DEALLOC(inputBuffer);
			inputBuffer = NULL;
		}

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}

		return ret;
	}

	TEE_STATUS CommandInvoker::JhisListInstalledTAs(IN SD_SESSION_HANDLE sdHandle, OUT	UUID_LIST* uuidList)
	{
		//**********************Command Buffer***********************//
		// JHI_COMMAND | JHI_CMD_LIST_INSTALLED_TAS | initBuffer     //

		TEE_STATUS ret = TEE_STATUS_INTERNAL_ERROR;
		JHI_RESPONSE* res;
		JHI_COMMAND cmd;
		uint8_t* inputBuffer = NULL;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize;
		uint32_t dataBufferLen = 0;
		JHI_RES_LIST_INSTALLED_TAS* resData = NULL;
		JHI_CMD_LIST_INSTALLED_TAS* cmd_data = NULL;

		if (uuidList == NULL)
		{
			ret = TEE_STATUS_INTERNAL_ERROR;
			goto error;
		}
		uuidList->uuids = NULL;

		cmd.id = LIST_INSTALLED_TAS;

		cmd.dataLength = sizeof(JHI_COMMAND) -1 + sizeof(JHI_CMD_LIST_INSTALLED_TAS);

		// build the command buffer
		inputBuffer = (uint8_t*)JHI_ALLOC(cmd.dataLength);
		if (inputBuffer == NULL)
		{
			TRACE0("CommandInvoker: failed to allocate inputBuffer memory.");
			ret = TEE_STATUS_INTERNAL_ERROR;
			goto error;
		}

		//fill the buffer
		*((JHI_COMMAND*) inputBuffer) = cmd;

		cmd_data = (JHI_CMD_LIST_INSTALLED_TAS*)(((JHI_COMMAND*) inputBuffer)->data);
		cmd_data->sdHandle = (uint64_t)sdHandle;

		// send the command buffer
		if (!InvokeCommand(inputBuffer, cmd.dataLength, &outputBuffer, &outputBufferSize))
		{
			ret = TEE_STATUS_SERVICE_UNAVAILABLE;
			goto error;
		}

		// validate buffer
		if ( (outputBufferSize < sizeof(JHI_RESPONSE)) || (outputBuffer == NULL) )
		{
			ret = TEE_STATUS_INTERNAL_ERROR;
			goto error;
		}

		res = (JHI_RESPONSE*) outputBuffer;

		if (res->retCode != JHI_SUCCESS)
		{
			ret = jhiErrorToTeeError(res->retCode);
			goto error;
		}

		if (outputBufferSize != res->dataLength)
		{
			ret = TEE_STATUS_INTERNAL_ERROR;
			goto error;
		}

		// The buffer containg all the UUIDs (including their null termination) concatenated one after the other.
		resData = (JHI_RES_LIST_INSTALLED_TAS*)res->data;
		dataBufferLen = res->dataLength - (sizeof(JHI_RESPONSE) - 1) - (sizeof(JHI_RES_LIST_INSTALLED_TAS) - 1) -1; // The length of the inner data.

		if (dataBufferLen != UUID_LEN * resData->count)
		{
			ret = TEE_STATUS_INTERNAL_ERROR;
			goto error;
		}

		uuidList->uuidCount = resData->count;
		uuidList->uuids = (UUID_STR*) JHI_ALLOC(res->dataLength);
		memcpy_s(uuidList->uuids, res->dataLength, resData->data, res->dataLength);

		// verify the uuids
		if (!validateUuidList(uuidList))
		{
			ret = TEE_STATUS_INTERNAL_ERROR;
			goto error;
		}

		// success
		ret = (TEE_STATUS) res->retCode;
		goto cleanup;

error:
		if (uuidList)
		{
			uuidList->uuidCount = 0;
			if (uuidList->uuids)
			{
				JHI_DEALLOC(uuidList->uuids);
			}
		}
cleanup:
		if (inputBuffer)
		{
			JHI_DEALLOC(inputBuffer);
			inputBuffer = NULL;
		}

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}

		return ret;
	}

	TEE_STATUS CommandInvoker::JhisListInstalledSDs(IN SD_SESSION_HANDLE sdHandle, OUT	UUID_LIST* uuidList)
	{
		//**********************Command Buffer***********************//
		// JHI_COMMAND | JHI_CMD_LIST_INSTALLED_SDS | initBuffer     //

		TEE_STATUS ret = TEE_STATUS_INTERNAL_ERROR;
		JHI_RESPONSE* res;
		JHI_COMMAND cmd;
		uint8_t* inputBuffer = NULL;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize;
		uint32_t dataBufferLen = 0;
		JHI_RES_LIST_INSTALLED_SDS* resData = NULL;
		JHI_CMD_LIST_INSTALLED_SDS* cmd_data = NULL;

		if (uuidList == NULL)
		{
			ret = TEE_STATUS_INTERNAL_ERROR;
			goto error;
		}
		uuidList->uuids = NULL;

		cmd.id = LIST_INSTALLED_SDS;

		cmd.dataLength = sizeof(JHI_COMMAND) - 1 + sizeof(JHI_CMD_LIST_INSTALLED_SDS);

		// build the command buffer
		inputBuffer = (uint8_t*)JHI_ALLOC(cmd.dataLength);
		if (inputBuffer == NULL)
		{
			TRACE0("CommandInvoker: failed to allocate inputBuffer memory.");
			ret = TEE_STATUS_INTERNAL_ERROR;
			goto error;
		}

		//fill the buffer
		*((JHI_COMMAND*)inputBuffer) = cmd;

		cmd_data = (JHI_CMD_LIST_INSTALLED_SDS*)(((JHI_COMMAND*)inputBuffer)->data);
		cmd_data->sdHandle = (uint64_t)sdHandle;

		// send the command buffer
		if (!InvokeCommand(inputBuffer, cmd.dataLength, &outputBuffer, &outputBufferSize))
		{
			ret = TEE_STATUS_SERVICE_UNAVAILABLE;
			goto error;
		}

		// validate buffer
		if ((outputBufferSize < sizeof(JHI_RESPONSE)) || (outputBuffer == NULL))
		{
			ret = TEE_STATUS_INTERNAL_ERROR;
			goto error;
		}

		res = (JHI_RESPONSE*)outputBuffer;

		if (res->retCode != JHI_SUCCESS)
		{
			ret = jhiErrorToTeeError(res->retCode);
			goto error;
		}

		if (outputBufferSize != res->dataLength)
		{
			ret = TEE_STATUS_INTERNAL_ERROR;
			goto error;
		}

		// The buffer containg all the UUIDs (including their null termination) concatenated one after the other.
		resData = (JHI_RES_LIST_INSTALLED_SDS*)res->data;
		dataBufferLen = res->dataLength - (sizeof(JHI_RESPONSE) - 1) - (sizeof(JHI_RES_LIST_INSTALLED_SDS) - 1) - 1; // The length of the inner data.
		if (
			(dataBufferLen != UUID_LEN * resData->count) ||
			(resData->data[dataBufferLen - 1] != '\0')
			)
		{
			ret = TEE_STATUS_INTERNAL_ERROR;
			goto error;
		}

		uuidList->uuidCount = resData->count;
		uuidList->uuids = (UUID_STR*)JHI_ALLOC(res->dataLength);
		memcpy_s(uuidList->uuids, res->dataLength, resData->data, res->dataLength);

		// verify the uuids
		if (!validateUuidList(uuidList))
		{
			ret = TEE_STATUS_INTERNAL_ERROR;
			goto error;
		}

		// success
		ret = (TEE_STATUS)res->retCode;
		goto cleanup;

	error:
		if (uuidList)
		{
			uuidList->uuidCount = 0;
			if (uuidList->uuids)
			{
				JHI_DEALLOC(uuidList->uuids);
			}
		}
	cleanup:
		if (inputBuffer)
		{
			JHI_DEALLOC(inputBuffer);
			inputBuffer = NULL;
		}

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}

		return ret;
	}
	
	TEE_STATUS CommandInvoker::JhisQueryTEEMetadata(OUT dal_tee_metadata* metadata, size_t max_length)
	{
		//**********************Response Buffer**************************//
		// JHI_RESPONSE | JHI_RES_QUERY_TEE_METADATA | dal_tee_metadata  //

		TEE_STATUS ret = TEE_STATUS_INTERNAL_ERROR;
		JHI_RESPONSE* res = NULL;
		JHI_COMMAND cmd = {0};
		uint8_t* inputBuffer = NULL;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize = 0;
		size_t dataBufferLen = 0;
		JHI_RES_QUERY_TEE_METADATA* resData = NULL;

		cmd.id = QUERY_TEE_METADATA;

		cmd.dataLength = sizeof(JHI_COMMAND);

		// build the command buffer
		inputBuffer = (uint8_t*)JHI_ALLOC(cmd.dataLength);
		if (inputBuffer == NULL)
		{
			TRACE0("CommandInvoker: failed to allocate inputBuffer memory.");
			ret = TEE_STATUS_INTERNAL_ERROR;
			goto error;
		}

		//fill the buffer
		*((JHI_COMMAND*) inputBuffer) = cmd;

		// send the command buffer
		if (!InvokeCommand(inputBuffer, cmd.dataLength, &outputBuffer, &outputBufferSize))
		{
			ret = TEE_STATUS_SERVICE_UNAVAILABLE;
			goto error;
		}

		// validate buffer
		if ((outputBufferSize < sizeof(JHI_RESPONSE)) || (outputBuffer == NULL))
		{
			ret = TEE_STATUS_INTERNAL_ERROR;
			goto error;
		}

		res = (JHI_RESPONSE*) outputBuffer;

		if (res->retCode != JHI_SUCCESS)
		{
			ret = jhiErrorToTeeError(res->retCode);
			goto error;
		}

		if (outputBufferSize != res->dataLength)
		{
			ret = TEE_STATUS_INTERNAL_ERROR;
			goto error;
		}

		resData = (JHI_RES_QUERY_TEE_METADATA*)res->data;
		dataBufferLen = res->dataLength - (sizeof(JHI_RESPONSE) - 1) - (sizeof(JHI_RES_QUERY_TEE_METADATA) - 1); // The length of the inner data.

		if (dataBufferLen >= max_length) // We are copying by the caller's expected size so we have to check that we have enough data to fill it
			memcpy_s(metadata, max_length, resData->metadata, max_length);
		else
		{
			ret = TEE_STATUS_INTERNAL_ERROR;
			TRACE0("JhisQueryTEEMetadata failed. Received data is shorter than expected");
			goto error;
		}

		if (dataBufferLen > max_length)
			TRACE2("JhisQueryTEEMetadata - Warning - Data truncated because of size mismatch. Expected: %d, Received: %d", max_length, dataBufferLen);

		// Success
		ret = (TEE_STATUS) res->retCode;

	error:
		if (inputBuffer)
		{
			JHI_DEALLOC(inputBuffer);
			inputBuffer = NULL;
		}

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}

		return ret;
	}

#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
	JHI_RET CommandInvoker::JhisGetSessionTable(JHI_SESSIONS_DATA_TABLE** SessionDataTable)
	{
		JHI_RET ret = JHI_INTERNAL_ERROR;
		JHI_RESPONSE* res;
		JHI_COMMAND cmd;
		uint8_t* inputBuffer = NULL;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize;

		cmd.id = GET_SESSIONS_DATA_TABLE;
		cmd.dataLength = sizeof(JHI_COMMAND);

		do
		{
			// build the command buffer
			inputBuffer = (uint8_t*)JHI_ALLOC(cmd.dataLength);
			if (inputBuffer == NULL)
			{
				TRACE0("CommandInvoker: failed to allocate inputBuffer memory.");
				ret = JHI_INTERNAL_ERROR;
				break;
			}
			
			//fill the buffer
			*((JHI_COMMAND*) inputBuffer) = cmd;

			// send the command buffer
			if (!InvokeCommand(inputBuffer,cmd.dataLength,&outputBuffer,&outputBufferSize))
			{
				ret = JHI_SERVICE_UNAVAILABLE;
				break;
			}

			//validating
			if (outputBufferSize < sizeof(JHI_RESPONSE) + sizeof(JHI_RES_GET_SESSIONS_DATA_TABLE))
				break;

			res = (JHI_RESPONSE*) outputBuffer;

			if (res->retCode != JHI_SUCCESS)
			{
				ret = res->retCode;
				break;
			}

			if (outputBufferSize != res->dataLength)
				break;

			uint32_t tableSize = res->dataLength - sizeof(JHI_RESPONSE);

			// update OUT parameters
			JHI_RES_GET_SESSIONS_DATA_TABLE* res_data = (JHI_RES_GET_SESSIONS_DATA_TABLE*) JHI_ALLOC(tableSize);

			memcpy_s(res_data, tableSize, res->data, tableSize);
			if (res_data->SessionDataTable.sessionsCount == 0)
			{
				*SessionDataTable = &res_data->SessionDataTable;
				ret = res->retCode;
				break;
			}

			// calculates sizes
			int sessionInfoSize = sizeof(JHI_SESSION_EXTENDED_INFO) * res_data->SessionDataTable.sessionsCount;

			/** arrange all the pointers from in outputbuffer **/

			//calculates where the sessions info starts
			JHI_SESSION_EXTENDED_INFO* sessionsOffset = (JHI_SESSION_EXTENDED_INFO*) ((uint8_t*)(res_data) + sizeof(JHI_RES_GET_SESSIONS_DATA_TABLE));
			res_data->SessionDataTable.dataTable = sessionsOffset;

			//calculates where the owners lists should start
			JHI_PROCESS_INFORMATION* OwnersListsOffset = (JHI_PROCESS_INFORMATION*)((uint8_t*)sessionsOffset + sessionInfoSize); //(starts at the end of the sessions)

			// updates the owners lists pointers
			for ( UINT32 i = 0; i < res_data->SessionDataTable.sessionsCount ; ++i )
			{
				res_data->SessionDataTable.dataTable[i].ownersList = OwnersListsOffset;
				OwnersListsOffset += res_data->SessionDataTable.dataTable[i].ownersListCount; //updates the offset
			}

			*SessionDataTable = &res_data->SessionDataTable;
			ret = res->retCode;
		}
		while (0);

		// cleanup:
		if (inputBuffer)
		{
			JHI_DEALLOC(inputBuffer);
			inputBuffer = NULL;
		}

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}

		return ret;
	}

	JHI_RET CommandInvoker::JhisGetLoadedAppletsList(JHI_LOADED_APPLET_GUIDS** appGUIDs)
	{
		JHI_RET ret = JHI_INTERNAL_ERROR;
		JHI_RESPONSE* res;
		JHI_COMMAND cmd;
		uint8_t* inputBuffer = NULL;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize;

		cmd.id = GET_LOADED_APPLETS;
		cmd.dataLength = sizeof(JHI_COMMAND);

		do
		{
			// build the command buffer
			inputBuffer = (uint8_t*)JHI_ALLOC(cmd.dataLength);
			if (inputBuffer == NULL)
			{
				TRACE0("CommandInvoker: failed to allocate inputBuffer memory.");
				ret = JHI_INTERNAL_ERROR;
				break;
			}
			
			//fill the buffer
			*((JHI_COMMAND*) inputBuffer) = cmd;

			// send the command buffer
			if (!InvokeCommand(inputBuffer,cmd.dataLength,&outputBuffer,&outputBufferSize))
			{
				ret = JHI_SERVICE_UNAVAILABLE;
				break;
			}

			//validating
			if (outputBufferSize < sizeof(JHI_RESPONSE) + sizeof(JHI_RES_GET_LOADED_APPLETS))
				break;

			res = (JHI_RESPONSE*) outputBuffer;
			if (res->retCode != JHI_SUCCESS)
			{
				ret = res->retCode;
				break;
			}

			if (outputBufferSize != res->dataLength)
				break;

			JHI_RES_GET_LOADED_APPLETS* res_data = ((JHI_RES_GET_LOADED_APPLETS*)res->data);
			JHI_LOADED_APPLET_GUIDS* loadedAppletsCopy = JHI_ALLOC_T(JHI_LOADED_APPLET_GUIDS);

			//fixing the pointer
			*loadedAppletsCopy = res_data->loadedApplets;

			//allocating the pointers of the GUIDs
			loadedAppletsCopy->appsGUIDs = JHI_ALLOC_T_ARRAY<char*>(res_data->loadedApplets.loadedAppletsCount);

			//the GUID offset in the received output buffer
			char* GUIDs = (char*)((uint8_t*)res_data + sizeof(JHI_RES_GET_LOADED_APPLETS));

			//allocated & copying the GUIDs
			for (UINT32 i = 0; i < res_data->loadedApplets.loadedAppletsCount; ++i)
			{
				loadedAppletsCopy->appsGUIDs[i] = (char*) JHI_ALLOC(LEN_APP_ID + 1);
				strcpy_s(loadedAppletsCopy->appsGUIDs[i], LEN_APP_ID + 1, GUIDs);
				//updating the offset in the received output buffer
				GUIDs += (LEN_APP_ID + 1);
			}

			// update OUT parameters
			*appGUIDs = loadedAppletsCopy;
			ret = res->retCode;
		}
		while (0);

		// cleanup:
		if (inputBuffer)
		{
			JHI_DEALLOC(inputBuffer);
			inputBuffer = NULL;
		}

		if (outputBuffer)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}

		return ret;
	}

#endif
}