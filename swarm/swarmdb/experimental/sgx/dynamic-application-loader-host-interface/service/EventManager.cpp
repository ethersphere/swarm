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

// The H-Files
#include "EventManager.h"
#include "SessionsManager.h"
#include "AppletsManager.h"
#include "EventLog.h"
#ifdef _WIN32
#include <io.h>
#endif // _WIN32
#include "string_s.h"

namespace intel_dal
{

#ifdef _WIN32
	DWORD JomEventListenerThread (LPVOID threadParam)
#else
	void* JomEventListenerThread (void* threadParam)
#endif // _WIN32
	{
		JHI_RET status;
		JHI_SESSION_ID target_session;
		JHI_EVENT_DATA* event_data = NULL;
		JhiEvent* eventHandle = NULL;
		bool doGlobalReset = false;
		VM_Plugin_interface* plugin = NULL;

		GlobalsManager::Instance().getPluginTable(&plugin); // ignore the return value

		if (plugin == NULL)
		{
#ifdef _WIN32
			return -1;
#else
			return NULL;
#endif
		}

		while (1)
		{	
			status = plugin->JHI_Plugin_WaitForSpoolerEvent(EventManager::Instance().spooler_handle, &event_data, &target_session);

			if (status == JHI_SUCCESS)
			{
				TRACE0("Event recieved from spooler");
				if (SessionsManager::Instance().getEventHandle(target_session, &eventHandle))
				{
					if (event_data != NULL)
					{
						//store the event data in the session event queue
						if (!SessionsManager::Instance().enqueueEventData(target_session, event_data))
						{
							TRACE0("internal error: failed to add event data to the session queue");

							if (event_data->data != NULL)
							{
								JHI_DEALLOC(event_data->data);
								event_data->data = NULL;
							}

							JHI_DEALLOC(event_data);
							event_data = NULL;
						}
					}

					// raise event to the application
					TRACE1("sending event to app, event handle: %d\n", eventHandle);
					if (!eventHandle->set())
					{
						TRACE0("internal error: failed to send event");
					}
				}
			}
			else if ( ((status == JHI_APPLET_FATAL) || (status == JHI_APPLET_BAD_STATE)) && (GlobalsManager::Instance().getJhiState() == JHI_INITIALIZED) )
			{
				TRACE0("Spooler applet crashed - trying to load it again\n");
				EventManager::Instance().spooler_handle = NULL;
				EventManager::Instance().Deinit();

#ifdef _WIN32
				// COM init - needed for working with the spooler dalp file
				CoInitialize(NULL);
#endif// _WIN32

				if (EventManager::Instance().Initialize() != JHI_SUCCESS)
				{
					doGlobalReset = true;
				}

#ifdef _WIN32
				// COM deinit
				CoUninitialize();
#endif// _WIN32

				break;
			}
			else
			{
				TRACE0("No connection to FW or a Spooler error");
				TRACE0("Performing global service reset...");
				doGlobalReset = true;
				break;
			}

		}

		if (doGlobalReset)
		{
			TRACE0("Calling JhiReset...");
			JhiReset();
		}

		return 0;
	}

	EventManager::EventManager()
	{
		initialized = false;
		spooler_handle = NULL;
	}

	EventManager::~EventManager()
	{
	}

	bool EventManager::GetSpoolerFullFilename(OUT FILESTRING& outFileName, OUT bool &isAcp)
	{
#ifdef _WIN32
		FILESTRING ServiceDir;
		GlobalsManager::Instance().getServiceFolder(ServiceDir);
		FILESTRING spoolerFileDalp = ServiceDir + FILESTRING(SPOOLER_APPLET_FILENAME) + ConvertStringToWString(dalpFileExt);
		FILESTRING spoolerFileAcp = ServiceDir + FILESTRING(SPOOLER_APPLET_FILENAME) + ConvertStringToWString(acpFileExt);
#else
		FILESTRING SpoolerDir;
		GlobalsManager::Instance().getSpoolerFolder(SpoolerDir);
		FILESTRING spoolerFileDalp = SpoolerDir + FILESTRING(SPOOLER_APPLET_FILENAME) + ConvertStringToWString(dalpFileExt);
		FILESTRING spoolerFileAcp = SpoolerDir + FILESTRING(SPOOLER_APPLET_FILENAME) + ConvertStringToWString(acpFileExt);
#endif
		//verify the spooler applet exist and has read access
		if (_waccess_s(spoolerFileDalp.c_str(),4) == 0)
		{
			outFileName = spoolerFileDalp;
			isAcp = false;
			return true;
		}
		else if (_waccess_s(spoolerFileAcp.c_str(),4) == 0)
		{
			outFileName = spoolerFileAcp;
			isAcp = true;
			return true;
		}
		else
		{
			LOG1("EventManager error: Spooler Applet file wasn't found, or no read access at: %s\n",ConvertWStringToString(spoolerFileDalp).c_str());
			WriteToEventLog(JHI_EVENT_LOG_ERROR, MSG_SPOOLER_NOT_FOUND);
			return false;
		}
	}

	JHI_RET EventManager::Initialize()
	{
		JHI_RET init_status = JHI_INVALID_SPOOLER;
		JHI_VM_TYPE vmType;
		
		if (initialized) return true;

		vmType = GlobalsManager::Instance().getVmType();

		FILESTRING spoolerFile;
		bool isAcp = false;

		bool weHaveASession = false;

		do
		{
			if (!GetSpoolerFullFilename(spoolerFile, isAcp))
			{
				init_status = JHI_SPOOLER_NOT_FOUND;
				break;
			}

			// Over Beihai V2, sometimes no need to install because installation is persistent
			if (vmType == JHI_VM_TYPE_BEIHAI_V2)
			{
				init_status = CreateSpoolerSession(spoolerFile, isAcp);
				if (init_status == JHI_SUCCESS)
					weHaveASession = true;
			}

			// Otherwise install and create session
			if (!weHaveASession)
			{
				init_status = InstallSpooler(spoolerFile, isAcp);
				if (init_status != JHI_SUCCESS)	break;

				init_status = CreateSpoolerSession(spoolerFile, isAcp);
				if (init_status != JHI_SUCCESS)	break;
			}

			init_status = CreateListenerThread();
		} while (0);

		initialized = (init_status == JHI_SUCCESS);
		return init_status;
	}

	void EventManager::Deinit()
	{
		if (initialized)
		{
			closeSpoolerSession();
		}
		else
		{
			TRACE0("error: the event manager is not initialized\n");
		}

		initialized = false;
	}

	void EventManager::closeSpoolerSession()
	{
		JHI_RET ulRetCode = JHI_INTERNAL_ERROR;

		if (spooler_handle != NULL)
		{
			VM_Plugin_interface* plugin = NULL;
			if ( (!GlobalsManager::Instance().getPluginTable(&plugin)) || (plugin == NULL) )
			{
				// we probably had a reset
				ulRetCode = JHI_NO_CONNECTION_TO_FIRMWARE;				
			}
			else
			{
				TRACE0("Force closing the spooler applet session...");
				ulRetCode = plugin->JHI_Plugin_ForceCloseSession(&spooler_handle);
			}

			if (ulRetCode != JHI_SUCCESS)
			{
				TRACE1("failed to close the spooler session. err: 0x%x\n",ulRetCode);
			}
		}

	}

	JHI_RET EventManager::SetSessionEventHandler(JHI_SESSION_ID SessionID, char* eventHandleName)
	{
		JhiEvent* eventHandle = NULL;
		JHI_SESSION_INFO info;
		JHI_SESSION_FLAGS flags;

		SessionsManager::Instance().getSessionInfo(SessionID,&info);

		if (info.state == JHI_SESSION_STATE_NOT_EXISTS)
		{
			return JHI_INVALID_SESSION_HANDLE;
		}

		flags.value = info.flags;

		if (flags.bits.sharedSession)
		{
			return JHI_EVENTS_NOT_SUPPORTED;
		}

		if (eventHandleName == NULL)
		{
			return JHI_INTERNAL_ERROR;
		}

		eventHandle = JHI_ALLOC_T(JhiEvent);
		if (eventHandle == NULL)
		{
			return JHI_INTERNAL_ERROR;
		}
#ifdef _WIN32
		if (strnlen_s(eventHandleName,JHI_EVENT_HANDLE_SIZE) != 0) // If length == 0, event will be unregistered.
#else
		if (strnlen(eventHandleName,JHI_EVENT_HANDLE_SIZE) != 0) // If length == 0, event will be unregistered.
#endif
		{
			if(!eventHandle->open(eventHandleName))
			{
				TRACE1("OpenEvent failure. Tried to open %s.", eventHandleName);
				JHI_DEALLOC_T(eventHandle);
				return JHI_INTERNAL_ERROR;
			}
		}

		if (!SessionsManager::Instance().setEventHandle(SessionID,eventHandle))
		{
			eventHandle->close();
			JHI_DEALLOC_T(eventHandle);
			return JHI_INTERNAL_ERROR;
		}

		return JHI_SUCCESS;
	}

	JHI_RET EventManager::InstallSpooler(const FILESTRING &spoolerFile, bool isAcp)
	{
		JHI_RET status = JHI_INVALID_SPOOLER;

		do
		{
			TRACE0("Installing the Spooler...");
			status = jhis_install(SPOOLER_APPLET_UUID, spoolerFile.c_str(), false, isAcp);

			if ((status != JHI_SUCCESS) && (status != JHI_FILE_IDENTICAL))
			{
				LOG0("failed downloading Spooler Applet to DAL FW\n");
				break;
			}

			TRACE0("Spooler is installed.");
		} while (0);
		
		return status;
	}

	JHI_RET EventManager::CreateSpoolerSession(const FILESTRING &spoolerFile, bool isAcp)
	{ 
		JHI_RET status = JHI_INVALID_SPOOLER;
		AppletsManager&  Applets = AppletsManager::Instance();
		VM_Plugin_interface* plugin = NULL;
		JHI_SESSION_ID spoolerID;

		JHI_VM_TYPE vmType = GlobalsManager::Instance().getVmType();

		list< vector<uint8_t> > spoolerBlobs;

		do
		{
			TRACE0("Creating the Spooler session...");
			GlobalsManager::Instance().getPluginTable(&plugin); // ignore the false return since jhi is not initialized at this point.

			if (plugin == NULL)
			{
				status = JHI_INTERNAL_ERROR;
				break;
			}

			if (!SessionsManager::Instance().generateNewSessionId(&spoolerID))
			{
				status = JHI_INTERNAL_ERROR;
				break;
			}

			DATA_BUFFER initBuffer; // no input to spooler on creation
			initBuffer.buffer = NULL;
			initBuffer.length = 0;

			if (vmType != JHI_VM_TYPE_BEIHAI_V2)
			{
				status = plugin->JHI_Plugin_CreateSession(SPOOLER_APPLET_UUID, &spooler_handle, NULL, 0, spoolerID, &initBuffer);
			}
			else // Create the session for CSE.
			{
				status = Applets.getAppletBlobs(spoolerFile, spoolerBlobs, isAcp); // in CSE we need the blobs for the create session API
				if (status != JHI_SUCCESS)
				{
					TRACE0("Failed getting applet blobs from dalp file\n");
					break;
				}

				for (list<vector<uint8_t> >::iterator it = spoolerBlobs.begin(); it != spoolerBlobs.end(); ++it)
				{
					status = plugin->JHI_Plugin_CreateSession(SPOOLER_APPLET_UUID, &spooler_handle, &(*it)[0], (unsigned int)(*it).size(), spoolerID, &initBuffer);

					if (status == JHI_SUCCESS)
					{
						break;
					}
				}

				if (status != JHI_SUCCESS) // we didn't find a working blob
				{
					TRACE0("No suitable blobs found for Spooler session creation");
					break;
				}
			}
		} while (0);

		if (status == JHI_SUCCESS)
		{
			TRACE0("Spooler session created successfully");
		}
		else
		{
			LOG0("Failed to create the Spooler Session");
		}

		return status;
	}

	// Create a listener thread that will listen for events form DAL FW
	JHI_RET EventManager::CreateListenerThread()
	{
		JHI_RET status = JHI_INTERNAL_ERROR;

		VM_Plugin_interface* plugin = NULL;
		GlobalsManager::Instance().getPluginTable(&plugin); // ignore the false return since jhi is not initialized at this point.

		TRACE0("Creating the event listener thread...");

#ifdef _WIN32
		listenerThreadHandle = CreateThread(NULL, 0, (LPTHREAD_START_ROUTINE)&JomEventListenerThread, NULL, 0, NULL);
		if (listenerThreadHandle == NULL)
#else
		if(pthread_create(&listenerThreadHandle, NULL, JomEventListenerThread, NULL))
#endif
		{
			TRACE0("Failed creating event handle thread\n");
			status = JHI_INTERNAL_ERROR;
			plugin->JHI_Plugin_CloseSession(&spooler_handle);
		}
		else
		{
#ifndef _WIN32
			pthread_detach(listenerThreadHandle);
#endif //!_WIN32
			TRACE0("Event listener thread created successfully");
			status = JHI_SUCCESS;
		}

		return status;
	}
}
