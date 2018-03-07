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
**    @file CommandsClientSocketsAndroid.h
**
**    @brief  Contains sockets implementation of ICommandsClient
**
**    @author Alexander Usyskin
**
********************************************************************************
*/
#ifndef _COMMANDS_CLIENT_SOCKETS_LINUX_H_
#define _COMMANDS_CLIENT_SOCKETS_LINUX_H_

#include "typedefs.h"
#include "CSTypedefs.h"
#include "ICommandsClient.h"
#include <unistd.h>

namespace intel_dal
{

	typedef int SOCKET;

	class CommandsClientSocketsAndroid : public ICommandsClient
	{
	private:
		SOCKET _socket;
	public:
	   CommandsClientSocketsAndroid();
	   ~CommandsClientSocketsAndroid();
	   bool Connect();
	   bool Disconnect();
	   bool Invoke(IN const uint8_t *inputBuffer, IN uint32_t inputBufferSize,
		OUT uint8_t **outputBuffer, OUT uint32_t* outputBufferSize);
	private:
	   uint32_t blockedRecv(SOCKET socket, uint8_t *buffer, uint32_t length);
	   uint32_t blockedSend(SOCKET socket, uint8_t *buffer, uint32_t length);
	};
}

#endif /* _COMMANDS_CLIENT_SOCKETS_LINUX_H_ */
