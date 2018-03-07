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
*    @file        TEETransportTeeLibWrapper.c
*    @brief       Implementation of the external interface for TEELib library.
*                 This code is a wrapper above the TEE APIs that provides 
*                 stream like behavior since TEE APIs are not.
*    @author      Artum Tchachkin
*    @date        August 2014
*/

#include "teetransport_internal.h"
#include "teetransport_libtee_wrapper.h"
#include "teetransport_libtee_client_metadata.h"

#ifndef _WIN32
#define min(a, b)  (((a) < (b)) ? (a) : (b))
#endif

/**
* DAL Host Interface protocol GUIDs. MUST correspond to JomClientDefs.h in the FW.
*/
#define TEE_LIB_IVM_PROTOCOL_GUID        { 0x3C4852D6, 0xD47B, 0x4F46, 0xB0, 0x5E, 0xB5, 0xED, 0xC1, 0xAA, 0x44, 0x0E }
#define TEE_LIB_SDM_PROTOCOL_GUID        { 0xDBA4D603, 0xD7ED, 0x4931, 0x88, 0x23, 0x17, 0xAD, 0x58, 0x57, 0x05, 0xD5 }
#define TEE_LIB_LAUNCHER_PROTOCOL_GUID   { 0x5565A099, 0x7FE2, 0x45C1, 0xA2, 0x2B, 0xD7, 0xE9, 0xDF, 0xEA, 0x9A, 0x2E }
#define TEE_LIB_SVM_PROTOCOL_GUID        { 0xF47ACC04, 0xD94B, 0x49CA, 0x87, 0xA6, 0x7F, 0x7D, 0xC0, 0x3F, 0xBA, 0xF3 }


typedef struct _TEE_LIB_LOOKUP_ENTRY
{
    int port;
    GUID guid;
} TEE_LIB_LOOKUP_ENTRY;

/**
* FW HECI GUID numbers. The values must match the values defined in BeihaiHAL.h.
* This table is built as look-up table for TEE_TRANSPORT_PORT enum, any change in the enum
* is affecting directly this table and should be synced.
*/
static const TEE_LIB_LOOKUP_ENTRY gTeeLibGuidLookupTable[TEE_TRANSPORT_ENTITY_COUNT] = 
{
    {TEE_TRANSPORT_ENTITY_IVM, TEE_LIB_IVM_PROTOCOL_GUID},
    {TEE_TRANSPORT_ENTITY_SDM, TEE_LIB_SDM_PROTOCOL_GUID},
    {TEE_TRANSPORT_ENTITY_RTM, TEE_LIB_LAUNCHER_PROTOCOL_GUID},
    {TEE_TRANSPORT_ENTITY_SVM, TEE_LIB_SVM_PROTOCOL_GUID}
};

static const GUID* FindHeciGuid(int port)
{
    int i;
    for(i = 0; i < TEE_TRANSPORT_ENTITY_COUNT; ++i)
    {
        if(gTeeLibGuidLookupTable[i].port == port)
        {
            return &(gTeeLibGuidLookupTable[i].guid);
        }
    }

    return NULL;
}

static TEE_CLIENT_META_DATA_CONTEXT gClientContext = { NULL, NULL, 0, 0 };


TEE_COMM_STATUS TEELIB_Setup(IN OUT TEE_TRANSPORT_INTERFACE_PTR pInterface)
{
    if(!pInterface)
    {
        return TEE_COMM_INVALID_PARAMS;
    }

    if(TEE_INTERFACE_STATE_NOT_INITIALIZED == pInterface->state)
    {

        if(TEE_COMM_SUCCESS != SetupContext(&gClientContext))
        {
            return TEE_COMM_INTERNAL_ERROR;
        }

        // Currently there is no Setup/Teardown flows in HECI

        pInterface->state = TEE_INTERFACE_STATE_INITIALIZED;
    }

    return TEE_COMM_SUCCESS;
}

TEE_COMM_STATUS TEELIB_Teardown(IN OUT TEE_TRANSPORT_INTERFACE_PTR pInterface)
{
    if(!pInterface)
    {
        return TEE_COMM_INVALID_PARAMS;
    }

    if(TEE_INTERFACE_STATE_INITIALIZED == pInterface->state)
    {
        if(TEE_COMM_SUCCESS != TeardownContext(&gClientContext))
        {
            // Log error
        }

        pInterface->state = TEE_INTERFACE_STATE_NOT_INITIALIZED;
    }

    return TEE_COMM_SUCCESS;
}

TEE_COMM_STATUS TEELIB_Connect(IN TEE_TRANSPORT_INTERFACE_PTR pInterface, IN TEE_TRANSPORT_ENTITY entity, IN const char* params, OUT TEE_TRANSPORT_HANDLE* handle)
{
    const GUID* pGuid = NULL;
    TEESTATUS stat = TEE_INTERNAL_ERROR;
    TEE_CLIENT_META_DATA* pClient = NULL;
    GUID guid = {0};

    if((NULL == handle) ||  (NULL == pInterface))
    {
        return TEE_COMM_INVALID_PARAMS;
    }

    if ((entity == TEE_TRANSPORT_ENTITY_CUSTOM) && (NULL == params))
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
        pGuid = ParseGUID(params, &guid);
    }
    else
    {
        // TODO: In normal world, the ENTITY should be enum and here we should translate it to the actual HECI GUID...
        pGuid = FindHeciGuid(entity);
    }

    if(NULL == pGuid)
    {  
        return TEE_COMM_INTERNAL_ERROR;
    }

    pClient = NewClient();
    if(NULL == pClient)
    {
        return TEE_COMM_OUT_OF_MEMORY;
    }

	stat = TeeInit(&(pClient->tee_context), pGuid, NULL);
	if (!TEE_IS_SUCCESS(stat))
	{
		free(pClient);
        return TEE_COMM_INTERNAL_ERROR;
	}
	
    stat = TeeConnect( &(pClient->tee_context));
    if(!TEE_IS_SUCCESS(stat))
    {
        free(pClient);
        return TEE_COMM_INTERNAL_ERROR;
    }

    if(TEE_COMM_SUCCESS != RegisterClient(&gClientContext, pClient))
    {
        free(pClient);
        return TEE_COMM_INTERNAL_ERROR;
    }

    *handle = (TEE_TRANSPORT_HANDLE)pClient->handle;

    return TEE_COMM_SUCCESS;
}

TEE_COMM_STATUS TEELIB_Disconnect(IN TEE_TRANSPORT_INTERFACE_PTR pInterface, IN TEE_TRANSPORT_HANDLE* handle)
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
        TEE_CLIENT_META_DATA* pClient = NULL;
        if(TEE_COMM_SUCCESS != UnregisterClient(&gClientContext, (uintptr_t)*handle, &pClient))
        {
            // Log error
        }

        if(NULL == pClient)
        {
            // Log error
        }
        else
        {
            TeeCancel(&(pClient->tee_context));
            TeeDisconnect(&(pClient->tee_context));
            DeleteClient(pClient); // release allocated memory for client metadata
            pClient = NULL;
        }

        *handle = TEE_TRANSPORT_INVALID_HANDLE_VALUE;
    }
    else
    {
        return TEE_COMM_INVALID_HANDLE;
    }	

    return TEE_COMM_SUCCESS;
}

TEE_COMM_STATUS TEELIB_Send(IN TEE_TRANSPORT_INTERFACE_PTR pInterface, IN TEE_TRANSPORT_HANDLE handle, IN const uint8_t* buffer, IN uint32_t length)
{
    size_t bytes_written = 0;
    size_t total_written = 0;
    TEE_CLIENT_META_DATA* pClient = NULL;
	size_t bytes_to_write = 0;
	size_t client_mtu = 0;

    if((TEE_TRANSPORT_INVALID_HANDLE_VALUE == handle) || (NULL == buffer) || (NULL == pInterface))
    {
        return TEE_COMM_INVALID_PARAMS;
    }

    if(TEE_INTERFACE_STATE_INITIALIZED != pInterface->state)
    {
        return TEE_COMM_NOT_INITIALIZED;
    }

    pClient = GetClientByHandle(&gClientContext, (uintptr_t)handle);
    if(NULL == pClient)
    {
        return TEE_COMM_INTERNAL_ERROR;
    }

	client_mtu = (size_t)pClient->tee_context.maxMsgLen;

    // Since TeeWrite might write only part of the wanted content, 
    // this loop will continue sending the remaining data until all done.
    while(total_written < length)
    {
        TEESTATUS stat = TEE_INTERNAL_ERROR;
        const char* ptr = (const char*)&buffer[total_written];

		bytes_to_write = min((size_t)(length - total_written), client_mtu);

        stat = TeeWrite(& (pClient->tee_context), ptr, bytes_to_write, &bytes_written);
        if(!TEE_IS_SUCCESS(stat))
        {
            return TEE_COMM_INTERNAL_ERROR;
        }

        total_written += bytes_written;
    }

    return TEE_COMM_SUCCESS;
}

TEE_COMM_STATUS TEELIB_Recv(IN TEE_TRANSPORT_INTERFACE_PTR pInterface, IN TEE_TRANSPORT_HANDLE handle, OUT uint8_t* buffer, IN OUT uint32_t* length)
{
    TEE_CLIENT_META_DATA* pClient = NULL;

    if((TEE_TRANSPORT_INVALID_HANDLE_VALUE == handle) || (NULL == buffer) || (NULL == length) || (NULL == pInterface))
    {
        return TEE_COMM_INVALID_PARAMS;
    }

    if(TEE_INTERFACE_STATE_INITIALIZED != pInterface->state)
    {
        return TEE_COMM_NOT_INITIALIZED;
    }

    pClient = GetClientByHandle(&gClientContext, (uintptr_t)handle);
    if(NULL == pClient)
    {
        return TEE_COMM_INTERNAL_ERROR;
    }
    
    // If no data is cached, read it from the device (TeeRead)
    if(0 == pClient->capacity)
    {
        TEESTATUS stat = TEE_INTERNAL_ERROR;
        pClient->capacity = 0;
        pClient->curr_pos = 0;

        stat = TeeRead(& (pClient->tee_context), pClient->buffer, pClient->tee_context.maxMsgLen, &pClient->capacity);
        if(!TEE_IS_SUCCESS(stat))
        {
            return TEE_COMM_INTERNAL_ERROR;
        }
    }

    // enought cached data to fill output buffer completly
    if(pClient->capacity >= (*length))
    {
        const char* ptr = (const char*)pClient->buffer;
        memcpy_s(buffer, (*length), &(ptr[pClient->curr_pos]), (*length));
        pClient->capacity -= (*length);
        pClient->curr_pos += (*length);
    }
    // just part of the output buffer will be filled with the data
    else
    {
        const char* ptr = (const char*)pClient->buffer;
        memcpy_s(buffer, (*length), &(ptr[pClient->curr_pos]), pClient->capacity);
        (*length) = (uint32_t)pClient->capacity;
		pClient->capacity = 0;
    }

    return TEE_COMM_SUCCESS;
}
