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

#ifndef __SESSIONSMANAGER_H
#define __SESSIONSMANAGER_H

#ifdef _WIN32
#include <windows.h>
#else
#include <string.h>
#endif

// The H-Files
#include <map>
#include <list>
#include <algorithm>
#include <cstdint>
#include <string>
#include <ostream>
#include <queue>
#include "jhi.h"
#include "Locker.h"
#include "typedefs.h"
#include "Singleton.h"
#include "GlobalsManager.h"
#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
#include "jhi_sdk.h"
#endif
#include "jhi_event.h"

namespace intel_dal
{
	using std::map;
	using std::pair;
	using std::list;
	using std::string;
	using std::ostream;
	using std::queue;

#define MAX_EVENTS_DATA_IN_QUEUE 100
#define	MAX_SESSION_OWNERS 20

typedef union
{
	UINT32 value;
	struct
	{
		UINT32 sharedSession : 1;
		UINT32 unused : 31;
	} bits;
} JHI_SESSION_FLAGS;


/* Session Table Record */
struct  SessionRecord 
{
	JHI_SESSION_ID			sessionId;
	VM_SESSION_HANDLE		vmSessionHandle;
	string					appId;
	JHI_SESSION_FLAGS		sessionFlags;
	JHI_SESSION_STATE		state;
	list<JHI_PROCESS_INFO>	ownersList;
	queue<JHI_EVENT_DATA*>  eventsDataQueue;
	Locker* 				sessionLock;
	uint32_t    			lastUsedTime;
	JhiEvent*				eventHandle;
};


/**
This class manage the Sessions Table of JHI
**/
class SessionsManager : public Singleton<SessionsManager>
{
	friend class Singleton<SessionsManager>;
public:
	/** 
		Add New Session to the Table with Default API Version

		Parameters:
			appId				- App ID
			vmSessionHandle		- assosiated VM session Handle to add
			sessionID			- the designated ID for this session
			processInfo			- pid and timestamp of the appliction that creates the session
		Return:
			true if added successfuly, false otherwise
	**/
	bool add(const string& appId, VM_SESSION_HANDLE vmSessionHandle,JHI_SESSION_ID sessionID,JHI_SESSION_FLAGS flags,JHI_PROCESS_INFO* processInfo);
	
	/** 
		Delete Session from the Table.

		Parameters:
			sessionId		- sessionId

		Return:
			true	- able to delete
			false	- unable to delete
	**/	
	bool remove(JHI_SESSION_ID sessionId);

	/** 
		Check if Session present in the Table. 

		Parameters:
			sessionId		- sessionId

		Return:
			true	- exists
			false	- not exists
	**/	
	bool isSessionPresent(JHI_SESSION_ID sessionId);

	/**
		add a session owner to the session owners list.
		owners list is limited to MAX_SESSION_OWNERS
		
		Parameters:
			sessionId		- sessionId
			info			- the process info

		Return:
			true if owner added, false otherwise
	**/
	bool addSessionOwner(JHI_SESSION_ID sessionID,JHI_PROCESS_INFO* info);


	/**
		remove a session owner from the session owners list.
		
		Parameters:
			sessionId		- sessionId
			info			- the process info

		Return:
			true if the owner removed, false otherwise.
	**/
	bool removeSessionOwner(JHI_SESSION_ID sessionID,JHI_PROCESS_INFO* info);


	bool isSessionOwnerValid(JHI_SESSION_ID sessionID,JHI_PROCESS_INFO* info);

	/**
		returns the number of session owners
		
		Parameters:
			sessionId		- sessionId

	**/
	int getOwnersCount(JHI_SESSION_ID sessionID);

	/** 
		Return the VM Session Handle of a given Session ID

		Parameters:
			IN sessionId		- sessionId
			OUT vm_handle	    - the VM session handle

		Return:
			true on success, false otherwise
	**/	
	 bool getVMSessionHandle(JHI_SESSION_ID sessionId,VM_SESSION_HANDLE* vm_handle);

	/** 
		Return a list VM Session Handles of a given applet ID and API Version

		Parameters:
			appID	- Applet ID

		Return:
			VM Session Handle
	**/	
	list<VM_SESSION_HANDLE> getVMSessionHandles(const string& appId);

	/** 
		Return a list JHI Session Handles of a given applet ID and API Version

		Parameters:
			appID	- Applet ID

		Return:
			VM Session Handle
	**/	
	list<JHI_SESSION_ID> getJHISessionHandles(const string& appId);

	/** 
		gets an applet shared session ID if there is such session

		Parameters:
			appID	- Applet ID
			sessionId	- the session id 

		Return:
			true if shared session was found, false otherwise
	**/	
	bool getSharedSessionID(JHI_SESSION_ID* sessionId,const string& appId);

	/** 
		Return is there are sessions of a given applet

		Parameters:
			appID	- Applet ID

		Return:
			VM Session Handle
	**/	
	bool hasLiveSessions(const string& appId);


	/** 
		add an event data to a session queue

		Parameters:
			sessionId		- sessionId	
			pEventData		- pointer to the event data

		Return:
			true on success, false on failure
	**/	
	bool enqueueEventData(const JHI_SESSION_ID sessionId,JHI_EVENT_DATA* pEventData);


	/** 
		this function is called by the application in order
		to recieve data of a raised event as a JHI_EVENT_DATA struct 

		Parameters:

			IN SessionID - Session ID
			OUT pEventData - pointer to the event data struct

		Return:
			true is init succeded, false otherwise.
	**/	
	JHI_RET getSessionEventData(const JHI_SESSION_ID SessionID,JHI_EVENT_DATA* pEventData);

	bool GetSessionLock(JHI_SESSION_ID sessionID);
	
	void ReleaseSessionLock(JHI_SESSION_ID sessionID);
	
	/** 
		Return information of a given session

		Parameters:
			IN sessionId		- sessionId
			OUT pSessionInfo    - the session info
		
	**/	
	void getSessionInfo(JHI_SESSION_ID sessionId,JHI_SESSION_INFO* pSessionInfo);

#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
	/**
		Return information of all sessions
		OUT pSessionsDataTable    - the sessions data table

	**/
	void getSessionsDataTable(JHI_SESSIONS_DATA_TABLE* pSessionsDataTable);
#endif

	/** 
		Reset the session table and the nextSessionId counter.
	**/	
	void resetSessionManager();

	/* 
		returns a new session ID.
	*/
	bool generateNewSessionId(JHI_SESSION_ID* sessionId);


	/** 
		try to remove dead sessions owners of a all sessions in the sessions table.
		this function is called on any type of session cleanup

		Return:
			true if removed at least one owner, false otherwise
		
	**/	
	bool ClearSessionsDeadOwners();

	/** 
		try to remove non shared sessions that thier owning applicaion no longer exists
		by calling jhi_closeSession with their session id.

		Return:
			true if removed at least one session, false otherwise
		
	**/	
	bool ClearAbandonedNonSharedSessions();

	/** 
		try to remove a shared session of a given applet that has no owning applicaion
		by calling jhi_closeSession with their session id.

		Return:
			true if the shared session removed, false otherwise
		
	**/	
	bool ClearAppletSharedSession(const string& appId);

	/** 
		check if an applet has existing non-shared sessions 

		Return:
			true if non-shared sessions exist, false otherwise
		
	**/	
	bool AppletHasNonSharedSessions(const string& appId);

	/** 
		try to remove one shared sessions that is not active (has no owning applicaion)
		using an LRU algorithem.

		Parameters:
			IN allowNonSharedSessions - set to true to include shared sessions of applets that has non-shared sessions as well, false otherwise  

		Return:
			true if removed one shared session, false otherwise
		
	**/	
	bool TryRemoveUnusedSharedSession(bool allowNonSharedSessions);


	/** 
		this function returns the session id as a string

		Parameters:
			IN sessionId		- sessionId
		
		Return:
			the session id as string

	**/		
	string sessionIdToString(JHI_SESSION_ID sessionId);

	/** 
		sets an event handler that is associated with a session

		Parameters:
			IN sessionId		- sessionId
			IN eventHandle		- the handle to the event

		Return:
			true on success, false on failure
	**/	
	bool setEventHandle(JHI_SESSION_ID SessionId, JhiEvent* eventHandle);

	/** 
		return an event handler that is associated with a given session ID

		Parameters:
			IN sessionId		- sessionId
			OUT eventHandle		- the handle to the event

		Return:
			true on success, false on failure
	**/	
	bool getEventHandle(JHI_SESSION_ID SessionId, JhiEvent** eventHandle);

	/*
		used in order to close all sessions in VM before resetting the service
	*/
	void closeSessionsInVM();

private:
	
	struct lt_sessionId
	{
	  bool operator()(const JHI_SESSION_ID sid1, const JHI_SESSION_ID sid2) const
	  {
		  //return strcmp(s1, s2) < 0;
		  return memcmp(&sid1, &sid2, sizeof(JHI_SESSION_ID)) < 0;
	  }
	};

	
	map<JHI_SESSION_ID, SessionRecord, lt_sessionId>	_sessionList;
	Locker	_locker;
	
	unsigned long sharedSessionLRUCounter; 
	
	// Default Constructor
	SessionsManager(void);

	// Destructor 
	~SessionsManager(void);	


	// convert a string (mostly appid) to upper case
	string toUpperCase(const string& str); 

	void clearEventsQueue(queue<JHI_EVENT_DATA*>&  eventsDataQueue);

	/** 
		try to remove dead session owners (crashed applications) of a given session from the session owners list.

		Return:
			true if removed at least one owner, false otherwise
		
	**/	
	bool ClearSessionDeadOwners(JHI_SESSION_ID sessionID);

	/** 
		this function updates a session last usage timestamp given a session Id

		Parameters:
			IN sessionId		- sessionId

	**/	
	void updateSessionLastUsage(SessionRecord* sessionRecord);
	
};

}

#endif