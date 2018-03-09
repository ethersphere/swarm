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
**    @brief  Defines functions for the JHI session creation interface
**
**    @author Elad Dabool
**
********************************************************************************
*/
#include "jhi_service.h"
#include "misc.h"
#include "dbg.h"   
#include "SessionsManager.h"
#include "GlobalsManager.h"

using namespace intel_dal;

//-------------------------------------------------------------------------------
// Function: jhis_close_session
//		    Used to close an active session of an applet
// IN	  : pSessionHandle - the session handle
// IN     : processInfo - the calling process, in case of NULL this is an internal JHI request.
// RETURN : JHI_RET - success or any failure returns
//-------------------------------------------------------------------------------
JHI_RET_I
jhis_close_session(
	JHI_SESSION_ID*		  pSessionID,
	JHI_PROCESS_INFO*	processInfo,
	bool force,
	bool removeFromVM
	)
{
	SessionsManager& Sessions = SessionsManager::Instance();
	VM_SESSION_HANDLE VMSessionHandle;
	UINT32 ulRetCode = JHI_INTERNAL_ERROR;
	bool removeSession = false;	
	
	JHI_SESSION_INFO info;
	JHI_SESSION_FLAGS sessionFlags;
	
	TRACE0("dispatching JHIS CLOSE_SESSION\n") ;

	// check that the session exists

	Sessions.getSessionInfo(*pSessionID,&info);

	sessionFlags.value = info.flags;

	if (info.state == JHI_SESSION_STATE_NOT_EXISTS)
		return JHI_INVALID_SESSION_HANDLE;

	// if processInfo == NULL or there are no more owners, remove the session
	if (processInfo == NULL)
	{
		removeSession = true;
	}
	else
	{
		if (Sessions.isSessionOwnerValid(*pSessionID,processInfo))
		{
			if ((Sessions.getOwnersCount(*pSessionID) == 1) && (!sessionFlags.bits.sharedSession))
			{
				removeSession = true;
			}
			else
			{
				Sessions.removeSessionOwner(*pSessionID,processInfo);
			}
			ulRetCode = JHI_SUCCESS;
		}
		else
		{
			// error there is no such session owner.
			ulRetCode = JHI_INTERNAL_ERROR;
		}
	}

	if (removeSession)
	{

		// Acquire a lock unless you want to force the closure of a session.
		// We already checked that the session exists above.
		if (force == false)
		{
			if (!Sessions.GetSessionLock(*pSessionID))
				return JHI_INVALID_SESSION_HANDLE;
		}

		if (removeFromVM)
		{
			if (!Sessions.getVMSessionHandle(*pSessionID,&VMSessionHandle))
				return JHI_INTERNAL_ERROR;

			VM_Plugin_interface* plugin = NULL;
			if ( (!GlobalsManager::Instance().getPluginTable(&plugin)) || (plugin == NULL) )
			{
				// probably a reset
				ulRetCode = JHI_NO_CONNECTION_TO_FIRMWARE;		
			}
			else
			{
				ulRetCode = force ? plugin->JHI_Plugin_ForceCloseSession(&VMSessionHandle) : plugin->JHI_Plugin_CloseSession(&VMSessionHandle);
				// In case of forced closure, where we didn't acquire a lock before, acquire it now.
				// If it succeeds, it means that no other thread removed the session record yet, we should do that.
				// If it fails, it means that the session record has already been deleted. We can stop here and return the result of the operation.
				if (force == true)
				{
					if (!Sessions.GetSessionLock(*pSessionID))
						return ulRetCode;
				}
			}
		}
		else
		{
			// no need to remove the session from the VM
			// usualy when the session has crashed during SendAndRecieve
			ulRetCode = JHI_SUCCESS;
		}

		if (ulRetCode == JHI_SUCCESS || ulRetCode == JHI_APPLET_FATAL)
		{
			// FW closed the session, remove its entry from our session table
			// In case of forced closure, the entry could have been already removed
			// so a falure of the function is not an error.
			if (!Sessions.remove(*pSessionID))
				if (force == false)
					ulRetCode = JHI_INTERNAL_ERROR;
		}
		else
		{	
			Sessions.ReleaseSessionLock(*pSessionID);
		}

	}
	
	return ulRetCode ;
}