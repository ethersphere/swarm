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
*    @file        TEETransport.c
*    @brief       Defines transport interface that is used by BH communication
*                 plugin. Defines factory method to create transport interface.
*    @author      Artum Tchachkin
*    @version     0.1
*/
#include "teetransport.h"
#include "teetransport_libtee.h"
#include "teetransport_socket.h"

#if defined(_WIN32)
#include <Windows.h>
#elif defined(__ANDROID__)
#elif defined(__linux__)
#include "teetransport_dal_device.h"
#endif

TEE_COMM_STATUS TEE_Transport_Create(IN TEE_TRANSPORT_TYPE type, OUT TEE_TRANSPORT_INTERFACE* pInterface)
{
	if(!pInterface)
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	switch(type)
	{
	case TEE_TRANSPORT_TYPE_TEE_LIB:
		return TEE_Transport_TeeLib_Create(pInterface);

	case TEE_TRANSPORT_TYPE_SOCKET:
#if defined(_WIN32)
		return TEE_Transport_Socket_Create(pInterface);
#elif defined(__ANDROID__)
		return TEE_COMM_INVALID_PARAMS;
#elif defined(__linux__)
        return TEE_Transport_Socket_Create(pInterface);
#endif	

#if defined(__linux__) && !defined(__ANDROID__)
	case TEE_TRANSPORT_TYPE_DAL_DEVICE:
		return TEE_Transport_DAL_Device_Create(pInterface);
#endif
		
	default:
		return TEE_COMM_INVALID_PARAMS;
	}
}


