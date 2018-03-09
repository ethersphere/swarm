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

#ifdef EXTERNAL_API_INSTRUMENTATION
#ifndef _EXTERNAL_API_LIST_H_
#define _EXTERNAL_API_LIST_H_

#include <windows.h>

#include <winsock2.h>
#include "Ws2tcpip.h"

//getaddrinfo
typedef WINSOCK_API_LINKAGE
INT
WSAAPI
PFN_GET_ADDRESS_INFO(
    __in_opt        PCSTR               pNodeName,
    __in_opt        PCSTR               pServiceName,
    __in_opt        const ADDRINFOA *   pHints,
    __deref_out     PADDRINFOA *        ppResult
    );

//socket
typedef WINSOCK_API_LINKAGE
__checkReturn
SOCKET
WSAAPI
PFN_SOCKET(
    __in int af,
    __in int type,
    __in int protocol
    );

//bind
typedef WINSOCK_API_LINKAGE
int
WSAAPI
PFN_BIND(
    __in SOCKET s,
    __in_bcount(namelen) const struct sockaddr FAR * name,
    __in int namelen
    );

//getsockname
typedef WINSOCK_API_LINKAGE 
int
WSAAPI
PFN_GETSOCKNAME(
    __in SOCKET s,
    __out_bcount_part(*namelen,*namelen) struct sockaddr FAR * name,
    __inout int FAR * namelen
    );

#endif //_EXTERNAL_API_LIST_H_
#endif //EXTERNAL_API_INSTRUMENTATION
