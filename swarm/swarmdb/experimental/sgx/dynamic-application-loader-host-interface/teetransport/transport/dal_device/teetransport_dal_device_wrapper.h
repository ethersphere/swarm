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
*    @file        TEETransportDalDeviceWrapper.h
*    @brief       Defines internal interface for DAL Device library.
*    @author      Adam Shitrit
*    @date        December 2015
*/
#ifndef _TEE_TRANSPORT_DAL_DEVICE_WRAPPER_H_
#define _TEE_TRANSPORT_DAL_DEVICE_WRAPPER_H_

#include "teetransport.h"

#ifdef __cplusplus
extern "C" {
#endif

   TEE_COMM_STATUS DAL_Device_Teardown(IO TEE_TRANSPORT_INTERFACE_PTR pInterface);
   TEE_COMM_STATUS DAL_Device_Connect(IN TEE_TRANSPORT_INTERFACE_PTR pInterface, IN TEE_TRANSPORT_ENTITY entity, IN const char* params, OUT TEE_TRANSPORT_HANDLE* handle);
   TEE_COMM_STATUS DAL_Device_Disconnect(IN TEE_TRANSPORT_INTERFACE_PTR pInterface, IN TEE_TRANSPORT_HANDLE* handle);
   TEE_COMM_STATUS DAL_Device_Send(IN TEE_TRANSPORT_INTERFACE_PTR pInterface, IN TEE_TRANSPORT_HANDLE handle, IN const uint8_t* buffer, IN uint32_t length);
   TEE_COMM_STATUS DAL_Device_Recv(IN TEE_TRANSPORT_INTERFACE_PTR pInterface, IN TEE_TRANSPORT_HANDLE handle, OUT uint8_t* buffer, IO uint32_t* length);

#ifdef __cplusplus
};
#endif

#endif //_TEE_TRANSPORT_DAL_DEVICE_WRAPPER_H_