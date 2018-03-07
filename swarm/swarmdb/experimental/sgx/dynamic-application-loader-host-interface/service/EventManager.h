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

#ifndef __EVENTMANAGER_H
#define __EVENTMANAGER_H

#ifdef _WIN32
#include <windows.h>
#endif
// The H-Files
#include <string>
#include <vector>
#include "jhi.h"
#include "jhi_i.h"
#include "typedefs.h"
#include "Singleton.h"
#include "GlobalsManager.h"



namespace intel_dal
{
	using std::string;
	using std::vector;

// Spooler Applet
//#define SPOOLER_APPLET_UUID "BA8D164350B649CC861D2C01BED14BE8"
#define SPOOLER_APPLET_FILENAME FILEPREFIX("/SpoolerApplet")

/**
This Class Will Manage the handling of events raised from applets to applications
using the spooler applet.
**/
class EventManager : public Singleton<EventManager>
{
	friend class Singleton<EventManager>;

public:

	/** 
		This init function initializes the event mechanism by
		downloading the spooler applet to JOM and opening a session to it,
		and in addition starts the event listener thread.

		Return:
			JHI_SUCCESS if succeded, error code otherwise
	**/	
	JHI_RET Initialize();

	/**
		This function closes the spooler session in JOM
		resulting the spooler thread to reset JHI.
		This function is called when a heci disable event is recieved by the JHI service.
	**/
	void closeSpoolerSession();

	bool GetSpoolerFullFilename(OUT FILESTRING& outFileName, OUT bool &isAcp);

	void Deinit();

	/** 
		this function set an eventHandler of a session
		in order to be later used when an event is raised from an applet
		to that session

		Parameters:

			IN SessionID - Session ID
			IN HandleName - the event handle name.  if NULL, event will be removed.

		Return:
			true is init succeded, false otherwise.
	**/	
	JHI_RET SetSessionEventHandler(JHI_SESSION_ID SessionID, char* eventHandleName);

	VM_SESSION_HANDLE spooler_handle;

private:
#ifdef _WIN32
	HANDLE listenerThreadHandle;
#else
	pthread_t listenerThreadHandle;
#endif //WIN32
	bool initialized;

	JHI_RET InstallSpooler(const FILESTRING &spoolerFile, bool isAcp);
	JHI_RET CreateSpoolerSession(const FILESTRING &spoolerFile, bool isAcp);
	JHI_RET CreateListenerThread();

	// Default Constructor
	EventManager(void);

	// Destructor 
	~EventManager(void);
};

}

#endif