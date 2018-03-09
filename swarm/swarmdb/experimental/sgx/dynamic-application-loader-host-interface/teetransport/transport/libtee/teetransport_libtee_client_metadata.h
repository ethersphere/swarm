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
*    @file        TEETransportTeeLibClientMetadata.h
*    @brief       Defines structures and functions to manage TEE clients meta data.
*    @author      Artum Tchachkin
*    @date        August 2014
*/
#ifndef _TEE_TRANSPORT_TEE_LIB_WRAPPER_CLIENT_METADATA_H_
#define _TEE_TRANSPORT_TEE_LIB_WRAPPER_CLIENT_METADATA_H_

#include "teetransport.h"
#include "teetransport_internal.h"

#ifndef _WIN32
#include <stdlib.h>
#include "string_s.h"
#endif

#include "libtee.h"

#ifdef __cplusplus
extern "C" {
#endif

	struct _TEE_CLIENT_META_DATA;
	typedef struct _TEE_CLIENT_META_DATA* TEE_CLIENT_META_DATA_PTR;
	typedef struct _TEE_CLIENT_META_DATA
	{
		TEE_CLIENT_META_DATA_PTR pNext;	// pointer to next object - linked list
		uintptr_t handle;				// handle of this client, used by public transport APIs
		TEEHANDLE tee_context;			// context used with LibTee APIs
		size_t capacity;				// amount of data in the 'buffer'
		size_t curr_pos;				// index of the first byte in the 'buffer'
		void* buffer;					// used to cache data. buffer is allocated on CONNECT and released on DISCONNECT
	} TEE_CLIENT_META_DATA;

	typedef struct _TEE_CLIENT_META_DATA_CONTEXT
	{
		TEE_CLIENT_META_DATA_PTR client_list;		// connected clients, implemented as linked list
		TEE_MUTEX_HANDLE client_mutex;			// lock used to protect 'client list' modification
		size_t client_count;						// number of connected clients, mainly for debug purposes
		unsigned int internal_counter;			// this monotonic counter is used to generate 'handle' for new connected client
	} TEE_CLIENT_META_DATA_CONTEXT;

	TEE_COMM_STATUS SetupContext(IN TEE_CLIENT_META_DATA_CONTEXT* pContext);
	TEE_COMM_STATUS TeardownContext(IN TEE_CLIENT_META_DATA_CONTEXT* pContext);
	TEE_COMM_STATUS RegisterClient(IN TEE_CLIENT_META_DATA_CONTEXT* pContext, IN TEE_CLIENT_META_DATA* pClient);
	TEE_COMM_STATUS UnregisterClient(IN TEE_CLIENT_META_DATA_CONTEXT* pContext, IN uintptr_t handle, OUT TEE_CLIENT_META_DATA** ppClient);
	TEE_CLIENT_META_DATA* GetClientByHandle(IN TEE_CLIENT_META_DATA_CONTEXT* pContext, IN uintptr_t handle);
	TEE_CLIENT_META_DATA* NewClient();
	TEE_COMM_STATUS DeleteClient(IN TEE_CLIENT_META_DATA* pClient);

#ifdef __cplusplus
};
#endif

#endif //_TEE_TRANSPORT_TEE_LIB_WRAPPER_CLIENT_METADATA_H_