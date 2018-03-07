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
*    @file        TEETransportDalDevice.h
*    @brief       Implementation of the factory method to create DAL Device transport interface.
*				  This device exists only on linux.
*    @author      Adam Shitrit
*    @date        December 2015
*/
#ifndef _TEE_TRANSPORT_DAL_DEVICE_H_
#define _TEE_TRANSPORT_DAL_DEVICE_H_

#include "teetransport.h"

#ifdef __cplusplus
extern "C" {
#endif

   /**
   * This function is used to populate the transport interface with the DAL Device
   * function pointers.
   *
   * @param[out]  pInterface
   *              caller should provide valid pointer to get the result.
   *
   * @return      TEE_COMM_SUCCESS - on success,
   *              TEE_COMM_INVALID_PARAMS - in case pInterface is NULL,
   *              TEE_COMM_NOT_IMPLEMENTED -  in case the DAL device doesn't exist ( exists only on linux ).
   */
   TEE_COMM_STATUS TEE_Transport_DAL_Device_Create(OUT TEE_TRANSPORT_INTERFACE* pInterface);

#ifdef __cplusplus
};
#endif

#endif //_TEE_TRANSPORT_DAL_DEVICE_H_