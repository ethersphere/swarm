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
**    @file jhi_sdk.h
**
**    @brief  Defines exported interfaces for JHI.DLL
**
**    @author Oded Angel
**
********************************************************************************
*/

#ifndef __JHI_H_SDK__
#define __JHI_H_SDK__

#include "jhi.h"
#include "typedefs_i.h"
#ifdef _WIN32
#include <rpc.h>
#endif // _WIN32

#ifdef __cplusplus
extern "C" {
#endif


typedef  UUID   JHI_SESSION_ID;

typedef struct
{
	UINT32 pid;				// represent the application process id.
	FILETIME creationTime;	// represent the application creation time
} JHI_PROCESS_INFORMATION;

// this struct contains extended information for a given session
typedef struct 
{
	JHI_SESSION_ID				sessionId;
	char						appId[32];
	UINT32						flags;			// the flags used when this session created
	JHI_SESSION_STATE			state;			// the session state
	UINT32						ownersListCount;
	JHI_PROCESS_INFORMATION*	ownersList;
	UINT32						reserved[20];	// reserved bits
} JHI_SESSION_EXTENDED_INFO;

typedef struct 
{
	UINT32					sessionsCount;
	JHI_SESSION_EXTENDED_INFO*	dataTable;
} JHI_SESSIONS_DATA_TABLE;


typedef struct 
{
	UINT32					loadedAppletsCount;
	char**					appsGUIDs;
} JHI_LOADED_APPLET_GUIDS;



//------------------------------------------------------------
// Function: JHI_GetSessionTable
//------------------------------------------------------------
// Note: to avoid memory leaks the application must call 
//	the JHI_FreeSessionTable API after the end of use.
//------------------------------------------------------------
JHI_EXPORT
JHI_GetSessionTable(
	OUT JHI_SESSIONS_DATA_TABLE** SessionDataTable
);

//------------------------------------------------------------
// Function: JHI_FreeSessionTable
//------------------------------------------------------------
JHI_EXPORT
JHI_FreeSessionTable(
	IN JHI_SESSIONS_DATA_TABLE** SessionDataTable
);

//------------------------------------------------------------
// Function: JHI_GetLoadedAppletsList
//------------------------------------------------------------
// Note: to avoid memory leaks the application must call 
//	the JHI_FreeLoadedAppletsListl API after the end of use.
//------------------------------------------------------------
JHI_EXPORT
JHI_GetLoadedAppletsList(
	OUT JHI_LOADED_APPLET_GUIDS** appGUIDs
);

//------------------------------------------------------------
// Function: JHI_FreeLoadedAppletsList
//------------------------------------------------------------
JHI_EXPORT
JHI_FreeLoadedAppletsList(
	IN JHI_LOADED_APPLET_GUIDS** appGUIDs
);

#ifdef __cplusplus
};
#endif


#endif