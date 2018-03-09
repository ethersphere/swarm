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

#include "jhi.h"
#include "jhi_i.h"
#include "jhi_plugin_loader.h"
#include "dbg.h"
#include "DLL_Loader.h"
#include "GlobalsManager.h"
#include "AppletsManager.h"
#include "misc.h"
#include "string_s.h"

#ifdef __linux__
#include <dlfcn.h>
#endif //__linux__

using namespace intel_dal;
HMODULE loadedPluginDLL = NULL;
bool pluginLoaded = false;

#ifdef JHI_MEMORY_PROFILING //for now we are not transferring the memory profiling to the plugin so we must reconfigure the memory functions.
#undef JHI_ALLOC(x)
#undef JHI_DEALLOC(x)
void * JHI_ALLOC(uint32_t bytes_alloc);
void JHI_DEALLOC(void* handle);
#endif //JHI_MEMORY_PROFILING


JHI_RET	JhiPlugin_Unregister(VM_Plugin_interface** plugin)
{
	if (plugin != NULL)
	{
		*plugin = NULL;
	}

	if (pluginLoaded)
	{
		pluginLoaded = false;
		return DLL_Loader::UnloadDll(&loadedPluginDLL);
	}
	return JHI_SUCCESS;
}


// Registers the correct plugin
JHI_RET JhiPlugin_Register (VM_Plugin_interface** plugin)
{
	JHI_RET retCode = JHI_VM_DLL_VERIFY_FAILED;
	if (NULL == plugin)
	{
		return retCode;
	}
	FILESTRING jhi_service_folder;
	PFN_pluginRegister pPluginRegister = NULL;
	bool verifySignature = true;
	JhiPlugin_Unregister(plugin);
	JHI_VM_TYPE vmType = GlobalsManager::Instance().getVmType();
	JHI_PLUGIN_TYPE pluginTypeToLoad = JHI_PLUGIN_TYPE_INVALID;
	JHI_PLUGIN_TYPE loadedPluginType = JHI_PLUGIN_TYPE_INVALID;

#ifdef _WIN32
    GlobalsManager::Instance().getServiceFolder(jhi_service_folder);
#else
    GlobalsManager::Instance().getPluginFolder(jhi_service_folder);
#endif

#ifndef _WIN32
    const char *error;
#endif

#if defined(SCHANNEL_OVER_SOCKET) || defined(DEBUG) || (JHI_ENGINEERING==12)
	verifySignature = false;
#else //validate in HECI release mode only
	verifySignature = true;
#endif

	FILESTRING vendorName;
	FILESTRING dllName;

	switch (vmType)
	{
	case JHI_VM_TYPE_TL:
		pluginTypeToLoad = JHI_PLUGIN_TYPE_TL;
		vendorName = TEE_VENDORNAME;
		dllName = TEE_FILENAME;
		break;

	case JHI_VM_TYPE_BEIHAI_V1:
		pluginTypeToLoad = JHI_PLUGIN_TYPE_BEIHAI_V1;
		vendorName = BH_VENDORNAME;
		dllName = BH_FILENAME;
		break;

	case JHI_VM_TYPE_BEIHAI_V2:
		pluginTypeToLoad = JHI_PLUGIN_TYPE_BEIHAI_V2;
		vendorName = BH_VENDORNAME;
		dllName = BH_V2_FILENAME;
		break;

	default:
		TRACE0("Error: Invalid VM type\n");
		retCode = JHI_INTERNAL_ERROR;
		goto cleanup;
	}

	TRACE1("Loading Plugin DLL, filename: %s\n",ConvertWStringToString(dllName).c_str());

	retCode = DLL_Loader::LoadDll(jhi_service_folder, dllName, vendorName, verifySignature, &loadedPluginDLL);

	if (retCode != JHI_SUCCESS)
	{
#ifndef _WIN32
		if ((error = dlerror()) != NULL)
			TRACE2("Failed to load %s, error:%s\n", JHI_PLUGIN_REGISTER_FUNCTION, error);
		else
			TRACE2("Failed to load %s, line:%d\n", JHI_PLUGIN_REGISTER_FUNCTION, __LINE__);
#endif //_WIN32

		goto cleanup;
	}

	pluginLoaded = true;

	//Get a pointer to the register function

#ifdef _WIN32
	pPluginRegister = (PFN_pluginRegister) GetProcAddress(GetModuleHandle(dllName.c_str()), JHI_PLUGIN_REGISTER_FUNCTION);
#else
    dlerror();
	pPluginRegister = (PFN_pluginRegister) dlsym(loadedPluginDLL, JHI_PLUGIN_REGISTER_FUNCTION);
#endif// _WIN32

	if (pPluginRegister == NULL)
	{
		TRACE2("Failed to get %s, line:%d\n", JHI_PLUGIN_REGISTER_FUNCTION, __LINE__);
#ifdef __linux__
		TRACE0(dlerror());
#endif
		retCode = JHI_VM_DLL_VERIFY_FAILED;
		goto cleanup;
	}

	retCode = pPluginRegister(plugin);
	if ( (retCode != JHI_SUCCESS) || (*plugin == NULL) )
	{
		TRACE2("Failed to register using %s line %d\n", JHI_PLUGIN_REGISTER_FUNCTION, __LINE__);
		retCode = JHI_VM_DLL_VERIFY_FAILED;
		goto cleanup;
	}

	loadedPluginType = (JHI_PLUGIN_TYPE)(*plugin)->JHI_Plugin_GetPluginType();
	if (loadedPluginType != pluginTypeToLoad)
	{
		TRACE2("Invalid PluginType %d line %d\n", "JHI_Plugin_GetPluginType()", __LINE__);
		retCode = JHI_VM_DLL_VERIFY_FAILED;
		goto cleanup;
	}

	return JHI_SUCCESS;

cleanup:
	JhiPlugin_Unregister(plugin);
	return retCode;
}

#ifdef JHI_MEMORY_PROFILING // for now we are not transferring the memory profiling to the plugin so we must reconfigure the memory functions.
// these are copies of the original memory functions
void * JHI_ALLOC(uint32_t bytes_alloc)
{
	void* var = NULL;
#ifdef _WIN32
	try 
	{
#endif
		var = (void*) new uint8_t[bytes_alloc];
#ifdef _WIN32
	}
	catch (...)
	{
#endif
		TRACE1("JHI memory allocation of size %d failed .",bytes_alloc);
#ifdef _WIN32	
	}
#endif
	return var;
}

//------------------------------------------------------------------------------
//
//------------------------------------------------------------------------------
void JHI_DEALLOC(void* handle)
{
#ifdef _WIN32
	try
	{
#endif
		if (handle != NULL)
			delete [] (uint8_t*)handle;
#ifdef _WIN32
	}
	catch (...) 
	{
#endif
		TRACE0("JHI free memory failed.");
#ifdef _WIN32
	}
#endif
}
#endif //JHI_MEMORY_PROFILING