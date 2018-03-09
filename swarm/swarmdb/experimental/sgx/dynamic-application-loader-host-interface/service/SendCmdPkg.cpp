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
#include "string_s.h"
#include "dbg.h"
#include "SessionsManager.h"
#include "AppletsManager.h"
#include "GlobalsManager.h"
#include "AppletsPackageReader.h"

using namespace intel_dal;

enum AC_CMD_ID {
	AC_CMD_INVALID,
	AC_INSTALL_SD,
	AC_UNINSTALL_SD,
	AC_INSTALL_JTA,
	AC_UNINSTALL_JTA,
	AC_INSTALL_NTA,
	AC_UNINSTALL_NTA,
	AC_UPDATE_SVL,
	AC_INSTALL_JTA_PROP,
	AC_CMD_NUM
};

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

JHI_RET_I cmd_pkg_install_jta (string& pAppId, const SD_SESSION_HANDLE handle, vector<uint8_t>& blob)
{
	UINT32 ulRetCode = JHI_INTERNAL_ERROR;
	JHI_APPLET_STATUS appStatus;
	VM_Plugin_interface* plugin = NULL;

	SessionsManager& Sessions = SessionsManager::Instance();
	AppletsManager&  Applets = AppletsManager::Instance();

	appStatus = Applets.getAppletState(pAppId);

	// check if there is allready an applet record in the applet table
	if ( !( (appStatus >= 0) && (appStatus < MAX_APP_STATES) ) )
	{
		TRACE2 ("AppState incorrect: %d for appid: %s \n", appStatus, pAppId.c_str());
		goto cleanup ;
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

	ulRetCode = Applets.prepareInstallFromBuffer(blob, pAppId);
	if (ulRetCode != JHI_SUCCESS) 
		goto cleanup;

	//call the Download function using the plugin
	ulRetCode = plugin->JHI_Plugin_SendCmdPkg(handle, blob);

	if (TEE_STATUS_IDENTICAL_PACKAGE == ulRetCode) // the applet version is already exists in the VM
	{
		// Force re-install:
		plugin->JHI_Plugin_UnloadApplet(pAppId.c_str());
		ulRetCode = plugin->JHI_Plugin_SendCmdPkg(handle, blob);
	}

	// in case of applet overflow, try to perform shared session cleanup
	// using LRU algorithem and download the applet again.

	if( JHI_MAX_INSTALLED_APPLETS_REACHED == ulRetCode) 
	{
		// try to unload one applet that doesnt have an active session
		if (TryUnloadUnusedApplet())
		{
			//call the Download function using the plugin
			ulRetCode = plugin->JHI_Plugin_SendCmdPkg(handle, blob);
		}
	}

	if (ulRetCode != JHI_SUCCESS)
	{
		//if (ulRetCode != JHI_MAX_INSTALLED_APPLETS_REACHED && ulRetCode != JHI_INSTALL_FAILURE_SESSIONS_EXISTS)
		//	ulRetCode = JHI_INSTALL_FAILED; // return a general error since we cannot return just the error from the last download try

		TRACE1("failed to install applet from DALP, error code: 0x%x\n", ulRetCode);
		goto errorRemoveApplet;
	}

	// Mark the Applet as Installed
	if (!Applets.completeInstall(pAppId, true))
	{
		ulRetCode = JHI_INTERNAL_ERROR;
		goto errorDeleteFromFW;
	}

	ulRetCode = JHI_SUCCESS;
	goto cleanup;


errorDeleteFromFW:

	//call the delete function using the plugin
	plugin->JHI_Plugin_UnloadApplet(pAppId.c_str());

errorRemoveApplet:

	// delete pending file from repository
	_wremove(AppletsManager::Instance().getPendingFileName(pAppId, true).c_str());

	// remove applet record if it is in pending state
	if (Applets.getAppletState(pAppId) == PENDING_INSTALL)
	{
		Applets.remove(pAppId);
	}

cleanup:

	return ulRetCode ;
}

JHI_RET_I jhis_send_cmd_pkg (const SD_SESSION_HANDLE handle, vector<uint8_t>& blob)
{
	UINT32 ulRetCode = TEE_STATUS_INTERNAL_ERROR;
	PACKAGE_INFO pkgInfo;
	string uuid;
	VM_Plugin_interface* plugin = NULL;

	if (blob.size() == 0)
	{
		ulRetCode = TEE_STATUS_INVALID_PARAMS;
		goto cleanup;
	}

	if ( (!GlobalsManager::Instance().getPluginTable(&plugin)) || (plugin == NULL) )
	{
		// probably a reset
		ulRetCode = TEE_STATUS_NO_FW_CONNECTION;
		goto cleanup;
	}

	ulRetCode = plugin->JHI_Plugin_ParsePackage(&blob[0], (uint32_t)blob.size(), pkgInfo);
	if (ulRetCode != TEE_STATUS_SUCCESS)
	{
		goto cleanup;
	}

	//block any command that is involved with the Spooler Applet UUID

#ifdef _WIN32
	if (_stricmp((char*)pkgInfo.uuid, SPOOLER_APPLET_UUID) == 0)
#else
	if (strcmp ((const char *)pkgInfo.uuid, SPOOLER_APPLET_UUID) == 0)
#endif
	{
		TRACE0("illegal use of spooler applet UUID\n");
		ulRetCode = TEE_STATUS_INVALID_UUID;
		goto cleanup;
	}

    uuid = strToUppercase(string(reinterpret_cast <char*>(pkgInfo.uuid)));

	switch (pkgInfo.packageType)
	{
	case AC_INSTALL_SD:
	case AC_UNINSTALL_SD:
		ulRetCode = plugin->JHI_Plugin_SendCmdPkg(handle, blob);
		break;
	case AC_INSTALL_NTA:
	case AC_UNINSTALL_NTA:
	case AC_INSTALL_JTA_PROP:
		ulRetCode = TEE_STATUS_UNSUPPORTED_PLATFORM;
		break;
	case AC_INSTALL_JTA:
		ulRetCode = cmd_pkg_install_jta (uuid, handle, blob);
		ulRetCode = jhiErrorToTeeError(ulRetCode);
		break;
	case AC_UNINSTALL_JTA:
		ulRetCode = jhis_uninstall(uuid.c_str(), handle, &blob);
		ulRetCode = jhiErrorToTeeError(ulRetCode);
		break;
	case AC_UPDATE_SVL:
		ulRetCode = plugin->JHI_Plugin_SendCmdPkg(handle, blob);
		break;
	case AC_CMD_INVALID:
	default:
		ulRetCode = TEE_STATUS_INVALID_PARAMS;
	}

cleanup:
	return ulRetCode ;
}