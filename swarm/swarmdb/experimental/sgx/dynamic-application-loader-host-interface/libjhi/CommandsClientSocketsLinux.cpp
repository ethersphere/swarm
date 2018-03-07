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

#include "CommandsClientSocketsLinux.h"
#include "dbg.h"
#include "reg.h"
#include <iostream>
#include <sstream>
#include <sys/types.h>
#include <sys/socket.h>
#include <netdb.h>
#include <errno.h>
#include <string.h>
#include <sys/un.h>
#include <cstddef>
#include <string>
#include <linux/limits.h>

using namespace std;

namespace intel_dal
{

CommandsClientSocketsLinux::CommandsClientSocketsLinux()
{
	_socket = INVALID_SOCKET;
}

CommandsClientSocketsLinux::~CommandsClientSocketsLinux()
{
	if (_socket != INVALID_SOCKET)
	{
		close(_socket);
		_socket = INVALID_SOCKET;
	}
}

bool CommandsClientSocketsLinux::Connect()
{
	bool status = false;

	sockaddr_un addr;

	char socket_path[PATH_MAX] = {0};
	JhiQueryDaemonSocketPathFromRegistry(socket_path);

	do
	{
        _socket = socket(AF_UNIX, SOCK_STREAM, PF_UNSPEC);

		if (_socket == INVALID_SOCKET)
		{
			TRACE1("Couldn't create a socket. error: %d\n", errno);
			break;
		}

        addr.sun_family = AF_UNIX;

        if (strlen(socket_path)+1 > sizeof(addr.sun_path))
        {
        	status = false;
            LOG1("socket path too long. path: %s", socket_path);
            break;
        }
        strncpy(addr.sun_path, socket_path, sizeof(addr.sun_path));

		if (connect(_socket, (sockaddr *)&addr, offsetof(sockaddr_un, sun_path) + strlen(addr.sun_path) + 1) == SOCKET_ERROR)
		{
			TRACE1("connection failed. error: %d\n", errno);
			break;
		}

		status = true;

	}
	while (0);

	//cleanup
	if (status == false)
	{
		if (_socket != INVALID_SOCKET)
		{
			close(_socket);
			_socket = INVALID_SOCKET;
		}
	}

	return status;
}

bool CommandsClientSocketsLinux::Disconnect()
{
	if (close(_socket) == SOCKET_ERROR)
	{
		_socket = INVALID_SOCKET;
		return false;
	}

	_socket = INVALID_SOCKET;
	return true;
}

bool CommandsClientSocketsLinux::Invoke(IN const uint8_t *inputBuffer, IN uint32_t inputBufferSize,
			OUT uint8_t **outputBuffer, OUT uint32_t* outputBufferSize)
{
	int iResult;
	char* RecvOutBuff = NULL;

	if (inputBufferSize == 0 || inputBuffer == NULL || outputBuffer == NULL || outputBufferSize == NULL)
		return false;


	// sending the InputBufferSize
	iResult = blockedSend(_socket,(char*) &inputBufferSize, sizeof(uint32_t));
	if (iResult != sizeof(uint32_t))
	{
		TRACE1("send inputBufferSize failed: %d\n", errno);
		return false;
	}

	// sending the InputBuffer
	iResult = blockedSend(_socket,(char*) inputBuffer, inputBufferSize);
	if (iResult != (int32_t)inputBufferSize)
	{
		TRACE1("send inputBuffer failed: %d\n", errno);
		return false;
	}

	// Receive until the peer closes the connection
	iResult = blockedRecv(_socket, (char*) outputBufferSize, sizeof(uint32_t));
	if (iResult !=  sizeof(uint32_t))
	{
		TRACE1("recv outputBufferSize failed: %d\n", errno);
		return false;
	}

	if ((*outputBufferSize >= sizeof(JHI_RESPONSE)) && (*outputBufferSize < JHI_MAX_TRANSPORT_DATA_SIZE))
	{
		// allocate new buffer
		try {
			RecvOutBuff = new char[*outputBufferSize];
		}
		catch (...) {
			LOG0("failed to allocate outputBufferSize memory.");
			return false;
		}

		iResult = blockedRecv(_socket, RecvOutBuff, *outputBufferSize);
		if (iResult !=  (int32_t)*outputBufferSize)
		{
			TRACE1("recv RecvOutBuff failed: %d\n", errno);
			delete [] RecvOutBuff;
			return false;
		}
	}
	else
	{
		TRACE0("invalid response recieved from JHI service");
		return false;
	}

	*outputBuffer = (uint8_t*) RecvOutBuff;

	return true;
}

int CommandsClientSocketsLinux::blockedRecv(SOCKET socket, char* buffer, int length)
{
	int bytesRecieved = 0;
	int count;

	while (bytesRecieved != length)
	{
			count = recv(socket, buffer + bytesRecieved, length - bytesRecieved, 0);

			if (count == SOCKET_ERROR || count == 0) // JHI service closed the connection
				break;

			bytesRecieved += count;
	}

	return bytesRecieved;
}

int CommandsClientSocketsLinux::blockedSend(SOCKET socket, char* buffer, int length)
{
	int bytesSent = 0;
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
}

