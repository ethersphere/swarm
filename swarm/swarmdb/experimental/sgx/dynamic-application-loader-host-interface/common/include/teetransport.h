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
*    @file        TEETransport.h
*    @brief       Defines transport interface which is used to communicate with DAL.
*    @author      Artum Tchachkin
*    @date        August 2014
*/
#ifndef _TEE_TRANSPORT_H_
#define _TEE_TRANSPORT_H_

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif


#ifndef IN
#define IN
#endif

#ifndef OUT
#define OUT
#endif

#ifndef IO
#define IO
#endif


   /**
    * These are the return values from the transport interface functions.
    */
    typedef enum _TEE_COMM_STATUS
    {
        TEE_COMM_SUCCESS                            =   0,
        TEE_COMM_INTERNAL_ERROR                     =  -1,
        TEE_COMM_INVALID_PARAMS                     =  -2,
        TEE_COMM_INVALID_HANDLE                     =  -3,
        TEE_COMM_ILLEGAL_USAGE                      =  -4,
        TEE_COMM_NOT_INITIALIZED                    =  -5,
        TEE_COMM_NO_FW_CONNECTION                   =  -6,
        TEE_COMM_NOT_AVAILABLE                      =  -7,
        TEE_COMM_ALREADY_EXISTS                     =  -8,
        TEE_COMM_PLUGIN_FAILED                      =  -9,
        TEE_COMM_TRANSPORT_FAILED                   = -10,
        TEE_COMM_OUT_OF_MEMORY                      = -11,
        TEE_COMM_BUFFER_IS_TOO_SHORT                = -12,
        TEE_COMM_BUFFER_IS_CORRUPTED                = -13,
        TEE_COMM_NOT_IMPLEMENTED                    = -14,
        TEE_COMM_OUT_OF_RESOURCE                    = -15,
        TEE_COMM_NOT_FOUND                          = -16,
        TEE_COMM_SECURITY_VERSION_ERROR             = -17
    } TEE_COMM_STATUS;


   /**
    * The interface state.
    */
    typedef enum _TEE_INTERFACE_STATE
    {       
        /* Interface was not initialized. */
        TEE_INTERFACE_STATE_NOT_INITIALIZED,

        /* Interface was initialized successfully. */
        TEE_INTERFACE_STATE_INITIALIZED
    } TEE_INTERFACE_STATE;


   /**
    * These are the valid transport entities for connection.
    */
    typedef enum _TEE_TRANSPORT_ENTITY
    {
        // IVM - Intel/Issuer Virtual Machine
        TEE_TRANSPORT_ENTITY_IVM        = 10002,

        // SDM - Security Domain Manager
        TEE_TRANSPORT_ENTITY_SDM        = 10001,

        // RTM - Run Time Manager (Launcher)
        TEE_TRANSPORT_ENTITY_RTM        = 10000,

        // SVM - Secondary Virtual Machine
        TEE_TRANSPORT_ENTITY_SVM        = 10003,

        // Custom entity (will be set in the 'params' argument in the PFN_TEE_TRANSPORT_CONNECT.
        TEE_TRANSPORT_ENTITY_CUSTOM     = 10100,
    } TEE_TRANSPORT_ENTITY;


   /**
    * These are the supported transport types.
    */
	typedef enum _TEE_TRANSPORT_TYPE
	{
		TEE_TRANSPORT_TYPE_INVALID = 0,
		TEE_TRANSPORT_TYPE_SOCKET  = 1,
		TEE_TRANSPORT_TYPE_TEE_LIB  = 2,
		TEE_TRANSPORT_TYPE_DAL_DEVICE = 3
	} TEE_TRANSPORT_TYPE;


   /**
    * The transport functions are stateless. This handle is used to 
    * pass the state between function invocations.
    */
    typedef void*  TEE_TRANSPORT_HANDLE;

    #define TEE_TRANSPORT_INVALID_HANDLE_VALUE       ((TEE_TRANSPORT_HANDLE)-1)

    // Forward declaration
    struct _TEE_TRANSPORT_INTERFACE;
    typedef struct _TEE_TRANSPORT_INTERFACE* TEE_TRANSPORT_INTERFACE_PTR;


   /**
    * This function should be called when interface is no longer needed. After the call to this function, the interface will be invalidated.
    *
    * @param[in/out]    pInterface  Pointer to the interface.
    *
    * @return      TEE_COMM_SUCCESS - on success, TEE_COMM_STATUS error code otherwise
    */
    typedef TEE_COMM_STATUS (*PFN_TEE_TRANSPORT_TEARDOWN)
    (
        IO  TEE_TRANSPORT_INTERFACE_PTR     pInterface
    );


   /**
    * This function is used to connect to a specific client in the FW.
    *
    * @param[in]    pInterface      Pointer to the interface.
    * @param[in]    entity          The client to connect to.
    * @param[in]    params          If the entity parameter is TEE_TRANSPORT_ENTITY_CUSTOM,
    *				                then the params parameter must be a string representation of the required client
    *				                (GUID for HECI or port number for sockets).
    * @param[out]   handle          The client handle that should be used in the subsequent functions invocation.
    *
    * @return      TEE_COMM_SUCCESS - on success, TEE_COMM_STATUS error code otherwise
    */
    typedef TEE_COMM_STATUS (*PFN_TEE_TRANSPORT_CONNECT)
    (
        IN  TEE_TRANSPORT_INTERFACE_PTR     pInterface, 
        IN  TEE_TRANSPORT_ENTITY            entity, 
        IN  const char*                     params,
        OUT TEE_TRANSPORT_HANDLE*           handle
    );


   /**
    * Disconnects from a previously connected client.
    *
    * @param[in]  pInterface        Pointer to the interface.
    * @param[in]  handle            The client handle returned by connect.
    *
    * @return      TEE_COMM_SUCCESS - on success, TEE_COMM_STATUS error code otherwise
    */
    typedef TEE_COMM_STATUS (*PFN_TEE_TRANSPORT_DISCONNECT) 
    (
        IN TEE_TRANSPORT_INTERFACE_PTR      pInterface, 
        IN TEE_TRANSPORT_HANDLE*            handle
    );


   /**
    * Sends a buffer to a previously connected client.
    *
    * @param[in]  pInterface        Pointer to the interface.
    * @param[in]  handle            The client handle returned by connect.
    * @param[in]  buffer            The data buffer to send to the client.
    * @param[in]  length            The size of the buffer argument.
    *
    * @return      TEE_COMM_SUCCESS - on success, TEE_COMM_STATUS error code otherwise
    */
    typedef TEE_COMM_STATUS (*PFN_TEE_TRANSPORT_SEND)
    (
        IN TEE_TRANSPORT_INTERFACE_PTR      pInterface, 
        IN TEE_TRANSPORT_HANDLE             handle, 
        IN const uint8_t*                   buffer, 
        IN uint32_t                         length
    );


   /**
    * Receives a buffer from a previously connected client.
    *
    * @param[in]        pInterface        Pointer to the interface.
    * @param[in]        handle            The client handle returned by connect.
    * @param[out]       buffer            The buffer to accept the client data.
    * @param[in/out]    length
    *				                      input: the size of the output buffer.
    *				                      output: the size of the copied data.
    *
    * @return      TEE_COMM_SUCCESS - on success, TEE_COMM_STATUS error code otherwise
    */
    typedef TEE_COMM_STATUS (*PFN_TEE_TRANSPORT_RECV)
    (
        IN TEE_TRANSPORT_INTERFACE_PTR      pInterface, 
        IN TEE_TRANSPORT_HANDLE             handle, 
        OUT uint8_t*                        buffer, 
        IO uint32_t*                        length
    );


   /**
    * This is the transport interface definition.
    */
    typedef struct _TEE_TRANSPORT_INTERFACE
    {
        PFN_TEE_TRANSPORT_TEARDOWN       pfnTeardown;
        PFN_TEE_TRANSPORT_CONNECT        pfnConnect;
        PFN_TEE_TRANSPORT_DISCONNECT     pfnDisconnect;
        PFN_TEE_TRANSPORT_SEND           pfnSend;
        PFN_TEE_TRANSPORT_RECV           pfnRecv;

        TEE_INTERFACE_STATE              state;
    } TEE_TRANSPORT_INTERFACE;


   /**
    * Factory method used to receive the required
    * transport interface based on the user's input.
    * This method must be called before any other API in the interface.
    *
    * @param[in]   type         The required transport interface type     
    * @param[out]  pInterface   The output interface that will be populated with the wanted transport.
    *				    
    * @return      TEE_COMM_SUCCESS - on success, TEE_COMM_STATUS error code otherwise
    */
#ifdef _WIN32	
    __declspec(dllexport) // Visibility outside of dll. On Linux it's public by default.
#endif
	TEE_COMM_STATUS TEE_Transport_Create
    (
        IN  TEE_TRANSPORT_TYPE              type, 
        OUT TEE_TRANSPORT_INTERFACE*        pInterface
    );


#ifdef __cplusplus
};
#endif


#endif      //_TEE_TRANSPORT_H_