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
**    @file jhi.cpp
**
**    @brief  Defines exported interfaces for JHI.DLL
**
**    @author Elad Dabool
**
********************************************************************************
*/

#ifdef _WIN32
#include <windows.h>
#include <process.h>
#include <io.h>
#include "ServiceManager.h"
#else
#include <uuid/uuid.h>
#endif //_WIN32

#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <list>

#include "jhi.h"
#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
#include "jhi_sdk.h"
#endif

#include "jhi_i.h"

#include "misc.h"
#include "reg.h"
#include "Locker.h"


#include "CommandInvoker.h"
#include "string_s.h"

using std::list;

using namespace intel_dal;

JHI_I_HANDLE* appHandle = NULL;		 // a handle that is passed by the application when calling any jhi API function.
Locker appHandleLock;				 // a lock for syncronization of appHandle

int g_logFlag = 0;

// compare the appHandle to the handle the application is using in it calls
// returns true if the handle valid, false otherwise
// note: cannot assume that the handle will remain valid afterwards, use the appHandleLock to assure that.
bool ValidateJHIhandle(JHI_HANDLE handle)
{
	if (handle == NULL)
		return false;

	if ((JHI_I_HANDLE*)handle == appHandle)
	{
		return true;
	}
	return false;
}

// search for the SessionHandle pointer in the session list
// note: this call should be performed only after aquiring a appHandleLock for thread safty
bool SessionHandleValid(JHI_I_SESSION_HANDLE* SessionHandle)
{
	list<JHI_I_SESSION_HANDLE*>* SessionsList;
	list<JHI_I_SESSION_HANDLE*>::iterator it;

	bool valid = false;

	if (appHandle == NULL)
		return false;

	if (appHandle->SessionsList == NULL)
		return false;

	if (SessionHandle == NULL)
		return false;

	SessionsList = (list<JHI_I_SESSION_HANDLE*> *) appHandle->SessionsList;

	// search the pointer in the session list
	for ( it = SessionsList->begin(); it != SessionsList->end(); ++it )
	{
		if ((*it) == SessionHandle)
		{
			valid = true;
			break;
		}
	}
	
	return valid;
}

// add a SessionHandle pointer to the sessions list
// note: this call should be performed only after aquiring a appHandleLock for thread safty
bool addSessionHandle(JHI_I_SESSION_HANDLE* SessionHandle)
{
	list<JHI_I_SESSION_HANDLE*>* SessionsList;

	if (appHandle == NULL)
		return false;

	if (appHandle->SessionsList == NULL)
		return false;

	if (SessionHandle == NULL)
		return false;

	SessionsList = (list<JHI_I_SESSION_HANDLE*> *) appHandle->SessionsList;
	SessionsList->push_back(SessionHandle);

	return true;
}


// remove a sessionHandle form the session list
// return true if removed, false otherwise
// note: this call should be performed only after aquiring a appHandleLock for thread safty
bool removeSessionHandle(JHI_I_SESSION_HANDLE* SessionHandle)
{
	list<JHI_I_SESSION_HANDLE*>* SessionsList;

	if (!SessionHandleValid(SessionHandle))
		return false;

	SessionsList = (list<JHI_I_SESSION_HANDLE*> *) appHandle->SessionsList;
	SessionsList->remove(SessionHandle);

	return true;
}

// most jhi API's just need the session ID in order to work, this function
// retreives the sessionID from a session handle. it ensures that the session handle is valid 
// before extracting the session ID. the function returns true when the session ID extracted successfuly, false otherwise.
// the function assumes that the sessionID parameter points to a valid session id struct.
bool getSessionID(JHI_SESSION_HANDLE SessionHandle,JHI_SESSION_ID* sessionID)
{
	bool status = false;
	JHI_I_SESSION_HANDLE* iSessionHandle = (JHI_I_SESSION_HANDLE*) SessionHandle;


	appHandleLock.Lock();

	if (SessionHandleValid(iSessionHandle))
	{
		*sessionID = iSessionHandle->sessionID;
		status = true;
	}

	appHandleLock.UnLock();

	return status;
}

// signal session event thread to close itself and free its allocated resources
// note: this call should be performed only after aquiring a appHandleLock for thread safty
void closeSessionEventThread(JHI_I_SESSION_HANDLE* iSessionHandle)
{
	// signal the thread to wake and close itself
	TRACE0("Closing thread and event handles..");

	if (!SessionHandleValid(iSessionHandle))
		return;

	iSessionHandle->callback = NULL;

	// close handle to the thread and the event
	if (iSessionHandle->eventHandle != NULL && iSessionHandle->eventHandle->is_created())
	{
		if (iSessionHandle->threadNeedToEnd != NULL)
			*(iSessionHandle->threadNeedToEnd) = 1;
#ifdef _WIN32
		iSessionHandle->eventHandle->set(); //callback is eventListenerThread
#else
		iSessionHandle->eventHandle->close();
		TRACE0("JHIDLL: close event handler\n");
#endif //WIN32
		iSessionHandle->eventHandle = 0;
	}

	if (iSessionHandle->threadHandle != 0)
	{
#ifdef _WIN32
		CloseHandle(iSessionHandle->threadHandle);
#else
	//	close(iSessionHandle->threadHandle);
#endif //WIN32
		iSessionHandle->threadHandle = 0;
	}
}

#ifdef __ANDROID__
void clearDeadOwnersSessions ()
{
	list<JHI_I_SESSION_HANDLE*>* SessionsList;
	list<JHI_I_SESSION_HANDLE*>::iterator it;
	CommandInvoker cInvoker;
	if (appHandle == NULL)
		return;
	if (appHandle->SessionsList == NULL)
		return;
	SessionsList = (list<JHI_I_SESSION_HANDLE*> *) appHandle->SessionsList;
	appHandleLock.Lock();
	for ( it = SessionsList->begin(); it != SessionsList->end(); )
	{
		if ((*it) == NULL) {
			SessionsList->erase(it);
			continue;
		}
		if ((*it)->processInfo.pid == appHandle->processInfo.pid) {
			++it;
			continue;
		}
		if (isProcessDead((*it)->processInfo.pid, (*it)->processInfo.creationTime) == false) {
			++it;
			continue;
		}
		if ((*it)->eventHandle != NULL && (*it)->eventHandle->is_created())
		{
			TRACE0 ("JHIDLL: removing dead session event registration\n");
			closeSessionEventThread((*it));
		}
		TRACE1 ("JHIDLL:close dead session %x\n", (*it)->sessionID);
		if (JHI_SUCCESS != cInvoker.JhisCloseSession(&((*it)->sessionID), &(appHandle->processInfo), false)) {
			TRACE0 ("JHIDLL: Can't remove Dead Session from the daemon/FW list\n");
		}
		JHI_DEALLOC((*it));
		it = SessionsList->erase(it);
		TRACE0 ("JHIDLL: Dead Session Close Complete\n");
	}
	appHandleLock.UnLock();
}

void clearDestroyedSessions (int DestroyedAppPid)
{
	list<JHI_I_SESSION_HANDLE*>* SessionsList;
	list<JHI_I_SESSION_HANDLE*>::iterator it;
	CommandInvoker cInvoker;
	TRACE2 ("Process to clear sessions: pid %d iter = 0 appHandle %p\n", DestroyedAppPid, appHandle);
	int iter = 0;
	if (appHandle == NULL)
		return;
	TRACE1 ("iter cnt %d\n", iter);
	if (appHandle->SessionsList == NULL)
		return;
	TRACE1 ("iter cnt %d\n", iter);
	SessionsList = (list<JHI_I_SESSION_HANDLE*> *) appHandle->SessionsList;
	appHandleLock.Lock();
	for ( it = SessionsList->begin(); it != SessionsList->end(); )
	{
		TRACE1 ("iter cnt %d\n", iter);

		if ((*it) == NULL) {
			SessionsList->erase(it);
			continue;
		}
		TRACE1 ("Current session pid %d\n", (*it)->processInfo.pid);
		if ((*it)->processInfo.pid != (UINT32)DestroyedAppPid) {
			++it;
			continue;
		}
		if ((*it)->eventHandle != NULL && (*it)->eventHandle->is_created())
		{
			TRACE0 ("JHIDLL: removing destroyed session event registration\n");
			closeSessionEventThread((*it));
		}
		TRACE1 ("JHIDLL:close destroyed session %x\n", (*it)->sessionID);
		if (JHI_SUCCESS != cInvoker.JhisCloseSession(&((*it)->sessionID), &(appHandle->processInfo), false)) {
			TRACE0 ("JHIDLL: Can't remove Destroyed Session from the daemon/FW list\n");
		}

		JHI_DEALLOC((*it));
		it = SessionsList->erase(it);
		TRACE0 ("JHIDLL: Destroyed Session Close Complete\n");
	}
	appHandleLock.UnLock();
}
#endif //__ANDROID__

//------------------------------------------------------------------------------
// Function: JHI_Initialize
//		  First interface to be called by IHA or any external vendor
//        to initialize data structs and set up COMMS with JoM
// IN	: context (not used)
// IN	: flags (not used)
// OUT	: ppHandle - returns a handle back to the caller to be used in future
// RETURN: JHI_RET - success or any failure returns
//------------------------------------------------------------------------------

JHI_RET
JHI_Initialize (
	OUT JHI_HANDLE* ppHandle,
	IN  PVOID       context,
	IN  UINT32      flags
)
{          
	UINT32           rc;

	CommandInvoker cInvoker;

	rc = JHI_INTERNAL_ERROR ;

	if (ppHandle == NULL)
		return JHI_INVALID_HANDLE;

	//Get log flag from registry if present
	JhiQueryLogLevelFromRegistry (&g_jhiLogLevel);

	// If debug prints are enabled, inform the user
	if (g_jhiLogLevel == JHI_LOG_LEVEL_DEBUG)
		TRACE0("JHI client - debug trace and release prints are enabled\n");

	appHandleLock.Lock();

	do
	{
		if (appHandle != NULL) // init was done before. return the existing appHandle.
		{
			*ppHandle = appHandle;
			appHandle->ReferenceCount++; // increase the reference count by one
			rc = JHI_SUCCESS;
			break;
		}

#ifdef _WIN32
		// verify the service is started before connecting it
		startJHIService();
#endif
#ifdef __ANDROID__
		if (isServiceRunning())
			TRACE0 ("JHI Service Running");
		else {
			TRACE0 ("JHI Service Stopped");
			rc = RestartJhiService ();
			TRACE1 ("JHI Service Restart %s\n", (rc == JHI_SUCCESS ? "SUCCES" : "ERROR"));
			sleep (1);
		}
#endif

		rc = cInvoker.JhisInit();
		
		if (JHI_SUCCESS != rc)
		{
			LOG0("JHI init at server side failed");
			break;
		}

		//Allocate jhi handle for internal operations
		appHandle = JHI_ALLOC_T(JHI_I_HANDLE);
		if(!appHandle) 
		{
			LOG2 ("%s: Malloc failure - line: %d\n", __FUNCTION__, __LINE__ );
			rc = JHI_INTERNAL_ERROR;
			break;
		}

		// set Reference count to 1;
		appHandle->ReferenceCount = 1;

		// allocate memory for the Sessions List		
		appHandle->SessionsList = JHI_ALLOC_T(list<JHI_I_SESSION_HANDLE*>);
		if (appHandle->SessionsList == NULL)
		{
			LOG2 ("%s: Malloc failure - line: %d\n", __FUNCTION__, __LINE__ );
			JHI_DEALLOC_T(appHandle);
			appHandle = NULL;
			rc = JHI_INTERNAL_ERROR;
			break;
		}

		// set process ID and timestamp
#ifdef _WIN32
		appHandle->processInfo.pid = _getpid();

		TRACE1("current process pid: %d\n",appHandle->processInfo.pid);
		
		FILETIME unusedVar;
		if (GetProcessTimes(GetCurrentProcess(),&(appHandle->processInfo.creationTime),&unusedVar,&unusedVar,&unusedVar) == FALSE)
		{
			TRACE1("Error: failed to get process creation time, windows error: %d\n",GetLastError());
			rc = JHI_INTERNAL_ERROR;
			break;
		}
#else //!WIN32
		appHandle->processInfo.pid = getpid();

		TRACE1("current process pid: %d\n", appHandle->processInfo.pid);

		if (JHI_SUCCESS != getProcStartTime(appHandle->processInfo.pid, appHandle->processInfo.creationTime))
		{
			LOG0("Error: failed to get process creation time\n");
			rc = JHI_INTERNAL_ERROR;
			break;
		}
#endif //WIN32

		*ppHandle = appHandle;
		
	
	} while(0);


	if (rc != JHI_SUCCESS)
	{
		LOG1("JHI init failed. Status: %d\n", rc);
		JHI_DEALLOC_T(appHandle);
		appHandle = NULL;
	}

	appHandleLock.UnLock();

	return rc;
}


//------------------------------------------------------------------------------
// Function: JHI_Deinit
//		  Interface to be called by IHA or any external vendor
//        to de-initialize all data structs. Currently Deinit not used.
//        used to dealloc handle
// IN	: handle 
// OUT	: none
// RETURN: JHI_RET - success or any failure returns
//------------------------------------------------------------------------------

JHI_RET 
JHI_Deinit(IN JHI_HANDLE handle)
{
	list<JHI_I_SESSION_HANDLE*>::iterator it;
	list<JHI_I_SESSION_HANDLE*>* SessionsList;

	if(!ValidateJHIhandle(handle))
		return JHI_INVALID_HANDLE;

	appHandleLock.Lock();

	if (appHandle != NULL)
	{
		do
		{
			appHandle->ReferenceCount--;

			if (appHandle->ReferenceCount > 0) 
				break;	// do not perform deinit

			// remove the session List
			if (appHandle->SessionsList != NULL)
			{

				SessionsList = (list<JHI_I_SESSION_HANDLE*> *) appHandle->SessionsList;

				// free all allocated session handles and event threads if exists
				for ( it = SessionsList->begin(); it != SessionsList->end(); ++it )
				{
					// remove events registration
					if ((*it)->eventHandle != NULL && (*it)->eventHandle->is_created())
					{
						closeSessionEventThread(*it);
					}
				
					// remove allocated session struct
					JHI_DEALLOC(*it);
					*it = NULL;
				}

				// clear list items
				((list<JHI_I_SESSION_HANDLE>*) appHandle->SessionsList)->clear();
				JHI_DEALLOC_T(appHandle->SessionsList); 
				appHandle->SessionsList = NULL;
			}

			JHI_DEALLOC_T(appHandle); 
			appHandle = NULL;
		}
		while(0);
	}

	appHandleLock.UnLock();

	return JHI_SUCCESS;
}


static JHI_RET 
JHI_CreateSession_handler(
	IN const JHI_HANDLE handle, 
	IN const char* AppId,
	IN UINT32 flags,
	IN DATA_BUFFER* initBuffer,
	OUT JHI_SESSION_HANDLE* pSessionHandle
)
{
	UINT32              rc = JHI_INTERNAL_ERROR;
	UINT8			    ucAppId[LEN_APP_ID+1] ; // Local copy
	JHI_I_SESSION_HANDLE* pHandle;
	DATA_BUFFER tmpBuffer;

	CommandInvoker cInvoker;
	
	if(!ValidateJHIhandle(handle))
		return JHI_INVALID_HANDLE;

	if (pSessionHandle == NULL)
		return JHI_INVALID_SESSION_HANDLE;

	if (initBuffer == NULL) // allow passing NULL buffer data
	{
		tmpBuffer.buffer = NULL;
		tmpBuffer.length = 0;
		initBuffer = &tmpBuffer;
	}

	if ( !(AppId && (strlen(AppId) == LEN_APP_ID) &&
		(JhiUtilUUID_Validate(AppId, ucAppId) == JHI_SUCCESS)) )
	{
		TRACE0 ("Either Appname is bad or illegal length ..\n");
		return JHI_INVALID_APPLET_GUID;
	}

	// Validate the incoming values
	if ((initBuffer->length > 0) && (initBuffer->buffer == NULL))
	{
		TRACE0 ("Illegal argument supplied.. Check the input values..\n");
		return JHI_INVALID_INIT_BUFFER;
	}
	
	if (initBuffer->length > JHI_BUFFER_MAX)
	{
		TRACE0 ("init buffer exceeds JHI_BUFFER_MAX limit\n");
		return JHI_INVALID_BUFFER_SIZE;
	}

	do 
	{
		// allocate memory for the session handle
		pHandle = JHI_ALLOC_T(JHI_I_SESSION_HANDLE);
		if(!pHandle) 
		{
			LOG2 ("%s: Malloc failure - line: %d\n", __FUNCTION__, __LINE__ );
			rc = JHI_INTERNAL_ERROR;
			break;
		}

		// init JHI_I_SESSION_HANDLE struct
		pHandle->sessionFlags = flags;
		pHandle->eventHandle = NULL;
		pHandle->threadHandle = 0;
		pHandle->callback = NULL;
		pHandle->threadNeedToEnd = NULL;

		// retrieve process info from the appHandle.
		appHandleLock.Lock();

		if (appHandle != NULL)
			pHandle->processInfo = appHandle->processInfo;
		else
		{
			appHandleLock.UnLock();
			rc = JHI_INVALID_HANDLE;
			JHI_DEALLOC_T(pHandle);
			pHandle = NULL;
			break;
		}

		appHandleLock.UnLock();


		// call for create session at the service
		rc  = cInvoker.JhisCreateSession((char *)ucAppId,&(pHandle->sessionID),flags,initBuffer,&(pHandle->processInfo));

		if (JHI_SUCCESS != rc )
		{
			// release allocated memory
			JHI_DEALLOC_T(pHandle);
			pHandle = NULL;
			TRACE1 ("JHDLL: Session creation failure, retcode: %08x\n", rc);
			break;
		}
		else
		{
			
			// add the session pointer to the sessions list
			appHandleLock.Lock();
			
			if (addSessionHandle(pHandle))
			{
				*pSessionHandle = pHandle;
			}
			else
			{
				// only reason for failure is that JHI_Deinit occured
				JHI_DEALLOC_T(pHandle);
				pHandle = NULL;
				rc = JHI_INVALID_HANDLE;
			}

			appHandleLock.UnLock();

			TRACE0 ("JHIDLL: Session Creation Complete\n");
		}

	} while (0);

	return rc;
}

#ifdef __ANDROID__
static JHI_RET
JHI_CreateSessionProcess_handler(
	IN const JHI_HANDLE handle,
	IN const char* AppId,
	IN int SessionPid,
	IN UINT32 flags,
	IN DATA_BUFFER* initBuffer,
	OUT JHI_SESSION_HANDLE* pSessionHandle
)
{
	UINT32              rc = JHI_INTERNAL_ERROR;
	UINT8			    ucAppId[LEN_APP_ID+1] ; // Local copy
	JHI_I_SESSION_HANDLE* pHandle;
	DATA_BUFFER tmpBuffer;

	CommandInvoker cInvoker;

	if(!ValidateJHIhandle(handle))
		return JHI_INVALID_HANDLE;
	
	clearDeadOwnersSessions();

	if (pSessionHandle == NULL)
		return JHI_INVALID_SESSION_HANDLE;

	if (initBuffer == NULL) // allow passing NULL buffer data
	{
		tmpBuffer.buffer = NULL;
		tmpBuffer.length = 0;
		initBuffer = &tmpBuffer;
	}

	if ( !(AppId && (strlen(AppId) == LEN_APP_ID) &&
		(JhiUtilUUID_Validate(AppId, ucAppId) == JHI_SUCCESS)) )
	{
		TRACE0 ("Either Appname is bad or illegal length ..\n");
		return JHI_INVALID_APPLET_GUID;
	}

	// Validate the incoming values
	if ((initBuffer->length > 0) && (initBuffer->buffer == NULL))
	{
		TRACE0 ("Illegal argument supplied.. Check the input values..\n");
		return JHI_INVALID_INIT_BUFFER;
	}

	if (initBuffer->length > JHI_BUFFER_MAX)
	{
		TRACE0 ("init buffer exceeds JHI_BUFFER_MAX limit\n");
		return JHI_INVALID_BUFFER_SIZE;
	}

	do
	{
		// allocate memory for the session handle
		pHandle = JHI_ALLOC_T(JHI_I_SESSION_HANDLE);
		if(!pHandle)
		{
			TRACE2 ("%s: Malloc failure - line: %d\n", __FUNCTION__, __LINE__ );
			rc = JHI_INTERNAL_ERROR;
			break;
		}

		// init JHI_I_SESSION_HANDLE struct
		pHandle->sessionFlags = flags;
		pHandle->eventHandle = NULL;
		pHandle->threadHandle = 0;
		pHandle->callback = NULL;
		pHandle->threadNeedToEnd = NULL;
		pHandle->processInfo.pid = SessionPid;
		/* FIXME: SEAndroid blocks us from getting process start time.
		meanwile ignore the risk of reuse of PID, and do not achive the start time */
		/*
		if (JHI_SUCCESS != getProcStartTime(pHandle->processInfo.pid, pHandle->processInfo.creationTime))
		{
			TRACE0("Error: failed to get Session process creation time\n");
			rc = JHI_INTERNAL_ERROR;
			break;
		}
		*/
		long long unsigned int data = 0;
		memcpy(&pHandle->processInfo.creationTime, &data, sizeof(data));

		// call for create session at the service
		rc  = cInvoker.JhisCreateSession((char *)ucAppId,&(pHandle->sessionID),flags,initBuffer,&(appHandle->processInfo));

		if (JHI_SUCCESS != rc )
		{
			// release allocated memory
			JHI_DEALLOC_T(pHandle);
			pHandle = NULL;
			TRACE1 ("JHDLL: Session creation failure, retcode: %08x\n", rc);
			break;
		}
		else
		{
			// add the session pointer to the sessions list
			appHandleLock.Lock();

			if (addSessionHandle(pHandle))
			{
				*pSessionHandle = pHandle;
			}
			else
			{
				// only reason for failure is that JHI_Deinit occured
				JHI_DEALLOC(pHandle);
				pHandle = NULL;
				rc = JHI_INVALID_HANDLE;
			}

			appHandleLock.UnLock();

			TRACE0 ("JHIDLL: Session Creation Complete\n");
		}
	} while (0);

	return rc;
}
#endif //__ANDROID__

//------------------------------------------------------------------------------
// Function: JHI_CreateSession
//		  Interface to be called by IHA or any external vendor
//        to an create a new session of installed applet.
// IN	: handle to the jhi
// IN	: pAppId - AppId of the package to be used
// IN   : flags - defines the session behaviour
// IN   : initBuffer - Some initialization data that will be passed to the applet onInit function.
// OUT  : ppSessionHandle - a handle to the created session 
// RETURN: JHI_RET - success or any failure returns
//------------------------------------------------------------------------------
JHI_RET 
JHI_CreateSession (
	IN const JHI_HANDLE handle, 
	IN const char* AppId,
	IN UINT32 flags,
	IN DATA_BUFFER* initBuffer,
	OUT JHI_SESSION_HANDLE* pSessionHandle
)
{
	return JHI_CreateSession_handler(handle,AppId,flags,initBuffer,pSessionHandle);
}

//------------------------------------------------------------------------------
// Function: JHI_CreateSessionProcess
//		  Interface to be called by IHA or any external vendor
//        to an create a new session of installed applet for defined process
// IN	: handle to the jhi
// IN	: pAppId - AppId of the package to be used
// IN   : SessionPid - PID of caller process
// IN   : flags - defines the session behaviour
// IN   : initBuffer - Some initialization data that will be passed to the applet onInit function.
// OUT  : ppSessionHandle - a handle to the created session
// RETURN: JHI_RET - success or any failure returns
//------------------------------------------------------------------------------

#ifdef __ANDROID__
JHI_RET
JHI_CreateSessionProcess (
	IN const JHI_HANDLE handle,
	IN const char* AppId,
	IN int SessionPid,
	IN UINT32 flags,
	IN DATA_BUFFER* initBuffer,
	OUT JHI_SESSION_HANDLE* pSessionHandle
)
{
	return JHI_CreateSessionProcess_handler(handle,AppId,SessionPid, flags,initBuffer,pSessionHandle);
}
#endif //__ANDROID__


//------------------------------------------------------------------------------
// Function: JHI_SendAndRecv2
//		  Interface to be called by IHA or any external vendor
//        to send/recv data to/from JoM
// IN	: handle
// IN	: pAppId - AppId of the package to be used in communication with JoM
// IN	: nCommandId - Command ID
// IN   : pComm - data buffer of type JVM_COMM_BUFFER to be used to send/recv 
//		  data. Tx buffer used to send data and Rx buffer to receive data
// RETURN: JHI_RET - success or any failure returns
//------------------------------------------------------------------------------
JHI_RET   
JHI_SendAndRecv2(
	IN JHI_HANDLE       handle,
	IN JHI_SESSION_HANDLE SessionHandle,
	IN INT32			nCommandId,
	INOUT JVM_COMM_BUFFER* pComm,
	OUT INT32* responseCode)
{
	UINT32 ulRetCode = JHI_INTERNAL_ERROR; 
	JHI_SESSION_ID sessionID;

	CommandInvoker cInvoker;

	// Validate the JHI handle
	if(!ValidateJHIhandle(handle))
		return JHI_INVALID_HANDLE;

	if (!getSessionID(SessionHandle,&sessionID))
		return JHI_INVALID_SESSION_HANDLE;

	if(pComm == NULL)
		return JHI_INVALID_COMM_BUFFER;

	
	// Validate the incoming values
	if ( ((pComm->TxBuf->length > 0) && (pComm->TxBuf->buffer == NULL) ) ||
		 ((pComm->RxBuf->length > 0) &&  (pComm->RxBuf->buffer == NULL)) )
	{
		TRACE0 ("Illegal argument supplied.. Check the input values..\n");
		return JHI_INVALID_COMM_BUFFER;
	}

	if ((pComm->TxBuf->length > JHI_BUFFER_MAX) || (pComm->RxBuf->length > JHI_BUFFER_MAX))
	{
		TRACE0 ("buffer sent exceeds JHI_BUFFER_MAX limit\n");
		return JHI_INVALID_BUFFER_SIZE;
	}

	TRACE0 ("calling SVC SAR..\n");
	ulRetCode = cInvoker.JhisSendAndRecv(&sessionID, nCommandId, (const uint8_t*) pComm->TxBuf->buffer, pComm->TxBuf->length, (uint8_t*)pComm->RxBuf->buffer, &pComm->RxBuf->length, responseCode);

	// if the session crashed we remove its allocated resources by calling closeSession
	if (ulRetCode == JHI_APPLET_FATAL || ulRetCode == JHI_INVALID_SESSION_HANDLE)
	{
		JHI_CloseSession(handle,&SessionHandle); // we ignore the return code
	}

	if (JHI_SUCCESS != ulRetCode)
	{
		TRACE1 ("JHIDLL: Service SAR failure, ulRetCode: %08x\n", ulRetCode);
	}

	return ulRetCode;
}


JHI_RET
JHI_Install2(
	IN const JHI_HANDLE handle, 
	IN const char*      AppId,
	IN const FILECHAR*     pInstallFile 
)
{
	UINT32 rc = JHI_INTERNAL_ERROR;
	UINT8 ucAppId[LEN_APP_ID+1]; // Local copy

	CommandInvoker cInvoker;

#ifdef __ANDROID__
	clearDeadOwnersSessions();
#endif //__ANDROID__

	// Validate the JHI handle
	if(!ValidateJHIhandle(handle))
		return JHI_INVALID_HANDLE;

	if ( !(AppId && (strlen(AppId) == LEN_APP_ID) &&
		(JhiUtilUUID_Validate(AppId, ucAppId) == JHI_SUCCESS)) )
	{
		TRACE0 ("Either Appname is bad or illegal length ..\n");
		return JHI_INVALID_APPLET_GUID;
	}

	if(pInstallFile == NULL || FILECHARLEN(pInstallFile) > FILENAME_MAX)
		return JHI_INVALID_INSTALL_FILE;


	TRACE0 ("calling SVC Install..\n") ;
	rc = cInvoker.JhisInstall((char *)ucAppId, pInstallFile) ;
		
	if (JHI_SUCCESS != rc ) 
	{
		TRACE1 ("JHDLL: Service Install failure, retcode: %08x\n", rc);
	}
	else
	{
		TRACE0 ("JHDLL: Service Install Complete\n");
	}

	return rc;
}

//------------------------------------------------------------
// Function: JHI_Uninstall
//		  Interface to be called by IHA or any external vendor
//        to un-install package from JoM
// IN	: handle 
// IN	: AppId - app ID of the package to be un-installed
// RETURN: JHI_RET - success or any failure returns
//------------------------------------------------------------

JHI_RET   JHI_Uninstall( 
	IN JHI_HANDLE   handle, 
	IN const char*        AppId
)
{
	UINT32 rc = JHI_INTERNAL_ERROR;
	UINT8 ucAppId[LEN_APP_ID+1] ; // Local copy

	CommandInvoker cInvoker;

#ifdef __ANDROID__
	clearDeadOwnersSessions();
#endif //__ANDROID__

	// Validate the JHI handle
	if(!ValidateJHIhandle(handle))
		return JHI_INVALID_HANDLE;

	if ( !(AppId && (strlen(AppId) == LEN_APP_ID) &&
		(JhiUtilUUID_Validate(AppId, ucAppId) == JHI_SUCCESS)) )
	{
		TRACE0 ("Either Appname is bad or illegal length ..\n");
		return JHI_INVALID_APPLET_GUID;
	}


	rc  = cInvoker.JhisUninstall((char*)ucAppId);
	
	if (JHI_SUCCESS != rc )
	{
		TRACE1 ("JHDLL: Applet Uninstall failure, retcode: %08x\n", rc);
	}
	else
	{
		TRACE0 ("JHIDLL: Applet Uninstall complete\n");
	}

	return rc;
}

//------------------------------------------------------------------------------
// Function: JHI_GetAppletProperty
//		  Interface to be called by IHA or any external vendor
//        to get version info of installed package in JoM
// IN	: handle
// IN	: pAppId - AppId of the package to be used in communication with JoM
// IN   : pComm - data buffer of type JVM_COMM_BUFFER to be used to send/recv 
//		  data. Tx buffer used to query version and Rx buffer to receive version
//		  response
// RETURN: JHI_RET - success or any failure returns
//------------------------------------------------------------------------------

JHI_RET   
JHI_GetAppletProperty(
	IN JHI_HANDLE       handle,
	IN const char*            pAppId,
	IN JVM_COMM_BUFFER* pComm
)
{ 
	UINT32 ulRetCode = JHI_INTERNAL_ERROR;
	UINT8 ucAppId[LEN_APP_ID+1] ; // Local copy
	DATA_BUFFER TxBuf;
	TxBuf.buffer = NULL;
	DATA_BUFFER RxBuf;
	RxBuf.buffer = NULL;
 

	CommandInvoker cInvoker;

	// Validate the JHI handle
	if(!ValidateJHIhandle(handle))
		return JHI_INVALID_HANDLE;

	if ( !(pAppId && (strlen(pAppId) == LEN_APP_ID) &&
		(JhiUtilUUID_Validate(pAppId, ucAppId) == JHI_SUCCESS)) )
	{
		TRACE0 ("Either Appname is bad or illegal length ..\n");
		return JHI_INVALID_APPLET_GUID;
	}

	// Validate the incoming values
	if(pComm == NULL)
	{
		return JHI_INVALID_COMM_BUFFER;
	}

	if ( ((pComm->TxBuf->length > 0) && (pComm->TxBuf->buffer == NULL) ) ||
		 ((pComm->RxBuf->length > 0) && (pComm->RxBuf->buffer == NULL) ))
	{
		TRACE0 ("Illegal argument supplied.. Check the input values..\n");
		return JHI_INVALID_COMM_BUFFER;
	}

	if (pComm->TxBuf->length == 0)
	{
		return JHI_APPLET_PROPERTY_NOT_SUPPORTED;
	}

	if ((pComm->TxBuf->length > JHI_BUFFER_MAX / sizeof(FILECHAR)) || (pComm->RxBuf->length > JHI_BUFFER_MAX / sizeof(FILECHAR)))
	{
		TRACE0 ("buffer sent exceeds JHI_BUFFER_MAX limit\n");
		return JHI_INVALID_BUFFER_SIZE;
	}

	//Convert from wchar_t to char if needed
	TxBuf.length = (pComm->TxBuf->length);
	TxBuf.buffer = (char*)JHI_ALLOC(TxBuf.length + 1);
	if (NULL == TxBuf.buffer)
	{
		LOG0 ("Failed to allocate buffer\n");
		ulRetCode = JHI_INTERNAL_ERROR;
		goto cleanup;
	}
	ZeroMemory(TxBuf.buffer, TxBuf.length + 1);
	memcpy_s(TxBuf.buffer, pComm->TxBuf->length , ConvertWStringToString((FILECHAR*)pComm->TxBuf->buffer).c_str(), TxBuf.length);
	
	RxBuf.length = pComm->RxBuf->length;
	if (0 != RxBuf.length)
	{
		RxBuf.buffer = (char*)JHI_ALLOC(RxBuf.length + 1);
		if (NULL == RxBuf.buffer)
		{
			LOG0 ("Failed to allocate buffer\n");
			ulRetCode = JHI_INTERNAL_ERROR;
			goto cleanup;
		}
	}
	TRACE0 ("calling SVC JhisGetAppletProperty..\n");

	ulRetCode = cInvoker.JhisGetAppletProperty((char*)ucAppId, (const uint8_t*)TxBuf.buffer, TxBuf.length, (uint8_t*)RxBuf.buffer, &RxBuf.length);

	pComm->RxBuf->length = RxBuf.length;

	if (JHI_SUCCESS != ulRetCode) 
	{
		TRACE1 ("JHIDLL: Service GetAppletProperty failure, ulRetCode: %08x\n", ulRetCode);
	}
	else
	{
		if (RxBuf.buffer != NULL)
		{
			((char*)RxBuf.buffer)[RxBuf.length] = '\0';
			FILESTRCPY((FILECHAR*)pComm->RxBuf->buffer, pComm->RxBuf->length + 1, ConvertStringToWString((char*)RxBuf.buffer).c_str());
		}
	}

cleanup:
	if (NULL != TxBuf.buffer)
		JHI_DEALLOC(TxBuf.buffer);
	if (NULL != RxBuf.buffer)
		JHI_DEALLOC(RxBuf.buffer);

	return ulRetCode;
}

JHI_RET 
JHI_GetSessionsCount(
	IN const JHI_HANDLE handle, 
	IN const char* AppId, 
	OUT UINT32* SessionsCount
)
{
	UINT32              rc = JHI_INTERNAL_ERROR;
	UINT8			    ucAppId[LEN_APP_ID+1] ; // Local copy

	CommandInvoker cInvoker;

	// Validate the JHI handle
	if(!ValidateJHIhandle(handle))
		return JHI_INVALID_HANDLE;

	if (SessionsCount == NULL)
		return JHI_INVALID_PARAMS;

	if ( !(AppId && (strlen(AppId) == LEN_APP_ID) &&
		(JhiUtilUUID_Validate(AppId, ucAppId) == JHI_SUCCESS)) )
	{
		TRACE0 ("Either Appname is bad or illegal length ..\n");
		return JHI_INVALID_APPLET_GUID;
	}

	// call for JhisGetSessionsCount at the service
	rc  = cInvoker.JhisGetSessionsCount((char *)ucAppId, SessionsCount);
	
	if (JHI_SUCCESS != rc )
		TRACE1 ("JHDLL: get sessions count failure, retcode: %08x\n", rc);
	else
		TRACE0 ("JHIDLL: Get Sessions Count Complete\n");

	return rc;
}

#ifdef __ANDROID__
JHI_RET
JHI_ClearSessions(
	IN const JHI_HANDLE handle,
	IN int ApplicationPid
)
{
	UINT32 rc = JHI_SUCCESS;
	if(!ValidateJHIhandle(handle))
		return JHI_INVALID_HANDLE;
	clearDestroyedSessions (ApplicationPid);
	return rc;
}
#endif //__ANDROID__

JHI_RET
JHI_CloseSessionInternal(
IN const JHI_HANDLE handle,
IN JHI_SESSION_HANDLE* pSessionHandle,
IN bool force
)
{
	UINT32 rc = JHI_SUCCESS;
	JHI_I_SESSION_HANDLE* iSessionHandle;

	CommandInvoker cInvoker;

	// Validate the JHI handle
	if (!ValidateJHIhandle(handle))
		return JHI_INVALID_HANDLE;

	if (pSessionHandle == NULL)
		return JHI_INVALID_SESSION_HANDLE;
	
	appHandleLock.Lock();

	do
	{
		iSessionHandle = (JHI_I_SESSION_HANDLE*) *pSessionHandle;

		if (!SessionHandleValid(iSessionHandle))
		{
			rc = JHI_INVALID_SESSION_HANDLE;
			break;
		}

		//remove event registration  the session is indead registered for events
		if (iSessionHandle->eventHandle != NULL && iSessionHandle->eventHandle->is_created())
		{
			TRACE0 ("JHIDLL: removing session event registration\n");
#ifdef __ANDROID__
			TRACE2 ("JHIDLL CloseSession: socket counters tx %d rx %d\n", iSessionHandle->eventHandle->tx_cnt, iSessionHandle->eventHandle->rx_cnt);
#endif
			closeSessionEventThread(iSessionHandle);
		}

		// call for close session at the service
		rc = cInvoker.JhisCloseSession(&iSessionHandle->sessionID, &appHandle->processInfo, force);

		// JHI may return JHI_INVALID_SESSION_HANDLE which is expected.
		if (rc == JHI_INVALID_SESSION_HANDLE)
			rc = JHI_SUCCESS;

		if (rc == JHI_SUCCESS)
		{
			// remove the session handle form the list.
			if (removeSessionHandle(((JHI_I_SESSION_HANDLE*)(*pSessionHandle))))
			{
				// release allocated memory
				JHI_DEALLOC(*pSessionHandle);
				*pSessionHandle = NULL;
			}
		}
		TRACE0 ("JHIDLL: Session Close Complete\n");

	} while (0);

	appHandleLock.UnLock();

	return rc;
}

JHI_RET
JHI_CloseSession(
	IN const JHI_HANDLE handle,
	IN JHI_SESSION_HANDLE* pSessionHandle
)
{
	return JHI_CloseSessionInternal(handle, pSessionHandle, false);
}

JHI_RET
JHI_ForceCloseSession(
	IN const JHI_HANDLE handle,
	IN JHI_SESSION_HANDLE* pSessionHandle
)
{
	return JHI_CloseSessionInternal(handle, pSessionHandle, true);
}

JHI_RET 
JHI_GetSessionInfo(
	IN const JHI_HANDLE handle, 
	IN JHI_SESSION_HANDLE SessionHandle, 
	OUT JHI_SESSION_INFO* SessionInfo
)
{
	UINT32 rc = JHI_INTERNAL_ERROR;
	JHI_SESSION_ID sessionID;

	CommandInvoker cInvoker;

	// Validate the JHI handle
	if(!ValidateJHIhandle(handle))
		return JHI_INVALID_HANDLE;

	if (!getSessionID(SessionHandle,&sessionID))
		return JHI_INVALID_SESSION_HANDLE;

	if (SessionInfo == NULL)
		return JHI_INVALID_PARAMS;


	// call for get session info at the service
	rc  = cInvoker.JhisGetSessionInfo(&sessionID,SessionInfo);
	
	// if the session crashed we remove its allocated resources by calling closeSession
	if (rc == JHI_INVALID_SESSION_HANDLE)
	{
		JHI_CloseSession(handle,&SessionHandle); // we ignore the return code
	}

	if (JHI_SUCCESS != rc )
		TRACE1 ("JHDLL: GetSessionStatus failure, retcode: %08x\n", rc);
	else
		TRACE0 ("JHIDLL: Get Session Status Complete\n");
	
	return rc;
}

#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
JHI_RET
JHI_GetSessionTable(OUT JHI_SESSIONS_DATA_TABLE** SessionDataTable)
{
	UINT32 rc = JHI_INTERNAL_ERROR;

	CommandInvoker cInvoker;

	// calls the setvice to get session data table
	rc  = cInvoker.JhisGetSessionTable(SessionDataTable);
	
	if (JHI_SUCCESS != rc )
	{
		TRACE1 ("JHDLL: GetSessionTable failure, retcode: %08x\n", rc);
	}
	else
	{
		TRACE0 ("JHIDLL: GetSessionTable Complete\n");
	}

	return rc;
}

JHI_RET
JHI_FreeSessionTable(IN JHI_SESSIONS_DATA_TABLE** SessionDataTable)
{
	UINT32 rc = JHI_INTERNAL_ERROR;
	if (*SessionDataTable)
	{
		JHI_DEALLOC(*SessionDataTable); // not using JHI_DEALLOC_T_ARRAY because JHI_ALLOC was used
		*SessionDataTable = NULL;
	}
	rc = JHI_SUCCESS;

	return rc;
}

JHI_RET
JHI_GetLoadedAppletsList(OUT JHI_LOADED_APPLET_GUIDS** appGUIDs)
{
	UINT32 rc = JHI_INTERNAL_ERROR;

	CommandInvoker cInvoker;

	// calls the setvice to get session data table
	rc  = cInvoker.JhisGetLoadedAppletsList(appGUIDs);
	
	if (JHI_SUCCESS != rc )
	{
		TRACE1 ("JHDLL: GetLoadedAppletsList failure, retcode: %08x\n", rc);
	}
	else
	{
		TRACE0 ("JHIDLL: Get Loaded Applets List Complete\n");
	}

	return rc;
}

JHI_RET
JHI_FreeLoadedAppletsList(IN JHI_LOADED_APPLET_GUIDS** appGUIDs)
{
	UINT32 rc = JHI_INTERNAL_ERROR;
	if (appGUIDs && *appGUIDs)
	{
		rc = freeLoadedAppletsList(*appGUIDs);
		if (rc == JHI_SUCCESS)
		{
			JHI_DEALLOC_T(*appGUIDs);
			*appGUIDs = NULL;
		}
	}
	return rc;
}

#endif

void retriveEventData(JHI_I_SESSION_HANDLE* pSession, JHI_SESSION_ID sessionID, JHI_EventFunc callback,	UINT8* threadNeedToEnd)
{
	UINT32 rc = JHI_INTERNAL_ERROR;
	UINT8 retrieveEventData;
	UINT8 dataType;
	JHI_EVENT_DATA event_data;
	CommandInvoker cInvoker;

	retrieveEventData = 1; 

	// we try to get events data until the service signals there are no more
	// or until the thread needs to end.
	while (retrieveEventData && !(*threadNeedToEnd))
	{
		event_data.data = NULL;
		event_data.datalen = 0;

		// call get event data in order to retrieve that data related to this event.
		rc  = cInvoker.JhisGetEventData(&sessionID, &event_data.datalen, &event_data.data, &dataType);

		if (rc == JHI_GET_EVENT_FAIL_NO_EVENTS)
		{
			// no more events stored in JHI for this session.
			retrieveEventData = 0;
		}
		else if (JHI_SUCCESS != rc )
		{
			// failed to recieve event data from the service,
			// the thread will keep listening until session is unregisterd by the application
			TRACE1("failed to retreive event! err: %d\n",rc);
			retrieveEventData = 0;
		}

		if (retrieveEventData)
		{
			// invoke the application callback and pass the event data
			TRACE0("event recieved!\n");
			event_data.dataType = (JHI_EVENT_DATA_TYPE) dataType;
			callback((JHI_SESSION_HANDLE) pSession,event_data);
		}

		if (event_data.data)
		{
			JHI_DEALLOC(event_data.data);
			event_data.data = NULL;
		}
	}
}

#ifdef _WIN32
DWORD eventListenerThread (LPVOID threadParam)
{
	DWORD success = 0, error = -1;
#else
void* eventListenerThread(void* threadParam)
{
	void * success = NULL, * error = NULL;
#endif //WIN32

	JHI_I_SESSION_HANDLE* pSession = (JHI_I_SESSION_HANDLE*) threadParam;
	if (pSession == NULL)
	{
		return error;
	}

	UINT8* threadNeedToEnd = pSession->threadNeedToEnd;
	JhiEvent* eventHandle = pSession->eventHandle;
	JHI_SESSION_ID sessionID = pSession->sessionID;
	JHI_EventFunc callback = pSession->callback;

	if (threadNeedToEnd == NULL || eventHandle == NULL)
	{
		return error;
	}
	
#ifndef _WIN32
	if (!eventHandle->listenCl()) {
	  TRACE3("socket srv lstn, l%d, %s \n",
		  __FILE__,__LINE__, strerror(errno));
	} else {
#endif
	  while(!(*threadNeedToEnd))
	  {
		// wait for an event
		if(eventHandle->wait())
			retriveEventData(pSession, sessionID, callback, threadNeedToEnd);
	  }
#ifndef _WIN32
	}
	TRACE0("JHIDLL: listener thread finishing...\n");
#endif

	if (threadNeedToEnd != NULL)
	{
		JHI_DEALLOC_T(threadNeedToEnd);
		threadNeedToEnd = NULL;
	}

	if (eventHandle != NULL)
	{
		eventHandle->close();
		JHI_DEALLOC_T(eventHandle);
		eventHandle = NULL;
	}

	return success;
}

char* generateHandleUUID(JHI_SESSION_ID sessionID)
{
	char * hName = NULL;
		
#ifdef _WIN32

	UUID uuid;
	char * uuidStr = NULL;
	RPC_STATUS status;

	status = UuidCreate(&uuid);
	if (status != RPC_S_OK && status != RPC_S_UUID_LOCAL_ONLY)
	{
		TRACE0("UuidCreate failed");
		return NULL;
	}

	string hNameAsString("Global\\");

	if (UuidToStringA(&uuid, (RPC_CSTR*)&uuidStr) != RPC_S_OK)
	{
		TRACE0("failed to generate eventhandle uuid");
		return NULL;
	}
	else
	{
		hNameAsString += string(uuidStr);

		hName = (char*) JHI_ALLOC((uint32_t)hNameAsString.length() + 1);
		if (hName == NULL)
		{
			TRACE0("failed to allocate memory for event handle name");
		}
		else
		{
			strcpy_s(hName,hNameAsString.length()+1,hNameAsString.c_str());
		}

		RpcStringFreeA((RPC_CSTR*)&uuidStr);
	}

#else //!_WIN32
    uuid_t id;
	string hNameAsString;
#ifdef __ANDROID__
	static string eventPath ("\0");
	if (eventPath.length() == 0 )
	{
		FILECHAR jhiEventSocketLocation[FILENAME_MAX+1]={0};
		if( JHI_SUCCESS != JhiQueryEventSocketsLocationFromRegistry(
			jhiEventSocketLocation,
			(FILENAME_MAX-1) * sizeof(FILECHAR)))
		{
			TRACE0( "unable to find dynamic sockets folder from registry") ;
			eventPath +="/data/intel/dal/dynamic_sockets/jhievent-";
		}
		else if (_waccess_s(jhiEventSocketLocation,0) != 0)
		{
			TRACE0("Init failed - cannot find sockets directory");
			eventPath += "/data/data/jhievent-";
		}
		else
		{
			eventPath += jhiEventSocketLocation;
			eventPath += "/jhievent-";
		}
	}
	hNameAsString += eventPath.c_str();
#else
	hNameAsString += "/tmp/jhievent-";
#endif //__ANDROID__
	uuid_generate(id);
	char out[37];
	uuid_unparse(id, out);

	hNameAsString += out;

	hName = (char*) JHI_ALLOC(hNameAsString.length() + 1);
	if (hName == NULL)
	{
		LOG0("failed to allocate memory for event handle name");
	}
	else
	{
		strcpy_s(hName, hNameAsString.length() + 1, hNameAsString.c_str());
	}
#endif //_WIN32
	
	TRACE1("jhi event name %s", hName);
	return hName;
}

JHI_RET 
JHI_RegisterEvents(IN const JHI_HANDLE handle, 
				  IN JHI_SESSION_HANDLE SessionHandle,
				  IN JHI_EventFunc pEventFunction) 
{
	UINT32 rc = JHI_INTERNAL_ERROR;
	char* HandleName = NULL;

	JHI_I_SESSION_HANDLE* iSessionHandle = (JHI_I_SESSION_HANDLE*) SessionHandle;

	CommandInvoker cInvoker;

	// Validate the JHI handle
	if(!ValidateJHIhandle(handle))
		return JHI_INVALID_HANDLE;

	if (pEventFunction == NULL)
		return JHI_INVALID_PARAMS;


	appHandleLock.Lock();

	do 
	{
		
		if (!SessionHandleValid(iSessionHandle))
		{
			rc = JHI_INVALID_SESSION_HANDLE;
			break;
		}

		if ((iSessionHandle->sessionFlags & JHI_SHARED_SESSION) == JHI_SHARED_SESSION)
		{
			rc = JHI_EVENTS_NOT_SUPPORTED;
			break;
		}

		// Check if event is already allocated
		if (iSessionHandle->eventHandle != NULL && iSessionHandle->eventHandle->is_created())
		{
			rc = JHI_SESSION_ALREADY_REGSITERED;
			break;
		}

		iSessionHandle->threadNeedToEnd = JHI_ALLOC_T(UINT8);

		if (iSessionHandle->threadNeedToEnd == NULL)
		{
			rc = JHI_INTERNAL_ERROR;
			break;
		}

		*(iSessionHandle->threadNeedToEnd) = 0;
		iSessionHandle->callback = pEventFunction;

		// create the os event the event thread will use
		HandleName = generateHandleUUID(iSessionHandle->sessionID);

		if (HandleName == NULL)
		{
			TRACE0("failed to generate event handle name");
			JHI_DEALLOC_T(iSessionHandle->threadNeedToEnd);
			iSessionHandle->threadNeedToEnd = NULL;
			rc = JHI_INTERNAL_ERROR;
			break;
		}

		iSessionHandle->eventHandle = JHI_ALLOC_T(JhiEvent);
		if (iSessionHandle->eventHandle == NULL)
		{
			LOG0("failed to allocate event handle");
			JHI_DEALLOC_T(iSessionHandle->threadNeedToEnd);
			iSessionHandle->threadNeedToEnd = NULL;
			rc = JHI_INTERNAL_ERROR;
			break;
		}

		if(!iSessionHandle->eventHandle->create(HandleName))
		{
			TRACE0("failed to create OS event");
			JHI_DEALLOC_T(iSessionHandle->threadNeedToEnd);
			iSessionHandle->threadNeedToEnd = NULL;
			JHI_DEALLOC_T(iSessionHandle->eventHandle);
			iSessionHandle->eventHandle = NULL;
			rc = JHI_INTERNAL_ERROR;
			break;
		}

		// create a thread that will listen for events
#ifdef _WIN32
		iSessionHandle->threadHandle = CreateThread(NULL, 0,(LPTHREAD_START_ROUTINE)&eventListenerThread, iSessionHandle,0,NULL);

		if (iSessionHandle->threadHandle == NULL)
#else
		rc = pthread_create(&iSessionHandle->threadHandle, NULL, eventListenerThread, iSessionHandle);
		if(rc)
#endif //WIN32
		{
			TRACE0("failed to create event listener thread");
			JHI_DEALLOC_T(iSessionHandle->threadNeedToEnd);
			iSessionHandle->threadNeedToEnd = NULL;
			JHI_DEALLOC_T(iSessionHandle->eventHandle);
			iSessionHandle->eventHandle = NULL;
			rc = JHI_INTERNAL_ERROR;
			break;
		}

		// call for Register Event at the service
		rc  = cInvoker.JhisSetSessionEventHandler(&(iSessionHandle->sessionID),HandleName);

	} while (0) ;


	// cleanup

	if (HandleName != NULL)
	{
		JHI_DEALLOC(HandleName);
		HandleName = NULL;
	}

	if ((rc != JHI_SUCCESS) && (rc != JHI_SESSION_ALREADY_REGSITERED))
	{
		closeSessionEventThread(iSessionHandle);
		TRACE1 ("JHDLL: Register Event failure, retcode: %08x\n", rc);

		if (rc == JHI_INVALID_SESSION_HANDLE)
		{
			// remove the session handle form the list.
			if (removeSessionHandle(iSessionHandle))
			{
				// release allocated memory
				JHI_DEALLOC(iSessionHandle);
				iSessionHandle = NULL;
			}
		}

	}
	else
	{
		TRACE0 ("JHIDLL: Register Event Complete\n");
	}

	appHandleLock.UnLock();
	
	return rc;
}


JHI_RET 
JHI_UnRegisterEvents(
	IN const JHI_HANDLE handle, 
	IN JHI_SESSION_HANDLE SessionHandle)
{
	UINT32 rc = JHI_INTERNAL_ERROR;
	JHI_I_SESSION_HANDLE* iSessionHandle = (JHI_I_SESSION_HANDLE*) SessionHandle;	


	CommandInvoker cInvoker;

	// Validate the JHI handle
	if(!ValidateJHIhandle(handle))
		return JHI_INVALID_HANDLE;

	appHandleLock.Lock();

	do
	{
		if (!SessionHandleValid(iSessionHandle))
		{
			rc = JHI_INVALID_SESSION_HANDLE;
			break;
		}

		if ((iSessionHandle->sessionFlags & JHI_SHARED_SESSION) == JHI_SHARED_SESSION)
		{
			rc = JHI_EVENTS_NOT_SUPPORTED;
			break;
		}

		//check that the session is indead registered for events
		if (iSessionHandle->eventHandle == NULL || !iSessionHandle->eventHandle->is_created())
		{
			TRACE0("Trying to unregister an unregistered session");
			rc = JHI_SESSION_NOT_REGISTERED;
			break;
		}
#ifdef __ANDROID__
		TRACE2 ("JHIDLL unregister: socket counters tx %d rx %d\n", iSessionHandle->eventHandle->tx_cnt, iSessionHandle->eventHandle->rx_cnt);
#endif
		closeSessionEventThread(iSessionHandle);

		// send an Unregister Event command to the JHI_service
		// call for JhisSetSessionEventHandler at the service with HandleName = "" to unregister it.
		rc  = cInvoker.JhisSetSessionEventHandler(&(iSessionHandle->sessionID),"");
		
		if (JHI_SUCCESS != rc )
		{
			TRACE1 ("JHDLL: Unregister Event failure, retcode: %08x\n", rc);
		}
		else
		{
			TRACE0 ("JHIDLL: Unregister Event Complete\n");
		}

		// signal the thread to wake and close itself

		if (rc == JHI_INVALID_SESSION_HANDLE)
		{
			// remove the session handle form the list.
			if (removeSessionHandle(iSessionHandle))
			{
				// release allocated memory
				JHI_DEALLOC(iSessionHandle);
				iSessionHandle = NULL;
			}
		}

	}
	while(0);

	appHandleLock.UnLock();

	return rc;
}

JHI_RET
JHI_GetVersionInfo (
   IN const JHI_HANDLE handle,
   OUT JHI_VERSION_INFO* pVersionInfo
)
{
	UINT32 rc = JHI_INTERNAL_ERROR;	

	CommandInvoker cInvoker;

	// Validate the JHI handle
	if(!ValidateJHIhandle(handle))
		return JHI_INVALID_HANDLE;

	do 
	{

		if (pVersionInfo == NULL)
		{
			rc =  JHI_INVALID_PARAMS;
			break;
		}


		// send an Get Version Info  command to the JHI_service
		rc  = cInvoker.JhisGetVersionInfo(pVersionInfo);

		if (JHI_SUCCESS != rc )
		{
			TRACE1 ("JHDLL: VersionInfo failure, retcode: %08x\n", rc);
			break;
		}
		else
		{
			TRACE0 ("JHIDLL: Get Version Info Complete\n");
		}

		rc = JHI_SUCCESS;
	}
	while (0);

	return rc;
}