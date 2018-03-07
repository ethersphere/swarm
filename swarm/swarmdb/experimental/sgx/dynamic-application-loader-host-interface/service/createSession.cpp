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
#include "AppletsManager.h"

using namespace intel_dal;

//-------------------------------------------------------------------------------
// Function: jhis_create_session
//		    Used to create a new session of an installed applet 
// IN	  : pAppId - AppId of package to be used for the creation of the session
// IN     : flags - detrmine the session properties
// RETURN : JHI_RET - success or any failure returns
//-------------------------------------------------------------------------------
JHI_RET_I
jhis_create_session(
		const char*				pAppId,
		JHI_SESSION_ID*		pSessionID,
		UINT32				flags,
		DATA_BUFFER*		initBuffer,
		JHI_PROCESS_INFO*	processInfo
)
{
	UINT32 ulRetCode = JHI_INTERNAL_ERROR;

	do
	{
		// Make sure we have a plugin to work with, and get it.
		VM_Plugin_interface* plugin = NULL;
		if ( (!GlobalsManager::Instance().getPluginTable(&plugin)) || (plugin == NULL) )
		{
			// probably we had a reset
			ulRetCode = JHI_NO_CONNECTION_TO_FIRMWARE;
			break;
		}

		AppletsManager&  Applets = AppletsManager::Instance();
		SessionsManager& Sessions = SessionsManager::Instance();

		// Needed only in CSE to get the applet blob from the repository
		list<vector<uint8_t> > appletBlobs;

		// Different flows depending on FW type. ME and SEC vs CSE
		JHI_VM_TYPE vmType = GlobalsManager::Instance().getVmType();

		// For ME/SEC make sure that the applet is installed by looking at the repository.
		// In CSE we must have the applet in the repository because the applet blob is needed for session creation.
		FILESTRING filename;
		bool isAcp = false;
		if (!Applets.appletExistInRepository(pAppId, &filename, isAcp))
		{
			ulRetCode = JHI_APPLET_NOT_INSTALLED;
			break;
		}
		// In CSE we need the blobs for the create session API
		if (vmType == JHI_VM_TYPE_BEIHAI_V2)
		{
			ulRetCode = Applets.getAppletBlobs(filename, appletBlobs, isAcp);
			if (ulRetCode != JHI_SUCCESS)
			{
				TRACE0("failed getting applet blobs from dalp file\n");
				break;
			}
			// If ulRetCode == JHI_SUCCESS appletBlobs will not be empty and VMSessionHandle will always be initialized
		}

		// In ME/SEC, verify the applet is installed before trying to create a session, and if it's not, install it.
		if (vmType != JHI_VM_TYPE_BEIHAI_V2)
		{
			JHI_APPLET_STATUS appStatus = Applets.getAppletState(pAppId);

			if ( !( (appStatus >= 0) && (appStatus < MAX_APP_STATES) ) )
			{
				TRACE2 ("AppState incorrect: %d for appid: %s \n", appStatus, pAppId);
				ulRetCode = JHI_INTERNAL_ERROR;
				break;
			}

			if (appStatus == NOT_INSTALLED)
			{
				// Applet is not installed but applet file exists in the repository, try to install it.
				ulRetCode = jhis_install(pAppId, filename.c_str(), true, isAcp);

				if (ulRetCode != JHI_SUCCESS)
				{
					ulRetCode = JHI_APPLET_NOT_INSTALLED;
					break;
				}
			}
		}

		// Verify all sessions owners and perform abandoned non shared sessions clean-up
		Sessions.ClearSessionsDeadOwners();
		Sessions.ClearAbandonedNonSharedSessions();

		/*
		** Shared session
		** Check if shared session requested and there is already such session
		*/
		JHI_SESSION_FLAGS sessionFlags;
		sessionFlags.value = flags;

		if (sessionFlags.bits.sharedSession)
		{
			// In SKL and BXT, checking for Shared Session support is too heavy to be practical.
			// In these cases the check is disabled since it is not mandatory.
			// In KBL and later the enforcement is disabled completely.
			if (vmType != JHI_VM_TYPE_BEIHAI_V2 && !Applets.isSharedSessionSupported(pAppId))
			{
				ulRetCode = JHI_SHARED_SESSION_NOT_SUPPORTED;
				break;
			}
			else if (Sessions.getSharedSessionID(pSessionID,pAppId))
			{
				// Add the calling application to the session owners
				if (Sessions.addSessionOwner(*pSessionID,processInfo))
					ulRetCode = JHI_SUCCESS;
				else
					ulRetCode = JHI_MAX_SHARED_SESSION_REACHED;
				break; // No need to create a new session
			}
		}

		/*
		**  Non shared session is requested, or a shared session is requested but none for the applet exists yet.
		**  Create a new session
		*/
		VM_SESSION_HANDLE VMSessionHandle;
		JHI_SESSION_ID newSessionID;

		if (!Sessions.generateNewSessionId(&newSessionID))
		{
			ulRetCode = JHI_INTERNAL_ERROR;
			break;
		}

		// First try to create a session
		if (vmType != JHI_VM_TYPE_BEIHAI_V2)
			ulRetCode = plugin->JHI_Plugin_CreateSession(pAppId, &VMSessionHandle, NULL, 0, newSessionID, initBuffer);
		else // CSE
		{
			for (list<vector<uint8_t> >::iterator it = appletBlobs.begin(); it != appletBlobs.end(); ++it)
			{
				ulRetCode = plugin->JHI_Plugin_CreateSession(pAppId, &VMSessionHandle, &(*it)[0], (unsigned int)(*it).size(), newSessionID, initBuffer);
				if (ulRetCode == JHI_SUCCESS || ulRetCode == JHI_MAX_INSTALLED_APPLETS_REACHED || ulRetCode == JHI_MAX_SESSIONS_REACHED)
				{
					break; // break just out of the for loop
				}
			}
		}

		// If session creation failed because of MAX_SESSIONS try to close unused shared sessions
		if (ulRetCode == JHI_MAX_SESSIONS_REACHED || ulRetCode == JHI_MAX_INSTALLED_APPLETS_REACHED) // WHY MAX APPLETS REACHED??
		{
			if (Sessions.TryRemoveUnusedSharedSession(true))
			{
				// Then try to create the session again.
				if (vmType == JHI_VM_TYPE_BEIHAI_V2) // Create the session for CSE.
				{
					for (list<vector<uint8_t> >::iterator it = appletBlobs.begin(); it != appletBlobs.end(); ++it)
					{
						ulRetCode = plugin->JHI_Plugin_CreateSession(pAppId, &VMSessionHandle, &(*it)[0], (unsigned int)(*it).size(), newSessionID, initBuffer);
						if (ulRetCode == JHI_SUCCESS || ulRetCode == JHI_MAX_INSTALLED_APPLETS_REACHED || ulRetCode == JHI_MAX_SESSIONS_REACHED)
						{
							break;
						}
					}
				}
				else // not CSE
				{
					ulRetCode = plugin->JHI_Plugin_CreateSession(pAppId, &VMSessionHandle, NULL, 0, newSessionID, initBuffer);
				}
			}
		}

		if (ulRetCode == JHI_MAX_INSTALLED_APPLETS_REACHED) // AGAIN, WHY MAX APPLETS REACHED??
		{
			ulRetCode = JHI_MAX_SESSIONS_REACHED;
		}

		if (ulRetCode == JHI_SUCCESS)
		{
			// session created in FW, add an entry in the session table and return a session handle
			if (Sessions.add(pAppId,VMSessionHandle,newSessionID,sessionFlags,processInfo))
			{
				*pSessionID = newSessionID;
				break; //success, end loop
			}
			else
			{
				ulRetCode = JHI_INTERNAL_ERROR;
				plugin->JHI_Plugin_CloseSession(&VMSessionHandle);
				break;
			}
		}
	}
	while(0);

	return ulRetCode;
}
