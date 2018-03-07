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

// The H-Files
#include <cstdlib>
#include <algorithm>
#include <map>
#include "SessionsManager.h"

#ifdef __linux__
#include <sys/stat.h>
#include <uuid/uuid.h>
#include <signal.h>
#include <ctype.h>
#endif//__linux__
#include "string_s.h"

namespace intel_dal
{

	// Default Constructor
	SessionsManager::SessionsManager(void)  : _sessionList()
	{
		TRACE0("in SessionsManager constructor\n");
		sharedSessionLRUCounter = 1;
	}

	// Destructor
	SessionsManager::~SessionsManager(void)
	{
		TRACE0("in SessionsManager destructor\n");
	}


	bool SessionsManager::generateNewSessionId(JHI_SESSION_ID* sessionId)
	{
#ifdef _WIN32

		RPC_STATUS status;

		status = UuidCreate(sessionId);
		if (status != RPC_S_OK && status != RPC_S_UUID_LOCAL_ONLY)
		{
			TRACE0("Failed to generate a session uuid\n");
			return false;
		}

#else
		uuid_generate((unsigned char*)sessionId);
#endif // _WIN32

		return true;
	}

	string SessionsManager::sessionIdToString(JHI_SESSION_ID sessionId)
	{
		string returnedUuid = "";

#ifdef _WIN32
		char* uuidStr = NULL;

		if (UuidToStringA(&sessionId, (RPC_CSTR*)&uuidStr) != RPC_S_OK)
		{
			TRACE0("UuidToStringA failed");
		}
		else
		{
			returnedUuid = uuidStr;
			RpcStringFreeA((RPC_CSTR*)&uuidStr);
			uuidStr = NULL;
		}
#else
		char out[37];
		uuid_unparse((unsigned char*)&sessionId, out);
		returnedUuid = out;
#endif // _WIN32

		return returnedUuid;
	}

	bool SessionsManager::add(const string& appId, VM_SESSION_HANDLE vmSessionHandle, JHI_SESSION_ID sessionID,JHI_SESSION_FLAGS flags,JHI_PROCESS_INFO* processInfo)
	{
		bool status = true;

		_locker.Lock();

		do
		{
			if (isSessionPresent(sessionID))
			{
				status = false;
				break;
			}

			if (processInfo == NULL)
			{
				status = false;
				break;
			}

			SessionRecord newRecord;
			newRecord.sessionFlags.value = flags.value;
			newRecord.vmSessionHandle = vmSessionHandle;
			newRecord.appId = toUpperCase(appId);
			newRecord.state = JHI_SESSION_STATE_ACTIVE;
			newRecord.ownersList.push_back(*processInfo);
			newRecord.sessionId = sessionID; 
			newRecord.eventHandle = NULL;
			newRecord.sessionLock = JHI_ALLOC_T(Locker);
			newRecord.lastUsedTime = 0;

			_sessionList.insert(pair<JHI_SESSION_ID, SessionRecord>(sessionID, newRecord));

			TRACE2("session record added to session table,session id: [%s]\n current session count: %d\n",sessionIdToString(sessionID).c_str(),_sessionList.size());
		}
		while(0);

		_locker.UnLock();

		return status;
	}

	bool SessionsManager::remove(JHI_SESSION_ID sessionId)
	{
		size_t ret = 0;
		Locker* sessionLock;
		JhiEvent* eventHandle;

		_locker.Lock();

		do
		{
			if (!isSessionPresent(sessionId))
				break;

			// remove stored events
			clearEventsQueue(_sessionList[sessionId].eventsDataQueue);

			// remove the session record and free the session lock
			sessionLock = _sessionList[sessionId].sessionLock;
			eventHandle = _sessionList[sessionId].eventHandle;
			
			ret = _sessionList.erase(sessionId);

			TRACE2("session record removed to session table,session id: [%s]\n current session count: %d\n",sessionIdToString(sessionId).c_str(),_sessionList.size());

			sessionLock->UnLock();
			if (eventHandle != NULL)
			{
				eventHandle->close();
				JHI_DEALLOC_T(eventHandle);
			}

			JHI_DEALLOC_T(sessionLock);
			sessionLock = NULL;
		}
		while(0);

		_locker.UnLock();

		return (ret != 0);
	}

	bool SessionsManager::getVMSessionHandle(JHI_SESSION_ID sessionId,VM_SESSION_HANDLE* vm_handle)
	{
		bool ret = false;

		if (vm_handle == NULL)
			return false;

		_locker.Lock();

		do
		{
			if (!isSessionPresent(sessionId))
				break;

			SessionRecord record = _sessionList[sessionId];
			*vm_handle = record.vmSessionHandle;
			ret = true;
		}
		while (0);

		_locker.UnLock();

		return ret;
	}

	list<VM_SESSION_HANDLE> SessionsManager::getVMSessionHandles(const string& appId)
	{
		list<VM_SESSION_HANDLE> VMhandles;

		map<JHI_SESSION_ID, SessionRecord,lt_sessionId>::iterator it;

		string appid = toUpperCase(appId);

		_locker.Lock();

		for ( it=_sessionList.begin() ; it != _sessionList.end(); it++ )
		{
			if (appid == it->second.appId)
				VMhandles.push_back(it->second.vmSessionHandle);
		}

		_locker.UnLock();

		return VMhandles;
	}

	void SessionsManager::closeSessionsInVM()
	{
		map<JHI_SESSION_ID, SessionRecord, lt_sessionId>::iterator it;
		VM_SESSION_HANDLE sHandle;

		VM_Plugin_interface* plugin = NULL;

		if ( (!GlobalsManager::Instance().getPluginTable(&plugin)) || (plugin == NULL) )
		{
			return;	
		}

		_locker.Lock();

		for ( it=_sessionList.begin(); it != _sessionList.end(); it++ )
		{
			sHandle = it->second.vmSessionHandle;
			plugin->JHI_Plugin_CloseSession(&sHandle);
		}

		_locker.UnLock();

	}

	list<JHI_SESSION_ID> SessionsManager::getJHISessionHandles(const string& appId)
	{
		list<JHI_SESSION_ID> JHIhandles;

		map<JHI_SESSION_ID, SessionRecord, lt_sessionId>::iterator it;

		string appid = toUpperCase(appId);

		_locker.Lock();

		for ( it=_sessionList.begin() ; it != _sessionList.end(); it++ )
		{
			if (appid == it->second.appId)
				JHIhandles.push_back(it->first);
		}

		_locker.UnLock();

		return JHIhandles;
	}

	bool SessionsManager::isSessionPresent(JHI_SESSION_ID sessionId)
	{
		map<JHI_SESSION_ID, SessionRecord, lt_sessionId>::iterator it;

		it = _sessionList.find(sessionId);

		return (it != _sessionList.end());
	}

	bool SessionsManager::hasLiveSessions(const string& appId)
	{
		bool status = false;

		map<JHI_SESSION_ID, SessionRecord, lt_sessionId>::iterator it;

		string appid = toUpperCase(appId);

		_locker.Lock();

		for ( it=_sessionList.begin() ; it != _sessionList.end(); it++ )
		{
			if (appid == it->second.appId)
			{
				status = true;
				break;
			}
		}

		_locker.UnLock();

		return status;
	}

	string SessionsManager::toUpperCase(const string& str)
	{

		unsigned i;
		string ret_str = "";

		for (i = 0; i<str.length(); i++)
		{
			ret_str+=toupper(str[i]);
		}

		return ret_str;
	}

	void SessionsManager::resetSessionManager()
	{
		map<JHI_SESSION_ID, SessionRecord, lt_sessionId>::iterator it;

		TRACE0("Resetting Session Manager");

		_locker.Lock();

		for ( it=_sessionList.begin() ; it != _sessionList.end(); it++ )
		{
			// release all events stored by the session 
			clearEventsQueue((*it).second.eventsDataQueue);
		}

		_sessionList.clear();
		_locker.UnLock();
	}

	void SessionsManager::getSessionInfo(JHI_SESSION_ID sessionId,JHI_SESSION_INFO* pSessionInfo)
	{
		if (pSessionInfo == NULL)
			return;

		_locker.Lock();

		if (!isSessionPresent(sessionId))
		{
			pSessionInfo->state = JHI_SESSION_STATE_NOT_EXISTS;
			pSessionInfo->flags = 0;
		}
		else
		{
			pSessionInfo->state = _sessionList[sessionId].state;
			pSessionInfo->flags =  _sessionList[sessionId].sessionFlags.value;
			// add other info here
		}

		_locker.UnLock();
	}

#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)

	JHI_SESSION_EXTENDED_INFO getSessionExtendedInfo(SessionRecord session)
	{
		JHI_SESSION_EXTENDED_INFO tmp;
		int appIdLength = session.appId.length();
		strcpy_s(tmp.appId, appIdLength + 1, session.appId.c_str());
		tmp.flags = session.sessionFlags.value;
		tmp.sessionId = session.sessionId;
		tmp.state = session.state;
		tmp.ownersListCount = session.ownersList.size();
		tmp.ownersList = JHI_ALLOC_T_ARRAY<JHI_PROCESS_INFORMATION>(tmp.ownersListCount);

		list<JHI_PROCESS_INFO>::iterator it;

		int ownersCounter;
		for ( it = session.ownersList.begin(), ownersCounter = 0 ; it != session.ownersList.end(); it++ , ++ownersCounter)
		{
			tmp.ownersList[ownersCounter].creationTime = it->creationTime;
			tmp.ownersList[ownersCounter].pid = it->pid;
		}

		return tmp;
	}

	void SessionsManager::getSessionsDataTable(JHI_SESSIONS_DATA_TABLE* pSessionsDataTable)
	{
		_locker.Lock();

		map<JHI_SESSION_ID, SessionRecord, lt_sessionId>::iterator it;

		//copying data to the buffer
		pSessionsDataTable->sessionsCount = _sessionList.size();
		pSessionsDataTable->dataTable = JHI_ALLOC_T_ARRAY<JHI_SESSION_EXTENDED_INFO>(_sessionList.size());

		int sessionCounter = 0;
		for ( it = _sessionList.begin() ; it != _sessionList.end(); ++it, ++sessionCounter )
		{
			pSessionsDataTable->dataTable[sessionCounter] = getSessionExtendedInfo(it->second);
		}

		_locker.UnLock();
	}

#endif

	bool SessionsManager::setEventHandle(JHI_SESSION_ID SessionId, JhiEvent* eventHandle)
	{
		bool status = true;

		_locker.Lock();

		do
		{
			if (!isSessionPresent(SessionId))
			{
				status = false;
				break;
			}
			if (_sessionList[SessionId].eventHandle != NULL)
			{	
#ifdef __ANDROID__
				TRACE2 ("JHI_SetEventHandler: socket counters tx %d rx %d\n", _sessionList[SessionId].eventHandle->tx_cnt, _sessionList[SessionId].eventHandle->rx_cnt);
#endif
				_sessionList[SessionId].eventHandle->close();
				JHI_DEALLOC_T(_sessionList[SessionId].eventHandle); // close old handle
				_sessionList[SessionId].eventHandle = NULL;
			}

			_sessionList[SessionId].eventHandle = eventHandle;

			// in case of unregister (eventHandle == NULL) we clear all session events queue
			if (eventHandle == NULL || !eventHandle->is_created())
				clearEventsQueue(_sessionList[SessionId].eventsDataQueue);

		}
		while(0);

		_locker.UnLock();

		return status;
	}

	bool SessionsManager::getEventHandle(JHI_SESSION_ID SessionId, JhiEvent** eventHandle)
	{
		bool status = true;

		if (NULL == eventHandle)
			return false;

		_locker.Lock();

		do
		{
			if (!isSessionPresent(SessionId))
			{
				status = false;
				break;
			}

			*eventHandle = _sessionList[SessionId].eventHandle;

			if (*eventHandle == NULL || !(*eventHandle)->is_created())
				status = false;

		}
		while(0);

		_locker.UnLock();

		return status;
	}

	bool SessionsManager::enqueueEventData(const JHI_SESSION_ID sessionId,JHI_EVENT_DATA* pEventData)
	{
		bool status = true;

		//char* sessionIdStr = NULL;

		//if (UuidToStringA(&sessionId, (RPC_CSTR*)&sessionIdStr) == RPC_S_OK)
		//{
		//	TRACE1("event session id: %s",sessionIdStr);
		//	RpcStringFreeA((RPC_CSTR*)&sessionIdStr);
		//}
		TRACE1("event data size: %d",pEventData->datalen);

		_locker.Lock();

		do
		{
			if (!isSessionPresent(sessionId))
			{
				TRACE0("failed to add event data into queue, the session does not exists.\n");
				status = false;
				break;
			}

			TRACE1("Event data size: %d",pEventData->datalen);
			TRACE1("Number of event in session events queue (before add): %d",_sessionList[sessionId].eventsDataQueue.size());

			if (_sessionList[sessionId].eventsDataQueue.size() >= MAX_EVENTS_DATA_IN_QUEUE)
			{
				TRACE0("failed to add event data into queue, the queue is full.\n");
				status = false;
				break;
			}

			_sessionList[sessionId].eventsDataQueue.push(pEventData);
			TRACE0("event added successfuly.\n");
		}
		while(0);

		_locker.UnLock();

		return status;
	}

	JHI_RET SessionsManager::getSessionEventData(const JHI_SESSION_ID SessionID,JHI_EVENT_DATA* pEventData)
	{
		JHI_RET ret = JHI_SUCCESS;
		JHI_EVENT_DATA* pEventDataFromQ = NULL;

		_locker.Lock();

		do
		{
			if (!isSessionPresent(SessionID))
			{
				ret = JHI_INVALID_SESSION_HANDLE;
				break;
			}

			if (_sessionList[SessionID].eventHandle == NULL ||
				!_sessionList[SessionID].eventHandle->is_created()) // not registerd for events
			{
				ret = JHI_INTERNAL_ERROR;
				break;
			}

			if (_sessionList[SessionID].eventsDataQueue.empty())
			{
				ret = JHI_GET_EVENT_FAIL_NO_EVENTS;
				break;
			}
			else
			{
				pEventDataFromQ = _sessionList[SessionID].eventsDataQueue.front();
				_sessionList[SessionID].eventsDataQueue.pop(); // remove the event from the queue

				pEventData->datalen = pEventDataFromQ->datalen;
				pEventData->dataType = pEventDataFromQ->dataType;


				if (pEventDataFromQ->data != NULL)
				{
					pEventData->data = (uint8_t*) JHI_ALLOC(pEventDataFromQ->datalen);
					if (pEventData->data == NULL) {
						TRACE0("malloc of event data failed .\n");
						ret = JHI_INTERNAL_ERROR;
						break;
					}

					memcpy_s(pEventData->data,pEventDataFromQ->datalen,pEventDataFromQ->data,pEventDataFromQ->datalen);
					JHI_DEALLOC(pEventDataFromQ->data);
					pEventDataFromQ->data = NULL;
				}
				JHI_DEALLOC(pEventDataFromQ);
				pEventDataFromQ = NULL;
			}
		}
		while(0);
		_locker.UnLock();

		return ret;
	}

	bool SessionsManager::GetSessionLock(JHI_SESSION_ID sessionID)
	{
		bool status = true;
		Locker* sessionLock = NULL;
		_locker.Lock();

		do
		{
			if (!isSessionPresent(sessionID))
			{
				status = false;
				break;
			}

			sessionLock = _sessionList[sessionID].sessionLock;

		}
		while(0);

		_locker.UnLock();

		if (sessionLock)
		{
			sessionLock->Lock();
			if (!isSessionPresent(sessionID))
			{
				status = false;
				sessionLock->UnLock();
			}
		}

		return status;
	}

	void SessionsManager::ReleaseSessionLock(JHI_SESSION_ID sessionID)
	{
		_locker.Lock();

		do
		{
			if (!isSessionPresent(sessionID))
				break;

			_sessionList[sessionID].sessionLock->UnLock();
		}
		while(0);

		_locker.UnLock();
	}

	void SessionsManager::clearEventsQueue(queue<JHI_EVENT_DATA*>&  eventsDataQueue)
	{
		JHI_EVENT_DATA* pEventDataFromQ = NULL;

		while (!eventsDataQueue.empty())
		{
			pEventDataFromQ = eventsDataQueue.front();
			eventsDataQueue.pop(); // remove the event from the queue

			if (pEventDataFromQ != NULL)
			{
				if (pEventDataFromQ->data != NULL)
				{
					JHI_DEALLOC(pEventDataFromQ->data);
					pEventDataFromQ->data = NULL;
				}

				JHI_DEALLOC(pEventDataFromQ);
				pEventDataFromQ = NULL;
			}
		}
	}

	// verify that a process does not exists in the OS
	// returns false if exists, true otherwise.
	bool processIsDead(const JHI_PROCESS_INFO & pinfo)
	{
		bool isDead = true;
#ifdef _WIN32
		FILETIME creationTime;

		TRACE1("verifing if the process with pid %d is alive\n",pinfo.pid);

		FILETIME unusedVar;
		DWORD exitCode;

		// get the process handle by its id
		HANDLE processHandle = OpenProcess(PROCESS_QUERY_INFORMATION,FALSE,pinfo.pid);

		do
		{
			if (processHandle == NULL)
			{
				TRACE0("OpenProcess returned NULL\n");
				break; // there is no such process with the given id
			}

			if (GetExitCodeProcess(processHandle,&exitCode) == FALSE)
			{
				TRACE0("failed to determine process state");
				isDead = false; // internal error, it is better to leave the session alive.
				break; 
			}

			if (exitCode != STILL_ACTIVE)
				break; // the process with the given pid is dead


			if (GetProcessTimes(processHandle,&creationTime,&unusedVar,&unusedVar,&unusedVar) == FALSE)
			{
				TRACE0("failed to get process creation time\n");
				isDead = false; // internal error, it is better to leave the session alive.
				break;
			}

			if ((creationTime.dwHighDateTime != pinfo.creationTime.dwHighDateTime) || (creationTime.dwLowDateTime != pinfo.creationTime.dwLowDateTime))
				break; // same process id but diffrent creation times => this is not the process that created the session.

			isDead = false;
		}
		while(0);

		if (processHandle != NULL)
			CloseHandle(processHandle);
#else //!_WIN32
			
		isDead = isProcessDead (pinfo.pid, const_cast< FILETIME &>(pinfo.creationTime));
#endif // _WIN32

		if (isDead)
		{
			TRACE1("DAL process with pid %d is dead\n",pinfo.pid);
		}
		else
		{
			TRACE1("DAL process with pid %d is alive\n",pinfo.pid);
		}
		return isDead;
	}


	bool SessionsManager::ClearSessionsDeadOwners()
	{
		bool removed = false;
		map<JHI_SESSION_ID, SessionRecord, lt_sessionId>::iterator it;

		_locker.Lock();

		// iterating over all the sessions
		for ( it=_sessionList.begin() ; it != _sessionList.end(); it++ )
		{
			if (ClearSessionDeadOwners(it->second.sessionId))
				removed = true;
		}

		_locker.UnLock();

		return removed;
	}

	bool SessionsManager::ClearSessionDeadOwners(JHI_SESSION_ID sessionID)
	{
		list<JHI_PROCESS_INFO>::iterator owner_it;
		bool ownerRemoved = false;
		size_t sizeBefore,sizeAfter;

		do
		{
			if (!isSessionPresent(sessionID))
				break;

			if (_sessionList[sessionID].ownersList.size() == 0)
				break;

			sizeBefore = _sessionList[sessionID].ownersList.size();

			_sessionList[sessionID].ownersList.remove_if(processIsDead);

			sizeAfter = _sessionList[sessionID].ownersList.size();

			if (sizeAfter < sizeBefore)
			{
				TRACE2("Removed abandoned sessions from session id [%s], owners count: %d\n",sessionIdToString(sessionID).c_str(),sizeAfter);
				ownerRemoved = true;
				updateSessionLastUsage(&_sessionList[sessionID]);
			}
		}
		while(0);

		return ownerRemoved;
	}

	bool SessionsManager::AppletHasNonSharedSessions(const string& appId)
	{
		map<JHI_SESSION_ID, SessionRecord,lt_sessionId>::iterator it;
		bool status = false;
		string appid = toUpperCase(appId);

		_locker.Lock();

		for ( it=_sessionList.begin() ; it != _sessionList.end(); it++ )
		{
			if ((appid == it->second.appId) && (!it->second.sessionFlags.bits.sharedSession))
			{
				status = true;
				break;
			}
		}

		_locker.UnLock();

		return status;
	}

	bool SessionsManager::ClearAppletSharedSession(const string& appId)
	{
		map<JHI_SESSION_ID, SessionRecord,lt_sessionId>::iterator it;
		bool removed = false;
		string appid = toUpperCase(appId);

		for ( it=_sessionList.begin() ; it != _sessionList.end(); it++ )
		{
			// find the applet shared session. if exist and has no owners, try to close it.
			if ((appid == it->second.appId) && it->second.sessionFlags.bits.sharedSession && it->second.ownersList.empty())
			{
				if (jhis_close_session(&it->second.sessionId,NULL,false,true) == JHI_SUCCESS)
				{
					TRACE0("abandoned shared session removed\n");
					removed = true;
				}
				else
				{
					TRACE0("failed to remove abandoned shared session\n");
				}
				goto end;
			}
		}
end:
		return removed;
	}

	bool SessionsManager::ClearAbandonedNonSharedSessions()
	{
		bool removed = false;
		map<JHI_SESSION_ID, SessionRecord, lt_sessionId>::iterator it;
		list<JHI_SESSION_ID> slist;
		list<JHI_PROCESS_INFO>::iterator owner_it;

		_locker.Lock();

		// create a list of all non-shared sessions that has no owners
		for ( it=_sessionList.begin() ; it != _sessionList.end(); it++ )
		{
			if (it->second.ownersList.empty() && (!it->second.sessionFlags.bits.sharedSession))
				slist.push_back(it->first);
		}

		_locker.UnLock();

		// call for close session API with each of the sessions in the list

		JHI_SESSION_ID tmpHandle;
		for (list<JHI_SESSION_ID>::iterator sit = slist.begin(); sit != slist.end(); sit++)
		{
			tmpHandle = *sit;
			if (jhis_close_session(&tmpHandle,NULL,false,true) == JHI_SUCCESS)
			{
				removed = true;
			}
			else
			{
				TRACE0("failed to remove a non-shared session of a dead application\n");
			}
		}
		return removed;
	}

	bool SessionsManager::TryRemoveUnusedSharedSession(bool allowNonSharedSessions)
	{
		bool removed = false;
		map<JHI_SESSION_ID, SessionRecord, lt_sessionId>::iterator it;
		list<JHI_SESSION_ID> slist;
		JHI_SESSION_ID SessionToRemove;
		unsigned long SessionTimeStamp;

		// Create a list of all shared sessions that have no owners
		for ( it=_sessionList.begin() ; it != _sessionList.end(); it++ )
		{
			if (it->second.ownersList.empty() && (it->second.sessionFlags.bits.sharedSession))
			{
				if (allowNonSharedSessions || (!AppletHasNonSharedSessions(it->second.appId)))
					slist.push_back(it->first);
			}
		}

		if (slist.size() == 0)
			return false;

		SessionToRemove = *(slist.begin());
		SessionTimeStamp = _sessionList[SessionToRemove].lastUsedTime;

		for (list<JHI_SESSION_ID>::iterator sit = slist.begin(); sit != slist.end(); sit++)
		{
			if (_sessionList[*sit].lastUsedTime < SessionTimeStamp)
			{
				SessionToRemove = *sit;
				SessionTimeStamp = _sessionList[*sit].lastUsedTime;
			}
		}

		if (jhis_close_session(&SessionToRemove,NULL,false,true) == JHI_SUCCESS)
		{
			removed = true;
		}
		else
		{
			TRACE0("ERROR: failed to remove a shared session that has no owners\n");
		}

		return removed;
	}

	bool SessionsManager::addSessionOwner(JHI_SESSION_ID sessionID,JHI_PROCESS_INFO* info)
	{
		list<JHI_PROCESS_INFO>::iterator owner_it;
		bool ownerAdded = false;

		_locker.Lock();

		do
		{
			if (!isSessionPresent(sessionID))
				break;

			if (_sessionList[sessionID].ownersList.size() >= MAX_SESSION_OWNERS)
				break;

			_sessionList[sessionID].ownersList.push_back(*info);
			ownerAdded = true;

			TRACE2("Session owner added to shared session [%s], owners count: %d\n",sessionIdToString(sessionID).c_str(),_sessionList[sessionID].ownersList.size());

		}
		while(0);

		_locker.UnLock();

		return ownerAdded;
	}

	bool SessionsManager::removeSessionOwner(JHI_SESSION_ID sessionID,JHI_PROCESS_INFO* info)
	{
		bool removed = false;
		list<JHI_PROCESS_INFO>::iterator owner_it;

		_locker.Lock();

		do
		{
			if (!isSessionPresent(sessionID))
				break;

			for ( owner_it=_sessionList[sessionID].ownersList.begin() ; owner_it != _sessionList[sessionID].ownersList.end(); owner_it++ )
			{
				if (owner_it->pid == info->pid && 
					owner_it->creationTime.dwHighDateTime == info->creationTime.dwHighDateTime &&
					owner_it->creationTime.dwLowDateTime == info->creationTime.dwLowDateTime)
				{
					removed = true;
					_sessionList[sessionID].ownersList.erase(owner_it);
					TRACE2("Session owner removed from shared session [%s], owners count: %d\n",sessionIdToString(sessionID).c_str(),_sessionList[sessionID].ownersList.size());

					updateSessionLastUsage(&_sessionList[sessionID]);
					break;
				}
			}

		} while (0);

		_locker.UnLock();

		return removed;
	}

	bool SessionsManager::isSessionOwnerValid(JHI_SESSION_ID sessionID, JHI_PROCESS_INFO* info)
	{
		bool removed = false;
		list<JHI_PROCESS_INFO>::iterator owner_it;

		_locker.Lock();

		do
		{
			if (!isSessionPresent(sessionID))
				break;

			for ( owner_it=_sessionList[sessionID].ownersList.begin() ; owner_it != _sessionList[sessionID].ownersList.end(); owner_it++ )
			{
				if (owner_it->pid == info->pid && 
					owner_it->creationTime.dwHighDateTime == info->creationTime.dwHighDateTime &&
					owner_it->creationTime.dwLowDateTime == info->creationTime.dwLowDateTime)
				{
					removed = true;
					break;
				}
			}

		} while (0);

		_locker.UnLock();

		return removed;
	}

	int SessionsManager::getOwnersCount(JHI_SESSION_ID sessionID)
	{
		int count = -1;

		_locker.Lock();

		do
		{
			if (!isSessionPresent(sessionID))
				break;

			count = (int)_sessionList[sessionID].ownersList.size();

		} while (0);

		_locker.UnLock();

		return count;
	}

	bool SessionsManager::getSharedSessionID(JHI_SESSION_ID* sessionId, const string& appId)
	{
		bool exists = false;

		map<JHI_SESSION_ID, SessionRecord, lt_sessionId>::iterator it;

		_locker.Lock();

		// iterating over all the sessions
		for ( it=_sessionList.begin() ; it != _sessionList.end(); it++ )
		{
			if (it->second.appId == appId && it->second.sessionFlags.bits.sharedSession)
			{
				*sessionId = it->first;
				exists = true;
				break;
			}
		}

		_locker.UnLock();

		return exists;
	}

	void SessionsManager::updateSessionLastUsage(SessionRecord* sessionRecord)
	{
		if (sessionRecord == NULL)
			return;

		// we update the LRU counter whenever the session is a Shared Session and it has no owners.
		if ((sessionRecord->ownersList.size()==0) && (sessionRecord->sessionFlags.bits.sharedSession))
		{
			sessionRecord->lastUsedTime = sharedSessionLRUCounter;
			sharedSessionLRUCounter++;
			TRACE2("update shared session [%s] last used time to: %d\n",sessionIdToString(sessionRecord->sessionId).c_str(),sessionRecord->lastUsedTime);
		}
	}
}
