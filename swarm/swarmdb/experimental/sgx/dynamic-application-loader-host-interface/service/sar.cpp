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
**    @file sar.c
**
**    @brief  Defines functions for the JHI SendAndRecv interface
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

using namespace intel_dal;

//-------------------------------------------------------------------------------
// Function: jhis_txrx_raw
//		    Used to send/recv data to/from JoM
// IN		: pAppId - incoming AppId of package 
// IN		: nCommandId - Command ID to process the i/o buffer.
// IN/OUT   : pCommBuffer - i/o buffer used for input/output
// RETURN	: JHI_RET - success or any failure returns
//-------------------------------------------------------------------------------
//1.	On receiving a SAR command for an applet, look up APPID in app table to 
//		see if app is present
//		a.	If app is not present, then proceed to download applet by looking for
//			APPID.pack in the location specified by registry. 
//			i.	If download fails for some reason, or file not found, then return
//				appropriate error.
//			ii.	If success, create app table entry
//2.	If app is present, and is either INSTALLED or ACTIVE, then call TL 
//		SendAndRecv
//		a.	Check to see if session table has a previous session entry with handle
//		b.	If no, create a new session and update session table with corresponding
//			APPID and handle.
//		c.	Use the session handle to call SMStubPrepareInvokeOperation to allocate
//			resources for TL SAR.
//		d.	Use SMStubEncoderWriteUint8Array to write to the encoder
//		e.	Call SMStubPerformOperation to send data to the other side
//		f.	Call SMStubDecoderReadUint8Array to read the response back from the JoM.
//			i.	On successful return from this function, the calling buffer will be 
//	`			freed by calling SMFree. On failure, the SMAPI is assumed to perform 
//				cleanup.
//		g.	Call SMStubReleaseOperation to release all related resources
//		h.	Check to see if the response length is sufficient for the Rx buffer 
//			lengths specified by the host app. If not, only the correct response 
//			length is copied, and an error  JHI_INSUFFICENT_BUFFER is returned.
//3.	If SAR returns an error:
//		a.	If APPLET_FATAL: Close session and remove applet from JoM. Local cache 
//			of applet is not deleted
//		b.	For all other errors, return appropriate error code
//4.	If SAR successful:
//		a.	Add timestamp to the app table
//		b.	If session entry created in the session table, update app table entry 
//			to ACTIVE
//		c.	On return back to DLL, the response length is again checked to ensure 
//			that the Rx buffer lengths are sufficient before return to the calling 
//			host app. 
//-------------------------------------------------------------------------------

UINT32
jhis_txrx_raw( 
	JHI_SESSION_ID*		  pSessionID,
	INT32			      nCommandId,
	JVM_COMM_BUFFER*      pCommBuffer,
	INT32*				  pResponseCode
)
{
	VM_SESSION_HANDLE VMSessionHandle;

	SessionsManager& Sessions = SessionsManager::Instance();
	

	UINT32 ulRetCode = JHI_INTERNAL_ERROR;

	if ( ! (pSessionID && pCommBuffer && pResponseCode) )
		return ulRetCode ;

	do
	{
		// check that the session exists
		JHI_SESSION_INFO info;
		Sessions.getSessionInfo(*pSessionID,&info);

		if (info.state == JHI_SESSION_STATE_NOT_EXISTS)
			return JHI_INVALID_SESSION_HANDLE;

		if (!Sessions.GetSessionLock(*pSessionID))
		{
			ulRetCode = JHI_INVALID_SESSION_HANDLE;
			break;
		}

		// get the VM session handle
		if (!Sessions.getVMSessionHandle(*pSessionID,&VMSessionHandle))
		{
			ulRetCode = JHI_INTERNAL_ERROR;
			break;
		}

		VM_Plugin_interface* plugin = NULL;
		if ( (!GlobalsManager::Instance().getPluginTable(&plugin)) || (plugin == NULL) )
		{
			// probably a reset
			ulRetCode = JHI_NO_CONNECTION_TO_FIRMWARE;	
			break;
		}

		ulRetCode = plugin->JHI_Plugin_SendAndRecv(VMSessionHandle, 
											nCommandId,
											pCommBuffer,
											pResponseCode);

		if (JHI_APPLET_FATAL == ulRetCode) //if this is bad applet
		{
			// remove the session record
			if (jhis_close_session(pSessionID,NULL,false,false) != JHI_SUCCESS)
			{
				TRACE0("Failed to remove crashed session.");
			}
			
			// notify the application the applet has crashed
			ulRetCode = JHI_APPLET_FATAL;
			break;
		}

	}
	while (0);

	Sessions.ReleaseSessionLock(*pSessionID);

	return ulRetCode ;
}