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
*    @file        socket.h
*    @brief       Defines interface for Winsock implementation. This interface
*                 is used internally by the wrapper (TEETransportSocketWrapper).
*    @author      Artum Tchachkin
*    @date        August 2014
*/

#ifndef _TEE_TRANSPORT_SOCKET_LIBRARY_H_
#define _TEE_TRANSPORT_SOCKET_LIBRARY_H_

#if defined(_WIN32)
#define WIN32_LEAN_AND_MEAN
#include <Windows.h>
#include <WinSock2.h>
#define SOCKET_STATUS_SUCCESS ERROR_SUCCESS

#elif defined(__linux__)
#include <sys/socket.h>
#define SOCKET_STATUS_SUCCESS 0
#define SOCKET_STATUS_FAILURE 1
#define INVALID_SOCKET -1
#define SOCKET_ERROR -1
typedef intptr_t SOCKET;
typedef uint32_t DWORD;

#else
Unknown OS

#endif //_WIN32

#ifdef __cplusplus
extern "C" {
#endif

	// Localhost is for running over Win32 FW emulation which runs on the same OS as JHI
	// The other ip address is for running from inside a VirtualBox VM
	// This is not used on Linux. The address comes from reg-linux.cpp
    #define SOCK_DEFAULT_IP_ADDRESS     "127.0.0.1"
    //#define SOCK_DEFAULT_IP_ADDRESS     "192.168.56.1"
    #define SOCK_MIN_PORT_VALUE         (0x0400) // 1024 is min TCP port value that is not reserved to system use
    #define SOCK_MAX_PORT_VALUE         (0xFFFF) // 65535 is max value since port is SHORT = 2^16-1


   DWORD SocketSetup();
   DWORD SocketTeardown();
   DWORD SocketConnect(const char *c_ip, int port, SOCKET* sock);
   DWORD SocketDisconnect(SOCKET sock);
   DWORD SocketSend(SOCKET sock, const char* buffer, int* length);
   DWORD SocketRecv(SOCKET sock, char* buffer, int* length);

#ifdef __cplusplus
};
#endif


#endif //_TEE_TRANSPORT_SOCKET_LIBRARY_H_