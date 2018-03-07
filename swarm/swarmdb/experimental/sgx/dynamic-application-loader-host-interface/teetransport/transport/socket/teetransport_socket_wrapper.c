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
*    @file        TEETransportSocketWrapper.c
*    @brief       Implementation of the internal interface for socket library.
*    @author      Artum Tchachkin
*    @date        August 2014
*/

#include "teetransport_socket_wrapper.h"
#include "socket.h" // here is the raw implementation of the socket code
#include "teetransport_internal.h"
#include <stdlib.h>

TEE_COMM_STATUS SOCK_Setup(IO TEE_TRANSPORT_INTERFACE_PTR pInterface)
{
	if(!pInterface)
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	if(TEE_INTERFACE_STATE_NOT_INITIALIZED == pInterface->state)
	{
		if(SOCKET_STATUS_SUCCESS != SocketSetup())
		{
			return TEE_COMM_INTERNAL_ERROR;
		}

		pInterface->state = TEE_INTERFACE_STATE_INITIALIZED;
	}

	return TEE_COMM_SUCCESS;
}

TEE_COMM_STATUS SOCK_Teardown(IO TEE_TRANSPORT_INTERFACE_PTR pInterface)
{
	if(!pInterface)
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	if(TEE_INTERFACE_STATE_INITIALIZED == pInterface->state)
	{
		pInterface->state = TEE_INTERFACE_STATE_NOT_INITIALIZED;

		if(SOCKET_STATUS_SUCCESS != SocketTeardown())
		{
			return TEE_COMM_INTERNAL_ERROR;
		}
	}

	return TEE_COMM_SUCCESS;
}

TEE_COMM_STATUS SOCK_Connect(IN TEE_TRANSPORT_INTERFACE_PTR pInterface, IN TEE_TRANSPORT_ENTITY entity, IN const char* params, OUT TEE_TRANSPORT_HANDLE* handle)
{
	SOCKET s = INVALID_SOCKET;
	int port = -1;

	if((NULL == handle) || (NULL == pInterface))
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	if(!isEntityValid(entity))
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	if(TEE_INTERFACE_STATE_INITIALIZED != pInterface->state)
	{
		return TEE_COMM_NOT_INITIALIZED;
	}

	// set default value
	*handle = TEE_TRANSPORT_INVALID_HANDLE_VALUE;

	if (entity == TEE_TRANSPORT_ENTITY_CUSTOM)
	{
		port = atoi(params);
	}
	else
	{
		port = entity;
	}

	if ( (port < SOCK_MIN_PORT_VALUE) || (port > SOCK_MAX_PORT_VALUE) ) //out of valid range for socket ports.
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	if(SOCKET_STATUS_SUCCESS != SocketConnect(NULL, port, &s))
	{
		return TEE_COMM_INTERNAL_ERROR;
	}

	*handle = (TEE_TRANSPORT_HANDLE)s;

	return TEE_COMM_SUCCESS;
}

TEE_COMM_STATUS SOCK_Disconnect(IN TEE_TRANSPORT_INTERFACE_PTR pInterface, IN TEE_TRANSPORT_HANDLE* handle)
{
	if( (NULL == pInterface) || (NULL == handle) )
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	if(TEE_INTERFACE_STATE_INITIALIZED != pInterface->state)
	{
		return TEE_COMM_NOT_INITIALIZED;
	}

	if(TEE_TRANSPORT_INVALID_HANDLE_VALUE != *handle)
	{
		SOCKET s = (SOCKET)*handle;

		if(SOCKET_STATUS_SUCCESS != SocketDisconnect(s))
		{
			return TEE_COMM_INTERNAL_ERROR;
		}

        *handle = TEE_TRANSPORT_INVALID_HANDLE_VALUE;
	}
    else
    {
        return TEE_COMM_INVALID_HANDLE;
    }    	

	return TEE_COMM_SUCCESS;
}

TEE_COMM_STATUS SOCK_Send(IN TEE_TRANSPORT_INTERFACE_PTR pInterface, IN TEE_TRANSPORT_HANDLE handle, IN const uint8_t* buffer, IN uint32_t length)
{
	SOCKET s = INVALID_SOCKET;
    int bytes_written = 0;
    size_t total = 0;

	if((TEE_TRANSPORT_INVALID_HANDLE_VALUE == handle) || (NULL == buffer) || (NULL == pInterface))
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	if(TEE_INTERFACE_STATE_INITIALIZED != pInterface->state)
	{
		return TEE_COMM_NOT_INITIALIZED;
	}

	s = (SOCKET)handle;

    // Since WinsockSend might write only part of the wanted content, 
    // this loop will continue sending the remaining data until all done.
    while(total < length)
    {        
        const char* ptr = (const char*)&buffer[total];
        
        bytes_written = (int)(length - total);

        if(SOCKET_STATUS_SUCCESS != SocketSend(s, ptr, &bytes_written))
	    {
		    return TEE_COMM_INTERNAL_ERROR;
	    }

        total += bytes_written;
    }

    return TEE_COMM_SUCCESS;
}

TEE_COMM_STATUS SOCK_Recv(IN TEE_TRANSPORT_INTERFACE_PTR pInterface, IN TEE_TRANSPORT_HANDLE handle, IO uint8_t* buffer, IO uint32_t* length)
{
	SOCKET s = INVALID_SOCKET;

	if((TEE_TRANSPORT_INVALID_HANDLE_VALUE == handle) || (NULL == buffer) || (NULL == length) || (NULL == pInterface))
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	if(TEE_INTERFACE_STATE_INITIALIZED != pInterface->state)
	{
		return TEE_COMM_NOT_INITIALIZED;
	}

	s = (SOCKET)handle;

	if(SOCKET_STATUS_SUCCESS != SocketRecv(s, (char*)buffer, (int*)length))
	{
		return TEE_COMM_INTERNAL_ERROR;
	}

	return TEE_COMM_SUCCESS;
}
