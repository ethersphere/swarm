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

#include "CommandsClientSocketsAndroid.h"
#include "dbg.h"
#include "reg.h"
#include "misc.h"
#include <iostream>
#include <sstream>
#include <netdb.h>
#include <errno.h>
#include <string.h>
#include <netinet/in.h>
#include <cutils/sockets.h>

#define INVALID_SOCKET -1
#define SOCKET_ERROR   -1

#define SOCKET_NAME "jhid"

namespace intel_dal
{
	CommandsClientSocketsAndroid::CommandsClientSocketsAndroid()
	{
		_socket = INVALID_SOCKET;
	}

	CommandsClientSocketsAndroid::~CommandsClientSocketsAndroid()
	{
		if (_socket != INVALID_SOCKET)
		{
			close(_socket);
			_socket = INVALID_SOCKET;
		}
	}

	bool CommandsClientSocketsAndroid::Connect()
	{
		bool status = false;

		do
		{
			_socket = socket_local_client(SOCKET_NAME, ANDROID_SOCKET_NAMESPACE_RESERVED,
						       SOCK_STREAM);
			if (_socket < 0)
			{
				TRACE0("failed to get control socket\n");
				if (_socket != INVALID_SOCKET)
				{
				      close(_socket);
				      _socket = INVALID_SOCKET;
				}
				break;
			}

			status = true;

		}
		while (0);

		return status;
	}

	bool CommandsClientSocketsAndroid::Disconnect()
	{
		bool status = false;
		if (close(_socket) != SOCKET_ERROR)
		{
			status = true;
		}

		_socket = INVALID_SOCKET;
		return status;
	}

	bool CommandsClientSocketsAndroid::Invoke(IN const uint8_t *inputBuffer,
						   IN uint32_t inputBufferSize,
					           OUT uint8_t **outputBuffer,
					           OUT uint32_t* outputBufferSize)
	{
		int iResult;
		uint8_t* RecvOutBuff = NULL;

		if (inputBufferSize == 0 || inputBuffer == NULL
		    || outputBuffer == NULL || outputBufferSize == NULL)
			return false;


		// sending the InputBufferSize
		iResult = blockedSend(_socket, (uint8_t*)&inputBufferSize, sizeof(uint32_t));
		if (iResult != sizeof(uint32_t))
		{
			TRACE1("send inputBufferSize failed: %d\n", errno);
			return false;
		}

		// sending the InputBuffer
		iResult = blockedSend(_socket, (uint8_t*)inputBuffer, inputBufferSize);
		if (iResult != inputBufferSize)
		{
			TRACE1("send inputBuffer failed: %d\n", errno);
			return false;
		}

		// Receive until the peer closes the connection
		iResult = blockedRecv(_socket, (uint8_t*) outputBufferSize, sizeof(uint32_t));
		if (iResult !=  sizeof(uint32_t))
		{
			TRACE1("recv outputBufferSize failed: %d\n", errno);
			return false;
		}

		if ((*outputBufferSize >= sizeof(JHI_RESPONSE))
		   && (*outputBufferSize < JHI_MAX_TRANSPORT_DATA_SIZE))
		{
			// allocate new buffer
			RecvOutBuff = (uint8_t*)JHI_ALLOC(*outputBufferSize);
			if (NULL == RecvOutBuff)
			{
				TRACE0("failed to allocate outputBufferSize memory.");
				return false;
			}

			iResult = blockedRecv(_socket, RecvOutBuff, *outputBufferSize);
			if (iResult !=  *outputBufferSize)
			{
				TRACE1("recv RecvOutBuff failed: %d\n", errno);
				JHI_DEALLOC(RecvOutBuff);
				return false;
			}
		}
		else
		{
			TRACE0("invalid response recieved from JHI service");
			return false;
		}

		*outputBuffer = RecvOutBuff;

		return true;
	}

	uint32_t CommandsClientSocketsAndroid::blockedRecv(SOCKET socket, uint8_t* buffer,
							    uint32_t length)
	{
		uint32_t bytesRecieved = 0;
		int count;

		while (bytesRecieved != length)
		{
				count = recv(socket, buffer + bytesRecieved,
					     length - bytesRecieved, 0);

				if (count == SOCKET_ERROR || count == 0) // JHI service closed the connection
					break;

				bytesRecieved += count;
		}

		return bytesRecieved;
	}

	uint32_t CommandsClientSocketsAndroid::blockedSend(SOCKET socket, uint8_t* buffer,
							    uint32_t length)
	{
		uint32_t bytesSent = 0;
		int count;

		while (bytesSent != length)
		{
				count = send(socket, buffer + bytesSent, length - bytesSent, 0);

				if (count == SOCKET_ERROR) // JHI service closed the connection
					break;

				bytesSent += count;
		}

		return bytesSent;
	}
}//namespace intel_dal
