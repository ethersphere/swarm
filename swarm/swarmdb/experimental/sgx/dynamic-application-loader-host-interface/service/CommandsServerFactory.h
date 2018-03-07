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
**    @file CommandsServerFactory.h
**
**    @brief  Contains factory design pattern that creates ICommandsServer instances
**
**    @author Elad Dabool
**
********************************************************************************
*/
#ifndef _COMMANDS_SERVER_FACTORY_H_
#define _COMMANDS_SERVER_FACTORY_H_

#include "typedefs.h"
#include "misc.h"
#include "ICommandsServer.h"
#include "CommandDispatcher.h"

#ifdef _WIN32
#include "CommandsServerSocketsWin32.h"
#elif defined(__ANDROID__)
#include "CommandsServerSocketsAndroid.h"
#elif defined(__linux__)
#include "CommandsServerSocketsLinux.h"
#else
Unknown OS
#endif


namespace intel_dal
{
	class CommandsServerFactory
	{
	public:
		static ICommandsServer* createInstance()
		{

//#ifdef CS_SOCKETS_COMMUNICATION
			CommandDispatcher* dispatcher = JHI_ALLOC_T(CommandDispatcher); // is there a problem with inheritance?

#ifdef _WIN32
			return new CommandsServerSocketsWin32(dispatcher,JHI_MAX_CLIENTS_CONNECTIONS); 
#elif defined(__ANDROID__)
			return new CommandsServerSocketsAndroid(dispatcher,JHI_MAX_CLIENTS_CONNECTIONS);
#elif defined(__linux__)
			return new CommandsServerSocketsLinux(dispatcher,JHI_MAX_CLIENTS_CONNECTIONS);
#else
            UNKNOWN PLATFORM
#endif //_WIN32

//#else
			// other communication types here
//#endif //CS_SOCKETS_COMMUNICATION
		}
	};

}

#endif
