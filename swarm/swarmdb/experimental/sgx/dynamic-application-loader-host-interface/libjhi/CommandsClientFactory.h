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
**    @file CommandsClientFactory.h
**
**    @brief  Contains factory design pattern that creates ICommandsClient instances
**
**    @author Elad Dabool
**
********************************************************************************
*/
#ifndef _COMMANDS_CLIENT_FACTORY_H_
#define _COMMANDS_CLIENT_FACTORY_H_

#include "typedefs.h"
#include "ICommandsClient.h"
#include "misc.h"

#ifdef _WIN32
#include "CommandsClientSocketsWin32.h"
#elif defined(__ANDROID__)
#include "CommandsClientSocketsAndroid.h"
#elif defined(__linux__)
#include "CommandsClientSocketsLinux.h"
#else
Unknown OS
#endif


namespace intel_dal
{

	class CommandsClientFactory
	{
	public:
		ICommandsClient* createInstance()
		{
			ICommandsClient* instance = NULL;

//	#ifdef CS_SOCKETS_COMMUNICATION

		#ifdef _WIN32
			instance = JHI_ALLOC_T(CommandsClientSocketsWin32); // is there a problem with inheritance?
		#elif defined(__ANDROID__)
			instance = JHI_ALLOC_T(CommandsClientSocketsAndroid);
		#elif defined(__linux__)
			instance = JHI_ALLOC_T(CommandsClientSocketsLinux);
        #else
            Unknown OS
		#endif // _WIN32

//	#else
//		// other communication types here
//	#endif
		
			return instance;
		}
	};

}
#endif
