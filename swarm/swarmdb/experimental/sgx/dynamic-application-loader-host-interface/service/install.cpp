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
**    @file install.c
**
**    @brief  Defines functions for the JHI Install interface
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

// this function trys to find one applet in the loaded_applets list that has no sessions
// if it finds such applet the applet is unloaded
//
// Returns: true if succesfuly unloaded an applet, false otherwise
bool UnloadAppletWithNoSessions(list<string> loaded_applets)
{
	bool appletUnloaded = false;

	SessionsManager& Sessions = SessionsManager::Instance();

	for (list<string>::iterator it = loaded_applets.begin(); it != loaded_applets.end(); it++)
	{
		if (Sessions.getJHISessionHandles(*it).empty())
		{
			char Appid[LEN_APP_ID+1];
			strcpy_s(Appid,LEN_APP_ID + 1, (*it).c_str());

			if (jhis_unload(Appid) == JHI_SUCCESS)
			{
				TRACE1("unloaded applet with appid: %s\n",Appid);
				appletUnloaded = true;
			}
			else
			{
				TRACE0("ERROR: failed to unload applet that has no sessions!\n");
			}

			break;
		}
	}

	return appletUnloaded;
}

// this function trys to remove one unused applet
// an unused applet is an applet that has no sessions or an applet that its only session is a shared session with no owners.

// Returns: true if succesfuly unloaded an applet, false otherwise
bool TryUnloadUnusedApplet()
{
	SessionsManager& Sessions = SessionsManager::Instance();
	AppletsManager&  Applets = AppletsManager::Instance();

	list<string> loaded_applets;
	bool unloaded = false;

	Applets.getLoadedAppletsList(loaded_applets);

	// first, try to find applets that has no sessions
	if (UnloadAppletWithNoSessions(loaded_applets))
	{
		unloaded = true;
	}
	// if we reached here, there are no applets with no sessions,
	// try to find one that has only a shared session and that shared session has no owners.
	else if (Sessions.TryRemoveUnusedSharedSession(false))
	{
		if (UnloadAppletWithNoSessions(loaded_applets))
			unloaded = true;
	}

	return unloaded;
}

//-------------------------------------------------------------------------------
// Function: jhis_install
//		    Used to download an applet into JoM
// IN	  : pAppId - incoming AppId of package 
// IN     : pTFile - path of pack file to be installed
// IN     : visibleApp - determines whether to keep record of the app.
// RETURN : JHI_RET - success or any failure returns
//-------------------------------------------------------------------------------
//1.	On receiving an Install command for an applet with the specified APPID, 
//		JHI first looks up the app table to see if the applet has been installed 
//		before. 
//		a.	If the app already exists in the app table, and has been installed 
//			before, then this might be an upgrade or a re-install. In either case, 
//			JHI first closes the sessions corresponding to the applet in order to 
//			facilitate the JoM download.
//		b.	If not, then it is a fresh install
//2.	Next, JHI will download the PACK file specified in the Install command input, 
//		and see if JoM accepts the new one
//		a.	If successful:
//			i.	JHI will create a new app table entry for the applet and set the 
//				state to PRESENT
//			ii.	Copy the incoming PACK file to a location specified in the registry,
//				where all PACK files will be stored. The name of stored PACK file will 
//				be its corresponding APPID
//				1.	If copy fails for some reason, the applet will be unloaded from the
//					JoM and Copy file error returned.
//		b.	If JoM indicates the applet exists:
//			i.	If applet is present in the local app table - no change in state
//			ii.	If applet is not present in the local app table - indicates the 
//				incoming appid is faulty, return JHI_FILE_UUID_MISMATCH.
//		c.	If JoM indicates an overflow of more than 5 applets:
//			i.	If LRU implemented, then apply accordingly
//			ii.	If not, return error code
//		d.	If there is a UUID mismatch, return error code
//		e.	All other errors, return appropriate error code

//-------------------------------------------------------------------------------

JHI_RET_I
	jhis_install(
	const char* pAppId,
	const FILECHAR* pFile,
	bool visibleApp,
	bool isAcp
	)
{
	list<vector<uint8_t> > appletBlobs;
	UINT32 ulRetCode = JHI_INTERNAL_ERROR;
	JHI_APPLET_STATUS appStatus;
	string fileExtension;
	VM_Plugin_interface* plugin = NULL;

	SessionsManager& Sessions = SessionsManager::Instance();
	AppletsManager&  Applets = AppletsManager::Instance();

	TRACE2("Attempting to install - applet ID: %s\nPath: %s", pAppId, pFile);
	if (visibleApp)
	{
		appStatus = Applets.getAppletState(pAppId);

		// check if there is already an applet record in the applet table
		if ( !( (appStatus >= 0) && (appStatus < MAX_APP_STATES) ) )
		{
			TRACE2 ("AppState incorrect: %d for appid: %s \n", appStatus, pAppId);
			goto cleanup ;
		}
	}

	// try to perform sessions cleanup in order to avoid failure cause by abandoned session
	Sessions.ClearSessionsDeadOwners();
	Sessions.ClearAbandonedNonSharedSessions();

	if (!Sessions.AppletHasNonSharedSessions(pAppId))
	{
		// in case the applet was already installed,
		// try to remove the applet shared session in case it exist and not in use
		Sessions.ClearAppletSharedSession(pAppId);
	}

	if ( (!GlobalsManager::Instance().getPluginTable(&plugin)) || (plugin == NULL) )
	{
		// probably a reset
		ulRetCode = JHI_NO_CONNECTION_TO_FIRMWARE;	
		goto cleanup;
	}

	if (isAcp)
	{
		fileExtension = acpFileExt;
	}
	else
	{
		fileExtension = dalpFileExt;
	}

	// verify the applet file
	if (_waccess_s(pFile, 0) != 0)
	{
		TRACE0("prepare install failed - applet file not found");
		ulRetCode = JHI_FILE_NOT_FOUND;
		goto cleanup;
	}

	if (!Applets.compareFileExtention(pFile,fileExtension))
	{
		TRACE0("invalid applet file extension!\n");
		ulRetCode = JHI_INVALID_FILE_EXTENSION;
		goto cleanup;
	}

	if (visibleApp) // Package as temp name is the repository and mark as pending if we dont have a record in the app table
	{
		ulRetCode = Applets.prepareInstallFromFile(pFile, appletBlobs, pAppId, isAcp);
		if (ulRetCode != JHI_SUCCESS) 
			goto cleanup;
	}
	else
	{
		// get the applet blobs from the original dalp file
		ulRetCode = Applets.getAppletBlobs(pFile, appletBlobs, isAcp);
		if (ulRetCode != JHI_SUCCESS)
		{
			TRACE0("failed getting applet blobs from dalp file\n");
			goto cleanup;
		}
	}

	for (list<vector<uint8_t> >::iterator it = appletBlobs.begin(); it != appletBlobs.end(); it++)
	{
		//call the Download function using the plugin
		ulRetCode = plugin->JHI_Plugin_DownloadApplet(pAppId, &(*it)[0], (unsigned int)(*it).size() );

		if (JHI_FILE_IDENTICAL == ulRetCode) // the applet version is already exists in the VM
		{
			// Force re-install:
			plugin->JHI_Plugin_UnloadApplet(pAppId);
			ulRetCode = plugin->JHI_Plugin_DownloadApplet(pAppId, &(*it)[0], (unsigned int)(*it).size() );
			break;
		}
		else if (JHI_MAX_INSTALLED_APPLETS_REACHED == ulRetCode || ulRetCode == JHI_SUCCESS)
			break;

		TRACE1("failed to install applet from DALP, error code: 0x%x\n",ulRetCode);
	}

	// in case of applet overflow, try to perform shared session cleanup
	// using LRU algorithem and download the applet again.

	if( JHI_MAX_INSTALLED_APPLETS_REACHED == ulRetCode) 
	{
		// try to unload one applet that doesnt have an active session
		if (TryUnloadUnusedApplet())
		{
			// succeded to unload an applet, try to download the applet again  
			for (list<vector<uint8_t> >::iterator it = appletBlobs.begin(); it != appletBlobs.end(); it++)
			{
				//call the Download function using the plugin
				ulRetCode = plugin->JHI_Plugin_DownloadApplet(pAppId, &(*it)[0], (unsigned int)(*it).size() );

				if (JHI_SUCCESS == ulRetCode) 
					break;
			}
		}

		TRACE1("failed to install applet from DALP, error code: 0x%x\n", ulRetCode);
	}

	if (ulRetCode != JHI_SUCCESS)
	{
		//if (ulRetCode != JHI_MAX_INSTALLED_APPLETS_REACHED && ulRetCode != JHI_INSTALL_FAILURE_SESSIONS_EXISTS)
		//	ulRetCode = JHI_INSTALL_FAILED; // return a general error since we cannot return just the error from the last download try
		TRACE1("failed to install applet from DALP, error code: 0x%x\n", ulRetCode);

		goto errorRemoveApplet;
	}

	if (visibleApp)
	{
		// Mark the Applet as Installed
		if (!Applets.completeInstall(pAppId, isAcp)) 
		{
			ulRetCode = JHI_INTERNAL_ERROR;
			goto errorDeleteFromFW;
		}
	}

	ulRetCode = JHI_SUCCESS;
	goto cleanup;


errorDeleteFromFW:

	//call the delete function using the plugin
	plugin->JHI_Plugin_UnloadApplet(pAppId);

errorRemoveApplet:

	if (visibleApp)
	{
		// delete pending file from repository
		_wremove(AppletsManager::Instance().getPendingFileName(pAppId, isAcp).c_str()); 

		// remove applet record if it is in pending state
		if (Applets.getAppletState(pAppId) == PENDING_INSTALL)
		{
			Applets.remove(pAppId);
		}
	}

cleanup:

	if (ulRetCode != JHI_SUCCESS)
	{
		TRACE0("Applet installation failed");
	}

	return ulRetCode ;
}
