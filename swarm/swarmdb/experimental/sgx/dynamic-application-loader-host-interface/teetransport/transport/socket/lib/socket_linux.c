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

#include <stdio.h>
#include <netinet/in.h>
#include <netdb.h>
#include <sys/socket.h>
#include <string.h>
#include <string_s.h>
#include "socket.h"
#include "errno.h"

#include "reg.h"

DWORD SocketSetup()
{
    return SOCKET_STATUS_SUCCESS;
}

DWORD SocketTeardown()
{
    return SOCKET_STATUS_SUCCESS;
}

DWORD SocketConnect(const char *c_ip, int port, SOCKET* sock)
{
    struct addrinfo *addr = NULL;
    struct addrinfo hints;
    char port_cstr[20];
    int iResult = 0;

    // Since TEETransport doesn't allow passing an IP address along with the port
    // and adding it requires changes all over the code hierarchy, the IP address
    // retrieval is implemented here locally.
    char ip[16];
    strcpy(ip, SOCK_DEFAULT_IP_ADDRESS);

    JhiQuerySocketIpAddressFromRegistry(ip);

    if(!sock)
    {
      return SOCKET_STATUS_FAILURE;
    }

    // set default value
    *sock = INVALID_SOCKET;

    // check if port is in valid range
    if((port < SOCK_MIN_PORT_VALUE) || (port > SOCK_MAX_PORT_VALUE))
    {
      return SOCKET_STATUS_FAILURE;
    }

    memset( &hints, 0, sizeof(hints) );
    hints.ai_family = AF_UNSPEC;
    hints.ai_socktype = SOCK_STREAM;
    hints.ai_protocol = IPPROTO_TCP;

    if(-1 == sprintf_s(port_cstr, 20, "%d", port))
    {
      return SOCKET_STATUS_FAILURE;
    }

    // Resolve the server address and port
    iResult = getaddrinfo(ip, port_cstr, &hints, &addr);
    if (iResult != 0 || !addr)
    {
        return SOCKET_STATUS_FAILURE;
    }

    // Create a SOCKET for connecting to server
    *sock = socket(addr->ai_family, addr->ai_socktype, addr->ai_protocol);
    if (*sock == SOCKET_ERROR)
    {
      freeaddrinfo(addr);
      return SOCKET_STATUS_FAILURE;
    }

    // Connect to server.
    iResult = connect(*sock, addr->ai_addr, (int)addr->ai_addrlen);
    if (iResult == SOCKET_ERROR)
    {
        TRACE1("Couldn't connect. errno: %d\n", errno);
        close(*sock);
        freeaddrinfo(addr);
        return SOCKET_STATUS_FAILURE;
    }

    freeaddrinfo(addr);

    return SOCKET_STATUS_SUCCESS;
}

DWORD SocketDisconnect(SOCKET sock)
{
    if (INVALID_SOCKET == sock)
    {
        return SOCKET_STATUS_FAILURE;
    }
    else
    {
        shutdown(sock, SHUT_RDWR);
		close(sock);
        return SOCKET_STATUS_SUCCESS;
    }
}

DWORD SocketSend(SOCKET sock, const char* buffer, int* length)
{
    ssize_t bytes_written = 0;

    if ((INVALID_SOCKET == sock) || (NULL == buffer) || (NULL == length))
    {
        return SOCKET_STATUS_FAILURE;
    }

    bytes_written = send(sock, buffer, *length, 0);

    if (bytes_written == -1)
    {
        TRACE1("send failed. errno: %d", errno);
        *length = 0;
        return SOCKET_STATUS_FAILURE;
    }

    *length = bytes_written;

    return SOCKET_STATUS_SUCCESS;
}

DWORD SocketRecv(SOCKET sock, char* buffer, int* length)
{
    int iResult = 0;

    if ((INVALID_SOCKET == sock) || (NULL == buffer) || (NULL == length))
    {
      return SOCKET_STATUS_FAILURE;
    }

    iResult = recv(sock, buffer, *length, MSG_WAITALL);

    if ((iResult == SOCKET_ERROR) || (iResult < 0))
    {
        *length = 0;
        return SOCKET_STATUS_FAILURE;
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

    return SOCKET_STATUS_SUCCESS;
}
