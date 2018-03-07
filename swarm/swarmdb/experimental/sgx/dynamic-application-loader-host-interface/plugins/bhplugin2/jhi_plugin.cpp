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
#include <map>
#include <sstream>
#include <algorithm>
#include <thread>
#include <string>
#include <jhi.h>

using namespace std;

#include "bhp_exp.h"
#include "bh_acp_exp.h"
#include "jhi_plugin.h"
#include "BeihaiStatusHAL.h"
#include "dbg.h"
#include "bh_acp_util.h"
#include "teemanagement.h"
#include "misc.h"
#include "bh_shared_conf.h"

#ifndef _WIN32
#include "string_s.h"
#endif //_WIN32

//------------------------------------------------------------------------------
// first-time register of plugin callbacks
//------------------------------------------------------------------------------
extern "C"
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
	// a dirty patch for Android where this function is not present
#ifdef __ANDROID__
	template <typename T>
	std::string to_string(T value)
	{
	    std::ostringstream os ;
	    os << value ;
	    return os.str() ;
	}
#endif

	//declaring the static var
	TEE_TRANSPORT_INTERFACE BeihaiPlugin::transport_interface = {0};

	int BeihaiPlugin::sendWrapper(uintptr_t handle, uint8_t* buffer, uint32_t length)
	{
		return (int)transport_interface.pfnSend(&transport_interface, (TEE_TRANSPORT_HANDLE)handle, (const uint8_t*)buffer, (size_t)length);
	}

	int BeihaiPlugin::recvWrapper(uintptr_t handle, uint8_t* buffer, uint32_t* length)
	{
		return (int)transport_interface.pfnRecv(&transport_interface, (TEE_TRANSPORT_HANDLE)handle, (uint8_t*)buffer, length);
	}

	int BeihaiPlugin::connectWrapper(int heci_port, uintptr_t * handle) //not realy needed
	{
		return (int)transport_interface.pfnConnect(&transport_interface, (TEE_TRANSPORT_ENTITY)heci_port, NULL, (TEE_TRANSPORT_HANDLE*)handle);
	}

	int BeihaiPlugin::closeWrapper(uintptr_t handle)
	{
		return (int)transport_interface.pfnDisconnect(&transport_interface, (TEE_TRANSPORT_HANDLE*)&handle);
	}

	BeihaiPlugin::BeihaiPlugin()
	{
		memset (&memory_api, 0, sizeof(JHI_PLUGIN_MEMORY_API));
		memset (&bh_transport_APIs, 0, sizeof(BHP_TRANSPORT));
		memset (&transport_interface, 0, sizeof(TEE_TRANSPORT_INTERFACE));
		memset (&intel_sd_handle, 0, sizeof(SD_SESSION_HANDLE));
		is_intel_sd_open = false;
		is_oem_sd_open = false;
		oem_sd_handle = nullptr;
		plugin_type = JHI_PLUGIN_TYPE_BEIHAI_V2;
	}

	UINT32 BeihaiPlugin::JHI_Plugin_GetPluginType()
	{
		return plugin_type;
	}

	UINT32 BeihaiPlugin::JHI_Plugin_Set_Transport_And_Memory(unsigned int transportType, JHI_PLUGIN_MEMORY_API* plugin_memory_api)
	{
		JHI_RET ulRetCode = JHI_INVALID_PARAMS ;

		if(plugin_memory_api == NULL)
			goto end;

		memset (&memory_api, 0, sizeof(JHI_PLUGIN_MEMORY_API));
		BeihaiPlugin::memory_api = *plugin_memory_api;

		memset (&bh_transport_APIs, 0, sizeof(BHP_TRANSPORT));

		ulRetCode = TEE_Transport_Create((TEE_TRANSPORT_TYPE)transportType, &transport_interface);
		if (ulRetCode != TEE_COMM_SUCCESS)
		{
			return JHI_INTERNAL_ERROR;
		}

		//pass BH the wrappers that uses the transport_APIs
		bh_transport_APIs.pfnSend = sendWrapper;
		bh_transport_APIs.pfnRecv = recvWrapper;
		bh_transport_APIs.pfnConnect = connectWrapper;
		bh_transport_APIs.pfnClose = closeWrapper;

		ulRetCode = JHI_SUCCESS;

end:
		return ulRetCode ;
	}

#ifdef USE_LOCAL_ACP_FILE
	bool readFile(string path, char** buffer, unsigned int* length)
	{
		if (buffer == NULL || length == NULL || path.empty())
		{
			return false;
		}
		std::ifstream is (path, std::ifstream::binary);

		if (is) {
			// get length of file:
			is.seekg (0, is.end);
			*length = is.tellg();
			is.seekg (0, is.beg);

			if (*length >= MAX_APPLET_BLOB_SIZE)
			{
				return false;
			}

			*buffer = new char [*length];

			TRACE1("Reading %d characters... ", *length);
			// read data as a block:
			is.read (*buffer, *length);

			if (is)
			{
				TRACE0("all characters read successfully.");
			}
			else
			{
				TRACE1("error: only %d could be read", is.gcount());
				delete[] *buffer;
				*buffer = NULL;
				return false;
			}
			is.close();

			// ...buffer contains the entire file...

			return true;
		}
	}
#endif

	void BeihaiPlugin::setUninstallPack(const char *pAppId, char** uninstallPkg)
	{
		if (pAppId == NULL || uninstallPkg == NULL)
		{
			return;
		}

#ifdef USE_LOCAL_ACP_FILE
		/// *** use local acp file ***
		unsigned int length = 0;
		char* file = NULL;
		TRACE0("getting uninstall package from c:/EchoAppletUninstall.acp.");
		if ( (!readFile("c:/EchoAppletUninstall.acp", &file, &length) ) || (file == NULL) || (length == 0) )
		{
			return;
		}

		*uninstallPkg = (char*)memory_api.allocateMemory(length);
		memcpy(*uninstallPkg, file, length);
		delete[] file;
		return;
#endif

		///
		string appId = string(pAppId);
		std::transform(appId.begin(), appId.end(), appId.begin(), ::toupper);

		// copying the uninstall pack
		*uninstallPkg = (char*)memory_api.allocateMemory(UNINSTALL_PACK_LEN);
		if (*uninstallPkg == NULL)
			return;

		memcpy_s(*uninstallPkg, UNINSTALL_PACK_LEN, UNINSTALL_PACK, UNINSTALL_PACK_LEN);

		// replacing the uuid
		char* ptr = *uninstallPkg + 32 + JHI_CSS_HEADER_SIZE;
		for(int i=0; i< 32; i+=2)
		{
			string byte = appId.substr(i,2);
			*ptr = (char) (int)strtol(byte.c_str(), NULL, 16);
			++ptr;
		}
	}

	unsigned int BeihaiPlugin::getTotalSessionsCount()
	{
		BH_RET ret;
#ifndef OPEN_INTEL_SD_SESSION_ONCE
		//first open the SD
		ret = openIntelSD();
		if (ret != BH_SUCCESS)
		{
			return;
		}
#endif

		uint32_t appletsCount = 0;
		uint32_t appletSessionsCount = 0;
		uint32_t totalSessionsCount = 0;
		char** appIdStrs = NULL;
		JAVATA_SESSION_HANDLE* appletSessions = NULL;
		ret = BHP_ListInstalledTAs(intel_sd_handle, INTEL_SD_UUID, &appletsCount, &appIdStrs);

#ifndef OPEN_INTEL_SD_SESSION_ONCE
		//close the SD
		closeIntelSD();
#endif

		if (appIdStrs == NULL)
		{
			return 0;
		}

		if (ret == BH_SUCCESS)
		{
			for (uint32_t i=0; i < appletsCount; ++i)
			{
				if (appIdStrs[i] == NULL)
				{
					continue;
				}
				appletSessionsCount = 0;
				// getting TA sessions count.
				ret = BHP_ListTASessions(appIdStrs[i], &appletSessionsCount, &appletSessions);
				if (ret == BH_SUCCESS)
				{
					totalSessionsCount += appletSessionsCount;
				}
				if (appletSessions != NULL)
				{
					BHP_Free(appletSessions);
					appletSessions = NULL;
				}
				BHP_Free(appIdStrs[i]);
				appIdStrs[i] = NULL;
			}
		}

		if (appIdStrs != NULL)
		{
			BHP_Free(appIdStrs);
			appIdStrs = NULL;
		}

		return totalSessionsCount;
	}
	void BeihaiPlugin::uninstallAll()
	{
		BH_RET ret;
#ifndef OPEN_INTEL_SD_SESSION_ONCE
		//first open the SD
		ret = openIntelSD();
		if (ret != BH_SUCCESS)
		{
			return;
		}
#endif

		uint32_t appletsCount = 0;
		uint32_t appletSessionsCount = 0;
		char** appIdStrs = NULL;
		JAVATA_SESSION_HANDLE* appletSessions = NULL;
		ret = BHP_ListInstalledTAs(intel_sd_handle, INTEL_SD_UUID, &appletsCount, &appIdStrs);

#ifndef OPEN_INTEL_SD_SESSION_ONCE
		//close the SD
		closeIntelSD();
#endif

		if (appIdStrs == NULL)
		{
			return;
		}

		if (ret == BH_SUCCESS)
		{
			for (uint32_t i=0; i < appletsCount; ++i)
			{
				if (appIdStrs[i] == NULL)
				{
					continue;
				}
				appletSessionsCount = 0;
				// getting TA sessions and closing them.
				ret = BHP_ListTASessions(appIdStrs[i], &appletSessionsCount, &appletSessions);
				if (ret == BH_SUCCESS && appletSessions != NULL)
				{
					for (uint32_t j=0; j < appletSessionsCount; ++j)
					{
						ret = JHI_Plugin_CloseSession((VM_SESSION_HANDLE*)&(appletSessions[i]));
					}
				}

				if (appletSessions != NULL)
				{
					BHP_Free(appletSessions);
					appletSessions = NULL;
				}
				// uninstall the TA
				ret = JHI_Plugin_UnloadApplet(appIdStrs[i]);
				BHP_Free(appIdStrs[i]);
				appIdStrs[i] = NULL;
			}
		}
		if (appIdStrs != NULL)
		{
			BHP_Free(appIdStrs);
			appIdStrs = NULL;
		}
	}

	UINT32 BeihaiPlugin::JHI_Plugin_Init(bool do_vm_reset)
	{
		TRACE0("JHI_Plugin_Init start");

		int ret = BHP_Init(&bh_transport_APIs, do_vm_reset);

#ifdef OPEN_INTEL_SD_SESSION_ONCE
		openIntelSD();
#endif

		TRACE1("JHI_Plugin_Init end, result = 0x%X", ret);
		return beihaiToJhiError(ret,JHI_NO_CONNECTION_TO_FIRMWARE);
	}

	UINT32 BeihaiPlugin::JHI_Plugin_DeInit(bool do_vm_reset)
	{
		TRACE0("JHI_Plugin_DeInit start");

#ifdef OPEN_INTEL_SD_SESSION_ONCE
		closeIntelSD();
#endif

		// Close the OEM SD session if it is open
		// Ignore the return value because if it failed to be closed it is already invalid.
		if(is_oem_sd_open)
		{
			BHP_CloseSDSession(oem_sd_handle);
			is_oem_sd_open = false;
			oem_sd_handle = nullptr;
			oem_sd_id.clear();
		}

		int ret = BHP_Deinit(do_vm_reset);

        if(transport_interface.state != TEE_INTERFACE_STATE_NOT_INITIALIZED)
        {
            int ret2 = transport_interface.pfnTeardown(&transport_interface);
            if (ret2 != TEE_STATUS_SUCCESS || transport_interface.state != TEE_INTERFACE_STATE_NOT_INITIALIZED)
            {
                TRACE1("transport_interface Teardown error, result = 0x%X", ret2);
                return JHI_INTERNAL_ERROR;
            }
        }
        else
            TRACE0("transport_interface is not initialized, skipping deinitialization.");

		TRACE1("JHI_Plugin_DeInit end, result = 0x%X", ret);
		return beihaiToJhiError(ret,JHI_INTERNAL_ERROR);
	}

	bool BeihaiPlugin::isTAinstalled(const char *pAppId)
	{
		//TRACE0("isTAinstalled start");
		BH_RET ret;

#ifndef OPEN_INTEL_SD_SESSION_ONCE
		//first open the SD
		ret = openIntelSD();
		if (ret != BH_SUCCESS)
		{
			return false;
		}
#endif

		uint32_t appletsCount = 0;
		char** appIdStrs = NULL;
		bool result = false;
		ret = BHP_ListInstalledTAs(intel_sd_handle, INTEL_SD_UUID, &appletsCount, &appIdStrs);

#ifndef OPEN_INTEL_SD_SESSION_ONCE
		closeIntelSD();
#endif

		if (appIdStrs == NULL)
		{
			return false;
		}

		if (ret == BH_SUCCESS)
		{
			for (uint32_t i=0; i < appletsCount; ++i)
			{
				if (appIdStrs[i] == NULL)
				{
					continue;
				}
#ifdef _WIN32
				if (_stricmp(appIdStrs[i], pAppId) == 0)
#else
				if (strcasecmp(appIdStrs[i], pAppId) == 0)
#endif
				{
					//TRACE0("isTAinstalled end, result = true");
					result = true;
					//not breaking in order to perform the cleanup.
				}
				BHP_Free(appIdStrs[i]);
				appIdStrs[i] = NULL;
			}
		}
		if (!result)
		{
			//TRACE0("isTAinstalled end, result = false");
		}
		if (appIdStrs != NULL)
		{
			BHP_Free(appIdStrs);
            appIdStrs = NULL;
		}
		return result;
	}

	UINT32 BeihaiPlugin::getTA_SessionCount(const char *pAppId)
	{
		TRACE0("getTA_SessionCount start");
		BH_RET ret;
		unsigned int appletSessionsCount = 0;
		JAVATA_SESSION_HANDLE* appletSessions;

		if ((pAppId == NULL) || (!isTAinstalled(pAppId)))
		{
			TRACE1("getTA_SessionCount end, result = 0x%X", 0);
			return 0;
		}

		ret = BHP_ListTASessions(pAppId, &appletSessionsCount, &appletSessions);
		if (ret == BH_SUCCESS)
		{
			BHP_Free(appletSessions);
			appletSessions = NULL;
			return appletSessionsCount;
		}

		TRACE1("getTA_SessionCount end, result = 0x%X", 0);
		return 0;
	}

	UINT32 BeihaiPlugin::JHI_Plugin_ListInstalledTAs (const SD_SESSION_HANDLE handle, vector<string>& UUIDs)
	{
		BH_RET ret = BPE_INTERNAL_ERROR;
		UUIDs.clear();
		string sdId;
		uint32_t appletsCount = 0;
		char** appIdStrs = NULL;

		// validate inputs
		if (handle == NULL)
		{
			ret = BPE_INVALID_PARAMS;
			goto cleanup;
		}
		else if (handle == intel_sd_handle)
		{
			sdId = INTEL_SD_UUID;
		}
		else if (is_oem_sd_open && handle == oem_sd_handle)
		{
			sdId = oem_sd_id;
		}
		else
		{
			ret = BHE_SDM_NOT_FOUND;
			goto cleanup;
		}

		ret = BHP_ListInstalledTAs(handle, sdId.c_str(), &appletsCount, &appIdStrs);
		if (ret != BH_SUCCESS)
			goto cleanup;

		if (appletsCount != 0 && appIdStrs == NULL)
		{
			ret = BPE_INTERNAL_ERROR;
			goto cleanup;
		}

		for (uint32_t i = 0; i < appletsCount; ++i)
		{
			if (appIdStrs[i] == NULL || strnlen_s(appIdStrs[i], 32) != 32)
			{
				goto cleanup;
			}
			//else
			UUIDs.push_back(string(appIdStrs[i]));
		}

cleanup:
		if (appIdStrs != NULL)
		{
			for (uint32_t i = 0; i < appletsCount; ++i)
			{
				if (appIdStrs[i] != NULL)
				{
					BHP_Free(appIdStrs[i]);
                    appIdStrs[i] = NULL;
				}
			}
			BHP_Free(appIdStrs);
            appIdStrs = NULL;
		}

		return beihaiToTeeError(ret, TEE_STATUS_INTERNAL_ERROR);
	}

	UINT32 BeihaiPlugin::JHI_Plugin_ListInstalledSDs(const SD_SESSION_HANDLE handle, vector<string>& UUIDs)
	{
		BH_RET ret = BPE_INTERNAL_ERROR;
		UUIDs.clear();
		string sdId;
		uint32_t appletsCount = 0;
		char** appIdStrs = NULL;

		// validate inputs
		if (handle == intel_sd_handle)
		{
			sdId = INTEL_SD_UUID;
		}
		else
		{
			//get the ID from the map.
			ret = TEE_STATUS_UNSUPPORTED_PLATFORM;
			goto cleanup;
		}

		ret = BHP_ListInstalledSDs(handle, &appletsCount, &appIdStrs);
		if (ret != BH_SUCCESS)
		{
			return ret;
		}

		if (appIdStrs == NULL)
		{
			return BPE_INTERNAL_ERROR;
		}

		for (uint32_t i = 0; i < appletsCount; ++i)
		{
			if (appIdStrs[i] == NULL || strnlen_s(appIdStrs[i], 32) != 32)
			{
				goto cleanup;
			}
			//else
			UUIDs.push_back(string(appIdStrs[i]));
		}

	cleanup:
		if (appIdStrs != NULL)
		{
			for (uint32_t i = 0; i < appletsCount; ++i)
			{
				if (appIdStrs[i] != NULL)
				{
					BHP_Free(appIdStrs[i]);
					appIdStrs[i] = NULL;
				}
			}
			BHP_Free(appIdStrs);
			appIdStrs = NULL;
		}

		return beihaiToTeeError(ret, TEE_STATUS_INTERNAL_ERROR);
	}


	UINT32 BeihaiPlugin::JHI_Plugin_OpenSDSession(const string& sdId, SD_SESSION_HANDLE* pSession)
	{
		BH_RET ret = BPE_INTERNAL_ERROR;
		if (pSession == NULL || !validateUuidString(sdId))
		{
			ret = BPE_INVALID_PARAMS;
			goto end;
		}
#ifdef _WIN32
		if (_stricmp(INTEL_SD_UUID, sdId.c_str()) == 0)
#else
		if (strcasecmp(INTEL_SD_UUID, sdId.c_str()) == 0)
#endif
		{
			*pSession = intel_sd_handle;
			ret = BH_SUCCESS;
			goto end;
		}
		else
		{
			if(is_oem_sd_open && sdId == oem_sd_id)
			{
				*pSession = oem_sd_handle;
				ret = BH_SUCCESS;
			}
			else
			{
				ret = BHP_OpenSDSession(sdId.c_str(), pSession);
				if(ret == BH_SUCCESS)
				{
					is_oem_sd_open = true;
					oem_sd_handle = *pSession;
					oem_sd_id = sdId;
				}
			}
		}
end:
		return beihaiToTeeError(ret, TEE_STATUS_INTERNAL_ERROR);
	}

	UINT32 BeihaiPlugin::JHI_Plugin_CloseSDSession(SD_SESSION_HANDLE* pSession)
	{
		BH_RET ret = BPE_INTERNAL_ERROR;
		if (pSession == NULL)
		{
			ret = BPE_INVALID_PARAMS;
			goto end;
		}

		if (*pSession == intel_sd_handle)
		{
			ret = BH_SUCCESS;
			*pSession = NULL;
			goto end;
		}
		else if (is_oem_sd_open && *pSession == oem_sd_handle)
		{
			ret = BH_SUCCESS;
			*pSession = NULL;
			goto end;
		}
		else
		{
			ret = BHP_CloseSDSession(*pSession);
			/*
			The following is correct if the FW never returns 'success' without actually closing an SD session. Not sure about that.
			if (ret == BH_SUCCESS)
			{
				is_oem_sd_open = false;
				oem_sd_handle = NULL;
				oem_sd_id.clear();
			}
			*/
		}
end:
		return beihaiToTeeError(ret, TEE_STATUS_INTERNAL_ERROR);
	}

	UINT32 BeihaiPlugin::JHI_Plugin_SendCmdPkg(const SD_SESSION_HANDLE handle, vector<uint8_t>& blob)
	{
		TRACE0("JHI_Plugin_SendCmdPkg start");
		BH_RET ret = TEE_STATUS_INTERNAL_ERROR;
		if (blob.size() == 0)
		{
			return TEE_STATUS_INVALID_PARAMS;
		}

		ret = BHP_SendAdminCmdPkg(handle, (char*) &blob[0], (unsigned int)blob.size());

		TRACE1("JHI_Plugin_SendCmdPkg end, result = 0x%X", ret);
		return beihaiToTeeError(ret, TEE_STATUS_INTERNAL_ERROR);
	}

	UINT32 BeihaiPlugin::JHI_Plugin_QueryTeeMetadata(unsigned char ** metadata, unsigned int * length)
	{
		TRACE0("JHI_Plugin_QueryTeeMetadata start");
		BH_RET ret = TEE_STATUS_INTERNAL_ERROR;

		unsigned char * bh_metadata = NULL;

		if (metadata == NULL || length == NULL)
		{
			ret = BPE_INVALID_PARAMS;
			goto end;
		}

		// bhPlugin will allocate memory in metadata and lenght,
		// which should be freed using bhp_free
		ret = BHP_QueryTEEMetadata(&bh_metadata, length);

		if (ret == BH_SUCCESS)
		{
			*metadata = (unsigned char *)JHI_ALLOC(*length);
			memcpy_s(*metadata, *length, bh_metadata, *length);
			BHP_Free(bh_metadata);
		}

		TRACE1("JHI_Plugin_QueryTeeMetadata end, result = 0x%X", ret);
		
		end:
		return beihaiToTeeError(ret, TEE_STATUS_INTERNAL_ERROR);
	}

	UINT32 BeihaiPlugin::JHI_Plugin_ParsePackage(uint8_t* cmd_pkg, uint32_t pkg_len, OUT PACKAGE_INFO& pkgInfo)
	{
		BH_RET ret = BPE_INTERNAL_ERROR;
		pkgInfo.packageType = AC_CMD_INVALID;
		memset(pkgInfo.uuid, 0, sizeof(pkgInfo.uuid));

		int cmd_type = 0;

		if (cmd_pkg == NULL || pkg_len == 0) 
		{
			ret = BPE_INVALID_PARAMS;
			goto end;
		}

		// parse the package for the command type
		if (ACP_get_cmd_id(cmd_pkg, pkg_len, &cmd_type) != BH_SUCCESS)
		{
			ret = BPE_INVALID_PARAMS;
			goto end;
		}

		pkgInfo.packageType = cmd_type;

		switch (cmd_type){
#if (BEIHAI_ENABLE_SVM || BEIHAI_ENABLE_OEM_SIGNING_IOTG)
		case AC_INSTALL_SD:
		{
			ACInsSDPackExt installSDpack;

			// parse the package for the uuid
			ret = ACP_pload_ins_sd(cmd_pkg, (unsigned int)pkg_len, &installSDpack);

			if (ret != BH_SUCCESS)
			{
				ret = BPE_INVALID_PARAMS;
				goto end;
			}

			uuid_to_string((char*)&installSDpack.cmd_pack.head->sd_id, (char*)pkgInfo.uuid);

			ret = BH_SUCCESS;
			break;
		}
		case AC_UNINSTALL_SD:
		{
			ACUnsSDPackExt uninstallSDpack;

			// parse the package for the uuid
			ret = ACP_pload_uns_sd(cmd_pkg, (unsigned int)pkg_len, &uninstallSDpack);

			if (ret != BH_SUCCESS)
			{
				ret = BPE_INVALID_PARAMS;
				goto end;
			}

			uuid_to_string((char*)&uninstallSDpack.cmd_pack.p_sdid, (char*)pkgInfo.uuid);

			ret = BH_SUCCESS;
			break;
		}
#endif
#if BEIHAI_ENABLE_NATIVETA
		case AC_INSTALL_NTA:
			//ret = bh_do_install_nta(handle, cmd_pkg, pkg_len);
			break;
		case AC_UNINSTALL_NTA:
			//ret = bh_do_uninstall_nta(handle, cmd_pkg, pkg_len);
			break;
#endif
		case AC_INSTALL_JTA:
			{
				ACInsJTAPackExt installJTApack;

				// parse the package for the uuid
				ret = ACP_pload_ins_jta(cmd_pkg, (unsigned int)pkg_len, &installJTApack);
				if (ret != BH_SUCCESS)
				{
					ret = BPE_INVALID_PARAMS;
					goto end;
				}

				uuid_to_string((char*)&installJTApack.cmd_pack.head->ta_id, (char*)pkgInfo.uuid);

				ret = BH_SUCCESS;
				break;
			}
		case AC_UNINSTALL_JTA:
			{
				ACUnsTAPackExt uninstallJTApack;

				// parse the package for the uuid
				ret = ACP_pload_uns_jta(cmd_pkg, pkg_len, &uninstallJTApack);
				if (ret != BH_SUCCESS)
				{
					ret = BPE_INVALID_PARAMS;
					goto end;
				}

				uuid_to_string((char*)uninstallJTApack.cmd_pack.p_taid, (char*)pkgInfo.uuid);

				ret = BH_SUCCESS;
				break;
			}
		case AC_UPDATE_SVL:
			ret = BH_SUCCESS;
			break;
		default:
			ret = BHE_INVALID_BPK_FILE;
			break;
		}

		end:
		return beihaiToTeeError(ret, TEE_STATUS_INTERNAL_ERROR);
	}

	UINT32 BeihaiPlugin::JHI_Plugin_DownloadApplet(const char *pAppId, uint8_t* pAppBlob, unsigned int BlobSize)
	{
		TRACE0("JHI_Plugin_DownloadApplet start");
		BH_RET ret = BPE_INTERNAL_ERROR;

		// first check if there are open sessions:
		UINT32 appletSessionsCount = 0;
		appletSessionsCount = getTA_SessionCount(pAppId);
		if (appletSessionsCount > 0)
		{
			return JHI_INSTALL_FAILURE_SESSIONS_EXISTS;
		}
#ifndef OPEN_INTEL_SD_SESSION_ONCE
		//first open the SD
		ret = openIntelSD();
		if (ret != BH_SUCCESS)
		{
			return ret;
		}
#endif

		ret = BHP_SendAdminCmdPkg(intel_sd_handle, (char*) pAppBlob, BlobSize);

#ifndef OPEN_INTEL_SD_SESSION_ONCE
		//close the SD
		closeIntelSD();
#endif

		TRACE1("JHI_Plugin_DownloadApplet end, result = 0x%X", ret);
		return beihaiToJhiError(ret,JHI_INTERNAL_ERROR);
	}

	UINT32 BeihaiPlugin::JHI_Plugin_UnloadApplet(const char *pAppId)
	{
		TRACE0("JHI_Plugin_UnloadApplet start");
#ifndef OPEN_INTEL_SD_SESSION_ONCE
		//first open the SD
		BH_RET ret3 = openIntelSD();
		if (ret3 != BH_SUCCESS)
		{
			return ret3;
		}
#endif

		char* uninstallPkg = NULL;

		BH_RET ret;

		setUninstallPack(pAppId, &uninstallPkg);
		if (uninstallPkg == NULL)
		{
			return JHI_INTERNAL_ERROR;
		}

		TRACE1("uninstalling applet: %s.", pAppId);
		ret = BHP_SendAdminCmdPkg(intel_sd_handle, (const char*)uninstallPkg, UNINSTALL_PACK_LEN);

#ifndef OPEN_INTEL_SD_SESSION_ONCE
		//close the SD
		BH_RET ret2 = closeIntelSD();
#endif

		//cleanup no matter what
		memory_api.freeMemory(uninstallPkg);

		TRACE1("JHI_Plugin_UnloadApplet end, result = 0x%X", ret);
		return beihaiToJhiError(ret,JHI_INTERNAL_ERROR);
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

			BHP_Free(*output);
			*output = NULL;

			*output = (char*)memory_api.allocateMemory(6);

#ifdef _WIN32
			_itoa_s(versionUINT, *output, 6, 10);
#else
			string s_version = to_string(versionUINT);
			memcpy_s(*output, 6, s_version.c_str(), 6);
#endif
			return true;
		}
		catch (...)
		{
			return false;
		}
	}

#ifdef GET_APPLET_PROPERTY_NAMES_W_A
	void convertAppProperty(char* input, char** output, int* outputLen)
	{
		if (output == NULL)
		{
			return;
		}

		string newString;
		if (string(input).compare(string("security.version")) == 0)
		{
			newString = string("svn");
		}
		//made up
		/*
		else if (string(input).compare(string("applet.version")) == 0)
		{
		newString = string("ta_version");
		}
		else if (string(input).compare(string("applet.flash.quota")) == 0)
		{
		newString = string("flash_quota");
		}
		else if (string(input).compare(string("applet.version")) == 0)
		{
		newString = string("ta_version");
		}
		else if (string(input).compare(string("")) == 0)
		{
		newString = string("");
		}*/
		else 
		{
			newString = string(input);
		}

		*output = (char*)memory_api.allocateMemory(newString.length() + 1);
		*outputLen = newString.length();
		memset(*output, 0, newString.length() + 1);
		memcpy(*output, newString.c_str(), newString.length());
	}
#endif

	UINT32 BeihaiPlugin::JHI_Plugin_GetAppletProperty(const char *AppId, JVM_COMM_BUFFER *pIOBuffer)
	{
		TRACE0("JHI_Plugin_GetAppletProperty start");
		UINT32 ret = JHI_INTERNAL_ERROR;
		char* inputBuffer = (char*) pIOBuffer->TxBuf->buffer;
		int inputBufferLength = pIOBuffer->TxBuf->length -1;
		string AppProperty_Version = "applet.version";
		bool versionQuery = false;
		char* output = NULL;
		int outputLength = 0;

		char* outputBuffer = (char*) pIOBuffer->RxBuf->buffer;
		int* outputBufferLength = (int*) &pIOBuffer->RxBuf->length; // number of characters without /0, not size of buffer

#ifdef GET_APPLET_PROPERTY_NAMES_W_A
		char* newProperty = NULL;
		int newPropertyLen = 0;
		convertAppProperty(inputBuffer, &newProperty, &newPropertyLen);
#endif


#ifdef GET_APPLET_PROPERTY_NAMES_W_A
		ret = BHP_QueryTAProperty(const_cast<char*>(AppId), newProperty, newPropertyLen, &output);
#else
		ret = BHP_QueryTAProperty(const_cast<char*>(AppId), inputBuffer, inputBufferLength, &output);
#endif

		if (ret == BH_SUCCESS && output != NULL)
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
				ret = JHI_INSUFFICIENT_BUFFER;
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

		ret = beihaiToJhiError(ret,JHI_INTERNAL_ERROR);

cleanup:

		if (output != NULL)
		{
			if (versionQuery)
			{
				memory_api.freeMemory(output);
			}
			else
			{
				BHP_Free(output);
			}
		}

		TRACE1("JHI_Plugin_GetAppletProperty end, result = 0x%X", ret);
		return ret;
	}

	BH_RET BeihaiPlugin::sendSessionIDtoApplet(VM_SESSION_HANDLE* pSession, JHI_SESSION_ID SessionID, int* appletResponse)
	{
		//TRACE0("sendSessionIDtoApplet start");
		// TODO: BH bug w/a - unable to send null output buffer.
		char temp[] = "output\0";
		char* pOutput = temp;
		BH_RET outputLength = 0;

		char Uuid[sizeof(JHI_SESSION_ID)];
		memcpy_s(Uuid,sizeof(JHI_SESSION_ID),&SessionID,sizeof(JHI_SESSION_ID));
		// the value '1' in the 'what' field is internally reserved for passing the SessionID
        BH_RET ret = BHP_SendAndRecvInternal( *pSession, 1, 0, Uuid, sizeof(JHI_SESSION_ID), (void**)&pOutput, (unsigned int *)&outputLength, appletResponse);
		//TRACE1("sendSessionIDtoApplet end, result = 0x%X", ret);
		return ret;
	}

	BH_RET BeihaiPlugin::openIntelSD()
	{
		TRACE0("openIntelSD start");
		if (is_intel_sd_open)
		{
			TRACE1("openIntelSD end, result = 0x%X", BH_SUCCESS);
			return BH_SUCCESS;
		}
		BH_RET ret = BHP_OpenSDSession(INTEL_SD_UUID, &intel_sd_handle);
		if (ret == BH_SUCCESS)
		{
			is_intel_sd_open = true;
		}
		TRACE1("openIntelSD end, result = 0x%X", ret);
		return ret;
	}

	BH_RET BeihaiPlugin::closeIntelSD()
	{
		TRACE0("closeIntelSD start");
		if (!is_intel_sd_open)
		{
			TRACE1("closeIntelSD end, result = 0x%X", BH_SUCCESS);
			return BH_SUCCESS;
		}
		BH_RET ret = BHP_CloseSDSession(intel_sd_handle);
		// Ignoring the return value because even if it fails usually the
		// the SD session will not be valid.
		intel_sd_handle = NULL;
		is_intel_sd_open = false;

		TRACE1("closeIntelSD end, result = 0x%X", ret);
		return ret;
	}

	UINT32 BeihaiPlugin::JHI_Plugin_CreateSession(const char *AppId, VM_SESSION_HANDLE* pSession, const uint8_t* pAppBlob, unsigned int BlobSize, JHI_SESSION_ID SessionID, DATA_BUFFER* initBuffer)
	{
		TRACE1("JHI_Plugin_CreateSession start: %s", AppId);
		ACInsJTAPackExt ta_pack = { 0 };
		unsigned int ta_size = 0;

		ACP_pload_ins_jta(pAppBlob, BlobSize, &ta_pack);
		ta_size = BlobSize - (unsigned int)((char*)ta_pack.ta_pack - (char*)pAppBlob);

		BH_RET ret = BHP_OpenTASession(pSession, const_cast<char*>(AppId), ta_pack.ta_pack, ta_size, (char*)initBuffer->buffer, initBuffer->length);
		if (ret == BH_SUCCESS)
		{
			// sending the SessionID to the applet.
			int appletResponse = -1;
			BH_RET ret2 = sendSessionIDtoApplet(pSession, SessionID, &appletResponse);
			if (ret2 != BH_SUCCESS || appletResponse != 0)
			{
				TRACE1("JHI_Plugin_CreateSession->sendSessionIDtoApplet failed, result = 0x%X", ret2);
				return JHI_INTERNAL_ERROR;
			}
		}
		TRACE2("JHI_Plugin_CreateSession end, result = 0x%X Appid = %s", ret, AppId);
		return beihaiToJhiError(ret,JHI_INTERNAL_ERROR);
	}

	UINT32 BeihaiPlugin::JHI_Plugin_ForceCloseSession(VM_SESSION_HANDLE* pSession)
	{
		TRACE0("JHI_Plugin_CloseSpoolerSession start");
		int ret = BHP_ForceCloseTASession(*pSession);
		if (ret != BH_SUCCESS)
		{
			beihaiToJhiError(ret, JHI_INTERNAL_ERROR); //called just for the debug output
		}
		TRACE1("JHI_Plugin_ForceCloseSession end, result = 0x%X", ret);
		return beihaiToJhiError(ret, JHI_INTERNAL_ERROR);
	}

	UINT32 BeihaiPlugin::JHI_Plugin_CloseSession(VM_SESSION_HANDLE* pSession)
	{
		TRACE0("JHI_Plugin_CloseSession start");
		int ret = BHP_CloseTASession(*pSession);

		TRACE1("JHI_Plugin_CloseSession end, result = 0x%X", ret);
		return beihaiToJhiError(ret, JHI_INTERNAL_ERROR);
	}

	UINT32 BeihaiPlugin::JHI_Plugin_WaitForSpoolerEvent(VM_SESSION_HANDLE SpoolerSession,JHI_EVENT_DATA** ppEventData, JHI_SESSION_ID* targetSession)
	{
		TRACE0("JHI_Plugin_WaitForSpoolerEvent start");

		// When a SendAndReceive flow is active, the FW can't go down to PG because it changes to high performance mode.
		// This header for Spooler messages informs the FW that this is a Spooler SendAndReceive and that it shouldn't
		// change to the high performance mode.
        const char spoolerIdentifierMsg[] = {'S', 'P', 'L', 'R'};

		UINT32 ret = JHI_INTERNAL_ERROR;
		JVM_COMM_BUFFER IOBuffer;
		int responseCode = 0;

		// allocate output buffer
		IOBuffer.RxBuf->length = JHI_EVENT_DATA_BUFFER_SIZE + sizeof(JHI_SESSION_ID);
		IOBuffer.RxBuf->buffer = memory_api.allocateMemory(IOBuffer.RxBuf->length);

		if (!IOBuffer.RxBuf->buffer)
			return JHI_INTERNAL_ERROR;

		memset(IOBuffer.RxBuf->buffer, 0, IOBuffer.RxBuf->length);
        
        // allocate input buffer
		IOBuffer.TxBuf->length = sizeof(spoolerIdentifierMsg);
		IOBuffer.TxBuf->buffer = (PVOID)spoolerIdentifierMsg;

		*ppEventData = (JHI_EVENT_DATA*) memory_api.allocateMemory(sizeof(JHI_EVENT_DATA));

		if (!(*ppEventData))
		{
			TRACE0("WaitForSpoolerEvent: Memory allocation error!");
			memory_api.freeMemory(IOBuffer.RxBuf->buffer);
            IOBuffer.RxBuf->buffer = NULL;
            return JHI_INTERNAL_ERROR;
		}

		(*ppEventData)->data = NULL;
		(*ppEventData)->datalen = 0;

		ret = JHI_Plugin_SendAndRecv(SpoolerSession, SPOOLER_COMMAND_GET_EVENT, &IOBuffer,&responseCode);

		// check nError to see if all copied or need to extend the buffer...
		if (ret == JHI_INSUFFICIENT_BUFFER)
		{
			// reallocate the buffer
			memory_api.freeMemory(IOBuffer.RxBuf->buffer);
			IOBuffer.RxBuf->buffer = (UINT8*) memory_api.allocateMemory(IOBuffer.RxBuf->length);

			if (!IOBuffer.RxBuf->buffer)
			{
				TRACE0("WaitForSpoolerEvent: Memory allocation error!");
				memory_api.freeMemory(*ppEventData);
				*ppEventData = NULL;
				return JHI_INTERNAL_ERROR;
			}

			// call again with the larger buffer
			ret = JHI_Plugin_SendAndRecv(SpoolerSession, SPOOLER_COMMAND_GET_EVENT, &IOBuffer,&responseCode);
		}

		if (ret == JHI_SUCCESS && responseCode == JHI_SUCCESS)
		{
			if( IOBuffer.RxBuf->length < sizeof(JHI_SESSION_ID) )
			{
				TRACE1("Spooler data is too short - must contain session uuid at least. Length: %d", IOBuffer.RxBuf->length);
				return JHI_INTERNAL_ERROR;
			}

			(*targetSession) = *((JHI_SESSION_ID*) IOBuffer.RxBuf->buffer);

			(*ppEventData)->datalen = IOBuffer.RxBuf->length - sizeof(JHI_SESSION_ID);

			if ((*ppEventData)->datalen > 0)
			{
				(*ppEventData)->data = (uint8_t*) memory_api.allocateMemory((*ppEventData)->datalen);

				if ((*ppEventData)->data == NULL)
				{
					TRACE0("WaitForSpoolerEvent: Memory allocation error!");
					memory_api.freeMemory(*ppEventData);
					*ppEventData = NULL;

					memory_api.freeMemory(IOBuffer.RxBuf->buffer);
					IOBuffer.RxBuf->buffer = NULL;

					return JHI_INTERNAL_ERROR;
				}

				memcpy_s((*ppEventData)->data,(*ppEventData)->datalen, (UINT8*)IOBuffer.RxBuf->buffer + sizeof(JHI_SESSION_ID),(*ppEventData)->datalen);
			}

			(*ppEventData)->dataType = JHI_DATA_FROM_APPLET;
		}
		else
		{
			TRACE2("Spooler event retrieval failed. Return code: 0x%X, Response code: 0x%X", ret, responseCode);
			memory_api.freeMemory(*ppEventData);
			*ppEventData = NULL;
		}

		memory_api.freeMemory(IOBuffer.RxBuf->buffer);
		IOBuffer.RxBuf->buffer = NULL;

		TRACE0("JHI_Plugin_WaitForSpoolerEvent finished successfully");
		return ret;
	}

	UINT32 BeihaiPlugin::JHI_Plugin_SendAndRecv(VM_SESSION_HANDLE Session, INT32 nCommandId, JVM_COMM_BUFFER *pIOBuffer, INT32* pResponseCode)
	{
		TRACE0("JHI_Plugin_SendAndRecv start");
		UINT32 ret = JHI_INTERNAL_ERROR;
		char* inputBuffer = (char*) pIOBuffer->TxBuf->buffer;
		int inputBufferLength = pIOBuffer->TxBuf->length;

		char* outputBuffer = (char*) pIOBuffer->RxBuf->buffer;
		int* outputBufferLength = (int*) &pIOBuffer->RxBuf->length;

		char* output = NULL;
		int outputLength = *outputBufferLength; // TODO: tell BH to change this. no need to provide max buffer size.

		ret = BHP_SendAndRecv (Session, nCommandId, inputBuffer, inputBufferLength, (void **)&output, (unsigned int *)&outputLength, pResponseCode);

		if (ret == BH_SUCCESS && output != NULL)
		{
			// TODO: same as above
			//if (*outputBufferLength < outputLength)
			//{
			//	// buffer provided is too small for the response
			//	TRACE2("JHI_Plugin_SendAndRecv: insufficient buffer sent to VM, expected: %d, received: %d\n", outputLength, *outputBufferLength);
			//	ret = JHI_INSUFFICIENT_BUFFER;
			//	*outputBufferLength = outputLength;
			//	goto cleanup;
			//}

			// copy the output to the output buffer
			memcpy_s(outputBuffer, *outputBufferLength, output, outputLength);	
		}

		*outputBufferLength = outputLength;

		ret = beihaiToJhiError(ret,JHI_INTERNAL_ERROR);

		if (output)
			BHP_Free(output);

		TRACE1("JHI_Plugin_SendAndRecv end, result = 0x%X", ret);
		return ret;
	}

	UINT32 BeihaiPlugin::beihaiToJhiError(int bhError, UINT32 defaultError)
	{
		UINT32 jhiError = JHI_INTERNAL_ERROR;

		switch (bhError)
		{
		case BH_SUCCESS: 
			jhiError = JHI_SUCCESS;
			break;

		case BPE_INVALID_PARAMS:
			jhiError = JHI_INVALID_PARAMS;
			break;

			// SendAndRecv
		case BHE_INSUFFICIENT_BUFFER:
		case BHE_APPLET_SMALL_BUFFER:
		case HAL_BUFFER_TOO_SMALL:
			jhiError = JHI_INSUFFICIENT_BUFFER;
			break;

		case BPE_COMMS_ERROR:
		case BPE_NOT_INIT:
		case BPE_NO_CONNECTION_TO_FIRMWARE:
			jhiError = JHI_NO_CONNECTION_TO_FIRMWARE;
			break;

		case BHE_VM_INSTANCE_INIT_FAIL:
		case BHE_OUT_OF_MEMORY:
			jhiError = JHI_FIRMWARE_OUT_OF_RESOURCES;
			break;

		case HAL_OUT_OF_MEMORY:
		case BHE_UNCAUGHT_EXCEPTION:
		case BHE_APPLET_CRASHED:
		case BHE_WD_TIMEOUT:
		case HAL_TIMED_OUT:
		case BHE_APPLET_GENERIC: // not documented but sometimes recieved. (usually an exception thrown in onInit)
		case BHE_BAD_STATE: // not documented but sometimes recieved. (might be related to max heap)
			// TODO: 
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
		case BHE_SDM_SIGNATURE_VERIFY_FAIL:
			jhiError = JHI_FILE_ERROR_AUTH;
			break;

		case BHE_TA_PACKAGE_HASH_VERIFY_FAIL:
		case BHE_INVALID_BPK_FILE:
			jhiError = JHI_INVALID_PACKAGE_FORMAT;
			break;

		case HAL_ALREADY_INSTALLED:
		case BHE_SDM_ALREADY_EXIST:
			jhiError = JHI_FILE_IDENTICAL;
			break;

		case HAL_OUT_OF_RESOURCES:
		case BHE_SDM_TA_NUMBER_LIMIT:
			jhiError = JHI_MAX_INSTALLED_APPLETS_REACHED;
			break;

		case BHE_SDM_SVL_CHECK_FAIL:
			jhiError = JHI_SVL_CHECK_FAIL;
			break;

		case BHE_SDM_SVN_CHECK_FAIL:
			jhiError = JHI_SVN_CHECK_FAIL;
			break;

			// UnloadApplet
		case BHE_EXIST_LIVE_SESSION:
			jhiError = JHI_UNINSTALL_FAILURE_SESSIONS_EXISTS;
			break;

		case BHE_PACKAGE_NOT_FOUND:
		case BHE_SDM_NOT_FOUND:
			jhiError = JHI_APPLET_NOT_INSTALLED;
			break;

			// JHI_Plugin_GetAppletProperty
		case BHE_QUERY_PROP_NOT_SUPPORT:
			jhiError = JHI_APPLET_PROPERTY_NOT_SUPPORTED;
			break;

			// IAC errors
		case BHE_IAC_SERVICE_HOST_SESSION_NUM_EXCEED:
			jhiError = JHI_IAC_SERVER_SESSION_EXIST;
			break;

		case BHE_IAC_EXIST_INTERNAL_SESSION:
			jhiError = JHI_IAC_SERVER_INTERNAL_SESSIONS_EXIST;
			break;

			// Access control errors
		case BHE_GROUP_CHECK_FAIL:
			jhiError = JHI_MISSING_ACCESS_CONTROL;
			break;

		case BHE_SESSION_NUM_EXCEED:
			jhiError = JHI_MAX_SESSIONS_REACHED;
			break;

        case HAL_ILLEGAL_PLATFORM_ID:
            jhiError = JHI_ILLEGAL_PLATFORM_ID;
            break;
            
        case BHE_ONLY_SINGLE_INSTANCE_ALLOWED:
            jhiError = JHI_ONLY_SINGLE_INSTANCE_ALLOWED;
            break;

		case BHE_SDM_SD_INTERFACE_DISABLED:
			jhiError = JHI_ERROR_OEM_SIGNING_DISABLED;
			break;
           
		case BHE_SDM_SD_PUBLICKEY_HASH_VERIFY_FAIL:
			jhiError = JHI_ERROR_SD_PUBLICKEY_HASH_FAILED;
			break;

		case BHE_SDM_SD_DB_NO_FREE_SLOT:
			jhiError = JHI_ERROR_SD_DB_NO_FREE_SLOT;
			break;

		case BHE_SDM_TA_INSTALL_UNALLOWED:
			jhiError = JHI_ERROR_SD_TA_INSTALLATION_UNALLOWED;
			break;

		case BHE_OPERATION_NOT_PERMITTED:
			jhiError = JHI_OPERATION_NOT_PERMITTED;
			break;

		default:
			jhiError = defaultError;
		}

		if (jhiError != JHI_SUCCESS)
		{
			TRACE4("beihaiToJhiError: BH Error received - 0x%X (%s), translated to JHI Error - 0x%X (%s)\n" ,bhError, BHErrorToString(bhError), jhiError, JHIErrorToString(jhiError));
		}
		return jhiError;	
	}

	UINT32 BeihaiPlugin::beihaiToTeeError(int bhError, UINT32 defaultError)
	{
		UINT32 teeError = TEE_STATUS_INTERNAL_ERROR;

		switch (bhError)
		{
		case BH_SUCCESS: 
			teeError = TEE_STATUS_SUCCESS;
			break; 

		case BPE_INVALID_PARAMS:
			teeError = TEE_STATUS_INVALID_PARAMS;
			break;

		case BPE_COMMS_ERROR:
		case BPE_NOT_INIT:
		case BPE_NO_CONNECTION_TO_FIRMWARE:
			teeError = TEE_STATUS_NO_FW_CONNECTION;
			break;

			// Send command package
		case HAL_UNSUPPORTED_CPU_TYPE:
		case HAL_UNSUPPORTED_PCH_TYPE:
		case HAL_UNSUPPORTED_FEATURE_SET:
		case HAL_UNSUPPORTED_PLATFORM_TYPE:
			teeError = TEE_STATUS_UNSUPPORTED_PLATFORM;
			break;

		case HAL_ILLEGAL_SIGNATURE:
		case HAL_ILLEGAL_VERSION:
		case HAL_FW_VERSION_MISMATCH:
		case BHE_SDM_SIGNATURE_VERIFY_FAIL:

			teeError = TEE_STATUS_INVALID_SIGNATURE;
			break;

		case BHE_INVALID_BPK_FILE:
		case BHE_TA_PACKAGE_HASH_VERIFY_FAIL:
			teeError = TEE_STATUS_INVALID_PACKAGE;
			break;

			// SVL errors
		case BHE_SDM_SVL_DB_NO_FREE_SLOT:
			teeError = TEE_STATUS_MAX_SVL_RECORDS;
			break;

		case BHE_SDM_SVL_CHECK_FAIL:
			teeError = TEE_STATUS_SVL_CHECK_FAIL;
			break;

		case BHE_SDM_SVN_CHECK_FAIL:
			teeError = TEE_STATUS_INVALID_TA_SVN;
			break;

			// DownloadApplet
		case HAL_OUT_OF_RESOURCES:
		case BHE_SDM_TA_NUMBER_LIMIT:
			teeError = TEE_STATUS_MAX_TAS_REACHED;
			break;

		case BHE_SDM_ALREADY_EXIST:
			teeError = TEE_STATUS_IDENTICAL_PACKAGE;
			break;
				
			// UnloadApplet
		case BHE_EXIST_LIVE_SESSION:
			teeError = TEE_STATUS_CMD_FAILURE_SESSIONS_EXISTS;
			break;

		case BHE_PACKAGE_NOT_FOUND:
		case BHE_SDM_NOT_FOUND:
			teeError = TEE_STATUS_TA_DOES_NOT_EXIST;
			break;

		case BHE_SDM_SD_NOT_FOUND:
			teeError = TEE_STATUS_SD_SD_DOES_NOT_EXIST;
			break;

			// Access control errors
		case BHE_GROUP_CHECK_FAIL:
			teeError = TEE_STATUS_MISSING_ACCESS_CONTROL;
			break;

        case HAL_ILLEGAL_PLATFORM_ID:
            teeError = TEE_STATUS_ILLEGAL_PLATFORM_ID;
            break;

		case BHE_SDM_SD_INTERFACE_DISABLED:
			teeError = TEE_STATUS_SD_INTERFCE_DISABLED;
			break;

		case BHE_SDM_SD_PUBLICKEY_HASH_VERIFY_FAIL:
			teeError = TEE_STATUS_SD_PUBLICKEY_HASH_VERIFY_FAIL;
			break;

		case BHE_SDM_SD_DB_NO_FREE_SLOT:
			teeError = TEE_STATUS_SD_DB_NO_FREE_SLOT;
			break;

		case BHE_SDM_SVL_UPDATE_UNALLOWED:
		case BHE_SDM_TA_INSTALL_UNALLOWED:
			teeError = TEE_STATUS_SD_TA_INSTALLATION_UNALLOWED;
			break;

		case BHE_SDM_SD_INVALID_PROPERTIES:
		case BHE_SDM_PERMGROUP_CHECK_FAIL:
			teeError = TEE_STATUS_SD_INVALID_PROPERTIES;
			break;

		case BHE_SDM_SD_INSTALL_UNALLOWED:
			teeError = TEE_STATUS_SD_SD_INSTALL_UNALLOWED;
			break;

		default:
			teeError = defaultError;
		}

		if (teeError != JHI_SUCCESS)
		{
			TRACE4("beihaiToTeeError: BH Error received - 0x%X (%s), translated to TEE Error - 0x%X (%s)\n" ,bhError, BHErrorToString(bhError), teeError, TEEErrorToString(teeError));
		}

		return teeError;	
	}

	const char* BeihaiPlugin::BHErrorToString(UINT32 bh_error)
	{
		const char* str = "";
		switch (bh_error)
		{
		default:
			str = "BH_UNKNOWN_ERROR";
			break;

			// Errors from BeihaiStatusHAL.h
		//case	HAL_SUCCESS:						   str = "HAL_SUCCESS";							break;	//0x00000000
		case	HAL_TIMED_OUT:						    str = "HAL_TIMED_OUT";						break;	//0x00001001
		case	HAL_FAILURE:						    str = "HAL_FAILURE";						break;	//0x00001002
		case	HAL_OUT_OF_RESOURCES:				    str = "HAL_OUT_OF_RESOURCES";				break;	//0x00001003
		case	HAL_OUT_OF_MEMORY:					    str = "HAL_OUT_OF_MEMORY";					break;	//0x00001004
		case	HAL_BUFFER_TOO_SMALL:				    str = "HAL_BUFFER_TOO_SMALL";				break;	//0x00001005
		case	HAL_INVALID_HANDLE:					    str = "HAL_INVALID_HANDLE";					break;	//0x00001006
		case	HAL_NOT_INITIALIZED:				    str = "HAL_NOT_INITIALIZED";				break;	//0x00001007
		case	HAL_INVALID_PARAMS:					    str = "HAL_INVALID_PARAMS";					break;	//0x00001008
		case	HAL_NOT_SUPPORTED:					    str = "HAL_NOT_SUPPORTED";					break;	//0x00001009
		case	HAL_NO_EVENTS:						    str = "HAL_NO_EVENTS";						break;	//0x0000100A
		case	HAL_NOT_READY:						    str = "HAL_NOT_READY";						break;	//0x0000100B
		case	HAL_CONNECTION_CLOSED:				    str = "HAL_CONNECTION_CLOSED";				break;	//0x0000100C
		case	HAL_INTERNAL_ERROR:					    str = "HAL_INTERNAL_ERROR";					break;	//0x00001100
		case	HAL_ILLEGAL_FORMAT:					    str = "HAL_ILLEGAL_FORMAT";					break;	//0x00001101
		case	HAL_LINKER_ERROR:					    str = "HAL_LINKER_ERROR";					break;	//0x00001102
		case	HAL_VERIFIER_ERROR:					    str = "HAL_VERIFIER_ERROR";					break;	//0x00001103
		// User defined applet & session errors to be returned to the host (should be exposed also in the host DLL)	
		case	HAL_FW_VERSION_MISMATCH:				str = "HAL_FW_VERSION_MISMATCH";			break;	//0x00002000
		case	HAL_ILLEGAL_SIGNATURE:					str = "HAL_ILLEGAL_SIGNATURE";				break;	//0x00002001
		case	HAL_ILLEGAL_POLICY_SECTION:				str = "HAL_ILLEGAL_POLICY_SECTION";			break;	//0x00002002
		case	HAL_OUT_OF_STORAGE:						str = "HAL_OUT_OF_STORAGE";					break;	//0x00002003
		case	HAL_UNSUPPORTED_PLATFORM_TYPE:			str = "HAL_UNSUPPORTED_PLATFORM_TYPE";		break;	//0x00002004
		case	HAL_UNSUPPORTED_CPU_TYPE:				str = "HAL_UNSUPPORTED_CPU_TYPE";			break;	//0x00002005
		case	HAL_UNSUPPORTED_PCH_TYPE:				str = "HAL_UNSUPPORTED_PCH_TYPE";			break;	//0x00002006
		case	HAL_UNSUPPORTED_FEATURE_SET:			str = "HAL_UNSUPPORTED_FEATURE_SET";		break;	//0x00002007
		case	HAL_ILLEGAL_VERSION:					str = "HAL_ILLEGAL_VERSION";				break;	//0x00002008
		case	HAL_ALREADY_INSTALLED:					str = "HAL_ALREADY_INSTALLED";				break;	//0x00002009
		case	HAL_MISSING_POLICY:						str = "HAL_MISSING_POLICY";					break;	//0x00002010
		case	HAL_ILLEGAL_PLATFORM_ID:				str = "HAL_ILLEGAL_PLATFORM_ID";			break;  //0x00002011
		case	HAL_UNSUPPORTED_API_LEVEL:				str = "HAL_UNSUPPORTED_API_LEVEL";			break;  //0x00002012
		case	HAL_LIBRARY_VERSION_MISMATCH:			str = "HAL_LIBRARY_VERSION_MISMATCH";		break;  //0x00002013


			// Errors from bh_shared_errcode.h

		case	BH_SUCCESS:							str = "BH_SUCCESS";							break;	//0x0											

			/////BHP specific error code section: //0x000		
		case	BPE_NOT_INIT:					    str = "BPE_NOT_INIT";						break;	//0x001											
		case	BPE_SERVICE_UNAVAILABLE:		    str = "BPE_SERVICE_UNAVAILABLE";			break;	//0x002											
		case	BPE_INTERNAL_ERROR:				    str = "BPE_INTERNAL_ERROR";					break;	//0x003											
		case	BPE_COMMS_ERROR:				    str = "BPE_COMMS_ERROR";					break;	//0x004											
		case	BPE_OUT_OF_MEMORY:				    str = "BPE_OUT_OF_MEMORY";					break;	//0x005											
		case	BPE_INVALID_PARAMS:				    str = "BPE_INVALID_PARAMS";					break;	//0x006											
		case	BPE_MESSAGE_TOO_SHORT:			    str = "BPE_MESSAGE_TOO_SHORT";				break;	//0x007											
		case	BPE_MESSAGE_ILLEGAL:			    str = "BPE_MESSAGE_ILLEGAL";				break;	//0x008											
		case	BPE_NO_CONNECTION_TO_FIRMWARE:	    str = "BPE_NO_CONNECTION_TO_FIRMWARE";		break;	//0x009											
		case	BPE_NOT_IMPLEMENT:				    str = "BPE_NOT_IMPLEMENT";					break;	//0x00A											
		case	BPE_OUT_OF_RESOURCE:			    str = "BPE_OUT_OF_RESOURCE";				break;	//0x00B											
		case	BPE_INITIALIZED_ALREADY:		    str = "BPE_INITIALIZED_ALREADY";			break;	//0x00C											
		case	BPE_CONNECT_FAILED:				    str = "BPE_CONNECT_FAILED";					break;	//0x00D			
			//////////////////////////////////////////////////


			//General error code section for Beihai on Firmware: //0x100	
		case	BHE_OUT_OF_MEMORY:					str = "BHE_OUT_OF_MEMORY";					break;	//0x101		
			/* Bad parameters to native */
		case	BHE_BAD_PARAMETER:					str = "BHE_BAD_PARAMETER";					break;	//0x102											
		case	BHE_INSUFFICIENT_BUFFER:			str = "BHE_INSUFFICIENT_BUFFER";			break;	//0x103											
		case	BHE_MUTEX_INIT_FAIL:				str = "BHE_MUTEX_INIT_FAIL";				break;	//0x104											
		case	BHE_COND_INIT_FAIL:					str = "BHE_COND_INIT_FAIL";					break;	//0x105	
			/* Watchdog time out */
		case	BHE_WD_TIMEOUT:						str = "BHE_WD_TIMEOUT";						break;	//0x106											
		case	BHE_FAILED:							str = "BHE_FAILED";							break;	//0x107											
		case	BHE_INVALID_HANDLE:					str = "BHE_INVALID_HANDLE";					break;	//0x108			
			/* IPC error code */
		case	BHE_IPC_ERR_DEFAULT:				str = "BHE_IPC_ERR_DEFAULT";				break;	//0x109											
		case	BHE_IPC_ERR_PLATFORM:				str = "BHE_IPC_ERR_PLATFORM";				break;	//0x10A											
		case	BHE_IPC_SRV_INIT_FAIL:				str = "BHE_IPC_SRV_INIT_FAIL";				break;	//0x10B			
			//////////////////////////////////////////////////


			//VM communication error code section: //0x200			
		case	BHE_MAILBOX_NOT_FOUND:				str = "BHE_MAILBOX_NOT_FOUND";				break;	//0x201											
			//case	BHE_APPLET_CRASHED:					str = "BHE_APPLET_CRASHED";					break;	//BHE_MAILBOX_NOT_FOUND											
		case	BHE_MSG_QUEUE_IS_FULL:				str = "BHE_MSG_QUEUE_IS_FULL";				break;	//0x202											
			/*	Mailbox			is	denied	by	firewall	*/							
		case	BHE_MAILBOX_DENIED:					str = "BHE_MAILBOX_DENIED";					break;	//0x203			
			//////////////////////////////////////////////////

			//Firmware thread/mutex error code section: //0x280	
		case	BHE_THREAD_ERROR:					str = "BHE_THREAD_ERROR";					break;	//0x281											
		case	BHE_THREAD_TIMED_OUT:				str = "BHE_THREAD_TIMED_OUT";				break;	//0x282											


			//Applet manager error code section: //0x300
			/* JEFF file load fail, OOM or file format error not distinct by VM*/
		case	BHE_LOAD_JEFF_FAIL:					str = "BHE_LOAD_JEFF_FAIL";					break;	//0x303				
			/* Request operation on a package, but it does not exist.*/
		case	BHE_PACKAGE_NOT_FOUND:				str = "BHE_PACKAGE_NOT_FOUND";				break;	//0x304		
			/* Uninstall package fail because of live session exist.*/
		case	BHE_EXIST_LIVE_SESSION:				str = "BHE_EXIST_LIVE_SESSION";				break;	//0x305			
			/* VM instance init fail when create session.*/
		case	BHE_VM_INSTANCE_INIT_FAIL:			str = "BHE_VM_INSTANCE_INIT_FAIL";			break;	//0x306		
			/* Query applet property that Beihai does not support.*/
		case	BHE_QUERY_PROP_NOT_SUPPORT:			str = "BHE_QUERY_PROP_NOT_SUPPORT";			break;	//0x307	
			/* Incorrect Beihai package format */
		case	BHE_INVALID_BPK_FILE:				str = "BHE_INVALID_BPK_FILE";				break;	//0x308		
			/* Download a package which has already exists in app manager*/
		case	BHE_PACKAGE_EXIST:					str = "BHE_PACKAGE_EXIST";					break;	//0x309
			/* VM instance not found */
		case	BHE_VM_INSTNACE_NOT_FOUND:			str = "BHE_VM_INSTNACE_NOT_FOUND";			break;	//0x312		
			/* JDWP agent starting fail */
		case	BHE_STARTING_JDWP_FAIL:				str = "BHE_STARTING_JDWP_FAIL";				break;	//0x313											
			/* Group access checking fail*/
		case	BHE_GROUP_CHECK_FAIL:				str = "BHE_GROUP_CHECK_FAIL";				break;	//0x314											
			/* package SDID dose not equal to the effective one in app manager*/
		case	BHE_SDID_UNMATCH:					str = "BHE_SDID_UNMATCH";					break;	//0x315											
		case	BHE_APPPACK_UNINITED:				str = "BHE_APPPACK_UNINITED";				break;	//0x316											
		case	BHE_SESSION_NUM_EXCEED:				str = "BHE_SESSION_NUM_EXCEED";				break;	//0x317											
			/* TA package verify failure */								
		case	BHE_TA_PACKAGE_HASH_VERIFY_FAIL:	str = "BHE_TA_PACKAGE_HASH_VERIFY_FAIL";	break;	//0x318											
			/*SDID has not been reset to invalid one
			case BHE_SDID_NOT_RESET:					str = "BHE_SDID_NOT_RESET";					break;	//0x316
			*/
			//////////////////////////////////////////////////														


			//VM Applet instance error code section: //0x400					
		case	BHE_UNCAUGHT_EXCEPTION:				str = "BHE_UNCAUGHT_EXCEPTION";				break;	//0x401											
			/* Bad parameters to applet */					
		case	BHE_APPLET_BAD_PARAMETER:			str = "BHE_APPLET_BAD_PARAMETER";			break;	//0x402											
			/* Small response buffer */		
		case	BHE_APPLET_SMALL_BUFFER:			str = "BHE_APPLET_SMALL_BUFFER";			break;	//0x403											

		case    BHE_ONLY_SINGLE_INSTANCE_ALLOWED:   str = "BHE_ONLY_SINGLE_INSTANCE_ALLOWED";   break;  //0x406
        
			/*TODO: Should be removed these UI error code when integrate with ME 9 */
		case	BHE_UI_EXCEPTION:					str = "BHE_UI_EXCEPTION";					break;	//0x501											
		case	BHE_UI_ILLEGAL_USE:					str = "BHE_UI_ILLEGAL_USE";					break;	//0x502											
		case	BHE_UI_ILLEGAL_PARAMETER:			str = "BHE_UI_ILLEGAL_PARAMETER";			break;	//0x503											
		case	BHE_UI_NOT_INITIALIZED:				str = "BHE_UI_NOT_INITIALIZED";				break;	//0x504											
		case	BHE_UI_NOT_SUPPORTED:				str = "BHE_UI_NOT_SUPPORTED";				break;	//0x505											
		case	BHE_UI_OUT_OF_RESOURCES:			str = "BHE_UI_OUT_OF_RESOURCES";			break;	//0x506				

			//////////////////////////////////////////////////

			//BeiHai VMInternalError code section: //0x600
		case	BHE_UNKOWN:							str = "BHE_UNKOWN";							break;	//0x602											
		case	BHE_MAGIC_UNMATCH:					str = "BHE_MAGIC_UNMATCH";					break;	//0x603											
		case	BHE_UNIMPLEMENTED:					str = "BHE_UNIMPLEMENTED";					break;	//0x604											
		case	BHE_INTR:							str = "BHE_INTR";							break;	//0x605											
		case	BHE_CLOSED:							str = "BHE_CLOSED";							break;	//0x606											
		case	BHE_BUFFER_OVERFLOW:				str = "BHE_BUFFER_OVERFLOW";				break;	//0x607	/* TODO: no used error, should remove*/				
		case	BHE_NOT_SUPPORTED:					str = "BHE_NOT_SUPPORTED";					break;	//0x608											
		case	BHE_WEAR_OUT_VIOLATION:				str = "BHE_WEAR_OUT_VIOLATION";				break;	//0x609											
		case	BHE_NOT_FOUND:						str = "BHE_NOT_FOUND";						break;	//0x610											
		case	BHE_INVALID_PARAMS:					str = "BHE_INVALID_PARAMS";					break;	//0x611											
		case	BHE_ACCESS_DENIED:					str = "BHE_ACCESS_DENIED";					break;	//0x612											
		case	BHE_INVALID:						str = "BHE_INVALID";						break;	//0x614											
		case	BHE_TIMEOUT:						str = "BHE_TIMEOUT";						break;	//0x615											


			//SDM specific error code section: //0x800
		case	BHE_SDM_FAILED:							str = "BHE_SDM_FAILED";								break;	//0x800
		case	BHE_SDM_NOT_FOUND:						str = "BHE_SDM_NOT_FOUND";							break;	//0x801
		case	BHE_SDM_ALREADY_EXIST:					str = "BHE_SDM_ALREADY_EXIST";						break;	//0x803
		case	BHE_SDM_TATYPE_MISMATCH:				str = "BHE_SDM_TATYPE_MISMATCH";					break;	//0x804
		case	BHE_SDM_TA_NUMBER_LIMIT:				str = "BHE_SDM_TA_NUMBER_LIMIT";					break;	//0x805
		case	BHE_SDM_SIGNATURE_VERIFY_FAIL:			str = "BHE_SDM_SIGNATURE_VERIFY_FAIL";				break;	//0x806
		case	BHE_SDM_PERMGROUP_CHECK_FAIL:			str = "BHE_SDM_PERMGROUP_CHECK_FAIL";				break;	//0x807
		case	BHE_SDM_INSTALL_CONDITION_FAIL:			str = "BHE_SDM_INSTALL_CONDITION_FAIL";				break;	//0x808
		case	BHE_SDM_SVN_CHECK_FAIL:					str = "BHE_SDM_SVN_CHECK_FAIL";						break;	//0x809
		case	BHE_SDM_TA_DB_NO_FREE_SLOT:				str = "BHE_SDM_TA_DB_NO_FREE_SLOT";					break;	//0x80A
		case	BHE_SDM_SD_DB_NO_FREE_SLOT:				str = "BHE_SDM_SD_DB_NO_FREE_SLOT";					break;	//0x80B
		case	BHE_SDM_SD_INTERFACE_DISABLED:			str = "BHE_SDM_SD_INTERFACE_DISABLED";				break;	//0x810
		case	BHE_SDM_SD_PUBLICKEY_HASH_VERIFY_FAIL:	str = "BHE_SDM_SD_PUBLICKEY_HASH_VERIFY_FAIL";		break;	//0x811
		case	BHE_SDM_TA_INSTALL_UNALLOWED:			str = "BHE_SDM_TA_INSTALL_UNALLOWED";				break;	//0x812
		case	BHE_SDM_SVL_DB_NO_FREE_SLOT:			str = "BHE_SDM_SVL_DB_NO_FREE_SLOT";				break;	//0x80C
		case	BHE_SDM_SVL_CHECK_FAIL:					str = "BHE_SDM_SVL_CHECK_FAIL";						break;	//0x80D
		case	BHE_SDM_DB_READ_FAIL:					str = "BHE_SDM_DB_READ_FAIL";						break;	//0x80E
		case	BHE_SDM_DB_WRITE_FAIL:					str = "BHE_SDM_DB_WRITE_FAIL";						break;	//0x80F
		case	BHE_SDM_SD_INSTALL_UNALLOWED:			str = "BHE_SDM_SD_INSTALL_UNALLOWED";				break;	//0x813
		case	BHE_SDM_SVL_UPDATE_UNALLOWED:			str = "BHE_SDM_SVL_UPDATE_UNALLOWED";				break;	//0x814
		case	BHE_SDM_SD_NOT_FOUND:					str = "BHE_SDM_SD_NOT_FOUND";						break;	//0x815
		case	BHE_SDM_SD_INVALID_PROPERTIES:			str = "BHE_SDM_SD_INVALID_PROPERTIES";				break;	//0x816
			// ......
			//////////////////////////////////////////////////


			//Launcher specific error code section: //0x900
		case	BHE_LAUNCHER_INIT_FAILED:			str = "BHE_LAUNCHER_INIT_FAILED";			break;	//0x901											
		case	BHE_SD_NOT_INSTALLED:				str = "BHE_SD_NOT_INSTALLED";				break;	//0x902											
		case	BHE_NTA_NOT_INSTALLED:				str = "BHE_NTA_NOT_INSTALLED";				break;	//0x903											
		case	BHE_PROCESS_SPAWN_FAILED:			str = "BHE_PROCESS_SPAWN_FAILED";			break;	//0x904											
		case	BHE_PROCESS_KILL_FAILED:			str = "BHE_PROCESS_KILL_FAILED";			break;	//0x905											
		case	BHE_PROCESS_ALREADY_RUNNING:		str = "BHE_PROCESS_ALREADY_RUNNING";		break;	//0x906											
		case	BHE_PROCESS_IN_TERMINATING:			str = "BHE_PROCESS_IN_TERMINATING";			break;	//0x907										
		case	BHE_PROCESS_NOT_EXIST:				str = "BHE_PROCESS_NOT_EXIST";				break;	//0x908											
		case	BHE_PLATFORM_API_ERR:				str = "BHE_PLATFORM_API_ERR";				break;	//0x909											
		case	BHE_PROCESS_NUM_EXCEED:				str = "BHE_PROCESS_NUM_EXCEED";				break;	//0x90A

			//////////////////////////////////////////////////
		}
		return str;
	}

}
