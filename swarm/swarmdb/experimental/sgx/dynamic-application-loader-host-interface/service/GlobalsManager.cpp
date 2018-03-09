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

#include "GlobalsManager.h"
#include "dbg.h"
#include "string_s.h"

namespace intel_dal
{

	GlobalsManager::GlobalsManager()
	{
		jhi_state = JHI_STOPPED;
		plugin_registered = false;
		transport_registered = false;
		plugin_table = NULL;
		transportType = TEE_TRANSPORT_TYPE_INVALID;
		vmType = JHI_VM_TYPE_INVALID;
		memset(&fwVersion, 0, sizeof(fwVersion));
#ifdef _WIN32
		resetCompleteEvent = CreateEvent(NULL,FALSE,FALSE,NULL);

		if (resetCompleteEvent == NULL)
		{
			TRACE0("ERROR: failed to create reset complete event!");
		}
#else
		reset_complete = false;
		pthread_mutex_init(&reset_complete_mutex, NULL);
		pthread_cond_init(&reset_complete_cond, NULL);
#endif //WIN32
	}

	GlobalsManager::~GlobalsManager()
	{
#ifdef _WIN32
		CloseHandle(resetCompleteEvent);
#else
		pthread_mutex_destroy(&reset_complete_mutex);
		pthread_cond_destroy(&reset_complete_cond);
#endif //WIN32
	}

	void GlobalsManager::getServiceFolder(FILESTRING & jhi_service_folder)
	{
		locker.Lock();
		jhi_service_folder = service_folder;
		locker.UnLock();
	}

	bool GlobalsManager::setServiceFolder(const FILESTRING& jhi_service_folder)
	{
		if (jhi_service_folder.empty()) 
			return false;

		locker.Lock();
		service_folder = jhi_service_folder;
		locker.UnLock();
		TRACE1("GlobalsManager - setServiceFolder = %s", service_folder.c_str());

		return true;
	}

	void GlobalsManager::getAppletsFolder(FILESTRING & jhi_applets_folder)
	{
		locker.Lock();
		jhi_applets_folder = applets_folder;
		locker.UnLock();
	}

	bool GlobalsManager::setAppletsFolder(const FILESTRING& jhi_applets_folder)
	{
		if (jhi_applets_folder.empty()) 
			return false;
		locker.Lock();
		applets_folder = jhi_applets_folder;
		locker.UnLock();
        TRACE1("GlobalsManager - setAppletsFolder = %s", applets_folder.c_str());
		return true;
	}

#ifndef _WIN32
	void GlobalsManager::getPluginFolder(FILESTRING & jhi_plugin_folder)
	{
		locker.Lock();
		jhi_plugin_folder = plugin_folder;
		locker.UnLock();
	}
	bool GlobalsManager::setPluginFolder(const FILESTRING& jhi_plugin_folder)
	{
		if (jhi_plugin_folder.empty())
			return false;
		locker.Lock();
		plugin_folder = jhi_plugin_folder;
		locker.UnLock();
		return true;
	}
	void GlobalsManager::getSpoolerFolder(FILESTRING & jhi_spooler_folder)
	{
		locker.Lock();
		jhi_spooler_folder = spooler_folder;
		locker.UnLock();
	}
	bool GlobalsManager::setSpoolerFolder(const FILESTRING& jhi_spooler_folder)
	{
		if (jhi_spooler_folder.empty())
			return false;
		locker.Lock();
		spooler_folder = jhi_spooler_folder;
		locker.UnLock();
		return true;
	}
#endif //!_WIN32

	bool GlobalsManager::getPluginTable(VM_Plugin_interface** plugin_table)
	{
		bool status = false;

		if (plugin_table == NULL)
			return false;

		locker.Lock();

		if (plugin_registered)
			status = true;

		// copy even if not in initialized state, because the EventManager init will call this method before jhi is initialized.
		*plugin_table = this->plugin_table;

		locker.UnLock();

		return status;
	}

	bool GlobalsManager::isPluginRegistered()
	{
		return this->plugin_registered;
	}

	JHI_RET GlobalsManager::PluginRegister()
	{
		locker.Lock();
		JHI_RET ulRetCode = JHI_INTERNAL_ERROR;
		ulRetCode = JhiPlugin_Register(&this->plugin_table);
		if (ulRetCode == JHI_SUCCESS)
		{
			this->plugin_registered = true;
		}
		locker.UnLock();
		return ulRetCode;
	}

	void GlobalsManager::PluginUnregister()
	{
		locker.Lock();
		if(this->plugin_registered)
		{
			this->plugin_registered = false;
			JhiPlugin_Unregister(&this->plugin_table);
		}
		locker.UnLock();
	}

	bool GlobalsManager::isTransportRegistered()
	{
		return this->transport_registered;
	}
	


	void GlobalsManager::setJhiState(jhi_states newState)
	{
		locker.Lock();

		this->jhi_state = newState;

		locker.UnLock();
	}

	jhi_states GlobalsManager::getJhiState()
	{
		return jhi_state;
	}



	void GlobalsManager::setTransportType(TEE_TRANSPORT_TYPE transportType)
	{
		locker.Lock();
		GlobalsManager::transportType = transportType;
		locker.UnLock();
		TRACE1("GlobalsManager - setTransportType = %d.", transportType);
	}

	TEE_TRANSPORT_TYPE GlobalsManager::getTransportType()
	{
		return GlobalsManager::transportType;
	}

	// VM type getter and setter
	JHI_VM_TYPE GlobalsManager::getVmType() { return vmType; }

	bool GlobalsManager::setVmType(JHI_VM_TYPE newVmType)
	{
		if(newVmType > JHI_VM_TYPE_INVALID && newVmType < JHI_VM_TYPE_MAX)
		{
			vmType = newVmType;
			return true;
		}
		else
			return false;
	}

	// FW version getter and setter
	VERSION GlobalsManager::getFwVersion() { return fwVersion; }

	void GlobalsManager::setFwVersion(VERSION fw_version)
	{
		fwVersion = fw_version;
	}

	bool GlobalsManager::getFwVersionString(char *fw_version)
	{
		return sprintf_s(fw_version, FW_VERSION_STRING_MAX_LENGTH, "%d.%d.%d.%d", fwVersion.Major, fwVersion.Minor, fwVersion.Hotfix, fwVersion.Build) == 4;
	}

	// Platform ID
	JHI_PLATFROM_ID GlobalsManager::getPlatformId()
	{
		uint16_t ver_major = getFwVersion().Major;

		if(ver_major == 0)
			return INVALID_PLATFORM_ID;
		else if(ver_major == 1 || ver_major == 2)
			return SEC;
		else if(ver_major >= 7 && ver_major <= 10)
			return ME;
		else
			return CSE;
	}

	// Reset event
	void GlobalsManager::sendResetCompleteEvent()
	{
		TRACE0("Sending reset complete event...\n");
#ifdef _WIN32
		SetEvent(resetCompleteEvent);
#else
		pthread_mutex_lock(&reset_complete_mutex);
		reset_complete = true;
		pthread_cond_signal(&reset_complete_cond);
		pthread_mutex_unlock(&reset_complete_mutex);
#endif // _WIN32
	}
	void GlobalsManager::waitForResetComplete()
	{
#ifdef _WIN32
		WaitForSingleObject(resetCompleteEvent,INFINITE);
#else
		pthread_mutex_lock(&reset_complete_mutex);
		while (!reset_complete)
		{
			pthread_cond_wait(&reset_complete_cond, &reset_complete_mutex);
		}
		reset_complete = false;
		pthread_mutex_unlock(&reset_complete_mutex);
#endif// _WIN32
		TRACE0("received reset complete event!\n");
	}
}