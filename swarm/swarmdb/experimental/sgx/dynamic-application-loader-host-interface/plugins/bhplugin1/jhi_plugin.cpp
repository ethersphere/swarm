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
**
**    @file jhi_plugin.h
**
**    @brief  Defines BEIHAI Client plugin implementation
**
**    @author Elad Dabool
**
********************************************************************************
*/
#include "beihai.h"
#include "jhi_plugin.h"
#include "dbg.h"
#include "teemanagement.h"

#ifndef _WIN32
#include "string_s.h"
#endif //_WIN32

#include <sstream>
using namespace std;

// NOTE: Enable this defintion in order to activate JHI memory profiling
//#define JHI_MEMORY_PROFILING

#ifdef JHI_MEMORY_PROFILING
    #define PROFILING_ARGS ,__FILE__, __LINE__
#else
    #define PROFILING_ARGS
#endif // JHI_MEMORY_PROFILING

//------------------------------------------------------------------------------
// first-time register of plugin callbacks
//------------------------------------------------------------------------------
UINT32 pluginRegister(VM_Plugin_interface** plugin)
{
	TRACE0("pluginRegister start");
	JHI_RET ulRetCode = JHI_INVALID_PARAMS ;

	if (plugin == NULL)
		goto end;

	*plugin = &Jhi_Plugin::BeihaiPlugin::Instance();

	ulRetCode = JHI_SUCCESS;

end:
	TRACE1("pluginRegister end, result = 0x%X", ulRetCode);
	return ulRetCode ;
}

namespace Jhi_Plugin
{
	TEE_TRANSPORT_INTERFACE BeihaiPlugin::transport_interface = {0};

	BeihaiPlugin::BeihaiPlugin()
	{
		memset (&memory_api, 0, sizeof(JHI_PLUGIN_MEMORY_API));
		memset (&bh_transport_APIs, 0, sizeof(BH_PLUGIN_TRANSPORT));
		memset (&transport_interface, 0, sizeof(TEE_TRANSPORT_INTERFACE));
		plugin_type = JHI_PLUGIN_TYPE_BEIHAI_V1;
	}


	int BeihaiPlugin::sendWrapper(uintptr_t handle, uint8_t* buffer, uint32_t length)
	{
		return (int)transport_interface.pfnSend(&transport_interface, (TEE_TRANSPORT_HANDLE)handle, (const uint8_t*)buffer, (size_t)length);
	}


	int BeihaiPlugin::recvWrapper(uintptr_t  handle, uint8_t* buffer, uint32_t* length)
	{
		return (int)transport_interface.pfnRecv(&transport_interface, (TEE_TRANSPORT_HANDLE)handle, (uint8_t*)buffer, length);
	}

	int BeihaiPlugin::closeWrapper(uintptr_t handle)
	{
		return (int)transport_interface.pfnDisconnect(&transport_interface, (TEE_TRANSPORT_HANDLE*)&handle);
	}
	UINT32 BeihaiPlugin::JHI_Plugin_GetPluginType()
	{
		return plugin_type;
	}

	UINT32 BeihaiPlugin::JHI_Plugin_Set_Transport_And_Memory(unsigned int transportType, JHI_PLUGIN_MEMORY_API* plugin_memory_api)
	{
		JHI_RET ulRetCode = JHI_INVALID_PARAMS ;

		if (plugin_memory_api == NULL)
			goto end;

		memset (&memory_api, 0, sizeof(JHI_PLUGIN_MEMORY_API));
		BeihaiPlugin::memory_api = *plugin_memory_api;

		memset (&bh_transport_APIs, 0, sizeof(BH_PLUGIN_TRANSPORT));
		memset (&transport_interface, 0, sizeof(TEE_TRANSPORT_INTERFACE));

		ulRetCode = TEE_Transport_Create((TEE_TRANSPORT_TYPE)transportType, &transport_interface);
		if (ulRetCode != TEE_COMM_SUCCESS)
		{
			return JHI_INTERNAL_ERROR;
		}

		TEE_TRANSPORT_ENTITY transport_entity;
		transport_entity = TEE_TRANSPORT_ENTITY_IVM;
		if (transportType == TEE_TRANSPORT_TYPE_SOCKET)
		{
			// When using sockets instead of HECI, this is the right port number
			transport_entity = TEE_TRANSPORT_ENTITY_RTM;
		}

		ulRetCode = transport_interface.pfnConnect(&transport_interface, transport_entity, NULL, (TEE_TRANSPORT_HANDLE*)&bh_transport_APIs.handle);
		if (ulRetCode != TEE_COMM_SUCCESS)
		{
			return JHI_COMMS_ERROR;
		}

		bh_transport_APIs.pfnSend = sendWrapper;
		bh_transport_APIs.pfnRecv = recvWrapper;
		bh_transport_APIs.pfnClose = closeWrapper;
		//bh_transport_APIs.handle = plugin_transport->handle;

		ulRetCode = JHI_SUCCESS;

end:
		return ulRetCode ;
	}

	UINT32 BeihaiPlugin::JHI_Plugin_Init(bool do_vm_reset)
	{
		BH_ERRNO bhRet = BH_PluginInit(&bh_transport_APIs, do_vm_reset);
		return JhiErrorTranslate(bhRet, JHI_INTERNAL_ERROR);
	}

	UINT32 BeihaiPlugin::JHI_Plugin_DeInit(bool do_vm_reset)
	{
		BH_ERRNO bhRet = BH_PluginDeinit();

		//deinit the transport

		int ret2 = transport_interface.pfnDisconnect(&transport_interface, (TEE_TRANSPORT_HANDLE*)&bh_transport_APIs.handle);
		if (ret2 != TEE_COMM_SUCCESS)
		{
			TRACE1("transport_interface Teardown error, result = 0x%X", ret2);
		}

		ret2 = transport_interface.pfnTeardown(&transport_interface);
		if (ret2 != TEE_COMM_SUCCESS || transport_interface.state != TEE_INTERFACE_STATE_NOT_INITIALIZED)
		{
			TRACE1("transport_interface Teardown error, result = 0x%X", ret2);
		}

		return JhiErrorTranslate(bhRet, JHI_INTERNAL_ERROR);
	}

	UINT32 BeihaiPlugin::JHI_Plugin_DownloadApplet(const char *pAppId, uint8_t* pAppBlob, unsigned int AppSize)
	{
		BH_ERRNO bhRet = BH_PluginDownload(pAppId, (char*) pAppBlob, AppSize);
		return JhiErrorTranslate(bhRet, JHI_INTERNAL_ERROR);
	}

	UINT32 BeihaiPlugin::JHI_Plugin_UnloadApplet(const char *AppId)
	{
		BH_ERRNO bhRet = BH_PluginUnload(const_cast<char*>(AppId));
		return JhiErrorTranslate(bhRet, JHI_INTERNAL_ERROR);
	}

	UINT32 BeihaiPlugin::JHI_Plugin_OpenSDSession (const string& SD_ID, VM_SESSION_HANDLE* pSession)
	{
		return TEE_STATUS_UNSUPPORTED_PLATFORM;
	}

	UINT32 BeihaiPlugin::JHI_Plugin_CloseSDSession (VM_SESSION_HANDLE* pSession)
	{
		return TEE_STATUS_UNSUPPORTED_PLATFORM;
	}

	UINT32 BeihaiPlugin::JHI_Plugin_ListInstalledTAs (const VM_SESSION_HANDLE handle, vector<string>& UUIDs)
	{
		return TEE_STATUS_UNSUPPORTED_PLATFORM;
	}

	UINT32 BeihaiPlugin::JHI_Plugin_ListInstalledSDs(const VM_SESSION_HANDLE handle, vector<string>& UUIDs)
	{
		return TEE_STATUS_UNSUPPORTED_PLATFORM;
	}

	UINT32 BeihaiPlugin::JHI_Plugin_SendCmdPkg (const VM_SESSION_HANDLE handle, vector<uint8_t>& blob)
	{
		return TEE_STATUS_UNSUPPORTED_PLATFORM;
	}

	UINT32 BeihaiPlugin::JHI_Plugin_ParsePackage(uint8_t* cmd_pkg, uint32_t pkg_len, OUT PACKAGE_INFO& pkgInfo)
	{
		return TEE_STATUS_UNSUPPORTED_PLATFORM;
	}

	UINT32 BeihaiPlugin::JHI_Plugin_QueryTeeMetadata(unsigned char** metadata, unsigned int* length)
	{
		return TEE_STATUS_UNSUPPORTED_PLATFORM;
	}

	bool BeihaiPlugin::convertAppProperty_Version(char** output)
	{
		try
		{
			string version = string(*output);

			size_t index = version.rfind('.');
			if (index == string::npos)
				return false;

			string majorSTR = version.substr(0, index);
			string minorSTR = version.substr(index + 1, version.length() -1);

			istringstream majorStreamSTR(majorSTR);
			istringstream minorStreamSTR(minorSTR);

			unsigned int majorUINT;
			unsigned int minorUINT;
			majorStreamSTR >> majorUINT;
			minorStreamSTR >> minorUINT;

			if ( (majorUINT > 255) || (minorUINT > 255) )
			{
				return false;	// not valid
			}
			minorUINT = minorUINT << 8;
			unsigned int versionUINT = majorUINT | minorUINT;

			BH_FREE(*output);
			*output = NULL;

			*output = (char*)memory_api.allocateMemory(6 PROFILING_ARGS);

#ifndef _WIN32
			snprintf(*output, 6, "%d", versionUINT);
#else
			_itoa_s(versionUINT, *output, 6, 10);
#endif
			return true;
		}
		catch (...)
		{
			return false;
		}
	}

	UINT32 BeihaiPlugin::JHI_Plugin_GetAppletProperty(const char *AppId, JVM_COMM_BUFFER *pIOBuffer)
	{
		UINT32 jhiRet = JHI_INTERNAL_ERROR;
		char* inputBuffer = (char*) pIOBuffer->TxBuf->buffer;
		int inputBufferLength = pIOBuffer->TxBuf->length;
		string AppProperty_Version = "applet.version";
		bool versionQuery = false;
		char* output = NULL;
		int outputLength = 0;

		char* outputBuffer = (char*) pIOBuffer->RxBuf->buffer;
		int* outputBufferLength = (int*) &pIOBuffer->RxBuf->length; // number of characters without /0, not size of buffer

		BH_ERRNO bhRet = BH_PluginQueryAPI(const_cast<char*>(AppId), inputBuffer, inputBufferLength, &output);

		if (bhRet == BH_SUCCESS && output != NULL)
		{
			if (AppProperty_Version == inputBuffer)
			{
				versionQuery = convertAppProperty_Version(&output); // convert to unsigned int like in TL.
			}

			outputLength = (int)strlen(output);

			if (*outputBufferLength < outputLength)
			{
				// buffer provided is too small for the response
				TRACE2("JHI_Plugin_GetAppletProperty: insufficient buffer sent to VM, expected: %d, received: %d\n", outputLength, *outputBufferLength);
				jhiRet = JHI_INSUFFICIENT_BUFFER;
				*outputBufferLength = outputLength;
				goto cleanup;
			}

			// copy the output to the output buffer
			strcpy_s(outputBuffer, *outputBufferLength + 1, output);
			*outputBufferLength = outputLength;
		}
		else
		{
			*outputBufferLength = 0;
		}

		jhiRet = JhiErrorTranslate(bhRet, JHI_INTERNAL_ERROR);

cleanup:

		if (output)
		{
			if (versionQuery)
				memory_api.freeMemory(output PROFILING_ARGS);
			else
				BH_FREE(output);
		}
		return jhiRet;
	}

	UINT32 BeihaiPlugin::sendSessionIDtoApplet(VM_SESSION_HANDLE* pSession, JHI_SESSION_ID SessionID, int *appletResponse)
	{
		// TODO: BH bug w/a - unable to send null output buffer.
		char temp[] = "output\0";
		char *pOutput = temp;
		int outputLength = 0;
		char Uuid[sizeof(JHI_SESSION_ID)];

		memcpy_s(Uuid,sizeof(JHI_SESSION_ID),&SessionID,sizeof(JHI_SESSION_ID));
		// the value '1' in the 'what' field is internally reserved for passing the SessionID
		BH_ERRNO bhRet = BH_PluginSendAndRecvInternal( *pSession, 1, 0, Uuid, sizeof(JHI_SESSION_ID), (void**)&pOutput, (unsigned int *)&outputLength, appletResponse);

		return JhiErrorTranslate(bhRet, JHI_INTERNAL_ERROR);
	}

	UINT32 BeihaiPlugin::JHI_Plugin_CreateSession (const char *AppId, VM_SESSION_HANDLE* pSession, const uint8_t* pAppBlob, unsigned int BlobSize, JHI_SESSION_ID SessionID,DATA_BUFFER* initBuffer)
	{
		BH_ERRNO bhRet = BH_PluginCreateSession (const_cast<char*>(AppId), pSession, (char*)initBuffer->buffer, initBuffer->length);
		if (bhRet == BH_SUCCESS)
		{
			// sending the SessionID to the applet.
			int appletResponse = -1;
			UINT32 jhiRet = sendSessionIDtoApplet(pSession, SessionID, &appletResponse);
			if (jhiRet != JHI_SUCCESS || appletResponse != 0)
				return JHI_INTERNAL_ERROR;
		}

		return JhiErrorTranslate(bhRet, JHI_INTERNAL_ERROR);
	}

	UINT32 BeihaiPlugin::JHI_Plugin_ForceCloseSession(VM_SESSION_HANDLE* pSession)
	{
		BH_ERRNO bhRet = BH_PluginForceCloseSession(*pSession);
		return JhiErrorTranslate(bhRet, JHI_INTERNAL_ERROR);
	}


	UINT32 BeihaiPlugin::JHI_Plugin_CloseSession(VM_SESSION_HANDLE* pSession)
	{
		BH_ERRNO bhRet = BH_PluginCloseSession(*pSession);

		return JhiErrorTranslate(bhRet, JHI_INTERNAL_ERROR);
	}

	UINT32 BeihaiPlugin::JHI_Plugin_WaitForSpoolerEvent(VM_SESSION_HANDLE SpoolerSession,JHI_EVENT_DATA** ppEventData, JHI_SESSION_ID* targetSession)
	{

		UINT32 jhiRet = JHI_INTERNAL_ERROR;
		JVM_COMM_BUFFER IOBuffer = {0};
		int responseCode = 0;

		// allocate output buffer
		IOBuffer.RxBuf->length = JHI_EVENT_DATA_BUFFER_SIZE + sizeof(JHI_SESSION_ID);
		IOBuffer.RxBuf->buffer = memory_api.allocateMemory(IOBuffer.RxBuf->length PROFILING_ARGS);

		if (!IOBuffer.RxBuf->buffer)
			return JHI_INTERNAL_ERROR;

		memset(IOBuffer.RxBuf->buffer, 0, IOBuffer.RxBuf->length);

		*ppEventData = (JHI_EVENT_DATA*) memory_api.allocateMemory(sizeof(JHI_EVENT_DATA) PROFILING_ARGS);

		if (!(*ppEventData))
		{
			TRACE0("WaitForSpoolerEvent: Memory allocation error!");
			memory_api.freeMemory(IOBuffer.RxBuf->buffer PROFILING_ARGS);
			return JHI_INTERNAL_ERROR;
		}

		(*ppEventData)->data = NULL;
		(*ppEventData)->datalen = 0;

		jhiRet = JHI_Plugin_SendAndRecv(SpoolerSession, SPOOLER_COMMAND_GET_EVENT, &IOBuffer,&responseCode);

		// check nError to see if all copied or need to extand the buffer..
		if (jhiRet == JHI_INSUFFICIENT_BUFFER)
		{
			// reallocate the buffer
			memory_api.freeMemory(IOBuffer.RxBuf->buffer PROFILING_ARGS);
			IOBuffer.RxBuf->buffer = (UINT8*) memory_api.allocateMemory(IOBuffer.RxBuf->length PROFILING_ARGS);

			if (!IOBuffer.RxBuf->buffer)
			{
				TRACE0("WaitForSpoolerEvent: Memory allocation error!");
				memory_api.freeMemory(*ppEventData PROFILING_ARGS);
				*ppEventData = NULL;
				return JHI_INTERNAL_ERROR;
			}

			// call again with the larger buffer
			jhiRet = JHI_Plugin_SendAndRecv(SpoolerSession, SPOOLER_COMMAND_GET_EVENT, &IOBuffer,&responseCode);
		}


		if (jhiRet == JHI_SUCCESS)
		{
			if( IOBuffer.RxBuf->length < sizeof(JHI_SESSION_ID) )
			{
				TRACE0("Spooler data is too short - must contain session uuid at least.");
				return JHI_INTERNAL_ERROR;
			}

			(*targetSession) = *((JHI_SESSION_ID*) IOBuffer.RxBuf->buffer);

			(*ppEventData)->datalen = IOBuffer.RxBuf->length - sizeof(JHI_SESSION_ID);

			if ((*ppEventData)->datalen > 0)
			{
				(*ppEventData)->data = (uint8_t*) memory_api.allocateMemory((*ppEventData)->datalen PROFILING_ARGS);

				if ((*ppEventData)->data == NULL)
				{
					TRACE0("WaitForSpoolerEvent: Memory allocation error!");
					memory_api.freeMemory(*ppEventData PROFILING_ARGS);
					*ppEventData = NULL;

					memory_api.freeMemory(IOBuffer.RxBuf->buffer PROFILING_ARGS);
					IOBuffer.RxBuf->buffer = NULL;

					return JHI_INTERNAL_ERROR;
				}

				memcpy_s((*ppEventData)->data,(*ppEventData)->datalen, (UINT8*)IOBuffer.RxBuf->buffer + sizeof(JHI_SESSION_ID),(*ppEventData)->datalen);
			}

			(*ppEventData)->dataType = JHI_DATA_FROM_APPLET;

		}
		else
		{
			memory_api.freeMemory(*ppEventData PROFILING_ARGS);
			*ppEventData = NULL;
		}

		memory_api.freeMemory(IOBuffer.RxBuf->buffer PROFILING_ARGS);
		IOBuffer.RxBuf->buffer = NULL;

		return jhiRet;
	}

	UINT32 BeihaiPlugin::JHI_Plugin_SendAndRecv(VM_SESSION_HANDLE Session, INT32 nCommandId, JVM_COMM_BUFFER *pIOBuffer, INT32* pResponseCode)
	{
		UINT32 jhiRet = JHI_INTERNAL_ERROR;
		char* inputBuffer = (char*) pIOBuffer->TxBuf->buffer;
		int inputBufferLength = pIOBuffer->TxBuf->length;

		char* outputBuffer = (char*) pIOBuffer->RxBuf->buffer;
		int* outputBufferLength = (int*) &pIOBuffer->RxBuf->length;

		char* output = NULL;
		int outputLength = *outputBufferLength; // TODO: tell BH to change this. no need to provide max buffer size.
		BH_ERRNO bhRet = BH_PluginSendAndRecv (Session, nCommandId, inputBuffer, inputBufferLength, (void **)&output, (unsigned int *)&outputLength, pResponseCode);

		if (bhRet == BH_SUCCESS && output != NULL)
		{
			// TODO: same as above
			//if (*outputBufferLength < outputLength)
			//{
			//	// buffer provided is too small for the response
			//	TRACE2("JHI_Plugin_SendAndRecv: insufficient buffer sent to VM, expected: %d, received: %d\n", outputLength, *outputBufferLength);
			//	jhiRet = JHI_INSUFFICIENT_BUFFER;
			//	*outputBufferLength = outputLength;
			//	goto cleanup;
			//}

			// copy the output to the output buffer
			memcpy_s(outputBuffer, *outputBufferLength, output,outputLength);
		}

		*outputBufferLength = outputLength;

		jhiRet = JhiErrorTranslate(bhRet, JHI_INTERNAL_ERROR);

		if (output)
			BH_FREE(output);

		return jhiRet;
	}

	const char* BeihaiPlugin::BHErrorToString(UINT32 retVal)
	{
		const char* str = "";
		switch (retVal)
		{
		case BH_SUCCESS:					str = "BH_SUCCESS";						break; // 0,

		case BPE_NOT_INIT:					str = "BPE_NOT_INIT";					break; // 0xF0001000,
		case BPE_SERVICE_UNAVAILABLE:		str = "BPE_SERVICE_UNAVAILABLE";		break; // 0xF0001001,
		case BPE_INTERNAL_ERROR:			str = "BPE_INTERNAL_ERROR";				break; // 0xF0001002,
		case BPE_COMMS_ERROR:				str = "BPE_COMMS_ERROR";				break; // 0xF0001003,
		case BPE_OUT_OF_MEMORY:				str = "BPE_OUT_OF_MEMORY";				break; // 0xF0001004,
		case BPE_INVALID_PARAMS:			str = "BPE_INVALID_PARAMS";				break; // 0xF0001005,
		case BPE_MESSAGE_TOO_SHORT:			str = "BPE_MESSAGE_TOO_SHORT";			break; // 0xF0001006,
		case BPE_MESSAGE_ILLEGAL:			str = "BPE_MESSAGE_ILLEGAL";			break; // 0xF0001007,
		case BPE_NO_CONNECTION_TO_FIRMWARE: str = "BPE_NO_CONNECTION_TO_FIRMWARE";	break; // 0xF0001008,
		case BPE_NOT_IMPLEMENT:				str = "BPE_NOT_IMPLEMENT";				break; // 0xF0001009,
		case BPE_OUT_OF_RESOURCE:			str = "BPE_OUT_OF_RESOURCE";			break; // 0xF000100A,
		case BPE_INITIALIZED_ALREADY:		str = "BPE_INITIALIZED_ALREADY";		break; // 0xF000100B,

			/* copied from errcode.h */
			/* General errors: 0x100 */
		case BHE_OUT_OF_MEMORY:				str = "BHE_OUT_OF_MEMORY";				break; // 0x101, /* Out of memory */
		case BHE_BAD_PARAMETER:				str = "BHE_BAD_PARAMETER";				break; // 0x102, /* Bad parameters to native */
		case BHE_INSUFFICIENT_BUFFER:		str = "BHE_INSUFFICIENT_BUFFER";		break; // 0x103,
		case BHE_MUTEX_INIT_FAIL:			str = "BHE_MUTEX_INIT_FAIL";			break; // 0x104,
		case BHE_COND_INIT_FAIL:			str = "BHE_COND_INIT_FAIL";				break; // 0x105, /* Cond init fail is not return to host now, it may be used later. */				      
		case BHE_WD_TIMEOUT:				str = "BHE_WD_TIMEOUT";					break; // 0x106, /* Watchdog time out */

			/* Communication: 0x200 */
		case BHE_MAILBOX_NOT_FOUND:			str = "BHE_APPLET_CRASHED or BHE_MAILBOX_NOT_FOUND"; break; // 0x201, /* Mailbox not found or applet crashed */
		case BHE_MSG_QUEUE_IS_FULL:			str = "BHE_MSG_QUEUE_IS_FULL";			break; // 0x202, /* Message queue is full */
		case BHE_MAILBOX_DENIED:			str = "BHE_MAILBOX_DENIED";				break; // 0x203, /* Mailbox is denied by firewall */

			/* Applet manager: 0x300 */
		case BHE_LOAD_JEFF_FAIL:			str = "BHE_LOAD_JEFF_FAIL";				break; // 0x303, /* JEFF file load fail, OOM or file format error not distinct by current JEFF loading process (bool jeff_loader_load). */
		case BHE_PACKAGE_NOT_FOUND:			str = "BHE_PACKAGE_NOT_FOUND";			break; // 0x304, /* Request operation on a package, but it does not exist. */
		case BHE_EXIST_LIVE_SESSION:		str = "BHE_EXIST_LIVE_SESSION";			break; // 0x305, /* Uninstall package fail because of live session exist. */
		case BHE_VM_INSTANCE_INIT_FAIL:		str = "BHE_VM_INSTANCE_INIT_FAIL";		break; // 0x306, /* VM instance init fail when create session */
		case BHE_QUERY_PROP_NOT_SUPPORT:	str = "BHE_QUERY_PROP_NOT_SUPPORT";		break; // 0x307, /* Query applet property that Beihai does not support. */
		case BHE_INVALID_BPK_FILE:			str = "BHE_INVALID_BPK_FILE";			break; // 0x308, /* Incorrect Beihai package format */
		case BHE_VM_INSTNACE_NOT_FOUND:		str = "BHE_VM_INSTNACE_NOT_FOUND";		break; // 0x312, /* VM instance not found */
		case BHE_STARTING_JDWP_FAIL:		str = "BHE_STARTING_JDWP_FAIL";			break; // 0x313, /* JDWP agent starting fail */

			/* Applet instance: 0x400 */
		case BHE_UNCAUGHT_EXCEPTION:		str = "BHE_UNCAUGHT_EXCEPTION";			break; // 0x401, /* uncaught exception */
		case BHE_APPLET_BAD_PARAMETER:		str = "BHE_APPLET_BAD_PARAMETER";		break; // 0x402, /* Bad parameters to applet */
		case BHE_APPLET_SMALL_BUFFER:		str = "BHE_APPLET_SMALL_BUFFER";		break; // 0x403, /* Small response buffer */
		case BHE_APPLET_BAD_STATE:			str = "BHE_APPLET_BAD_STATE";			break; // 0x404,

			/* copied from HAL.h */
		case HAL_TIMED_OUT:					str = "HAL_TIMED_OUT";					break; // 0x00001001,
		case HAL_FAILURE:					str = "HAL_FAILURE";					break; // 0x00001002,
		case HAL_OUT_OF_RESOURCES:			str = "HAL_OUT_OF_RESOURCES";			break; // 0x00001003,
		case HAL_OUT_OF_MEMORY:				str = "HAL_OUT_OF_MEMORY";				break; // 0x00001004,
		case HAL_BUFFER_TOO_SMALL:			str = "HAL_BUFFER_TOO_SMALL";			break; // 0x00001005,
		case HAL_INVALID_HANDLE:			str = "HAL_INVALID_HANDLE";				break; // 0x00001006,
		case HAL_NOT_INITIALIZED:			str = "HAL_NOT_INITIALIZED";			break; // 0x00001007,
		case HAL_INVALID_PARAMS:			str = "HAL_INVALID_PARAMS";				break; // 0x00001008,
		case HAL_NOT_SUPPORTED:				str = "HAL_NOT_SUPPORTED";				break; // 0x00001009,
		case HAL_NO_EVENTS:					str = "HAL_NO_EVENTS";					break; // 0x0000100A,
		case HAL_NOT_READY:					str = "HAL_NOT_READY";					break; // 0x0000100B,
			// ...etc

		case HAL_INTERNAL_ERROR:			str = "HAL_INTERNAL_ERROR";				break; // 0x00001100,
		case HAL_ILLEGAL_FORMAT:			str = "HAL_ILLEGAL_FORMAT";				break; // 0x00001101,
		case HAL_LINKER_ERROR:				str = "HAL_LINKER_ERROR";				break; // 0x00001102,
		case HAL_VERIFIER_ERROR:			str = "HAL_VERIFIER_ERROR";				break; // 0x00001103,

			// User defined applet & session errors to be returned to the host (should be exposed also in the host DLL)
		case HAL_FW_VERSION_MISMATCH:		str = "HAL_FW_VERSION_MISMATCH";		break; // 0x00002000,
		case HAL_ILLEGAL_SIGNATURE:			str = "HAL_ILLEGAL_SIGNATURE";			break; // 0x00002001,
		case HAL_ILLEGAL_POLICY_SECTION:	str = "HAL_ILLEGAL_POLICY_SECTION";		break; // 0x00002002,
		case HAL_OUT_OF_STORAGE:			str = "HAL_OUT_OF_STORAGE";				break; // 0x00002003,
		case HAL_UNSUPPORTED_PLATFORM_TYPE:	str = "HAL_UNSUPPORTED_PLATFORM_TYPE";	break; // 0x00002004,
		case HAL_UNSUPPORTED_CPU_TYPE:		str = "HAL_UNSUPPORTED_CPU_TYPE";		break; // 0x00002005,
		case HAL_UNSUPPORTED_PCH_TYPE:		str = "HAL_UNSUPPORTED_PCH_TYPE";		break; // 0x00002006,
		case HAL_UNSUPPORTED_FEATURE_SET:	str = "HAL_UNSUPPORTED_FEATURE_SET";	break; // 0x00002007,
		case HAL_ILLEGAL_VERSION:			str = "HAL_ILLEGAL_VERSION";			break; // 0x00002008,
		case HAL_ALREADY_INSTALLED:			str = "HAL_ALREADY_INSTALLED";			break; // 0x00002009,
		case HAL_MISSING_POLICY:			str = "HAL_MISSING_POLICY";				break; // 0x00002010,

		default:
			str = "BH_UNKNOWN_ERROR";
			break;
		}

		return str;
	}

	UINT32 BeihaiPlugin::JhiErrorTranslate(BH_ERRNO bhError, UINT32 defaultError)
	{
		UINT32 jhiError;

		switch (bhError)
		{
		case BH_SUCCESS:
			jhiError = JHI_SUCCESS;
			break;


			// SendAndRecv
		case BHE_INSUFFICIENT_BUFFER:
		case BHE_APPLET_SMALL_BUFFER:
		case HAL_BUFFER_TOO_SMALL:
			jhiError = JHI_INSUFFICIENT_BUFFER;
			break;

		case BHE_APPLET_BAD_STATE:
			jhiError = JHI_APPLET_BAD_STATE;
			break;

		case BPE_NO_CONNECTION_TO_FIRMWARE:
			jhiError = JHI_NO_CONNECTION_TO_FIRMWARE;
			break;

		case HAL_OUT_OF_MEMORY:
		case BHE_UNCAUGHT_EXCEPTION:
		case BHE_APPLET_CRASHED:
		case BHE_WD_TIMEOUT:
		case HAL_TIMED_OUT:
			//case BHE_APPLET_SMALL_BUFFER: //Oded - ???
			jhiError = JHI_APPLET_FATAL;
			break;



			// DownloadApplet
		case HAL_ILLEGAL_SIGNATURE:
		case HAL_ILLEGAL_VERSION:
		case HAL_FW_VERSION_MISMATCH:
		case HAL_UNSUPPORTED_CPU_TYPE:
		case HAL_UNSUPPORTED_PCH_TYPE:
		case HAL_UNSUPPORTED_FEATURE_SET:
		case HAL_UNSUPPORTED_PLATFORM_TYPE:

			jhiError = JHI_APPLET_AUTHENTICATION_FAILURE;
			break;

			//case :
			//jhiError = JHI_INTERNAL_ERROR;

		case BHE_INVALID_BPK_FILE:
			jhiError = JHI_BAD_APPLET_FORMAT;
			break;

		case HAL_ALREADY_INSTALLED:
			jhiError = JHI_FILE_IDENTICAL;
			break;

		case HAL_OUT_OF_RESOURCES:
		case HAL_OUT_OF_STORAGE:
			jhiError = JHI_MAX_INSTALLED_APPLETS_REACHED;
			break;

			// UnloadApplet
		case BHE_EXIST_LIVE_SESSION:
			jhiError = JHI_INSTALL_FAILURE_SESSIONS_EXISTS;
			break;

			// JHI_Plugin_GetAppletProperty
		case BHE_QUERY_PROP_NOT_SUPPORT:
			jhiError = JHI_APPLET_PROPERTY_NOT_SUPPORTED;
			break;

		case BHE_PACKAGE_NOT_FOUND:
			jhiError = JHI_APPLET_NOT_INSTALLED;
			break;

		default:
			jhiError = defaultError;
		}

		if (jhiError != JHI_SUCCESS)
		{
			TRACE4("beihaiToJhiError: BH Error received - 0x%X (%s), translated to JHI Error - 0x%X (%s)", bhError, BHErrorToString(bhError), jhiError, JHIErrorToString(jhiError));
		}

		return jhiError;
	}
}