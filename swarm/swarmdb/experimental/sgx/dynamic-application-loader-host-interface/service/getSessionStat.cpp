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
**    @file createSession.cpp
**
**    @brief  Defines functions for the JHI session interface
**
**    @author Elad Dabool
**
********************************************************************************
*/
#include "jhi_service.h"
#include "dbg.h"   
#include "SessionsManager.h"
#include "GlobalsManager.h"
#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
#include "jhi_sdk.h"
#endif

using namespace intel_dal;

//-------------------------------------------------------------------------------
// Function: jhis_get_session_info
//		    Used to get information of a session
// IN	  : pSessionHandle - handle to the session
// OUT    : pStatus - the session status
// RETURN : JHI_RET - success or any failure returns
//-------------------------------------------------------------------------------
JHI_RET_I
jhis_get_session_info(
	JHI_SESSION_ID*      pSessionID,
	JHI_SESSION_INFO*	 pSessionInfo
)
{
	SessionsManager& Sessions = SessionsManager::Instance();

	UINT32 ulRetCode = JHI_INTERNAL_ERROR;
	
	TRACE0("dispatching JHIS GET_SESSION_INFO\n") ;
	
	do
	{

		if (pSessionInfo == NULL || pSessionID == NULL)
		{
			ulRetCode = JHI_INVALID_PARAMS;
			break;
		}
		
		Sessions.getSessionInfo(*pSessionID,pSessionInfo);
		
		ulRetCode = JHI_SUCCESS;
	}
	while (0);

	return ulRetCode;
}

#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
//-------------------------------------------------------------------------------
// Function: jhis_get_sessions_data_table
//		    Used to get information of a session
// OUT    : JHI_SESSIONS_DATA_TABLE - the session data table
// RETURN : JHI_RET - success or any failure returns
//-------------------------------------------------------------------------------
JHI_RET_I
jhis_get_sessions_data_table(
	JHI_SESSIONS_DATA_TABLE*	 pSessionsDataTable
)
{
	SessionsManager& Sessions = SessionsManager::Instance();

	UINT32 ulRetCode = JHI_INTERNAL_ERROR;
	
	TRACE0("dispatching JHIS GET_SESSION_INFO\n") ;
	
	do
	{
		Sessions.getSessionsDataTable(pSessionsDataTable);
		
		ulRetCode = JHI_SUCCESS;
	}
	while (0);

	return ulRetCode;
}
#endif