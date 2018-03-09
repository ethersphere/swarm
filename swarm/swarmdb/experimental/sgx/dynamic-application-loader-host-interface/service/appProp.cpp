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
**    @file appProp.c
**
**    @brief  Defines functions for the JHI interface to get applet property
**
**    @author Niveditha Sundaram
**
********************************************************************************
*/
#include "jhi_service.h"
#include "misc.h"
#include "reg.h"
#include "dbg.h"
#include "string_s.h"
#include "AppletsManager.h"

#include "SessionsManager.h"
#ifdef _WIN32
#include <process.h>
#endif

using namespace intel_dal;

//-------------------------------------------------------------------------------
// Function: jhis_get_applet_property
//		    Used to query applet version info from JoM
// IN		: pAppId - incoming AppId of package 
// IN/OUT   : pCommBuffer - i/o buffer used for input/output
// RETURN	: JHI_RET - success or any failure returns
//-------------------------------------------------------------------------------
//1.	On receiving a GetAppletProperty command for an applet, look up APPID 
//		in app table to see if app is present
//		a.	If app is not present, then proceed to download applet by looking for
//			APPID.pack in the location specified by registry. 
//			i.	If download fails for some reason, or file not found, then return 
//				appropriate error.
//			ii.	If success, create app table entry
//2.    Compare the applet property against valid properties and block invalid ones.
//3.	If app is present, then call TL query API
//		a.	Call SMManagerGetServiceProperty to get corresponding service property
//		b.	On successful return, 
//			i.	free the response buffer using SMFree
//			ii.	Copy the response to the Rx buffer, after ensuring lengths are 
//				sufficient. If not, return JHI_INSUFFICIENT_BUFFER
//		c.	On failure
//			i.	Return corresponding error code
//-------------------------------------------------------------------------------

// compare the applet property request against supported 
// property values
bool isSupportedProperty(JVM_COMM_BUFFER* pCommBuffer)
{
	char* AppProperty = (char*) pCommBuffer->TxBuf->buffer;

	int AppPropertyLength;

	if (pCommBuffer->TxBuf->length < 1)
		return false;

	AppPropertyLength = pCommBuffer->TxBuf->length;

	if (AppProperty == NULL)
		return false;

	if (AppProperty[AppPropertyLength-1] != '\0')
		return false;


	if (strcmp(AppProperty, "applet.name") == 0)
		return true;
	if (strcmp(AppProperty, "applet.vendor") == 0)
		return true;
	if (strcmp(AppProperty, "applet.description") == 0)
		return true;
	if (strcmp(AppProperty, "applet.version") == 0)
		return true;
	if (strcmp(AppProperty, "security.version") == 0)
		return true;
	if (strcmp(AppProperty, "applet.flash.quota") == 0)
		return true;
	if (strcmp(AppProperty, "applet.debug.enable") == 0)
		return true;
	if (strcmp(AppProperty, "applet.shared.session.support") == 0)
		return true;
	if (strcmp(AppProperty, "applet.platform") == 0)
		return true;

	return false;
}

JHI_RET_I
	jhis_get_applet_property (
	const char* pAppId,
	JVM_COMM_BUFFER* pCommBuffer
	)
{
	UINT32 ulRetCode = JHI_INTERNAL_ERROR; 
	JHI_APPLET_STATUS appStatus;
	VM_Plugin_interface* plugin = NULL;

	AppletsManager&  Applets = AppletsManager::Instance();

	string AppPropertyStr;
	JVM_COMM_BUFFER requestBuffers;
	requestBuffers.TxBuf->buffer = NULL;
	requestBuffers.TxBuf->length = 0;
	requestBuffers.RxBuf->buffer = NULL;
	requestBuffers.RxBuf->length = 0;

	// Get applet property requires an open session over BHv2
	JHI_VM_TYPE vmType = GlobalsManager::Instance().getVmType();
	bool sessionCreated = false;
	SessionsManager& sessionsManager = SessionsManager::Instance();
	JHI_SESSION_ID session_id = {0};
	JHI_PROCESS_INFO processInfo;

	if ( ! (pAppId && pCommBuffer) )
		return ulRetCode ;

	if (!isSupportedProperty(pCommBuffer))
	{
		ulRetCode = JHI_APPLET_PROPERTY_NOT_SUPPORTED;
		goto error;
	}



	//get app status
	appStatus = Applets.getAppletState(pAppId);

	if ( ! ( (appStatus >= 0) && (appStatus < MAX_APP_STATES) ) )
	{
		TRACE2 ("AppState incorrect: %d for appid: %s \n", appStatus, pAppId);
		return ulRetCode ;
	}

	if (NOT_INSTALLED == appStatus)
	{
		// try to install the applet if the pack file is in our repository
		FILESTRING filename;
		bool isAcp;
		if (Applets.appletExistInRepository(pAppId,&filename, isAcp))
		{
			// applet is not installed but applet file exists in the repository, try to install it.
			ulRetCode = jhis_install(pAppId,filename.c_str(), true, isAcp);

			if (ulRetCode != JHI_SUCCESS)
			{
				ulRetCode = JHI_APPLET_NOT_INSTALLED;
				goto error;
			}
		}
		else
		{
			ulRetCode = JHI_APPLET_NOT_INSTALLED;
			goto error;
		}
	}


	// Get applet property requires an open session over BHv2
	if (vmType == JHI_VM_TYPE_BEIHAI_V2)
	{
		//check if there's an open session, if not, open one
		if (!sessionsManager.hasLiveSessions(pAppId))
		{
			TRACE0("Get applet property was callled for and applet without an open session. A session needs to be created.");
			DATA_BUFFER tmpBuffer;
			tmpBuffer.buffer = NULL;
			tmpBuffer.length = 0;
#ifdef _WIN32
			processInfo.pid = _getpid();
#else
			processInfo.pid = getpid();
#endif

			TRACE1("Creating session for %s", pAppId);
			ulRetCode = jhis_create_session(pAppId, &session_id, 0, &tmpBuffer, &processInfo);

			if (ulRetCode != JHI_SUCCESS)
			{
				ulRetCode = JHI_APPLET_NOT_INSTALLED;
				goto error;
			}
			sessionCreated = true;
		}
	}

	requestBuffers.TxBuf->buffer = JHI_ALLOC(pCommBuffer->TxBuf->length);
	requestBuffers.RxBuf->buffer = JHI_ALLOC(pCommBuffer->RxBuf->length);

	if (requestBuffers.TxBuf->buffer == NULL || requestBuffers.RxBuf->buffer == NULL) {
		TRACE0("malloc of requestBuffers failed .");
		ulRetCode = JHI_INTERNAL_ERROR;
		goto error;
	}

	TRACE1("Applet property request: %s\n", pCommBuffer->TxBuf->buffer);

	AppPropertyStr = string((char*)pCommBuffer->TxBuf->buffer);

	strcpy_s((char*) requestBuffers.TxBuf->buffer,pCommBuffer->TxBuf->length,AppPropertyStr.c_str());
	requestBuffers.TxBuf->length = (uint32_t)AppPropertyStr.length() +1;

	requestBuffers.RxBuf->length = pCommBuffer->RxBuf->length - 1;
	memset(requestBuffers.RxBuf->buffer,0,pCommBuffer->RxBuf->length);

	if ( (!GlobalsManager::Instance().getPluginTable(&plugin)) || (plugin == NULL) )
	{
		// probably a reset
		ulRetCode = JHI_NO_CONNECTION_TO_FIRMWARE;		
	}
	else
	{
		ulRetCode = plugin->JHI_Plugin_GetAppletProperty(pAppId, &requestBuffers);
	}

	if (JHI_SUCCESS != ulRetCode)
	{
		TRACE0 ("JHI unable to get applet property\n");

		if (ulRetCode == JHI_INSUFFICIENT_BUFFER)
		{
			pCommBuffer->RxBuf->length = requestBuffers.RxBuf->length + 1; // update needed buffer size in case of short buffer
		}
	}
	else
	{
		string Property = string((char*)requestBuffers.RxBuf->buffer);
		strcpy_s((char*)pCommBuffer->RxBuf->buffer, pCommBuffer->RxBuf->length, Property.c_str());

		pCommBuffer->RxBuf->length = requestBuffers.RxBuf->length + 1;

		TRACE1("Applet property result: \"%s\"\n", pCommBuffer->RxBuf->buffer);
	}


error:

	if (requestBuffers.TxBuf->buffer != NULL)
	{
		JHI_DEALLOC(requestBuffers.TxBuf->buffer);
		requestBuffers.TxBuf->buffer = NULL;
	}

	if (requestBuffers.RxBuf->buffer != NULL)
	{
		JHI_DEALLOC(requestBuffers.RxBuf->buffer);
		requestBuffers.RxBuf->buffer = NULL;
	}

	// Get applet property requires an open session over BHv2
	if (vmType == JHI_VM_TYPE_BEIHAI_V2)
	{
		if (sessionCreated)
		{
			TRACE1("Closing session for %s", pAppId);
			jhis_close_session(&session_id, &processInfo, false, true);
		}
	}

	return ulRetCode;

}

#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)

JHI_RET_I
	jhis_get_loaded_applets(JHI_LOADED_APPLET_GUIDS* loadedAppletsList)
{
	UINT32 retCode = JHI_INTERNAL_ERROR;
	AppletsManager&  Applets = AppletsManager::Instance();
	list<string>::const_iterator it;
	list<string> loaded_applets;
	Applets.getLoadedAppletsList(loaded_applets);
	loadedAppletsList->loadedAppletsCount = 0;
	if (loaded_applets.size() > 0)
	{
		loadedAppletsList->loadedAppletsCount = loaded_applets.size();
		loadedAppletsList->appsGUIDs = JHI_ALLOC_T_ARRAY<char*>(loaded_applets.size());
		if (loadedAppletsList->appsGUIDs == NULL)
		{
			retCode = JHI_MEM_ALLOC_FAIL;
			goto error;
		}
		int i = 0;
		for (it = loaded_applets.begin(); it != loaded_applets.end(); it++ )
		{
			loadedAppletsList->appsGUIDs[i] = (char*) JHI_ALLOC(LEN_APP_ID + 1);
			if (loadedAppletsList->appsGUIDs[i] == NULL)
			{
				retCode = JHI_MEM_ALLOC_FAIL;
				goto error;
			}
			memset(loadedAppletsList->appsGUIDs[i], 0, LEN_APP_ID + 1);
			strcpy_s(loadedAppletsList->appsGUIDs[i], LEN_APP_ID + 1, (*it).c_str());
			++i;
		}
	}
	retCode = JHI_SUCCESS;

error:
	if (JHI_MEM_ALLOC_FAIL == retCode)
		freeLoadedAppletsList(loadedAppletsList);

	return retCode;
}

#endif