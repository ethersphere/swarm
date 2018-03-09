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

#include "ReadWriteLock.h"

namespace intel_dal
{

	ReadWriteLock::ReadWriteLock()
	{
		_readersCount = 0;

		// create a no_readers condition
		pthread_cond_init(&no_readers_cond, NULL);
	}

	ReadWriteLock::~ReadWriteLock()
	{
		// before closing the readers event, verify that no one is using it.
		_readLock.Lock();
		while (_readersCount > 0)
		{
			pthread_cond_wait(&no_readers_cond, &_readLock.linuxmutex);
		}
		_readLock.UnLock();
		pthread_cond_destroy(&no_readers_cond);
	}

	void ReadWriteLock::aquireReaderLock()
	{
		// first, wait for writer (is exists) to do his job in order to prevent
		// writer starvation
		_writeLock.Lock();

		// try to get a read lock
		_readLock.Lock();

		// increment the readers count
		_readersCount++;

		// free the read lock
		_readLock.UnLock();

		// free the writer lock
		_writeLock.UnLock();

	}

	void ReadWriteLock::releaseReaderLock()
	{
		_readLock.Lock();

		// decrement the readers count
		_readersCount--;

		if (_readersCount == 0) // signal writer that there are no more readers
			pthread_cond_signal(&no_readers_cond);

		_readLock.UnLock();
	}

	void ReadWriteLock::aquireWriterLock()
	{
		_writeLock.Lock();

		// at this point, we ensured that no new readers will enter,
		// now we wait until all the readers are done
		_readLock.Lock();
		while (_readersCount > 0)
		{
			pthread_cond_wait(&no_readers_cond, &_readLock.linuxmutex);
		}
		_readLock.UnLock();
	}

	void ReadWriteLock::releaseWriterLock()
	{
		// release all waiting readers.
		_writeLock.UnLock();
	}
}
