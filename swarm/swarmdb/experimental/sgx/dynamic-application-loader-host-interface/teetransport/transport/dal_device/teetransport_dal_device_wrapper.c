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
*    @file        TEETransportDalDeviceWrapper.c
*    @brief       Implementation of the internal interface for dal device.
*    @author      Adam Shitrit
*    @date        December 2015
*/

#include "teetransport_dal_device_wrapper.h"
#include "teetransport_internal.h"
#include <stdlib.h>

#ifndef _WIN32
#include <fcntl.h>
#include <unistd.h>
#define min(a, b)  (((a) < (b)) ? (a) : (b))
#endif


#define DAL_IVM_FILE		"/dev/dal0"
#define DAL_SDM_FILE		"/dev/dal1"
#define DAL_RTM_FILE		"/dev/dal2"


TEE_COMM_STATUS DAL_Device_Teardown(IO TEE_TRANSPORT_INTERFACE_PTR pInterface)
{

#ifdef WIN32

	return TEE_COMM_NOT_IMPLEMENTED;

#else

	if(!pInterface)
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	if(TEE_INTERFACE_STATE_INITIALIZED == pInterface->state)
	{
		pInterface->state = TEE_INTERFACE_STATE_NOT_INITIALIZED;
	}

	return TEE_COMM_SUCCESS;

#endif
}


TEE_COMM_STATUS DAL_Device_Connect(IN TEE_TRANSPORT_INTERFACE_PTR pInterface, IN TEE_TRANSPORT_ENTITY entity, IN const char* params, OUT TEE_TRANSPORT_HANDLE* handle)
{

#ifdef WIN32

	return TEE_COMM_NOT_IMPLEMENTED;

#else

	intptr_t fd = -1;

	if ( (NULL == handle) || (NULL == pInterface) )
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	if ( !isEntityValid(entity) )
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	if( TEE_INTERFACE_STATE_INITIALIZED != pInterface->state )
	{
		return TEE_COMM_NOT_INITIALIZED;
	}

	// set default value
	*handle = TEE_TRANSPORT_INVALID_HANDLE_VALUE;

	if ( entity == TEE_TRANSPORT_ENTITY_IVM )
	{
		fd = open(DAL_IVM_FILE, O_RDWR);
	}
	
	else if ( entity == TEE_TRANSPORT_ENTITY_SDM )
	{
		fd = open(DAL_SDM_FILE, O_RDWR);
	}

	else if ( entity == TEE_TRANSPORT_ENTITY_RTM )
	{
		fd = open(DAL_RTM_FILE, O_RDWR);
	}

	//open failed
	if ( fd < 0 )
	{
		return TEE_COMM_INTERNAL_ERROR;
	}

	*handle = (TEE_TRANSPORT_HANDLE) fd;

	return TEE_COMM_SUCCESS;

#endif
}

TEE_COMM_STATUS DAL_Device_Disconnect(IN TEE_TRANSPORT_INTERFACE_PTR pInterface, IN TEE_TRANSPORT_HANDLE* handle)
{

#ifdef WIN32

	return TEE_COMM_NOT_IMPLEMENTED;

#else

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
		intptr_t fd = (intptr_t) *handle;

		if(close(fd) < 0 )
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

#endif

TEE_COMM_STATUS DAL_Device_Send(IN TEE_TRANSPORT_INTERFACE_PTR pInterface, IN TEE_TRANSPORT_HANDLE handle, IN const uint8_t* buffer, IN uint32_t length)
{

#ifdef WIN32

	return TEE_COMM_NOT_IMPLEMENTED;

#else

	intptr_t fd = -1;
    ssize_t bytes_written = 0;
    size_t total_written = 0;
    size_t bytes_to_write = 0;

    // Currently, the client MTU of KDI (of DAL in general) is 4K.
    // Can be changed to be queried from KDI if the need comes
    const size_t client_mtu = 4096;

	if((TEE_TRANSPORT_INVALID_HANDLE_VALUE == handle) || (NULL == buffer) || (NULL == pInterface))
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	if(TEE_INTERFACE_STATE_INITIALIZED != pInterface->state)
	{
		return TEE_COMM_NOT_INITIALIZED;
	}

	fd = (intptr_t) handle;

    while(total_written < length)
    {
        bytes_to_write = min((size_t)(length - total_written), client_mtu);

        bytes_written = write(fd, &buffer[total_written], bytes_to_write);

        if ( bytes_written <= 0 )
        {
            return TEE_COMM_INTERNAL_ERROR;
        }

        total_written += bytes_written;
    }

    return TEE_COMM_SUCCESS;

#endif
}


TEE_COMM_STATUS DAL_Device_Recv(IN TEE_TRANSPORT_INTERFACE_PTR pInterface, IN TEE_TRANSPORT_HANDLE handle, IO uint8_t* buffer, IO uint32_t* length)
{

#ifdef WIN32

	return TEE_COMM_NOT_IMPLEMENTED;

#else

	ssize_t status;
	intptr_t fd = -1;

	if((TEE_TRANSPORT_INVALID_HANDLE_VALUE == handle) || (NULL == buffer) || (NULL == length) || (NULL == pInterface))
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	if(TEE_INTERFACE_STATE_INITIALIZED != pInterface->state)
	{
		return TEE_COMM_NOT_INITIALIZED;
	}

	fd = (intptr_t)handle;

	status = read(fd, buffer, *length);
	if ( status == -1 )
	{
		return TEE_COMM_TRANSPORT_FAILED;
	}

	*length = status;

	return TEE_COMM_SUCCESS;

#endif
}
