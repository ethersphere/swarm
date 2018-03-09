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

#ifndef __BEIHAI_H__
#define __BEIHAI_H__

#ifdef __cplusplus
extern "C" {
#endif

/* The data structures are used in the DLL file. */
#ifdef _WIN32

#include <Windows.h>
# ifdef JHIDLL
#   define DLL_EXPORT __declspec(dllexport)
# else
#   define DLL_EXPORT
# endif

#else
#define DLL_EXPORT
#endif

#include <stdint.h>

typedef void *SHANDLE;

typedef	enum {
	BH_SUCCESS = 0,

	BPE_NOT_INIT = 0xF0001000,
	BPE_SERVICE_UNAVAILABLE = 0xF0001001,
	BPE_INTERNAL_ERROR = 0xF0001002,
	BPE_COMMS_ERROR = 0xF0001003,
	BPE_OUT_OF_MEMORY = 0xF0001004,
	BPE_INVALID_PARAMS = 0xF0001005,
	BPE_MESSAGE_TOO_SHORT = 0xF0001006,
	BPE_MESSAGE_ILLEGAL = 0xF0001007,
	BPE_NO_CONNECTION_TO_FIRMWARE = 0xF0001008,
	BPE_NOT_IMPLEMENT = 0xF0001009,
	BPE_OUT_OF_RESOURCE = 0xF000100A,
	BPE_INITIALIZED_ALREADY = 0xF000100B,

/* copied from errcode.h */
	/* General errors: 0x100 */
	BHE_OUT_OF_MEMORY           = 0x101, /* Out of memory */
	BHE_BAD_PARAMETER           = 0x102, /* Bad parameters to native */
	BHE_INSUFFICIENT_BUFFER     = 0x103,
	BHE_MUTEX_INIT_FAIL         = 0x104,
	BHE_COND_INIT_FAIL          = 0x105, /* Cond init fail is not return to
					      * host now, it may be used later.
					      */
	BHE_WD_TIMEOUT              = 0x106, /* Watchdog time out */

	/* Communication: 0x200 */
	BHE_MAILBOX_NOT_FOUND       = 0x201, /* Mailbox not found */
        BHE_APPLET_CRASHED          = BHE_MAILBOX_NOT_FOUND,
	BHE_MSG_QUEUE_IS_FULL       = 0x202, /* Message queue is full */
	BHE_MAILBOX_DENIED          = 0x203, /* Mailbox is denied by firewall */

	/* Applet manager: 0x300 */
	BHE_LOAD_JEFF_FAIL          = 0x303, /* JEFF file load fail, OOM or file
					      * format error not distinct by
					      * current JEFF loading
					      * process (bool jeff_loader_load).
					      */
	BHE_PACKAGE_NOT_FOUND       = 0x304, /* Request operation on a package,
					      * but it does not exist.
					      */
	BHE_EXIST_LIVE_SESSION      = 0x305, /* Uninstall package fail because of
					      * live session exist.
					      */
	BHE_VM_INSTANCE_INIT_FAIL   = 0x306, /* VM instance init fail when create
					      * session.
					      */
	BHE_QUERY_PROP_NOT_SUPPORT  = 0x307, /* Query applet property that Beihai
					      * does not support.
					      */
	BHE_INVALID_BPK_FILE        = 0x308, /* Incorrect Beihai package format */

	BHE_VM_INSTNACE_NOT_FOUND   = 0x312, /* VM instance not found */
	BHE_STARTING_JDWP_FAIL      = 0x313, /* JDWP agent starting fail */

	/* Applet instance: 0x400 */
	BHE_UNCAUGHT_EXCEPTION      = 0x401, /* uncaught exception */
	BHE_APPLET_BAD_PARAMETER    = 0x402, /* Bad parameters to applet */
	BHE_APPLET_SMALL_BUFFER     = 0x403, /* Small response buffer */
	BHE_APPLET_BAD_STATE        = 0x404,

/* copied from HAL.h */
	HAL_TIMED_OUT                    = 0x00001001,
	HAL_FAILURE                      = 0x00001002,
	HAL_OUT_OF_RESOURCES             = 0x00001003,
	HAL_OUT_OF_MEMORY                = 0x00001004,
	HAL_BUFFER_TOO_SMALL             = 0x00001005,
	HAL_INVALID_HANDLE               = 0x00001006,
	HAL_NOT_INITIALIZED              = 0x00001007,
	HAL_INVALID_PARAMS               = 0x00001008,
	HAL_NOT_SUPPORTED                = 0x00001009,
	HAL_NO_EVENTS                    = 0x0000100A,
	HAL_NOT_READY                    = 0x0000100B,
	// ...etc

	HAL_INTERNAL_ERROR               = 0x00001100,
	HAL_ILLEGAL_FORMAT               = 0x00001101,
	HAL_LINKER_ERROR                 = 0x00001102,
	HAL_VERIFIER_ERROR               = 0x00001103,

	// User defined applet & session errors to be returned to the host (should be exposed also in the host DLL)
	HAL_FW_VERSION_MISMATCH          = 0x00002000,
	HAL_ILLEGAL_SIGNATURE            = 0x00002001,
	HAL_ILLEGAL_POLICY_SECTION       = 0x00002002,
	HAL_OUT_OF_STORAGE               = 0x00002003,
	HAL_UNSUPPORTED_PLATFORM_TYPE    = 0x00002004,
	HAL_UNSUPPORTED_CPU_TYPE         = 0x00002005,
	HAL_UNSUPPORTED_PCH_TYPE         = 0x00002006,
	HAL_UNSUPPORTED_FEATURE_SET      = 0x00002007,
	HAL_ILLEGAL_VERSION              = 0x00002008,
	HAL_ALREADY_INSTALLED            = 0x00002009,
	HAL_MISSING_POLICY               = 0x00002010
	// ... etc

} BH_ERRNO;

typedef int (*PFN_BH_TRANSPORT_SEND)    (uintptr_t handle, unsigned char* buffer, unsigned int length);
typedef int (*PFN_BH_TRANSPORT_RECEIVE) (uintptr_t handle, unsigned char* buffer, unsigned int* length);
typedef int (*PFN_BH_TRANSPORT_CLOSE)	(uintptr_t handle);

typedef struct 
{
        PFN_BH_TRANSPORT_SEND  pfnSend;
        PFN_BH_TRANSPORT_RECEIVE pfnRecv;
		PFN_BH_TRANSPORT_CLOSE pfnClose;
        unsigned int handle;
} BH_PLUGIN_TRANSPORT;

/** 
 * Invoke this function before using other API. 
 * It will try to connect ME, create a receiving thread and issues a reset command to ME.
 * 
 * 
 * @return BH_SUCCESS if succuss
 *
 * @return BPE_NO_CONNECTION_TO_FIRMWARE if failed to HECI initialation
 * @return BPE_INTERNAL_ERROR if receiver thread cannot be created or failed to execute reset command
 */
DLL_EXPORT BH_ERRNO BH_PluginInit (BH_PLUGIN_TRANSPORT* transport, int do_vm_reset);



/** 
 * Invoke this function before exiting.
 * If BH_PluginInit is not called, this function will do nothing.
 * If anything goes wrong, please call this function to release resources.
 * 
 * @return BH_SUCCESS if success
 */
DLL_EXPORT BH_ERRNO BH_PluginDeinit (void);



/** 
 * Send a Reset command to VM. Reset command makes VM close all sessions and unload all packages. This function will be blocked until VM replies the response.
 * Please call BH_PluginDeinit() to clean up when anything goes wrong.
 * 
 * 
 * @return BH_SUCCESS if success.
 *
 * @return BPE_NOT_INIT if plugin is inited correctly
 * @return BPE_OUT_OF_MEMORY if plugin cannot allocate buffer to receive message.
 * @return BPE_OUT_OF_RESOURCE if plugin cannot allocate mutex.
 * @return BPE_MESSAGE_ILLEGAL if the format of message from VM is illegal.
 * @return BPE_COMMS_ERROR if communication goes wrong.
 * @return BPE_MESSAGE_TOO_SHORT if message from VM is too short to parse.
 * 
 * @return BHE_MSG_QUEUE_IS_FULL if VM cannot receive message due to message queue is full
 */
DLL_EXPORT BH_ERRNO BH_PluginReset (void);

/** 
 * Send a Download command to VM. This function will be blocked until VM replies the response.
 * 
 * @param pAppId [IN] the applet ID to create session.
 * @param pAppBlob [IN] the buffer of applet package.
 * @param AppSize [IN] the size of pAppBlob.
 * 
 * @return BH_SUCCESS if success
 *
 * @return BPE_INVALID_PARAMS if any parameters is invalid
 * @return BPE_NOT_INIT if plugin is inited correctly
 * @return BPE_OUT_OF_MEMORY if plugin cannot allocate buffer to receive message.
 * @return BPE_OUT_OF_RESOURCE if plugin cannot allocate mutex.
 * @return BPE_MESSAGE_ILLEGAL if the format of message from VM is illegal.
 * @return BPE_COMMS_ERROR if communication goes wrong.
 * @return BPE_MESSAGE_TOO_SHORT if message from VM is too short to parse.
 *
 * @return BHE_MSG_QUEUE_IS_FULL if VM cannot receive message due to message queue is full
 *
 * @return BHE_OUT_OF_MEMORY if VM cannot allocate buffer for the applet package
 * @return HAL_ALREADY_INSTALLED if the package has been installed before
 * @return BHE_INVALID_BPK_FILE if the package format is wrong.
 * @return HAL_ILLEGAL_SIGNATURE if the signature of package is wrong
 * @return HAL_ILLEGAL_POLICY_SECTION if the policy section of package is wrong
 * @return HAL_MISSING_POLICY if the policy section is incomplete.
 * @return HAL_ILLEGAL_VERSION if the version of package is too old
 * @return HAL_OUT_OF_RESOURCES if the number of installed package exceeds the limit.
 */
DLL_EXPORT BH_ERRNO BH_PluginDownload  ( const char *pAppId, const void* pAppBlob, unsigned int AppSize);

/** 
 * Send a Unload command to VM. The Unload command makes VM unload the package of the AppId. This function will be blocked until VM replies the response.
 * 
 * @param AppId [IN] the Apple ID to unload.
 * 
 * @return BH_SUCCESS if success
 *
 * @return BPE_INVALID_PARAMS if any parameters is invalid
 * @return BPE_NOT_INIT if plugin is inited correctly
 * @return BPE_OUT_OF_MEMORY if plugin cannot allocate buffer to receive message.
 * @return BPE_OUT_OF_RESOURCE if plugin cannot allocate mutex.
 * @return BPE_MESSAGE_ILLEGAL if the format of message from VM is illegal.
 * @return BPE_COMMS_ERROR if communication goes wrong.
 * @return BPE_MESSAGE_TOO_SHORT if message from VM is too short to parse.
 *
 * @return BHE_MSG_QUEUE_IS_FULL if VM cannot receive message due to message queue is full
 *
 * @return BHE_PACKAGE_NOT_FOUND if VM cannot find the applet package by AppId.
 * @return BHE_EXIST_LIVE_SESSION if the applet still has alive sessions.
 */
DLL_EXPORT BH_ERRNO BH_PluginUnload( const char *AppId );

/** 
 * Send a CreateSession command to VM. This function will be blocked until VM replies the response.
 * Please call BH_PluginDeinit() to clean up when anything goes wrong.
 * 
 * @param pAppId [IN] the applet ID to create session.
 * @param pSession [OUT] the session handle, which is used in the function BH_PluginSendAndRecv.
 * @param initBuffer [IN] the input buffer of the CreateSession command.
 * @param length [IN] the length of input buffer  
 * 
 * @return BH_SUCCESS if success
 *
 * @return BPE_INVALID_PARAMS if any parameters is invalid
 * @return BPE_NOT_INIT if plugin is inited correctly
 * @return BPE_OUT_OF_MEMORY if plugin cannot allocate buffer to receive message.
 * @return BPE_OUT_OF_RESOURCE if plugin cannot allocate mutex.
 * @return BPE_MESSAGE_ILLEGAL if the format of message from VM is illegal.
 * @return BPE_COMMS_ERROR if communication goes wrong.
 * @return BPE_MESSAGE_TOO_SHORT if message from VM is too short to parse.
 *
 * @return BHE_MSG_QUEUE_IS_FULL if VM cannot receive message due to message queue is full
 *
 * @return BHE_PACKAGE_NOT_FOUND if VM cannot find the applet package by AppId.
 * @return HAL_OUT_OF_RESOURCES if the number of all sessions exceeds session number limit.
 * @return BHE_VM_INSTANCE_INIT_FAIL if VM cannot create instance.
 *
 * @return BHE_OUT_OF_MEMORY if VM is runing out of memory
 * @return BHE_UNCAUGHT_EXCEPTION if uncaught exception is thrown in the IntelApplet.onInit() function
 * @return BHE_WD_TIMEOUT if watchdog times out in the the IntelApplet.onInit() function.
 */
DLL_EXPORT BH_ERRNO BH_PluginCreateSession ( const char *pAppId, SHANDLE* pSession, const void* initBuffer, unsigned int length);

/** 
 * Send a CloseSession command to VM. This function will be blocked until VM replies the response.
 * 
 * @param pSession [IN] the session handle to close.
 * 
 * @return BH_SUCCESS if success
 *
 * @return BPE_INVALID_PARAMS if any parameters is invalid
 * @return BPE_NOT_INIT if plugin is inited correctly
 * @return BPE_OUT_OF_MEMORY if plugin cannot allocate buffer to receive message.
 * @return BPE_OUT_OF_RESOURCE if plugin cannot allocate mutex.
 * @return BPE_MESSAGE_ILLEGAL if the format of message from VM is illegal.
 * @return BPE_COMMS_ERROR if communication goes wrong.
 * @return BPE_MESSAGE_TOO_SHORT if message from VM is too short to parse.
 *
 * @return BHE_APPLET_CRASHED if the session doesn't exist
 * @return BHE_MSG_QUEUE_IS_FULL if VM cannot receive message due to message queue is full
 *
 * @return BHE_OUT_OF_MEMORY if VM is runing out of memory
 * @return BHE_UNCAUGHT_EXCEPTION if uncaught exception is thrown in the IntelApplet.onClose() function
 * @return BHE_WD_TIMEOUT if watchdog times out in the the IntelApplet.onClose() function.
 */
DLL_EXPORT BH_ERRNO BH_PluginCloseSession (SHANDLE pSession);

/** 
 * Send a ForceCloseSession command to VM. This function will be blocked until VM replies the response.
 * 
 * @param pSession [IN] the session handle to close.
 * 
 * @return BH_SUCCESS if success
 *
 * @return BPE_INVALID_PARAMS if any parameters is invalid
 * @return BPE_NOT_INIT if plugin is inited correctly
 * @return BPE_OUT_OF_MEMORY if plugin cannot allocate buffer to receive message.
 * @return BPE_OUT_OF_RESOURCE if plugin cannot allocate mutex.
 * @return BPE_MESSAGE_ILLEGAL if the format of message from VM is illegal.
 * @return BPE_COMMS_ERROR if communication goes wrong.
 * @return BPE_MESSAGE_TOO_SHORT if message from VM is too short to parse.
 *
 * @return BHE_APPLET_CRASHED if the session doesn't exist
 * @return BHE_MSG_QUEUE_IS_FULL if VM cannot receive message due to message queue is full
 *
 * @return BHE_OUT_OF_MEMORY if VM is runing out of memory
 */
DLL_EXPORT BH_ERRNO BH_PluginForceCloseSession (SHANDLE pSession);

/** 
 * Send a SendAndRecv command to VM. This function will be blocked until VM replies the response.
 * 
 * @param pSession [IN] the destination session handle.
 * @param nCommandId [IN] the command ID.
 * @param input [IN] the input buffer to be sent to VM.
 * @param length [IN] the length of input buffer.
 * @param output [OUT] the pointer to output buffer.
 * @param output_length [IN/OUT] the expected maximum length of output buffer / the actually length of output buffer.
 * @param pResponseCode [OUT] the command result, which is set by IntelApplet.setResponseCode()
 * 
 * @return BH_SUCCESS if success
 *
 * @return BPE_INVALID_PARAMS if any parameters is invalid
 * @return BPE_NOT_INIT if plugin is inited correctly
 * @return BPE_OUT_OF_MEMORY if plugin cannot allocate buffer to receive message.
 * @return BPE_OUT_OF_RESOURCE if plugin cannot allocate mutex.
 * @return BPE_MESSAGE_ILLEGAL if the format of message from VM is illegal.
 * @return BPE_COMMS_ERROR if communication goes wrong.
 * @return BPE_MESSAGE_TOO_SHORT if message from VM is too short to parse.
 *
 * @return BHE_APPLET_CRASHED if the session doesn't exist
 * @return BHE_MSG_QUEUE_IS_FULL if VM cannot receive message due to message queue is full
 *
 * @return BHE_APPLET_SMALL_BUFFER if parameter output_length is too smaller than what we need
 * @return BHE_OUT_OF_MEMORY if VM is runing out of memory
 * @return BHE_UNCAUGHT_EXCEPTION if uncaught exception is thrown in the IntelApplet.invokeCommand() function
 * @return BHE_WD_TIMEOUT if watchdog times out in the the IntelApplet.invokeCommand() function.
 * 
 */
DLL_EXPORT BH_ERRNO BH_PluginSendAndRecv ( SHANDLE pSession, int nCommandId, const void* input, unsigned int length, void** output, unsigned int* output_length, int* pResponseCode);

/** 
 * Send a SendAndRecv command to VM. This function will be blocked until VM replies the response.
 * 
 * @param pSession [IN] the destination session handle.
 * @param nCommandId [IN] the command ID.
 * @param input [IN] the input buffer to be sent to VM.
 * @param length [IN] the length of input buffer.
 * @param output [OUT] the pointer to output buffer.
 * @param output_length [IN/OUT] the expected maximum length of output buffer / the actually length of output buffer.
 * @param pResponseCode [OUT] the command result, which is set by IntelApplet.setResponseCode()
 * 
 * @return BH_SUCCESS if success
 *
 * @return BPE_INVALID_PARAMS if any parameters is invalid
 * @return BPE_NOT_INIT if plugin is inited correctly
 * @return BPE_OUT_OF_MEMORY if plugin cannot allocate buffer to receive message.
 * @return BPE_OUT_OF_RESOURCE if plugin cannot allocate mutex.
 * @return BPE_MESSAGE_ILLEGAL if the format of message from VM is illegal.
 * @return BPE_COMMS_ERROR if communication goes wrong.
 * @return BPE_MESSAGE_TOO_SHORT if message from VM is too short to parse.
 *
 * @return BHE_APPLET_CRASHED if the session doesn't exist
 * @return BHE_MSG_QUEUE_IS_FULL if VM cannot receive message due to message queue is full
 *
 * @return BHE_APPLET_SMALL_BUFFER if parameter output_length is too smaller than what we need
 * @return BHE_OUT_OF_MEMORY if VM is runing out of memory
 * @return BHE_UNCAUGHT_EXCEPTION if uncaught exception is thrown in the IntelApplet.invokeCommand() function
 * @return BHE_WD_TIMEOUT if watchdog times out in the the IntelApplet.invokeCommand() function.
 * 
 */
DLL_EXPORT BH_ERRNO BH_PluginSendAndRecvInternal ( SHANDLE pSession, int what, int nCommandId, const void* input, unsigned int length, void** output, unsigned int* output_length, int* pResponseCode);

/** 
 * Send a Query command to VM. This function will be blocked until VM replies the response.
 * 
 * @param AppId [IN] the applet ID to Query.
 * @param input [IN] the input buffer to be sent to VM.
 * @param length [IN] the length of input buffer.
 * @param output [OUT] the pointer to output buffer, which is allocated by beihai plugin. User should free this buffer.
 * 
 * @return BH_SUCCESS if success
 *
 * @return BPE_INVALID_PARAMS if any parameters is invalid
 * @return BPE_NOT_INIT if plugin is inited correctly
 * @return BPE_OUT_OF_MEMORY if plugin cannot allocate buffer to receive message.
 * @return BPE_OUT_OF_RESOURCE if plugin cannot allocate mutex.
 * @return BPE_MESSAGE_ILLEGAL if the format of message from VM is illegal.
 * @return BPE_COMMS_ERROR if communication goes wrong.
 * @return BPE_MESSAGE_TOO_SHORT if message from VM is too short to parse.
 *
 * @return BHE_MSG_QUEUE_IS_FULL if VM cannot receive message due to message queue is full
 * 
 * @return BHE_PACKAGE_NOT_FOUND if no such applet.
 * @return BHE_QUERY_PROP_NOT_SUPPORT if the property doesn't exist in the applet.
 *
 * @return BHE_OUT_OF_MEMORY if VM is runing out of memory
 *
 */
DLL_EXPORT BH_ERRNO BH_PluginQueryAPI  ( const char *AppId, const void* input, unsigned int length, char** output);


/** 
 * Send a List Property Names Command to VM. This function return all property names of the specific applet. The result is stored in the parameter properties. 
 * 
 * @param AppId		[IN]  the applet id to query
 * @param number	[IN]  the number of applet properties
 * @param properties	[OUT] the result of property names
 * 
 * @return BH_SUCCESS if success
 *
 * @return BPE_INVALID_PARAMS if any parameters is invalid
 * @return BPE_NOT_INIT if plugin is inited correctly
 * @return BPE_OUT_OF_MEMORY if plugin cannot allocate buffer to receive message.
 * @return BPE_OUT_OF_RESOURCE if plugin cannot allocate mutex.
 * @return BPE_MESSAGE_ILLEGAL if the format of message from VM is illegal.
 * @return BPE_COMMS_ERROR if communication goes wrong.
 * @return BPE_MESSAGE_TOO_SHORT if message from VM is too short to parse.
 *
 * @return BHE_MSG_QUEUE_IS_FULL if VM cannot receive message due to message queue is full
 *
 * @return BHE_PACKAGE_NOT_FOUND if no such applet.
 *
 * @return BHE_OUT_OF_MEMORY if VM is runing out of memory
 */
DLL_EXPORT BH_ERRNO BH_PluginListProperties ( const char* AppId, int *number, char*** properties);

/** 
 * Send a List Sessions Command to VM. This function return count and array of Session Handle if success.
 * 
 * @param AppId [IN] Applet ID to query
 * @param count [OUT] count of Sessions
 * @param array [OUT] Array of Session Handle. Allocated by plugin, please release after used.
 * 
 * @return BH_SUCCESS if success
 *
 * @return BPE_INVALID_PARAMS if any parameters is invalid
 * @return BPE_NOT_INIT if plugin is inited correctly
 * @return BPE_OUT_OF_MEMORY if plugin cannot allocate buffer to receive message.
 * @return BPE_OUT_OF_RESOURCE if plugin cannot allocate mutex.
 * @return BPE_MESSAGE_ILLEGAL if the format of message from VM is illegal.
 * @return BPE_COMMS_ERROR if communication goes wrong.
 * @return BPE_MESSAGE_TOO_SHORT if message from VM is too short to parse.
 *
 * @return BHE_MSG_QUEUE_IS_FULL if VM cannot receive message due to message queue is full
 *
 * @return BHE_PACKAGE_NOT_FOUND if no such applet.
 *
 * @return BHE_OUT_OF_MEMORY if VM is runing out of memory
 */
DLL_EXPORT BH_ERRNO BH_PluginListSessions ( const char* AppId, int* count, SHANDLE** array);

/** 
 * Send a List Packages Command to VM. This function
 * 
 * @param number [OUT] number of packages
 * @param array [OUT] Strings Array of packages ID. Allocated by plugin, please release after used.
 * 
 * @return BH_SUCCESS if success
 *
 * @return BPE_INVALID_PARAMS if any parameters is invalid
 * @return BPE_NOT_INIT if plugin is inited correctly
 * @return BPE_OUT_OF_MEMORY if plugin cannot allocate buffer to receive message.
 * @return BPE_OUT_OF_RESOURCE if plugin cannot allocate mutex.
 * @return BPE_MESSAGE_ILLEGAL if the format of message from VM is illegal.
 * @return BPE_COMMS_ERROR if communication goes wrong.
 * @return BPE_MESSAGE_TOO_SHORT if message from VM is too short to parse.
 *
 * @return BHE_MSG_QUEUE_IS_FULL if VM cannot receive message due to message queue is full
 *
 * @return BHE_OUT_OF_MEMORY if VM is runing out of memory
 */
DLL_EXPORT BH_ERRNO BH_PluginListPackages (int *number, char*** array);

/** 
 * Function to free memory
 * 
 * @param ptr of buffer to free
 */
DLL_EXPORT void BH_FREE (void* ptr);

#ifdef __cplusplus
};
#endif


#endif

/* Local Variables: */
/* mode:c           */
/* c-basic-offset: 4 */
/* indent-tabs-mode: nil */
/* End:             */
