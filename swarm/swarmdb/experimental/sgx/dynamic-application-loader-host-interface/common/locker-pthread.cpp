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

#include "Locker.h"

using namespace intel_dal;

Locker::Locker(void)
{
	// By Default Lock is unlock
	Init();
}

Locker::Locker(bool lockOnCreation)
{
	Init();

	if (lockOnCreation)
		Lock();
}

void Locker::Init()
{
	pthread_mutexattr_t attr;
	pthread_mutexattr_init(&attr);
	pthread_mutexattr_settype(&attr, PTHREAD_MUTEX_RECURSIVE);
	pthread_mutex_init(&linuxmutex, &attr);
    pthread_mutexattr_destroy(&attr);
}

Locker::~Locker(void)
{
	UnLock();

	pthread_mutex_destroy(&linuxmutex);
}

void Locker::Lock()
{
	pthread_mutex_lock(&linuxmutex);
}

void Locker::UnLock()
{
	pthread_mutex_unlock(&linuxmutex);
}

