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
**    @file Win32Service.h
**
**    @brief  Contains win32 service implementation for JHI
**
**    @author Elad Dabool
**
********************************************************************************
*/
#ifndef _WIN32SERVICE_H_
#define _WIN32SERVICE_H_

#ifdef _WIN32

#include <windows.h>
#include <dbt.h>
#include <tchar.h>
#include <strsafe.h>
#include "JHIMain.h"
#include "FWInfoWin32.h"


#pragma comment(lib, "advapi32.lib")

#ifdef SCHANNEL_OVER_SOCKET // emulation mode

#define SVCNAME TEXT("jhi_service_emulation")
#define SVC_DISPLAY_NAME TEXT("Intel(R) Dynamic Application Loader Host Interface Service - EMULATION")

#else

#define SVCNAME TEXT("jhi_service")
#define SVC_DISPLAY_NAME TEXT("Intel(R) Dynamic Application Loader Host Interface Service")

#endif

// -----------------------------------------------------------------------
// JHI service command line error codes
// ------------------------------------------------------------------------
#define JHI_SERVICE_SUCCESS				0
#define JHI_SERVICE_GENERAL_ERROR		1
#define JHI_SERVICE_ACCESS_DENIED		2
#define JHI_SERVICE_ALREADY_EXISTS		3
#define JHI_SERVICE_NOT_EXISTS			4
#define JHI_SERVICE_ALREADY_STARTED		5
#define JHI_SERVICE_NOT_STARTED			6


// service globals
static SERVICE_STATUS          gSvcStatus; 
static SERVICE_STATUS_HANDLE   gSvcStatusHandle; 
static HANDLE                  ghSvcStopEvent = NULL;

static HANDLE				   heciDevice = NULL;
static HDEVNOTIFY			   heciNotifyHandle = NULL;


// service API
int SvcInstall(void);
int SvcUninstall(void);
int SvcStart(void);
int SvcStop(void);

// API for getting events from HECI device
bool RegisterHeciDeviceEvents();
bool UnRegisterHeciDeviceEvents();
 
// windows service events callback function
DWORD WINAPI SvcCtrlHandler(DWORD dwOpcode,DWORD evtype, PVOID evdata, PVOID Context);

// main function for JHI service
void WINAPI SvcMain( DWORD, LPTSTR * ); 

VOID ReportSvcStatus( DWORD, DWORD, DWORD );
VOID SvcInit( DWORD, LPTSTR * ); 














#endif

#endif