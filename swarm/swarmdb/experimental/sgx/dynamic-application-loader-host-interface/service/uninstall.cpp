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
**    @file uninstall.c
**
**    @brief  Defines functions for the JHI UnInstall interface
**
**    @author Niveditha Sundaram
**    @author Venky Gokulrangan
**
********************************************************************************
*/
#include "jhi_service.h"
#include "misc.h"
#include "dbg.h"
#include "SessionsManager.h"
#include "AppletsManager.h"
#include "GlobalsManager.h"
#include "string_s.h"

using namespace intel_dal;

//-------------------------------------------------------------------------------
// JHI_UNLOAD - removes applet from JoM only, local copy not removed
//-------------------------------------------------------------------------------

JHI_RET_I
jhis_unload(const char* pAppId, const SD_SESSION_HANDLE handle, vector<uint8_t>* blob)
{
	SessionsManager& Sessions = SessionsManager::Instance();
	AppletsManager&  Applets = AppletsManager::Instance();
	
	UINT32 ulRetCode = JHI_INTERNAL_ERROR;
	
	JHI_VM_TYPE vmType = GlobalsManager::Instance().getVmType();

	JHI_APPLET_STATUS appStatus = Applets.getAppletState(pAppId);

	if ( ! ( (appStatus >= 0) && (appStatus < MAX_APP_STATES) ) )
	{
		TRACE2 ("Uninstall: AppState incorrect-> %d for appid: %s \n", appStatus, pAppId);
		return JHI_INTERNAL_ERROR ;
	}
	
	if (NOT_INSTALLED == appStatus)
	{
		TRACE0 ("Uninstall: Invoked for an app that does not exist in app table ");
		if (vmType != JHI_VM_TYPE_BEIHAI_V2)
			return JHI_APPLET_NOT_INSTALLED;
	}

	// update sessions owners list and try to perform
	// sessions cleanup of non shared sessions in order to avoid failure cause by abandoned session.
	Sessions.ClearSessionsDeadOwners();
	Sessions.ClearAbandonedNonSharedSessions();

	if (!Sessions.AppletHasNonSharedSessions(pAppId))
	{
		// remove the applet shared session in case it is not in use
		Sessions.ClearAppletSharedSession(pAppId);
	}
	
	// do not allow app uninstall in case there are any type of sessions at this point
	if (Sessions.hasLiveSessions(pAppId))
	{
		ulRetCode = JHI_UNINSTALL_FAILURE_SESSIONS_EXISTS;
		return ulRetCode;
	}

	VM_Plugin_interface* plugin = NULL;
	if (( !GlobalsManager::Instance().getPluginTable(&plugin)) || (plugin == NULL) )
	{
		// probably a reset
		return JHI_NO_CONNECTION_TO_FIRMWARE;	
	}
	
	// Non-CSE - Guaranteed that the app state is 1 or 2.
	// In CSE, trying to unload even if it's not
	// known to be installed because installations
	// are persistent.
	TRACE0 ("Calling Plugin to unload the applet");
	if (blob == NULL)
	{
		ulRetCode = plugin->JHI_Plugin_UnloadApplet(pAppId);
	}
	else
	{
		ulRetCode = plugin->JHI_Plugin_SendCmdPkg(handle, *blob);
	}

	if (JHI_SUCCESS == ulRetCode)
	{
		//REMOVE THE ENTRY
		if (!Applets.remove(pAppId))
		{
			TRACE0 ("Unable to delete app table entry\n");
			// Delete failed, could be different reasons.
			// Over Beihai v2 ignore the error because applets may be installed without
			// JHI knowing about it because installations are persistent.
			if (vmType != JHI_VM_TYPE_BEIHAI_V2)
				ulRetCode = JHI_INTERNAL_ERROR; //command failed
		}
		TRACE0 ("JOM delete success");
	}
	else //JOM delete failed 
	{
		TRACE1 ("JOM delete failed: %08x\n", ulRetCode);
	}

	return ulRetCode;
}


//-------------------------------------------------------------------------------
// Function: jhis_uninstall
//		    Used to remove an applet from JoM and local disk
// IN	  : pAppId - AppId of package to be uninstalled
// IN	  : blob - optional - if null, run legacy uninstall.
// RETURN : JHI_RET - success or any failure returns
//-------------------------------------------------------------------------------
//1.	On receiving the Uninstall command, JHI checks to see if APPID is in 
//		app table
//		a.	If app is not present, send back APPID_NOT_EXIST
//		b.	If app is present, call TL Uninstall sequence:
//			i.	Look for any active sessions corresponding to the applet in the 
//				session table
//			ii.	If present, close the sessions and remove session table entry
//			iii.Remove the applet from the JoM by issuing remove service command 
//				to JoM
//2.	Remove corresponding app table entry from app table
//3.	Delete the corresponding PACK/DALP file from disk
//-------------------------------------------------------------------------------

JHI_RET_I
jhis_uninstall(const char* pAppId, const SD_SESSION_HANDLE handle, vector<uint8_t>* blob)
{
	UINT32 ulRetCode = JHI_INTERNAL_ERROR;

	TRACE0("dispatching JHIS Uninstall\n") ;

	ulRetCode = jhis_unload(pAppId, handle, blob);

	//Regardless of whether applet was present in JOM or not, we will attempt to remove 
	//the file just in case it is present in disk
	if ((JHI_SUCCESS == ulRetCode) || (JHI_APPLET_NOT_INSTALLED == ulRetCode) || (TEE_STATUS_TA_DOES_NOT_EXIST) == ulRetCode )
	{
		bool isAcp;
		FILESTRING filename;
		if (AppletsManager::Instance().appletExistInRepository(pAppId,&filename, isAcp))
		{
			if (_wremove(filename.c_str()) != 0)
			{
#ifdef _WIN32
				volatile DWORD x = GetLastError();
				TRACE1 (" JHI file removal from disk failed, error %d\n", x);
#else
				TRACE0 (" JHI file removal from disk failed\n");
#endif

				if (JHI_SUCCESS == ulRetCode)
					ulRetCode = JHI_DELETE_FROM_REPOSITORY_FAILURE;
			}
			else
			{
				ulRetCode = JHI_SUCCESS;
			}
		}
		
	}
	else //unload failed
	{
		TRACE0 ("JHI Unload failed\n");
	}

	return ulRetCode ;
}