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
*    @file        TEETransportSocket.h
*    @brief       Defines factory method to create SOCKET transport interface.
*    @author      Artum Tchachkin
*    @date        August 2014
*/
#ifndef _TEE_TRANSPORT_SOCKET_H_
#define _TEE_TRANSPORT_SOCKET_H_

#include "teetransport.h"

#ifdef __cplusplus
extern "C" {
#endif

   /**
   * This function is used to populate the transport interface with the SOCKET
   * function poiters.
   *
   * @param[out]  pInterface
   *              caller should provide valid pointer to get the result.
   *
   * @return      TEE_COMM_SUCCESS - on success,
   *              TEE_COMM_INVALID_PARAMS - if pointer is NULL,
   *              TEE_COMM_INTERNAL_ERROR - if SOCKET layer error happens
   */
   TEE_COMM_STATUS TEE_Transport_Socket_Create(OUT TEE_TRANSPORT_INTERFACE* pInterface);

#ifdef __cplusplus
};
#endif

#endif //_TEE_TRANSPORT_SOCKET_H_