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

#ifndef __LOCKER_H
#define __LOCKER_H

#include "typedefs.h"
#include "dbg.h"
#include "misc.h"

#ifdef _WIN32
#include "windows.h"
#else
 #include <pthread.h>
#endif

namespace intel_dal
{
#ifndef _WIN32
class ReadWriteLock;
#endif
/**
This class will be used to handle critical sections in our code by,
providing an easy Lock-Release (MUTEX) mechanizem.
**/
class Locker
{
public:

	// This Defualt Constructor will create a Locker with a lock on creation 
	Locker(void);

	// This Constractor allows to choose whether the Locker will acquire a lock when created or not.
	Locker(bool lockOnCreation);

	// Destructor 
	// Release active locks when being disposed.
	~Locker(void);	

	// This function is used in order to manually acquire a lock.
	// Use this function above a critical section.
	void Lock();

	// This function is used in order to manually unlock an lock.
	// Use this function at an end of a critical section.
	void UnLock();

private:

#ifdef _WIN32
	HANDLE win32mutex;
#else
	pthread_mutex_t linuxmutex;
	friend class ReadWriteLock;
#endif

	void Init();

};

}

#endif