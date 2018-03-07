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
*    @file        TEETransportTeeLib.c
*    @brief       Implementation of the factory method to create TEELib transport interface.
*    @author      Artum Tchachkin
*    @date        August 2014
*/
#include "teetransport_libtee.h"
#include "teetransport_libtee_wrapper.h"
#include "teetransport_internal.h"

TEE_COMM_STATUS TEE_Transport_TeeLib_Create(OUT TEE_TRANSPORT_INTERFACE* pInterface)
{
	TEE_COMM_STATUS status = TEE_COMM_INVALID_PARAMS;

	if(!pInterface)
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	pInterface->pfnTeardown = TEELIB_Teardown;

	pInterface->pfnConnect = TEELIB_Connect;
	pInterface->pfnDisconnect = TEELIB_Disconnect;

	pInterface->pfnSend = TEELIB_Send;
	pInterface->pfnRecv = TEELIB_Recv;

	pInterface->state = TEE_INTERFACE_STATE_NOT_INITIALIZED;

	// init the transport.
	status = TEELIB_Setup(pInterface);
	if (status != TEE_COMM_SUCCESS)
	{
		goto error;
	}

	if (pInterface->state != TEE_INTERFACE_STATE_INITIALIZED)
	{
		status = TEE_COMM_INTERNAL_ERROR;
		goto error;
	}
	return TEE_COMM_SUCCESS;

error:

	pInterface->pfnTeardown = NULL;
	pInterface->pfnConnect = NULL;
	pInterface->pfnDisconnect = NULL;
	pInterface->pfnSend = NULL;
	pInterface->pfnRecv = NULL;
	pInterface->state = TEE_INTERFACE_STATE_NOT_INITIALIZED;

	return status;
}
