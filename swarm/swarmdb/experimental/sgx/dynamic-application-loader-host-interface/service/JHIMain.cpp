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

#include "JHIMain.h"

#include <iostream>
#include "GlobalsManager.h"
#include "EventManager.h"
#include "jhi_service.h"
#include "dbg.h"
#include "EventLog.h"
#include "CommandsServerFactory.h"

namespace intel_dal
{
	static ICommandsServer* commandsServer;

	bool jhi_init()
	{
		LOG0("--> jhi start");
		GlobalsManager::Instance().setJhiState(JHI_STOPPED); // also calls the GlobalsManager constructor to avoid problems later
		
		commandsServer = CommandsServerFactory::createInstance();

		TRACE0("opening command server\n");
		if (!commandsServer->open())
		{
			LOG0("Error: command server has failed to open a connection\n");
			return false;
		}

		LOG0("<-- jhi start");
		return true;
	}

#ifdef _WIN32
	DWORD jhiMainThread(LPVOID threadParam)
	{
		jhi_main();
		return 0;
	}
#endif

	void jhi_start()
	{
#ifdef _WIN32
		LOG0("JHI service starting");
		jhi_main_thread_handle = CreateThread(NULL, 0,(LPTHREAD_START_ROUTINE)&jhiMainThread, NULL,0,NULL);
#endif

		WriteToEventLog(JHI_EVENT_LOG_INFORMATION, MSG_SERVICE_START);
	}


	void jhi_stop()
	{
		TRACE0("***** JHI STOP SERVICE *****\n");

		// first, stop accepting requests
		TRACE0("Closing command server\n");
		commandsServer->close();

#ifdef _WIN32
		CloseHandle(jhi_main_thread_handle);
		jhi_main_thread_handle = NULL;
#endif

		// if jhi is initialized, reset it.
		if (GlobalsManager::Instance().getJhiState() == JHI_INITIALIZED)
		{
			TRACE0("JHI is initialized. Resetting...");
			GlobalsManager::Instance().setJhiState(JHI_STOPPING);
			JhiReset();
		}

		LOG0("jhi stopping");
		WriteToEventLog(JHI_EVENT_LOG_INFORMATION, MSG_SERVICE_STOP);
	}


	void jhi_invoke_reset()
	{
		GlobalsManager::Instance().initLock.aquireReaderLock();

		// invoke reset only when JHI is initialized
		if (GlobalsManager::Instance().getJhiState() == JHI_INITIALIZED)
		{
			GlobalsManager::Instance().setJhiState(JHI_STOPPING);

			TRACE0("invoking JHI reset\n");

			// calling EventManager Deinit() will trigger a JHI reset by the spooler thread.
			EventManager::Instance().Deinit();
		}

		GlobalsManager::Instance().initLock.releaseReaderLock();

		// we are waiting for the reset being done by the spooler thread to be completed
		// before returning response to the OS event
		GlobalsManager::Instance().waitForResetComplete();
	}

	int jhi_main()
	{
#ifndef ANDROID
		try
		{
#endif
			commandsServer->waitForRequests(); // return when stopped listening.

			// we are waiting for the reset being done before exiting the main thread
			if(GlobalsManager::Instance().getJhiState() != JHI_STOPPED)
				GlobalsManager::Instance().waitForResetComplete();
#ifndef ANDROID
		}
		catch (std::exception ex)
		{
			TRACE0("Exception raised in JHI service:");
			TRACE1("%s\n",ex.what());
		}
#endif
		// returns only when stopped
		return 0;
	}
}