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
**    @file jhi_i.h
**
**    @brief  Defines common types and definitions for both JHI Service and JHI DLL
**
**    @author Elad Dabool
**
********************************************************************************
*/

#ifndef __JHI_I_H__
#define __JHI_I_H__

#ifdef _WIN32
#include <windows.h>
#include <Shlwapi.h>
#else
#include <pthread.h>
#endif // _WIN32

#include "typedefs.h"
#include "typedefs_i.h"
#include "jhi_version.h"
#include "jhi.h"
#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
#include "jhi_sdk.h"
#endif
#include <list>
using std::list;
#include "jhi_event.h"

#define JHI_RET_I UINT32
#define INTEL_SD_UUID "BD2FBA36A2D64DAB9390FF6DA2FEF31C"
#define SPOOLER_APPLET_UUID "BA8D164350B649CC861D2C01BED14BE8"

// internal events errors
#define JHI_GET_EVENT_FAIL_NO_EVENTS	0x2000		// For event data retrieval failure

#define MAX_APPLET_BLOB_SIZE 2097152 // applet blob cannot be more then 2MB
#define FW_VERSION_STRING_MAX_LENGTH 50

#ifdef __linux__
#define STR_COMMAND_BEGIN_DELIMITER "<"
#define CHR_COMMAND_BEGIN_DELIMITER '<'

#define STR_COMMAND_END_DELIMITER   ">"
#define CHR_COMMAND_END_DELIMITER '>'
#endif //__linux__

// APP RELATED
#define COMMAND_OTP	1
#define LEN_APP_ID  32 	// applet id without \0 and separators

typedef UUID JHI_SESSION_ID;

typedef struct
{
	UINT32 pid;				// represent the application process id.
	FILETIME creationTime;	// represent the application creation time
} JHI_PROCESS_INFO;


typedef struct
{
	JHI_SESSION_ID sessionID;
	intel_dal::JhiEvent* eventHandle;
#ifdef _WIN32
	HANDLE	threadHandle;
#else
	pthread_t threadHandle;
#endif // _WIN32
	JHI_EventFunc	callback;
	UINT8* threadNeedToEnd;
	UINT32 sessionFlags;
	JHI_PROCESS_INFO processInfo;
} JHI_I_SESSION_HANDLE;

typedef struct 
{
	JHI_PROCESS_INFO processInfo;
	list<JHI_I_SESSION_HANDLE*>* SessionsList;
	UINT32 ReferenceCount;
} JHI_I_HANDLE ;  // For internal usage

#ifdef _WIN32
#define FILECHARLEN wcslen
#define FILESTRCPY wcscpy_s
#define FILESTRING std::wstring	
#define FILEPREFIX(a) TEXT(a)
#define FILE_SEPERATOR L"\\"
#define FILETOSTRING std::to_wstring	
#else
#define FILECHARLEN strlen
#define FILESTRCPY strcpy_s
#define FILESTRING std::string
#define FILEPREFIX(a) (a)
#define FILE_SEPERATOR "/"
#endif // _WIN32

#endif // __JHI_I_H__
