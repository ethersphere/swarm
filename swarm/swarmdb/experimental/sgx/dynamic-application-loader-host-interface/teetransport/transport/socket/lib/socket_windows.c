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

#include "socket.h"

#include <WS2tcpip.h>
#include <stdio.h>

#pragma comment (lib, "Ws2_32.lib")


DWORD SocketSetup()
{
   WSADATA wsaData = {0};

   if (0 != WSAStartup(MAKEWORD(2,2), &wsaData)) 
   {
      return ERROR_INTERNAL_ERROR;
   }
   return ERROR_SUCCESS;
}

DWORD SocketTeardown()
{
   if (0 != WSACleanup()) 
   {
      return ERROR_INTERNAL_ERROR;
   }
   return ERROR_SUCCESS;
}

DWORD SocketConnect(const char *c_ip, int port, SOCKET* sock)
{
   struct addrinfo *addr = NULL;
   struct addrinfo hints;
   char port_cstr[20];
   int iResult = 0;

   if(!sock)
   {
      return ERROR_INVALID_PARAMETER;
   }
   
   // set default value
   *sock = INVALID_SOCKET;

   // check if port is in valid range
   if((port < SOCK_MIN_PORT_VALUE) || (port > SOCK_MAX_PORT_VALUE))
   {
      return ERROR_INVALID_PARAMETER;
   }

   memset( &hints, 0, sizeof(hints) );
   hints.ai_family = AF_UNSPEC;
   hints.ai_socktype = SOCK_STREAM;
   hints.ai_protocol = IPPROTO_TCP;

   if(-1 == sprintf_s(port_cstr, 20, "%d", port))
   {
      return ERROR_INVALID_PARAMETER;
   }

   // Resolve the server address and port
   iResult = getaddrinfo(SOCK_DEFAULT_IP_ADDRESS, port_cstr, &hints, &addr);
   if ( iResult != 0 ) 
   {
      return iResult;
   }

   if(!addr)
   {
      return ERROR_INTERNAL_ERROR;
   }

   // Create a SOCKET for connecting to server
   *sock = socket(addr->ai_family, addr->ai_socktype, addr->ai_protocol);
   if (INVALID_SOCKET == *sock) 
   {
      freeaddrinfo(addr);
      return ERROR_INTERNAL_ERROR;
   }

   // Connect to server.
   iResult = connect(*sock, addr->ai_addr, (int)addr->ai_addrlen);
   if (iResult == SOCKET_ERROR) 
   {
      closesocket(*sock);
      freeaddrinfo(addr);
      return ERROR_INTERNAL_ERROR;
   }

   freeaddrinfo(addr);

   if (INVALID_SOCKET == *sock) 
   {
      return ERROR_INTERNAL_ERROR;
   }

   return ERROR_SUCCESS;
}

DWORD SocketDisconnect(SOCKET sock)
{
   if (INVALID_SOCKET == sock) 
   {
      return ERROR_INVALID_PARAMETER;
   }
   
   shutdown(sock, SD_SEND);
   closesocket(sock);

   return ERROR_SUCCESS;
}

DWORD SocketSend(SOCKET sock, const char* buffer, int* length)
{
   int bytes_written = 0;

   if ((INVALID_SOCKET == sock) || (NULL == buffer) || (NULL == length))
   {
      return ERROR_INVALID_PARAMETER;
   }   

   bytes_written = send(sock, buffer, *length, 0);

   if (bytes_written == SOCKET_ERROR)
   {
      *length = 0;
      return ERROR_INTERNAL_ERROR;
   }

   *length = bytes_written;
   
   return ERROR_SUCCESS;
}

DWORD SocketRecv(SOCKET sock, char* buffer, int* length)
{
   int iResult = 0;

   if ((INVALID_SOCKET == sock) || (NULL == buffer) || (NULL == length))
   {
      return ERROR_INVALID_PARAMETER;
   }

   iResult = recv(sock, buffer, *length, 0);

   if ((iResult == SOCKET_ERROR) || (iResult < 0))
   {
        *length = 0;
        return ERROR_INTERNAL_ERROR;
   }

   if (iResult > 0)
   {
      // recv - return the number of bytes received
      *length = iResult;      
   }
   else if (iResult == 0) 
   {
      // recv - connection has been gracefully closed
      *length = 0;      
   }

   return ERROR_SUCCESS;
}
