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

#include "CommandsClientSocketsWin32.h"
#include "dbg.h"
#include "reg.h"
#include "misc.h"
#include <iostream>
#include <sstream>
#include <string>
using namespace std;

namespace intel_dal
{
	CommandsClientSocketsWin32::CommandsClientSocketsWin32()
	{
		_socket = INVALID_SOCKET;

		if (WSAStartup(MAKEWORD(2,2), &_wsaData) != 0)
		{
			TRACE0("Error: WSAStartup failed - failed to initialize sockets interface\n");
			throw std::exception("Error: failed to initialize sockets interface");
		}
	}

	CommandsClientSocketsWin32::~CommandsClientSocketsWin32()
	{
		if (_socket != INVALID_SOCKET)
		{
			closesocket(_socket);
			_socket = INVALID_SOCKET;
		}

		int ret = WSACleanup();

#ifdef DEBUG
		if (ret != 0)
			TRACE0("WSACleanup failed.\n");
#endif
	}

	bool CommandsClientSocketsWin32::Connect()
	{
		bool status = false;
		struct addrinfo hints;
		struct addrinfo *result = NULL;	
		uint32_t portNumber;
		uint32_t addressType;
		int ret;
		std::stringstream sstream;


		do
		{
			ret = JhiQueryServicePortFromRegistry(&portNumber);
			if (ret!= JHI_SUCCESS)
			{
				TRACE0("failed to get port number from registry\n");
				break;
			}

			// convert port number to string	
			sstream << portNumber;
			string portString = sstream.str();

			ret = JhiQueryAddressTypeFromRegistry(&addressType);
			if (ret!= JHI_SUCCESS)
			{
				TRACE0("failed to get address type from registry\n");
				break;
			}

			if ((addressType != AF_INET) && (addressType != AF_INET6))
			{
				TRACE0("invalid address type recieved from registry\n");
				break;
			}

			ZeroMemory( &hints, sizeof(hints) );
			hints.ai_socktype = SOCK_STREAM;
			hints.ai_protocol = IPPROTO_TCP;
			hints.ai_family = addressType;

			if (getaddrinfo("localhost",portString.c_str(),&hints,&result) != 0)
			{
				TRACE0("failed to get adderss info\n");
				break;
			}

			if (result == NULL)
			{
				TRACE0("no adderss info recieved\n");
				break;
			}

			_socket = socket(result->ai_family, result->ai_socktype, result->ai_protocol);

			if (_socket == INVALID_SOCKET)
			{
				TRACE1("Couldn't create a socket. error: %d\n", WSAGetLastError());
				break;
			}  

			if (connect(_socket, result->ai_addr, (int)result->ai_addrlen) == SOCKET_ERROR)
			{
				TRACE1("connection failed. error: %d\n", WSAGetLastError());
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
				closesocket(_socket);
				_socket = INVALID_SOCKET;
			}
		}

		if (result != NULL)
			freeaddrinfo(result);

		return status;
	}

	bool CommandsClientSocketsWin32::Disconnect()
	{
		if (closesocket(_socket) == SOCKET_ERROR)
		{
			_socket = INVALID_SOCKET;
			return false;
		}

		_socket = INVALID_SOCKET;
		return true;
	}
	bool CommandsClientSocketsWin32::Invoke(IN const uint8_t* inputBuffer,IN uint32_t inputBufferSize,OUT uint8_t** outputBuffer,OUT uint32_t* outputBufferSize)
	{
		int iResult;
		char* RecvOutBuff = NULL;

		if (inputBufferSize == 0 || inputBuffer == NULL || outputBuffer == NULL || outputBufferSize == NULL)
			return false;


		// sending the InputBufferSize
		iResult = blockedSend(_socket,(char*) &inputBufferSize, sizeof(uint32_t));
		if (iResult != sizeof(uint32_t)) 
		{
			TRACE1("send inputBufferSize failed: %d\n", WSAGetLastError());
			return false;
		}

		// sending the InputBuffer
		iResult = blockedSend(_socket,(char*) inputBuffer, inputBufferSize);
		if (iResult != inputBufferSize) 
		{
			TRACE1("send inputBuffer failed: %d\n", WSAGetLastError());
			return false;
		}

		// Receive until the peer closes the connection
		iResult = blockedRecv(_socket, (char*) outputBufferSize, sizeof(uint32_t));
		if (iResult !=  sizeof(uint32_t))
		{
			TRACE1("recv outputBufferSize failed: %d\n", WSAGetLastError());
			return false;
		}

		if ((*outputBufferSize >= sizeof(JHI_RESPONSE)) && (*outputBufferSize < JHI_MAX_TRANSPORT_DATA_SIZE))
		{
			// allocate new buffer
			RecvOutBuff = (char*) JHI_ALLOC(*outputBufferSize);
			if (RecvOutBuff == NULL) {
				TRACE0("failed to allocate outputBufferSize memory.");
				return false;
			}

			iResult = blockedRecv(_socket, RecvOutBuff, *outputBufferSize);
			if (iResult !=  *outputBufferSize)
			{
				TRACE1("recv RecvOutBuff failed: %d\n", WSAGetLastError());
				JHI_DEALLOC(RecvOutBuff);
				RecvOutBuff = NULL;
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

	int CommandsClientSocketsWin32::blockedRecv(SOCKET socket, char* buffer, int length)
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

	int CommandsClientSocketsWin32::blockedSend(SOCKET socket, char* buffer, int length)
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