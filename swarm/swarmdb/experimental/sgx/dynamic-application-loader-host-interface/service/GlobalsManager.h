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

#ifndef __GLOBALS_MANAGER_H
#define __GLOBALS_MANAGER_H

// The H-Files
#include <string>
#include "Locker.h"
#include "Singleton.h"
#include "ReadWriteLock.h"
#include "jhi_service.h"
#include "teetransport.h"

#ifdef __linux__
#include <pthread.h>
#endif//__linux__

namespace intel_dal
{
	/**
		This class holds JHI global variables and make them thread safe
	**/

	enum jhi_states
	{
		JHI_INITIALIZED = 0,
		JHI_STOPPING,
		JHI_STOPPED
	};

	class GlobalsManager : public Singleton<GlobalsManager>
	{
		friend class Singleton<GlobalsManager>;
	private:

		Locker	locker;				// used for thread safety within the class
		
		jhi_states jhi_state;		// marks the JHI status

		FILESTRING service_folder;		// contains a full path to the location of the service files
		FILESTRING applets_folder;		// contains a full path to the location of the applet repository
#ifndef _WIN32
		FILESTRING plugin_folder;		// contains a full path to the location of the plugin library
		FILESTRING spooler_folder;		// contains a full path to the location of the Spooler Applet directory
#endif
		bool plugin_registered;			// used to determine if vm plugin is loaded and registered.
		bool transport_registered;		// used to determine if the transport layer is loaded and registered.
		VM_Plugin_interface* plugin_table;	// contains the functions for communication with the VM

		TEE_TRANSPORT_TYPE transportType;	// the transport type to communicate with DAL (HECI / sockets)
		JHI_VM_TYPE vmType;				// the discovered DAL VM type in the FW
		VERSION fwVersion;

#ifdef _WIN32
		HANDLE resetCompleteEvent;
#else
		bool reset_complete;
		pthread_mutex_t reset_complete_mutex;
		pthread_cond_t reset_complete_cond;
#endif //WIN32

		// Default Constructor
		GlobalsManager(void);

		// Destructor 
		~GlobalsManager(void);

	public:

		ReadWriteLock initLock;

		/*
			get the full path to the JHI service directory
			Paramters:
				jhi_service_folder		[Out]			the path string
				
			Return:
				true - the path assigned
				false - failed to get the path. (jhi hasn't been initialized yet)
		*/
		void getServiceFolder(FILESTRING& jhi_service_folder);

		/*
			set the full path to the JHI service directory
			Paramters:
				jhi_service_folder		[In]			the path string
		*/
		bool setServiceFolder(const FILESTRING& jhi_service_folder);
		
		/*
			get the full path to the applets repository directory
			Paramters:
				jhi_applets_folder		[Out]			the path string
				
			Return:
				true - the path assigned
				false - failed to get the path. (jhi hasn't been initialized yet)
		*/
		void getAppletsFolder(FILESTRING& jhi_applets_folder);

		/*
			set the full path to the applets repository directory
			Paramters:
				jhi_applets_folder		[In]			the path string	
		*/
		bool setAppletsFolder(const FILESTRING& jhi_applets_folder);
#ifndef _WIN32
		/*
			get the full path to the JHI plugin directory
			Paramters:
				jhi_plugin_folder		[Out]			the path string

			Return:
				true - the path assigned
				false - failed to get the path. (jhi hasn't been initialized yet)
		*/
		void getPluginFolder(FILESTRING& jhi_plugin_folder);

		/*
			set the full path to the JHI plugin directory
			Paramters:
				jhi_plugin_folder		[In]			the path string
		*/
		bool setPluginFolder(const FILESTRING& jhi_plugin_folder);

		/*
			get the full path to the Spooler applet directory
			Paramters:
				jhi_spooler_folder		[Out]			the path string

			Return:
				true - the path assigned
				false - failed to get the path. (jhi hasn't been initialized yet)
		*/
		void getSpoolerFolder(FILESTRING& jhi_spooler_folder);

		/*
			set the full path to the Spooler applet directory
			Paramters:
				jhi_spooler_folder		[In]			the path string
		*/
		bool setSpoolerFolder(const FILESTRING& jhi_spooler_folder);
#endif //!_WIN32

		/*
			get the plugin table that contains the API to communicate with the FW
			Paramters:
				plugin_table		[Out]			the plugin table
				
			Return:
				true - the plugin table succesfuly assigned
				false - failed to get the plugin table. (jhi hasn't been initialized yet)
		*/
		bool getPluginTable(VM_Plugin_interface** plugin_table);

		/*
			Returns whether the VM plugin was registered or not.
		*/
		bool isPluginRegistered();

		/*
			sets the plugin table that contains the API to communicate with the VM
		*/
		JHI_RET PluginRegister();

		/*
			Removes the plugin table that contains the API to communicate with the VM
		*/
		void PluginUnregister();

		/*
			Returns whether the transport was registered or not.
		*/
		bool isTransportRegistered();
		/*
			set the jhi state
			Paramters:
				newState		[IN]			the new jhi state
		*/		
		void setJhiState(jhi_states newState);

		/*
			returns the jhi state
		*/	
		jhi_states getJhiState();

		/*
			get the transport type to communicate with DAL (HECI / sockets)
			Return:
				the transport type (enum)
		*/
		TEE_TRANSPORT_TYPE getTransportType();

		/*
			set the transport type to communicate with DAL (HECI / sockets)
			Paramters:
				transportType		[In]			the transport type (enum)
			Return:
				true - the path assigned
				false - failed to get the path. (jhi hasn't been initialized yet)
		*/
		void setTransportType(TEE_TRANSPORT_TYPE transportType);

		JHI_VM_TYPE getVmType();
		bool setVmType(JHI_VM_TYPE vm_type);

		VERSION getFwVersion();
		void setFwVersion(VERSION fw_version);
		bool getFwVersionString(char *fw_version);

		JHI_PLATFROM_ID getPlatformId();

		/*
			Notify that JHI reset has completed by sending an event
		*/
		void sendResetCompleteEvent();

		/*
			Wait for reset complete notification
		*/
		void waitForResetComplete();
	};
}

#endif 

