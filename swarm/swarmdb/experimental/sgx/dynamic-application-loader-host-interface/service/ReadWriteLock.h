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

#ifndef __READ_WRITE_LOCK_H
#define __READ_WRITE_LOCK_H

// The H-Files
#include "Locker.h"
#include "jhi_i.h"

#ifdef _WIN32
#include "windows.h"
#else //!WIN32
#include <pthread.h>
#endif //WIN32

namespace intel_dal
{
	/**
		This class is an implementation of read/write lock pattern in order to sync a shared object
		between multiple readers and one writer.
		for more info: http://en.wikipedia.org/wiki/Readers-writer_lock
	**/
	class ReadWriteLock
	{
	private:

		Locker	_readLock;
		Locker	_writeLock;
		
		UINT32 _readersCount;	// the number of readers that has aquired a lock

#ifdef _WIN32
		HANDLE no_readers_event;
#else
		pthread_cond_t no_readers_cond;
#endif

		// disabling copy constructor and assignment operator by declaring them as private
		ReadWriteLock&  operator = (const ReadWriteLock& other) { return *this; }
		ReadWriteLock(const ReadWriteLock& other) { }

	public:

		ReadWriteLock();
		~ReadWriteLock();

		/**
			Recieve a readers lock, we MUST use this before using the shared object
		**/
		void aquireReaderLock();

		/**
			Release a readers lock, we MUST use this after using the shared object
		**/
		void releaseReaderLock();

		/**
			Recieve a writer lock in order to update the shared object.
			we MUST use this before updating shared object
		**/
		void aquireWriterLock();

		/**
			Release a writer lock and let waiting readers to use the updated shared object
			we MUST use this after updating shared object
		**/
		void releaseWriterLock();
	};
}

#endif 

