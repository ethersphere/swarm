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
**    @file plugin_interface.h
**
**    @brief  Conatins VM plugin interface for JHI 
**
**    @author Elad Dabool
**
********************************************************************************
*/
#ifndef __JHI_PLUGIN_INTERFACE__
#define __JHI_PLUGIN_INTERFACE__

#include "jhi.h"
#include "jhi_i.h"
#include "dbg.h"
#include <string>
#include <vector>

using std::string;
using std::vector;

#ifdef __cplusplus
extern "C" {
#endif

	typedef  PVOID VM_SESSION_HANDLE;
#ifndef LEN_APP_ID
#define LEN_APP_ID  32 	// applet id without \0 and separators
#endif

	// JHI Memory managment API
#ifdef JHI_MEMORY_PROFILING
	typedef void* (*PFN_JHI_ALLOCATE_MEMORY) (UINT32 bytes_alloc, const char* file, int line);
	typedef void(*PFN_JHI_FREE_MEMORY) (void* handle, const char* file, int line);
#else
	typedef void* (*PFN_JHI_ALLOCATE_MEMORY) (UINT32 bytes_alloc);
	typedef void (*PFN_JHI_FREE_MEMORY) (void* handle);
#endif

	typedef struct 
	{
		PFN_JHI_ALLOCATE_MEMORY	allocateMemory;
		PFN_JHI_FREE_MEMORY freeMemory;
	} JHI_PLUGIN_MEMORY_API;

	typedef struct 
	{
		int	packageType;
		uint8_t uuid[LEN_APP_ID + 1];
	} PACKAGE_INFO;

	class VM_Plugin_interface
	{
	public:
		virtual UINT32 JHI_Plugin_Init(bool do_vm_reset = true) = 0;
		virtual UINT32 JHI_Plugin_DeInit(bool do_vm_reset = true) = 0;
		virtual UINT32 JHI_Plugin_Set_Transport_And_Memory(unsigned int transportType, JHI_PLUGIN_MEMORY_API* plugin_memory_api) = 0;
		virtual UINT32 JHI_Plugin_GetPluginType() = 0;
		virtual UINT32 JHI_Plugin_DownloadApplet (const char *pAppId, uint8_t* pAppBlob, unsigned int BlobSize) = 0;
		virtual UINT32 JHI_Plugin_UnloadApplet (const char *AppId ) = 0;
		virtual UINT32 JHI_Plugin_GetAppletProperty (const char *AppId, JVM_COMM_BUFFER *pIOBuffer) = 0;
		virtual UINT32 JHI_Plugin_CreateSession (const char *AppId, VM_SESSION_HANDLE* pSession, const uint8_t* pAppBlob, unsigned int BlobSize, JHI_SESSION_ID SessionID,DATA_BUFFER* initBuffer) = 0;
		virtual UINT32 JHI_Plugin_CloseSession (VM_SESSION_HANDLE* pSession) = 0;
		virtual UINT32 JHI_Plugin_ForceCloseSession(VM_SESSION_HANDLE* pSession) = 0;
		virtual UINT32 JHI_Plugin_WaitForSpoolerEvent (VM_SESSION_HANDLE SpoolerSession,JHI_EVENT_DATA** ppEventData,JHI_SESSION_ID* targetSession) = 0;
		virtual UINT32 JHI_Plugin_SendAndRecv (VM_SESSION_HANDLE Session, INT32 nCommandId, JVM_COMM_BUFFER *pIOBuffer,INT32* pResponseCode) = 0;
		virtual	UINT32 JHI_Plugin_OpenSDSession (const string& SD_ID, VM_SESSION_HANDLE* pSession) = 0;
		virtual UINT32 JHI_Plugin_CloseSDSession (VM_SESSION_HANDLE* pSession) = 0;
		virtual UINT32 JHI_Plugin_ListInstalledTAs (const VM_SESSION_HANDLE handle, vector<string>& UUIDs) = 0;
		virtual UINT32 JHI_Plugin_ListInstalledSDs(const VM_SESSION_HANDLE handle, vector<string>& UUIDs) = 0;
		virtual UINT32 JHI_Plugin_SendCmdPkg (const VM_SESSION_HANDLE handle, vector<uint8_t>& blob) = 0;
		virtual UINT32 JHI_Plugin_QueryTeeMetadata(unsigned char** metadata, unsigned int* length) = 0;
		virtual UINT32 JHI_Plugin_ParsePackage(uint8_t* cmd_pkg, uint32_t pkg_len, OUT PACKAGE_INFO& pkgInfo) = 0;
		virtual ~VM_Plugin_interface() {}
#ifdef _WIN32
		virtual void JHI_Plugin_SetLogLevel(JHI_LOG_LEVEL log_level) = 0;
#endif
	protected:
		PVOID pluginCtx;
	};

	// Plugin Register function
#ifdef JHI_PLUGIN
	// Register funtion that should be exported by the plugin
	UINT32 __declspec(dllexport) pluginRegister(OUT VM_Plugin_interface** plugin);
#else
	// used by jhi service to dynamically call the dll register function
	typedef int (*PFN_pluginRegister) (VM_Plugin_interface** plugin);
#endif // JHI_PLUGIN

#ifdef __cplusplus
};
#endif // __cplusplus


#endif // __JHI_PLUGIN_INTERFACE__
