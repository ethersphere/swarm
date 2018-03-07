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

#include "CommandsServerSocketsWin32.h"
#include "dbg.h"
#include "EventLog.h"
#include "reg.h"
#include "misc.h"
#include <iostream>


namespace intel_dal
{
	CommandsServerSocketsWin32::CommandsServerSocketsWin32(ICommandDispatcher* dispatcher, uint8_t maxClientNum)
		: ICommandsServer(dispatcher,maxClientNum)
	{
		int ret;

		_socket = INVALID_SOCKET;

		ret = WSAStartup(MAKEWORD(2,2), &_wsaData);

		if (ret != 0) 
		{
			LOG1("WSAStartup failed with error: %d\n", ret);
			throw std::exception("WSAStartup failed");
		}

		if (LOBYTE(_wsaData.wVersion) != 2 || HIBYTE(_wsaData.wVersion) != 2) 
		{
			LOG0("Could not find a usable version of Winsock.dll\n");
			WSACleanup();
			throw std::exception("Could not find a usable version of Winsock.dll");
		}

	}

	CommandsServerSocketsWin32::~CommandsServerSocketsWin32()
	{
		TRACE0("in ~CommandsServerSocketsWin32()\n");
		if (_socket != INVALID_SOCKET)
			closesocket(_socket);

		WSACleanup();
	}

	bool CommandsServerSocketsWin32::open()
	{
		int iResult;


		SOCKADDR_IN6 ipv6_socket_data;
		SOCKADDR_IN  ipv4_socket_data; 
		void* socket_data = NULL;

		int socket_data_size = 0;
		int port_number;
		bool status = false;


		struct addrinfo hints;
		struct addrinfo *result = NULL;
		struct addrinfo *ptr = NULL;

		ZeroMemory( &hints, sizeof(hints) );
		hints.ai_family = AF_UNSPEC;
		hints.ai_socktype = SOCK_STREAM;
		hints.ai_protocol = IPPROTO_TCP;

		ZeroMemory(&ipv4_socket_data, sizeof(ipv4_socket_data));
		ZeroMemory(&ipv6_socket_data, sizeof(ipv6_socket_data));

		do
		{

			if (!_dispatcher->init())
			{
				LOG0("dispatcher init failed\n");
				break;
			}


			if (getaddrinfo("localhost",NULL,&hints,&result) != 0)
			{
				LOG0("failed to get adderss info\n");
				WriteToEventLog(JHI_EVENT_LOG_ERROR, MSG_CONNECT_FAILURE);
				break;
			}

			if (result == NULL)
			{
				LOG0("no adderss info recieved\n");
				WriteToEventLog(JHI_EVENT_LOG_ERROR, MSG_CONNECT_FAILURE);
				break;
			}

			// select address of ipv4 or ipv6 family
			for(ptr=result; ptr != NULL ;ptr=ptr->ai_next)
			{
				if (ptr->ai_family == AF_INET || ptr->ai_family == AF_INET6)
					break;
			}

			if (ptr == NULL)
			{
				LOG0("failed to find IPV4 or IPV6 address\n");
				WriteToEventLog(JHI_EVENT_LOG_ERROR, MSG_CONNECT_FAILURE);
				break;
			}

			_socket = socket(ptr->ai_family, ptr->ai_socktype, ptr->ai_protocol);

			if (_socket == INVALID_SOCKET)
			{
				LOG1("socket() failed with error: %d\n", WSAGetLastError());
				WriteToEventLog(JHI_EVENT_LOG_ERROR, MSG_CONNECT_FAILURE);
				break;
			}

			if (bind(_socket, ptr->ai_addr, (int)ptr->ai_addrlen) == SOCKET_ERROR)
			{
				LOG1("bind() failed with error: %d\n", WSAGetLastError());
				WriteToEventLog(JHI_EVENT_LOG_ERROR, MSG_CONNECT_FAILURE);
				break;
			}

			if (ptr->ai_family == AF_INET)
			{
				socket_data = &ipv4_socket_data;
				socket_data_size = sizeof(ipv4_socket_data);
			}
			else // ipv6
			{
				socket_data = &ipv6_socket_data;
				socket_data_size = sizeof(ipv6_socket_data);
			}

			if (getsockname(_socket,(LPSOCKADDR)socket_data,&socket_data_size) != 0)
			{
				LOG1("getsockname() failed with error: %d\n", WSAGetLastError());
				WriteToEventLog(JHI_EVENT_LOG_ERROR, MSG_CONNECT_FAILURE);
				break;
			}

			if (ptr->ai_family == AF_INET)
			{
				port_number = ntohs(ipv4_socket_data.sin_port);
			}
			else // ipv6
			{
				port_number = ntohs(ipv6_socket_data.sin6_port);
			}


			iResult = JhiWritePortNumberToRegistry(port_number);
			if (iResult != JHI_SUCCESS)
			{
				LOG0("failed to write service port at registry.");
				WriteToEventLog(JHI_EVENT_LOG_ERROR, MSG_REGISTRY_WRITE_ERROR);
				break;
			}

			iResult = JhiWriteAddressTypeToRegistry(ptr->ai_family);
			if (iResult != JHI_SUCCESS)
			{
				LOG0("failed to write address type at registry.");
				WriteToEventLog(JHI_EVENT_LOG_ERROR, MSG_REGISTRY_WRITE_ERROR);
				break;
			}


			iResult = listen(_socket, SOMAXCONN);
			if (iResult == SOCKET_ERROR)
			{
				LOG1("listen failed with error: %d\n", WSAGetLastError());
				WriteToEventLog(JHI_EVENT_LOG_ERROR, MSG_CONNECT_FAILURE);
				break;
			}

			status = true;
		}
		while (0);

		// cleanup
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

	bool CommandsServerSocketsWin32::close()
	{
		if (closesocket(_socket) == SOCKET_ERROR)
		{
			TRACE0("failed to close socket\n:");
			return false;
		}

		if (!_dispatcher->deinit())
		{
			TRACE0("dispatcher deinit has failed\n:");
			return false;
		}

		return true;
	}

	void CommandsServerSocketsWin32::waitForRequests()
	{

		SOCKET clientSocket;

		while (_socket != INVALID_SOCKET)
		{
			// acquire max client semaphore 
			getSemaphore()->Acquire();

			// Accept a client socket
			clientSocket = accept(_socket, NULL, NULL);
			if (clientSocket == INVALID_SOCKET) {
				TRACE1("accept failed with error: %d\n", WSAGetLastError());
				getSemaphore()->Release();
				break;
			}

			startClientSession(clientSocket);	
		}
	}

	int blockedRecv(SOCKET socket, char* buffer, int length)
	{
		int bytesRecieved = 0;
		int count;

		while (bytesRecieved != length)
		{
			count = recv(socket, buffer + bytesRecieved, length - bytesRecieved, 0);

			if (count == SOCKET_ERROR || count == 0) // client closed the connection
				break;

			bytesRecieved += count;
		}

		return bytesRecieved;
	}

	int blockedSend(SOCKET socket, char* buffer, int length)
	{
		int bytesSent = 0;
		int count;

		while (bytesSent != length)
		{
			count = send(socket, buffer + bytesSent, length - bytesSent, 0);

			if (count == SOCKET_ERROR) // client closed the connection
				break;

			bytesSent += count;
		}

		return bytesSent;
	}

	DWORD ClientSessionThread(LPVOID threadParam)
	{
		int iResult;
		uint32_t inputBufferSize = 0;
		char * inputBuffer = NULL;
		char * outputBuffer = NULL;
		uint32_t outputBufferSize = 0;


		CS_ClientThreadParams* params = (CS_ClientThreadParams*) threadParam;
		SOCKET clientSocket = params->clientSocket;
		ICommandDispatcher* dispatcher = params->dispatcher;
		Semaphore* semaphore = params->semaphore;

		JHI_DEALLOC_T(params);
		params = NULL;

		// COM init
		CoInitializeEx(NULL, COINIT_MULTITHREADED);

		do
		{

			iResult = blockedRecv(clientSocket, (char*) &inputBufferSize, sizeof(uint32_t));
			if (iResult !=  sizeof(uint32_t))
			{
				TRACE1("recv inputBufferSize failed with error: %d\n", WSAGetLastError());
				break;
			}

			if ((inputBufferSize < sizeof(JHI_COMMAND)) || (inputBufferSize > JHI_MAX_TRANSPORT_DATA_SIZE))
				break;


			// allocate new buffer
			inputBuffer = (char*) JHI_ALLOC(inputBufferSize);
			if (inputBuffer == NULL) {
				TRACE0("malloc of InputBuffer failed .");
				break;
			}

			iResult = blockedRecv(clientSocket, inputBuffer, inputBufferSize);
			if (iResult !=  inputBufferSize)
			{
				TRACE1("recv InputBuffer failed with error: %d\n", WSAGetLastError());
				break;
			}


			// prosess command here using the dispatcher
			dispatcher->processCommand((const uint8_t*) inputBuffer,inputBufferSize,(uint8_t**) &outputBuffer,&outputBufferSize);

			// sending the OutputBufferSize
			iResult = blockedSend(clientSocket,(char*) &outputBufferSize,sizeof(uint32_t));
			if (iResult != sizeof(uint32_t)) 
			{
				TRACE1("send outputBufferSize failed with error: %d\n", WSAGetLastError());
				break;
			}

			if (outputBufferSize > 0)
			{
				// sending the outputBuffer
				iResult = blockedSend(clientSocket,(char*) outputBuffer, outputBufferSize);
				if (iResult != outputBufferSize) 
				{
					TRACE1("send outputBuffer failed with error: %d\n", WSAGetLastError());
					break;
				}

			}

			// closing the sockets for send operations, since no more data will be sent
			iResult = shutdown(clientSocket, SD_SEND);
			if (iResult == SOCKET_ERROR) 
			{
				TRACE1("shutdown for send operations failed with error: %d\n", WSAGetLastError());
				break;
			}

		}
		while(0);

		//cleanup:

		if (inputBuffer != NULL)
		{
			JHI_DEALLOC(inputBuffer);
			inputBuffer = NULL;
		}
		if (outputBuffer != NULL)
		{
			JHI_DEALLOC(outputBuffer);
			outputBuffer = NULL;
		}

		// closing the conection to client
		if (closesocket(clientSocket) == SOCKET_ERROR)
		{
			TRACE1("close client socket failed: %d\n", WSAGetLastError());
		}
		clientSocket = INVALID_SOCKET;

		// COM deinit
		CoUninitialize();

		//release Max Clients semaphore
		semaphore->Release();

		return 0;
	}

	void CommandsServerSocketsWin32::startClientSession(SOCKET clientSocket)
	{
		// create a thread to process the client request
		HANDLE clientThread;

		CS_ClientThreadParams* params = NULL;

		params = JHI_ALLOC_T(CS_ClientThreadParams);
		if (params == NULL)
		{
			TRACE0("CS_ClientThreadParams memory allocation failed");
			getSemaphore()->Release();
			return;
		}

		params->clientSocket = clientSocket;
		params->dispatcher = _dispatcher;
		params->semaphore = getSemaphore();

		clientThread = CreateThread(NULL, 0,(LPTHREAD_START_ROUTINE)&ClientSessionThread,(LPVOID)params,0,NULL);

		if (clientThread == NULL)
		{
			TRACE0("failed creating thread for client request\n");

			// cleanup
			if (params != NULL)
			{
				JHI_DEALLOC_T(params);
				params = NULL;
			}
		}
		else
		{
			// closing the clientThread handle, the thread stays alive.
			CloseHandle(clientThread);
		}
	}
}