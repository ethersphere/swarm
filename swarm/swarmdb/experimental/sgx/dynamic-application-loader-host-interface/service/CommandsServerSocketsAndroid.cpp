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

#include "CommandsServerSocketsAndroid.h"
#include "dbg.h"
#include "EventLog.h"
#include "reg.h"
#include "misc.h"
#include <iostream>
#include <sstream>
#include <netdb.h>
#include <errno.h>
#include <string.h>
#include <pthread.h>
#include <netinet/in.h>
#include <cutils/sockets.h>

#define INVALID_SOCKET -1
#define SOCKET_ERROR   -1

#define SOCKET_NAME "jhid"

namespace intel_dal
{
	CommandsServerSocketsAndroid::CommandsServerSocketsAndroid(ICommandDispatcher* dispatcher, uint8_t maxClientNum)
		: ICommandsServer(dispatcher,maxClientNum), _socket(INVALID_SOCKET)
	{
		TRACE0("in CommandsServerSocketsAndroid()\n");
	}

	CommandsServerSocketsAndroid::~CommandsServerSocketsAndroid()
	{
		TRACE0("in ~CommandsServerSocketsAndroid()\n");
		if (_socket != INVALID_SOCKET)
			::close(_socket);
			_socket = INVALID_SOCKET;
	}

	bool CommandsServerSocketsAndroid::open()
	{
		int iResult;
		bool status = false;

		do
		{
			if (!_dispatcher->init())
			{
				TRACE0("dispatcher init failed\n");
				break;
			}

			_socket = android_get_control_socket(SOCKET_NAME);
			if (_socket < 0)
			{
				TRACE0("failed to get control socket\n");
				_socket = INVALID_SOCKET;
				break;
			}
			
			TRACE0("listening Android control socket\n");

			iResult = listen(_socket, SOMAXCONN);
			if (iResult == SOCKET_ERROR)
			{
				TRACE1("listen failed with error: %d\n", errno);
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
				::close(_socket);
				_socket = INVALID_SOCKET;
			}
		}

		return status;
	}

	bool CommandsServerSocketsAndroid::close()
	{
	        bool rc = false;
		if (::close(_socket) == SOCKET_ERROR)
		{
			TRACE0("failed to close socket\n:");
		}
		else if (!_dispatcher->deinit())
		{
			TRACE0("dispatcher deinit has failed\n:");
		}
		else
		{
			rc = true;
		}

		_socket = INVALID_SOCKET;
		return rc;
	}

	void CommandsServerSocketsAndroid::waitForRequests()
	{
		SOCKET clientSocket;

		while (_socket != INVALID_SOCKET)
		{
			// acquire max client semaphore
			getSemaphore()->Acquire();

			//re-check socket after acquiring semaphore
			if (_socket == INVALID_SOCKET)
			{
				TRACE0("INVALID_SOCKET after acquiring semaphore");
				getSemaphore()->Release();
				break;
			}

			// Accept a client socket
			clientSocket = accept(_socket, NULL, NULL);
			if (clientSocket == INVALID_SOCKET) {
				TRACE1("accept failed with error: %d\n", errno);
				getSemaphore()->Release();
				break;
			}

			startClientSession(clientSocket);
		}
	}

	uint32_t blockedRecv(SOCKET socket, uint8_t* buffer, uint32_t length)
	{
		uint32_t bytesRecieved = 0;
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

	uint32_t blockedSend(SOCKET socket, char* buffer, uint32_t length)
	{
		uint32_t bytesSent = 0;
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

	void* ClientSessionThread(void* threadParam)
	{
		uint32_t iResult;
		uint32_t inputBufferSize = 0;
		uint8_t* inputBuffer = NULL;
		uint8_t* outputBuffer = NULL;
		uint32_t outputBufferSize = 0;


		CS_ClientThreadParams* params = (CS_ClientThreadParams*) threadParam;
		SOCKET clientSocket = params->clientSocket;
		ICommandDispatcher* dispatcher = params->dispatcher;
		Semaphore* semaphore = params->semaphore;

		JHI_DEALLOC_T(params);
		params = NULL;

		do
		{

			iResult = blockedRecv(clientSocket, (uint8_t*) &inputBufferSize, sizeof(uint32_t));
			if (iResult !=  sizeof(uint32_t))
			{
				TRACE1("recv inputBufferSize failed with error: %d\n", errno);
				break;
			}

			if ((inputBufferSize < sizeof(JHI_COMMAND)) || (inputBufferSize > JHI_MAX_TRANSPORT_DATA_SIZE))
				break;

			// allocate new buffer
			inputBuffer = (uint8_t*) JHI_ALLOC(inputBufferSize);
			if (NULL == inputBuffer)
			{
				TRACE0("malloc of InputBuffer failed .");
				break;
			}

			iResult = blockedRecv(clientSocket, inputBuffer, inputBufferSize);
			if (iResult !=  inputBufferSize)
			{
				TRACE1("recv InputBuffer failed with error: %d\n", errno);
				break;
			}


			// prosess command here using the dispatcher
			dispatcher->processCommand((const uint8_t*) inputBuffer,inputBufferSize,
						    (uint8_t**) &outputBuffer,&outputBufferSize);

			// sending the OutputBufferSize
			iResult = blockedSend(clientSocket,(char*) &outputBufferSize,
					       sizeof(uint32_t));
			if (iResult != sizeof(uint32_t))
			{
				TRACE1("send outputBufferSize failed with error: %d\n", errno);
				break;
			}

			if (outputBufferSize > 0 && outputBuffer != NULL)
			{
				// sending the outputBuffer
				iResult = blockedSend(clientSocket,(char*) outputBuffer, outputBufferSize);
				if (iResult != outputBufferSize)
				{
					TRACE1("send outputBuffer failed with error: %d\n", errno);
					break;
				}

			}

			// closing the sockets for send operations, since no more data will be sent
			if (shutdown(clientSocket, SHUT_WR) == SOCKET_ERROR)
			{
				TRACE1("shutdown for send operations failed with error: %d\n", errno);
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
		if (::close(clientSocket) == SOCKET_ERROR)
		{
			TRACE1("close client socket failed: %d\n", errno);
		}
		clientSocket = INVALID_SOCKET;

		//release Max Clients semaphore
		semaphore->Release();

		return 0;
	}

	void CommandsServerSocketsAndroid::startClientSession(SOCKET clientSocket)
	{
		// create a thread to process the client request
		pthread_t clientThread;

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

		int rc = pthread_create(&clientThread, NULL, ClientSessionThread, (void*)params);
		if(rc)
		{
			TRACE0("failed creating thread for client request\n");
			getSemaphore()->Release();
		}
		else
		{
			pthread_detach(clientThread);
		}
	}
}//namespace intel_dal
