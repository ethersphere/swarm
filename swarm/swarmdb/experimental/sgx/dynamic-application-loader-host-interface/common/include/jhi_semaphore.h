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

#ifndef __SEMAPHORE_H
#define __SEMAPHORE_H
#include "typedefs.h"

#ifdef _WIN32
#include "windows.h"
#else
#include <semaphore.h>
#endif

namespace intel_dal
{
	/**
	This class will be used to handle critical sections in our code by,
	providing a cross platform semaphore mechanizem.
	**/
	class Semaphore
	{
	public:

		// This Default Constructor will create a Locker with a lock on creation 
		Semaphore(uint8_t semaphore_count);

		// Destructor
		// Release active locks when being disposed.
		~Semaphore(void);

		// This function is used in order to manually acquire a lock.
		// Use this function above a critical section.
		void Acquire();

		// This function is used in order to manually unlock an lock.
		// Use this function at an end of a critical section.
		void Release();

	private:

	#ifdef _WIN32
		HANDLE win32semaphore;
	#else
		sem_t linuxsemaphore;
	#endif

	};
}

#endif