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
**    @file ICommandsServer.h
**
**    @brief  Contains interface for commands server
**
**    @author Elad Dabool
**
********************************************************************************
*/
#ifndef _ICOMMANDSSERVER_H_
#define _ICOMMANDSSERVER_H_

#include "typedefs.h"
#include "ICommandDispatcher.h"
#include "jhi_semaphore.h"

namespace intel_dal
{

	class ICommandsServer
	{
	protected:
		uint8_t _maxClientNum;

		Semaphore* _semaphore;

	public:

		ICommandDispatcher* _dispatcher;

		ICommandsServer(ICommandDispatcher* dispatcher, uint8_t maxClientNum) :
		_semaphore(new Semaphore(maxClientNum))
		{ 
			_dispatcher = dispatcher; 
			_maxClientNum = maxClientNum;
		};

		virtual bool open() = 0;
		virtual bool close() = 0;
		virtual void waitForRequests() = 0;

		Semaphore* getSemaphore()
		{
			return _semaphore;
		}

		virtual ~ICommandsServer() 
		{
			if (NULL != _semaphore)
				delete _semaphore;
			if (NULL != _dispatcher)
				delete _dispatcher;
		}
	};

}
#endif