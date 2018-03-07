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
**    @file jhi.cpp
**
**    @brief  Defines exported interfaces for JHI.DLL
**
**    @author Elad Dabool
**
********************************************************************************
*/

#ifdef _WIN32
#include <windows.h>
#include <process.h>
#endif

#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <list>

#include "jhi.h"
#include "jhi_i.h"

#include "Locker.h"


#ifdef _WIN32

#ifdef SCHANNEL_OVER_SOCKET // emulation mode
#define SVCNAME TEXT("jhi_service_emulation")
#else
#define SVCNAME TEXT("jhi_service")
#endif

#define JHI_SERVICE_NOTIFICATION_TIMEOUT 3000 

// dynamic definition of NotifyServiceStatusChange win API for use with StartService on delayed start
typedef DWORD (WINAPI *NotifyServiceStatusChangeFunc)(_In_ SC_HANDLE hService,_In_  DWORD dwNotifyMask,_In_  PSERVICE_NOTIFY pNotifyBuffe);
 
#endif

#ifdef _WIN32

static VOID CALLBACK onServiceChange(IN PVOID pParameter)
{
     SERVICE_NOTIFY* ServiceNotify;
     HANDLE EventHandle;
 
     ServiceNotify = (SERVICE_NOTIFY*) pParameter;
     EventHandle = *(HANDLE*)ServiceNotify->pContext;
     SetEvent(EventHandle);
}

// This function query the JHI service status and in case the service is not started it will start
// it by using StartService function. this makes sure that applications who runs at boot time will be able
// to use the service while it is configured as Delayed Start.
// Since Delayed Start is not supported in Win XP, we skip this operation on WinXP OS.
void startJHIService()
{
	//QueryServiceStatus

	SC_HANDLE schService = NULL;
	SC_HANDLE schSCManager = NULL;
    HANDLE EventHandle = NULL;
	SERVICE_NOTIFY ServiceNotify = {0};
	NotifyServiceStatusChangeFunc pNotifyServiceStatusChange = NULL;
	SERVICE_STATUS status = {0};
	DWORD ret = 0;

	if (!isVistaOrLater())
		goto cleanup;

	// Get a handle to the SCM database. 
    schSCManager = OpenSCManager( 
        NULL,                    // local computer
        NULL,                    // ServicesActive database 
        SC_MANAGER_CONNECT);  
 
	if (NULL == schSCManager) 
	{
		int errorCode = GetLastError();
			
		if (errorCode == ERROR_ACCESS_DENIED)
		{
			TRACE0("ACCESS DENIED: administrative privileges required.\n");
		}
		else
		{	
			TRACE1("OpenSCManager failed (%d)\n",errorCode );
		}
		goto cleanup;
	}


	// get a handle to the JHI Service
	schService = OpenService(schSCManager,SVCNAME, SERVICE_START | SERVICE_QUERY_STATUS);

	if (schService == NULL) 
    {
		int errorCode = GetLastError();
		if (errorCode == ERROR_ACCESS_DENIED)
		{
			TRACE0("ACCESS DENIED: administrative privileges required.\n");
		} 
		else if (errorCode == ERROR_SERVICE_DOES_NOT_EXIST)
		{
			TRACE0("Error: the service does not exist.\n");
		}
		else
		{
			TRACE1("OpenService failed: (%d)\n",GetLastError()); 
        }
		
		goto cleanup;
    }

	if (QueryServiceStatus(schService,&status) == FALSE)
	{
		TRACE0("Error: failed to query The service status!\n");
		goto cleanup;
	}

	TRACE1("Current Service State: %d\n",status.dwCurrentState);

	if (status.dwCurrentState == SERVICE_RUNNING)
	{
		TRACE0("Service is already running, no need to start it.\n");
		goto cleanup;
	}

	EventHandle = CreateEvent(NULL,FALSE,FALSE,NULL);

	if (EventHandle == NULL)
	{
		TRACE1("failed to create an event handle err: %d\n", GetLastError());
		goto cleanup;
	}
	
	ServiceNotify.dwVersion = SERVICE_NOTIFY_STATUS_CHANGE;
    ServiceNotify.pfnNotifyCallback = onServiceChange;
    ServiceNotify.pContext = &EventHandle;

	pNotifyServiceStatusChange = (NotifyServiceStatusChangeFunc) GetProcAddress(GetModuleHandle(TEXT("Advapi32.dll")),"NotifyServiceStatusChange");

	if (pNotifyServiceStatusChange == NULL)
	{
		TRACE0("Error: failed to retrieve pointer to NotifyServiceStatusChange\n");
		goto cleanup;
	}

	// note: in case the current thread is impersonating, it may not have the right privileges for this operation
	ret = pNotifyServiceStatusChange(schService,SERVICE_NOTIFY_RUNNING,&ServiceNotify);

	if (ret != ERROR_SUCCESS)
	{
		TRACE1("failed to register for service status event, reason: %d\n",ret);
		goto cleanup;
	}

	if (StartService(schService,NULL,NULL) == FALSE)
	{
		ret = GetLastError();
		TRACE1("Error: StartService failed, error: %d\n",ret);
		
		if (ret != ERROR_SERVICE_ALREADY_RUNNING)
		{
			TRACE0("stopping startJHIService flow\n");
			goto cleanup;
		}
	}

	while (true)
	{ 
		TRACE0("Waiting for service status event...\n");
		ret = WaitForSingleObjectEx(EventHandle, JHI_SERVICE_NOTIFICATION_TIMEOUT, TRUE);

		switch (ret)
		{
			case WAIT_IO_COMPLETION: 
				TRACE0("Awaken by RPC CALL, return to wait state\n"); 
				break;

			case WAIT_OBJECT_0: 
				TRACE0("JHI Service is in Runnning state\n");
				goto cleanup;

			case WAIT_TIMEOUT: 
				TRACE0("WaitForSingleObjectEx has timed out!\n");
				goto cleanup;

			default: 
				TRACE1("Unexpected WaitForSingleObjectEx error: %d\n",ret);
				goto cleanup;
		}
	}

cleanup:

	if (EventHandle != INVALID_HANDLE_VALUE)
		CloseHandle(EventHandle);
	
	if (schSCManager != NULL)
		CloseServiceHandle(schSCManager);

	if (schService != NULL)
		CloseServiceHandle(schService);
}

#endif