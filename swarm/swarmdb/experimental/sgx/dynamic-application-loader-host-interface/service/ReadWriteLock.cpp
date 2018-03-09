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

		// create a no_readers event, initialize it as signaled
		// in case a writer is locking before any readers.
		no_readers_event = CreateEvent(NULL,TRUE,TRUE,NULL);
	}

	ReadWriteLock::~ReadWriteLock()
	{
		// before closing the readers event, verify that no one is using it.
		WaitForSingleObject(no_readers_event,INFINITE);
		CloseHandle(no_readers_event);
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

		if (_readersCount == 1)	// when there is a reader, mark the no_readers_event to non-signaled.
			ResetEvent(no_readers_event);


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

		if (_readersCount == 0)	// signal writer that there are no more readers
			SetEvent(no_readers_event);

		_readLock.UnLock();
	}

	void ReadWriteLock::aquireWriterLock()
	{
		_writeLock.Lock();

		// at this point, we ensured that no new readers will enter,
		// now we wait until all the readers are done

		WaitForSingleObject(no_readers_event,INFINITE);
	}

	void ReadWriteLock::releaseWriterLock()
	{
		// release all waiting readers.
		_writeLock.UnLock();
	}
}