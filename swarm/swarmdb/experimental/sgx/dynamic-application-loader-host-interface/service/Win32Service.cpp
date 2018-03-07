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

#include "Win32Service.h"
#include "AccCtrl.h"
#include "Aclapi.h"
#include "dbg.h"
#include "misc.h"
#include <iostream>
using namespace intel_dal;
//
// Purpose: 
//   Entry point for the process
//
// Parameters:
//   None
// 
// Return value:
//   None
//
int __cdecl _tmain(int argc, TCHAR *argv[]) 
{ 
    // If command-line parameter is "install", install the service. 
    // Otherwise, the service is probably being started by the SCM.

    if( lstrcmpi( argv[1], TEXT("install")) == 0 )
    {
        return SvcInstall();
    }
	else if( lstrcmpi( argv[1], TEXT("uninstall")) == 0 )
    {
        return SvcUninstall();
    }
	else if( lstrcmpi( argv[1], TEXT("start")) == 0 )
    {
        return SvcStart();
    }
	else if( lstrcmpi( argv[1], TEXT("stop")) == 0 )
    {
        return SvcStop();
    }
	else if (lstrcmpi(argv[1], TEXT("-v")) == 0)
	{
		std::cout << JHI_VERSION << std::endl;
		return 0;
	}
	else if( lstrcmpi( argv[1], TEXT("console")) == 0 )
    {
#ifdef _DEBUG
		printf("Running JHI in console mode.\n");
				
		if (!jhi_init())
			return JHI_SERVICE_GENERAL_ERROR;

        return jhi_main();
#else
		printf("Running JHI in console mode is not supported in release mode.\n");
		return JHI_SERVICE_GENERAL_ERROR;
#endif
    }

    // TO_DO: Add any additional services for the process to this table.
    SERVICE_TABLE_ENTRY DispatchTable[] = 
    { 
        { SVCNAME, (LPSERVICE_MAIN_FUNCTION) SvcMain }, 
        { NULL, NULL } 
    }; 
 
    // This call returns when the service has stopped. 
    // The process should simply terminate when the call returns.

    if (!StartServiceCtrlDispatcher( DispatchTable )) 
    { 
        //SvcReportEvent(TEXT("StartServiceCtrlDispatcher")); 
    }

	return EXIT_SUCCESS;
} 

bool GetEveryoneGroupName(OUT LPTSTR* groupName)
{
	PSID pSidOwner = NULL;
	LPTSTR accountName = NULL;
	LPTSTR domainName = NULL;
	DWORD accountNameLength = 1, domainNameLength = 1;
	SID_NAME_USE eUse = SidTypeUnknown;
	BOOL functionSuccess = FALSE;
	bool error = false;

	SID_IDENTIFIER_AUTHORITY auth = SECURITY_WORLD_SID_AUTHORITY;
	functionSuccess = AllocateAndInitializeSid(&auth, 1,
		SECURITY_WORLD_RID, 0, 0, 0, 0, 0, 0, 0, &pSidOwner);
	if (!functionSuccess)
	{
		TRACE0("AllocateAndInitializeSid error!");
		error = true;
		goto cleanup;
	}

	// First call to LookupAccountSid to get the buffer sizes.
	functionSuccess = LookupAccountSid(
		NULL,           // local computer
		pSidOwner,
		accountName,
		(LPDWORD)&accountNameLength,
		domainName,
		(LPDWORD)&domainNameLength,
		&eUse
		);

	// Reallocate memory for the buffers.
	accountName = (LPTSTR) JHI_ALLOC(sizeof(WCHAR) * (accountNameLength + 1));
	if (accountName == NULL) {
		TRACE0("malloc of accountName failed.");
		error = true;
		goto cleanup;
	}

	domainName = (LPTSTR) JHI_ALLOC(sizeof(WCHAR) * (domainNameLength + 1));
	if (domainName == NULL) {
		TRACE0("malloc of domainName failed.");
		error = true;
		goto cleanup;
	}


	// Second call to LookupAccountSid to get the account name.
	functionSuccess = LookupAccountSid(
		NULL,                   		// name of local or remote computer
		pSidOwner,              		// security identifier
		accountName,               		// account name buffer
		(LPDWORD)&accountNameLength,   	// size of account name buffer 
		domainName,             		// domain name
		(LPDWORD)&domainNameLength, 	// size of domain name buffer
		&eUse);                 		// SID type

	// Check GetLastError for LookupAccountSid error condition.
	if (functionSuccess == FALSE) {
		DWORD dwErrorCode = 0;
		dwErrorCode = GetLastError();
		if (dwErrorCode == ERROR_NONE_MAPPED)
			TRACE0("Account owner not found for specified SID.\n");
		else 
			TRACE0("Error in LookupAccountSid.\n");
		error = true;
		goto cleanup;
	}

	// success

	*groupName = (LPTSTR) JHI_ALLOC(sizeof(WCHAR) * (accountNameLength + 1));
	if (*groupName == NULL) {
		TRACE0("malloc of groupName failed.");
		error = true;
		goto cleanup;
	}

	HRESULT res = StringCchCopyW(*groupName, accountNameLength + 1, accountName);
	if (res != S_OK)
	{
		TRACE0("Error copying account name.\n");
		error = true;
		goto cleanup;
	}

cleanup:
	if (pSidOwner != NULL)
	{
		FreeSid(pSidOwner);
		pSidOwner = NULL;
	}
	if (accountName != NULL)
	{
		JHI_DEALLOC(accountName);
		accountName = NULL;
	}
	if (domainName != NULL)
	{
		JHI_DEALLOC(domainName);
		domainName = NULL;
	}
	return !error;
}

bool SetServiceACL(SC_HANDLE schService)
{
	DWORD dwBytesNeeded = 0;
	PSECURITY_DESCRIPTOR psd = NULL;
	PACL pacl = NULL;
	PACL pNewAcl = NULL;
	BOOL bDaclPresent   = FALSE;
    BOOL bDaclDefaulted = FALSE;
	EXPLICIT_ACCESS ea;
    SECURITY_DESCRIPTOR sd;
	DWORD dwError = 0;
	bool status = false;
	LPTSTR everyoneGroupName = NULL;

	// set service ACL according to: http://msdn.microsoft.com/en-us/library/windows/desktop/ms684215%28v=vs.85%29.aspx

	// Get the current security descriptor.
	// first call to QueryServiceObjectSecurity will retun the buffer size we need
    if (!QueryServiceObjectSecurity(schService, DACL_SECURITY_INFORMATION,&psd,0, &dwBytesNeeded))
    {
        if (GetLastError() == ERROR_INSUFFICIENT_BUFFER)
        {
			psd = (PSECURITY_DESCRIPTOR) JHI_ALLOC(dwBytesNeeded);
			if (psd == NULL) {
				TRACE0("malloc of psd failed.");
				goto cleanup;
			}

			memset(psd,0,dwBytesNeeded);
  
            if (!QueryServiceObjectSecurity(schService,DACL_SECURITY_INFORMATION, psd, dwBytesNeeded, &dwBytesNeeded))
            {
                printf("QueryServiceObjectSecurity failed (%lu)\n", GetLastError());
                goto cleanup;
            }
        }
        else 
        {
            printf("QueryServiceObjectSecurity failed (%lu)\n", GetLastError());
            goto cleanup;
        }
    }

    // Get the DACL.

    if (!GetSecurityDescriptorDacl(psd, &bDaclPresent, &pacl, &bDaclDefaulted))
    {
        printf("GetSecurityDescriptorDacl failed(%lu)\n", GetLastError());
        goto cleanup;
    }

	bool functionSuccess = GetEveryoneGroupName(&everyoneGroupName);
	if (!functionSuccess)
	{
        printf("GetEveryoneGroupName failed(%lu)\n", GetLastError());		
        goto cleanup;
	}
	TRACE1("everyoneGroupName found = %S", everyoneGroupName);

    // Build the ACE.

    BuildExplicitAccessWithName(&ea, everyoneGroupName, SERVICE_START | SERVICE_QUERY_STATUS, SET_ACCESS, NO_INHERITANCE);
    
    dwError = SetEntriesInAcl(1, &ea, pacl, &pNewAcl);
    if (dwError != ERROR_SUCCESS)
    {
        printf("SetEntriesInAcl failed(%d)\n", dwError);
        goto cleanup;
    }

    // Initialize a new security descriptor.
    if (!InitializeSecurityDescriptor(&sd, SECURITY_DESCRIPTOR_REVISION))
    {
        printf("InitializeSecurityDescriptor failed(%lu)\n", GetLastError());
        goto cleanup;
    }

    // Set the new DACL in the security descriptor.
    if (!SetSecurityDescriptorDacl(&sd, TRUE, pNewAcl, FALSE))
    {
        printf("SetSecurityDescriptorDacl failed(%lu)\n", GetLastError());
        goto cleanup;
    }

    // Set the new DACL for the service object.
    if (!SetServiceObjectSecurity(schService, DACL_SECURITY_INFORMATION, &sd))
    {
        printf("SetServiceObjectSecurity failed(%lu)\n", GetLastError());
        goto cleanup;
    }
	
	status = true;
	
cleanup:

	if(everyoneGroupName != NULL)
	{
		JHI_DEALLOC(everyoneGroupName);
		everyoneGroupName = NULL;
	}

    if(NULL != pNewAcl)
        LocalFree((HLOCAL)pNewAcl);
    
	if(psd != NULL)
	{
        JHI_DEALLOC(psd);
		psd = NULL;
	}

	return status;
}

// Code taken from https://msdn.microsoft.com/en-us/library/windows/desktop/ms683500(v=vs.85).aspx
// Installs JHI service in the SCM database
int SvcInstall()
{
	SC_HANDLE schSCManager;
    SC_HANDLE schService;
    TCHAR szPath[MAX_PATH];
	memset(szPath, 0, MAX_PATH);
	DWORD pathLength = 0;

	try
	{
		pathLength = GetModuleFileName( NULL, szPath, MAX_PATH );
		if (pathLength == 0)
		{
			printf("Cannot install service (%lu)\n", GetLastError());
			return JHI_SERVICE_GENERAL_ERROR;
		}

		// Adding quotes to the path.

		TCHAR szQuotedPath[MAX_PATH + 3];
		memset(szQuotedPath, 0, MAX_PATH + 3);
		szQuotedPath[0] = '\"';
		wcsncpy_s(szQuotedPath + 1, MAX_PATH + 2, szPath, pathLength);
		szQuotedPath[pathLength + 1] = '\"';
		szQuotedPath[pathLength + 2] = 0;

		// Get a handle to the SCM database. 
 
		schSCManager = OpenSCManager( 
			NULL,                    // local computer
			NULL,                    // ServicesActive database 
			SC_MANAGER_CREATE_SERVICE); 
 
		if (NULL == schSCManager) 
		{
			int errorCode = GetLastError();
			
			if (errorCode == ERROR_ACCESS_DENIED)
			{
				printf("ACCESS DENIED: administrative privileges required.\n");
				return JHI_SERVICE_ACCESS_DENIED;
			} 

			printf("OpenSCManager failed (%d)\n",errorCode );
			return JHI_SERVICE_GENERAL_ERROR;
		}

		// Create the service

		schService = CreateService( 
			schSCManager,              // SCM database 
			SVCNAME,                   // name of service 
			SVC_DISPLAY_NAME,		   // service name to display 
			SERVICE_ALL_ACCESS,        // desired access 
			SERVICE_WIN32_OWN_PROCESS, // service type 
			SERVICE_AUTO_START,        // start type 
			SERVICE_ERROR_NORMAL,      // error control type 
			szQuotedPath,              // path to service's binary 
			NULL,                      // no load ordering group 
			NULL,                      // no tag identifier 
			NULL,                      // no dependencies 
			NULL,                      // LocalSystem account 
			NULL);                     // no password 
 
		if (schService == NULL) 
		{
			CloseServiceHandle(schSCManager);

			int errorCode = GetLastError();
			if (errorCode == ERROR_SERVICE_EXISTS)
			{
				printf("Install failed: service already exist.\n"); 
				return JHI_SERVICE_ALREADY_EXISTS;
			}
			
			printf("Install failed (%d)\n",errorCode); 
			return JHI_SERVICE_GENERAL_ERROR;
		}
		

		// change Service load time to delayed auto-start, in windows vista and above
		// Apps can force the service to start sooner by using the StartService function
		
		if (isVistaOrLater())
		{
			SERVICE_DELAYED_AUTO_START_INFO delayedStartInfo = { true };

			if (!ChangeServiceConfig2(schService,SERVICE_CONFIG_DELAYED_AUTO_START_INFO,&delayedStartInfo))
			{
				CloseServiceHandle(schService); 
				CloseServiceHandle(schSCManager);
			
				printf("Install error: Couldn't set the service to delayed auto-start.\n");
			
				return JHI_SERVICE_GENERAL_ERROR;
			}

			// Set Service ACL to allow everyone to be able to start the service sooner
			// using the StartService function
			if (!SetServiceACL(schService))
			{
				CloseServiceHandle(schService); 
				CloseServiceHandle(schSCManager);
				return JHI_SERVICE_GENERAL_ERROR;
			}
		}

		//Change description
		SERVICE_DESCRIPTION sd;
		sd.lpDescription = 
			TEXT("Intel(R) Dynamic Application Loader Host Interface Service - Allows applications to access the local Intel (R) DAL");
	
		if (!ChangeServiceConfig2(schService,SERVICE_CONFIG_DESCRIPTION,&sd))
		{
			CloseServiceHandle(schService); 
			CloseServiceHandle(schSCManager);
			
			printf("Install error: Couldn't change the description\n");
			
			return JHI_SERVICE_GENERAL_ERROR;
		}

		//// Set Recovery policy
		//SERVICE_FAILURE_ACTIONS serviceActions;
		//SC_ACTION actions[3];
		//
		//serviceActions.dwResetPeriod = INFINITE;
		//serviceActions.lpRebootMsg = NULL;
		//serviceActions.lpCommand = NULL;
		//serviceActions.cActions = 3;
		//serviceActions.lpsaActions = actions;

		//// on first 2 crashes, restart the service, afterwards do not restart automatically
		//actions[0].Type = SC_ACTION_RESTART;
		//actions[0].Delay = 0;

		//actions[1].Type = SC_ACTION_RESTART;
		//actions[1].Delay = 0;
		//
		//actions[2].Type = SC_ACTION_NONE;
		//actions[2].Delay = 0;

		//if (!ChangeServiceConfig2(schService,SERVICE_CONFIG_FAILURE_ACTIONS,&serviceActions))
		//{
		//	CloseServiceHandle(schService); 
		//	CloseServiceHandle(schSCManager);
		//	
		//	printf("Install error: Couldn't set service recovery policy\n");
		//	
		//	return JHI_SERVICE_GENERAL_ERROR;
		//}

		CloseServiceHandle(schService); 
		CloseServiceHandle(schSCManager);

	}
	catch (...)
	{
		printf("JHI Service install failed\n"); 
		return JHI_SERVICE_GENERAL_ERROR;
	}

	printf("JHI Service installed successfully\n"); 

	return JHI_SERVICE_SUCCESS;
}

// Uninstalls JHI service in the SCM database
int SvcUninstall(void)
{
	SC_HANDLE schService;
	SC_HANDLE schSCManager;

	try
	{
		// Get a handle to the SCM database. 
		schSCManager = OpenSCManager( 
			NULL,                    // local computer
			NULL,                    // ServicesActive database 
			SC_MANAGER_ALL_ACCESS);  // full access rights 
 
		if (NULL == schSCManager) 
		{
			int errorCode = GetLastError();
			
			if (errorCode == ERROR_ACCESS_DENIED)
			{
				printf("ACCESS DENIED: administrative privileges required.\n");
				return JHI_SERVICE_ACCESS_DENIED;
			}
			
			printf("OpenSCManager failed (%d)\n",errorCode );
			return JHI_SERVICE_GENERAL_ERROR;
		}


		schService = OpenService(schSCManager,SVCNAME,DELETE);

		if (schService == NULL) 
		{
			CloseServiceHandle(schSCManager);

			int errorCode = GetLastError();
			
			if (errorCode == ERROR_ACCESS_DENIED)
			{
				printf("ACCESS DENIED: administrative privileges required.\n");
				return JHI_SERVICE_ACCESS_DENIED;
			} 
			
			if (errorCode == ERROR_SERVICE_DOES_NOT_EXIST)
			{
				printf("Error: the service does not exist.\n");
				return JHI_SERVICE_NOT_EXISTS;
			}
			
			printf("Uninstall Service failed: (%d)\n",GetLastError()); 
			return JHI_SERVICE_GENERAL_ERROR;
		}

		if (DeleteService(schService) == FALSE)
		{
			CloseServiceHandle(schSCManager);
			CloseServiceHandle(schService);
			
			int errorCode = GetLastError();
			if (errorCode == ERROR_ACCESS_DENIED)
			{
				printf("ACCESS DENIED: administrative privileges required.\n");
				return JHI_SERVICE_ACCESS_DENIED;
			} 
			
			printf("Uninstall Service failed: (%d)\n",GetLastError()); 
			return JHI_SERVICE_GENERAL_ERROR;
		}
		
		printf("JHI Service removed successfully\n");

		CloseServiceHandle(schSCManager);
		CloseServiceHandle(schService);

	}
	catch (...)
	{
		printf("Uninstall JHI Service failed\n");
		return JHI_SERVICE_GENERAL_ERROR;
	}
	return JHI_SERVICE_SUCCESS;
}

// Start installed JHI service
int SvcStart()
{
	SC_HANDLE schService;
	SC_HANDLE schSCManager;

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
			printf("ACCESS DENIED: administrative privileges required.\n");
			return JHI_SERVICE_ACCESS_DENIED;
		}
			
		printf("OpenSCManager failed (%d)\n",errorCode );
		return JHI_SERVICE_GENERAL_ERROR;
	}


	schService = OpenService(schSCManager,SVCNAME, SERVICE_START);

	if (schService == NULL) 
    {
		CloseServiceHandle(schSCManager);

		int errorCode = GetLastError();
		if (errorCode == ERROR_ACCESS_DENIED)
		{
			printf("ACCESS DENIED: administrative privileges required.\n");
			return JHI_SERVICE_ACCESS_DENIED;
		} 
		
		if (errorCode == ERROR_SERVICE_DOES_NOT_EXIST)
		{
			printf("Error: the service does not exist.\n");
			return JHI_SERVICE_NOT_EXISTS;
		}
		
		printf("Open Service failed: (%d)\n",GetLastError()); 
        return JHI_SERVICE_GENERAL_ERROR;
    }

	if (StartService(schService,NULL,NULL) == FALSE)
	{
		CloseServiceHandle(schSCManager);
		CloseServiceHandle(schService);
		
		int errorCode = GetLastError();
		if (errorCode == ERROR_ACCESS_DENIED)
		{
			printf("ACCESS DENIED: administrative privileges required.\n");
			return JHI_SERVICE_ACCESS_DENIED;
		}
		
		if (errorCode == ERROR_SERVICE_ALREADY_RUNNING)
		{
			printf("Error: JHI service already running.\n");
			return JHI_SERVICE_ALREADY_STARTED;
		}
		
		printf("Start Service failed: (%d)\n",GetLastError()); 
        return JHI_SERVICE_GENERAL_ERROR;
	}
	
	printf("JHI Service started successfully\n");

	CloseServiceHandle(schSCManager);
	CloseServiceHandle(schService);

	return JHI_SERVICE_SUCCESS;
}

// Stop installed JHI service
int SvcStop()
{
	SC_HANDLE schService;
	SC_HANDLE schSCManager;

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
			printf("ACCESS DENIED: administrative privileges required.\n");
			return JHI_SERVICE_ACCESS_DENIED;
		}
			
		printf("OpenSCManager failed (%d)\n",errorCode );
		return JHI_SERVICE_GENERAL_ERROR;
	}


	schService = OpenService(schSCManager,SVCNAME,GENERIC_EXECUTE);

	if (schService == NULL) 
    {
		CloseServiceHandle(schSCManager);

		int errorCode = GetLastError();
		if (errorCode == ERROR_ACCESS_DENIED)
		{
			printf("ACCESS DENIED: administrative privileges required.\n");
			return JHI_SERVICE_ACCESS_DENIED;
		} 
		
		if (errorCode == ERROR_SERVICE_DOES_NOT_EXIST)
		{
			printf("Error: the service does not exist.\n");
			return JHI_SERVICE_NOT_EXISTS;
		}
		
		printf("Stop Service failed: (%d)\n",GetLastError()); 
		return JHI_SERVICE_GENERAL_ERROR;
    }

	SERVICE_STATUS sStatus;
	if (ControlService(schService,SERVICE_CONTROL_STOP,&sStatus) == FALSE)
	{
		CloseServiceHandle(schSCManager);
		CloseServiceHandle(schService);
		
		int errorCode = GetLastError();
		if (errorCode == ERROR_ACCESS_DENIED)
		{
			printf("ACCESS DENIED: administrative privileges required.\n");
			return JHI_SERVICE_ACCESS_DENIED;
		}
		
		if (errorCode == ERROR_SERVICE_NOT_ACTIVE)
		{
			printf("Error: JHI service has not been started.\n");
			return JHI_SERVICE_NOT_STARTED;
		}
		
		printf("Stop Service failed: (%d)\n",GetLastError()); 
        return JHI_SERVICE_GENERAL_ERROR;
	}
	
	printf("JHI Service terminated successfully\n");

	CloseServiceHandle(schSCManager);
	CloseServiceHandle(schService);

	return JHI_SERVICE_SUCCESS;
}

bool RegisterHeciDeviceEvents()
{
	FWInfoWin32 fwInfo;
	WCHAR DevicePath[256] = { 0 };

	int heciMaxAttempts = 100;
	int attemptsCounter = 0;

	if (heciDevice == NULL)
	{
		if (!fwInfo.GetHeciDeviceDetail(&DevicePath[0]))
		{
			TRACE0("failed getting heci device details\n");
			return false;
		}

		for (; attemptsCounter < heciMaxAttempts; ++attemptsCounter)
		{
			heciDevice = fwInfo.GetHandle(DevicePath);
			if (heciDevice != INVALID_HANDLE_VALUE)
				break;
			if (attemptsCounter < heciMaxAttempts - 1) // if not last attempt.
			{
				TRACE0("***JHI_SERVICE- Failed to get heci device handle.\nSleeping then retrying...\n");
				Sleep(50);
				TRACE1("***JHI_SERVICE- Attempt #%i to get heci device handle.\n", attemptsCounter + 2);
			}
		}

		if (heciDevice == INVALID_HANDLE_VALUE)
		{
			TRACE0("failed to get heci device handle\n");
			heciDevice = NULL;
			return false;
		}

		heciNotifyHandle = NULL;
    
		DEV_BROADCAST_HANDLE filter;
		memset(&filter, 0, sizeof(filter));

		filter.dbch_size = sizeof(filter);
		filter.dbch_devicetype = DBT_DEVTYP_HANDLE;
		filter.dbch_handle = heciDevice;
	
		heciNotifyHandle = RegisterDeviceNotification(gSvcStatusHandle, &filter, DEVICE_NOTIFY_SERVICE_HANDLE);
	}

	return true;
}

bool UnRegisterHeciDeviceEvents()
{
	if (heciDevice != NULL)
	{
		if (CloseHandle(heciDevice) == FALSE)
		{
			TRACE0("failed to close heci handle");
			return false;
		}
		heciDevice = NULL;

		if (UnregisterDeviceNotification(heciNotifyHandle) == FALSE)
			return false;
	}

	return true;
}

// Code taken from https://msdn.microsoft.com/en-us/library/windows/desktop/bb540475(v=vs.85).aspx
//
// Purpose: 
//   Entry point for the service
//
// Parameters:
//   dwArgc - Number of arguments in the lpszArgv array
//   lpszArgv - Array of strings. The first string is the name of
//     the service and subsequent strings are passed by the process
//     that called the StartService function to start the service.
// 
// Return value:
//   None.
//
void WINAPI SvcMain( DWORD dwArgc, LPTSTR *lpszArgv )
{
    // Register the handler function for the service

    gSvcStatusHandle = RegisterServiceCtrlHandlerEx( 
        SVCNAME, 
        (LPHANDLER_FUNCTION_EX)SvcCtrlHandler,NULL);

    if( !gSvcStatusHandle )
    { 
        //SvcReportEvent(TEXT("RegisterServiceCtrlHandler")); 
        return; 
    } 

    // These SERVICE_STATUS members remain as set here
    gSvcStatus.dwServiceType = SERVICE_WIN32_OWN_PROCESS; 
    gSvcStatus.dwServiceSpecificExitCode = 0;    

    // Report initial status to the SCM
	ReportSvcStatus(SERVICE_START_PENDING, NO_ERROR, 0 );


    if (!jhi_init())
    {
 	   ReportSvcStatus( SERVICE_STOPPED, NO_ERROR, 0 );
    }
    else
    { 
	    // start the service
	    jhi_start();
	    ReportSvcStatus( SERVICE_RUNNING, NO_ERROR, 0 );
    }
}



//
// Purpose: 
//   Sets the current service status and reports it to the SCM.
//
// Parameters:
//   dwCurrentState - The current state (see SERVICE_STATUS)
//   dwWin32ExitCode - The system error code
//   dwWaitHint - Estimated time for pending operation, 
//     in milliseconds
// 
// Return value:
//   None
//
void ReportSvcStatus( DWORD dwCurrentState,
                      DWORD dwWin32ExitCode,
                      DWORD dwWaitHint)
{
    static DWORD dwCheckPoint = 1;

    // Fill in the SERVICE_STATUS structure.

    gSvcStatus.dwCurrentState = dwCurrentState;
    gSvcStatus.dwWin32ExitCode = dwWin32ExitCode;
    gSvcStatus.dwWaitHint = dwWaitHint;

    if (dwCurrentState == SERVICE_START_PENDING)
        gSvcStatus.dwControlsAccepted = 0;
    else gSvcStatus.dwControlsAccepted = SERVICE_ACCEPT_STOP;

    if ( (dwCurrentState == SERVICE_RUNNING) ||
           (dwCurrentState == SERVICE_STOPPED) )
        gSvcStatus.dwCheckPoint = 0;
    else gSvcStatus.dwCheckPoint = dwCheckPoint++;

    // Report the status of the service to the SCM.
    SetServiceStatus( gSvcStatusHandle, &gSvcStatus );
}

//
// Purpose: 
//   Called by SCM whenever a control code is sent to the service
//   using the ControlService function.
//
// Parameters:
//   dwCtrl - control code
// 
// Return value:
//   None
//
DWORD WINAPI SvcCtrlHandler(DWORD dwOpcode,DWORD evtype, PVOID evdata, PVOID Context)
{
   // Handle the requested control code. 

   switch(dwOpcode) 
   {  
      case SERVICE_CONTROL_STOP: 
         ReportSvcStatus(SERVICE_STOP_PENDING, NO_ERROR, 0);

         jhi_stop();

         ReportSvcStatus(SERVICE_STOPPED, NO_ERROR, 0);
         
         break;

		 // register anyway, if HECI doesn't exist, or JHI won't register this event will not occur.
	  case SERVICE_CONTROL_DEVICEEVENT:
		  
		  switch (evtype)
		  {
			case DBT_DEVICEQUERYREMOVE: TRACE0("Removing HECI device..."); jhi_invoke_reset(); break;

			case DBT_DEVICEREMOVECOMPLETE: TRACE0("HECI device removed"); break;
			
			case DBT_DEVICEQUERYREMOVEFAILED: TRACE0("HECI device removal failed"); break;
		  }

		  break;

      case SERVICE_CONTROL_INTERROGATE: 
         break; 

	// power events can be added here
 
      default: 
         break;
   } 
   
   return NO_ERROR;
}

