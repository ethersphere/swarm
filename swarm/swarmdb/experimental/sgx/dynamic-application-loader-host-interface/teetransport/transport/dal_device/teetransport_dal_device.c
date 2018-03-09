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
*    @file        TEETransportDalDevice.c
*    @brief       Implementation of the factory method to create DAL Device transport interface.
*                 This device exists only on linux.
*    @author      Adam Shitrit
*    @date        December 2015
*/
#include "teetransport_dal_device.h"
#include "teetransport_dal_device_wrapper.h"
#include "teetransport_internal.h"

TEE_COMM_STATUS TEE_Transport_DAL_Device_Create(OUT TEE_TRANSPORT_INTERFACE* pInterface)
{

#ifdef WIN32
	
	return TEE_COMM_NOT_IMPLEMENTED;

#else

	if(!pInterface)
	{
		return TEE_COMM_NOT_IMPLEMENTED;
	}

	pInterface->pfnTeardown = DAL_Device_Teardown;

	pInterface->pfnConnect = DAL_Device_Connect;
	pInterface->pfnDisconnect = DAL_Device_Disconnect;

	pInterface->pfnSend = DAL_Device_Send;
	pInterface->pfnRecv = DAL_Device_Recv;

	pInterface->state = TEE_INTERFACE_STATE_INITIALIZED;


	//no further init' is required, dal device is initialized when it's loaded to the kernel.

	return TEE_COMM_SUCCESS;

#endif
}
