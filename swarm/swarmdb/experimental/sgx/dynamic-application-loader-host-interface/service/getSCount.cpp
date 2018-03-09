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
#include "misc.h"
#include "dbg.h"   
#include "SessionsManager.h"
#include "AppletsManager.h"

using namespace intel_dal;

//-------------------------------------------------------------------------------
// Function: jhis_get_sessions_count
//		    Used to get the number of active sessions of an installed applet
// IN	  : pAppId - AppId of the applet
// OUT    : pSessionsCount - the sessions count
// RETURN : JHI_RET - success or any failure returns
//-------------------------------------------------------------------------------
JHI_RET_I
jhis_get_sessions_count(
	const char* pAppId,
	UINT32* pSessionsCount
)
{
	SessionsManager& Sessions = SessionsManager::Instance();
	AppletsManager&  Applets = AppletsManager::Instance();

	UINT32 ulRetCode = JHI_INTERNAL_ERROR;
	JHI_APPLET_STATUS appStatus;

	TRACE0("dispatching jhis_get_sessions_count\n") ;

	do
	{

		if (pSessionsCount == NULL)
		{
			ulRetCode = JHI_INVALID_PARAMS;
			break;
		}

		appStatus = Applets.getAppletState(pAppId);

		if ( !( (appStatus >= 0) && (appStatus < MAX_APP_STATES) ) )
		{
			TRACE2 ("AppState incorrect: %d for appid: %s \n", appStatus, pAppId);
			ulRetCode = JHI_INTERNAL_ERROR;
			break;
		}

		if (appStatus == NOT_INSTALLED)
		{
			bool isAcp;
			FILESTRING filename;
			if (Applets.appletExistInRepository(pAppId, &filename, isAcp))
			{
				ulRetCode = JHI_SUCCESS;
			}
			else
			{
				ulRetCode = JHI_APPLET_NOT_INSTALLED;
			}

			*pSessionsCount = 0;
			
			break;
		}

		list<JHI_SESSION_ID> Slist = Sessions.getJHISessionHandles(pAppId);
		*pSessionsCount = (UINT32)Slist.size();
		TRACE2 ("jhis_get_sessions_count - session count for applet: %s = %u\n", pAppId, *pSessionsCount);
		ulRetCode = JHI_SUCCESS;

	}
	while(0);

	return ulRetCode ;
}