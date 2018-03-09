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

/*
 *
 * @file  bhp_exp.h
 * @brief This file declares the exported interface from Beihai Host Proxy (BHP) module(.dll or .so).
 *        BHP is updated version of previous Beihai Plugin.
 * @author
 * @version
 *
 */

#ifndef __BHP_EXP_H__
#define __BHP_EXP_H__

#include <cstdint>

#ifdef __cplusplus
extern "C" {
#endif

#ifdef _WIN32
#include <Windows.h>
#ifdef JHIDLL
#define DLL_EXPORT __declspec(dllexport)
#else
#define DLL_EXPORT
#endif
#else
#define DLL_EXPORT
#endif

#include "bh_shared_errcode.h"

typedef void* SD_SESSION_HANDLE; //SD session handle
typedef void* JAVATA_SESSION_HANDLE; //Java TA session handle

typedef int (*PFN_BHP_TRANSPORT_SEND)    (uintptr_t handle, unsigned char* buffer, unsigned int length);
typedef int (*PFN_BHP_TRANSPORT_RECEIVE) (uintptr_t handle, unsigned char* buffer, unsigned int* length);
typedef int (*PFN_BHP_TRANSPORT_CONNECT) (int heci_port, uintptr_t * handle);
typedef int (*PFN_BHP_TRANSPORT_CLOSE)  (uintptr_t handle);

typedef struct
{
    PFN_BHP_TRANSPORT_SEND  pfnSend;
    PFN_BHP_TRANSPORT_RECEIVE pfnRecv;
    PFN_BHP_TRANSPORT_CONNECT pfnConnect; //connect to specified heci_guid in FW
    PFN_BHP_TRANSPORT_CLOSE pfnClose; //close the connection
} BHP_TRANSPORT;

/**
 * Invoke this function before using other API.
 * It will try to connect ME processes(Launcher, SDM and I-VM), and create receiving threads 
 * for those process and do other initialization.
 *
 * do_vm_reset: if true performs a vm reset. otherwise doesn't perform a vm reset.
 *
 * @return BH_SUCCESS if succuss
 *
 * @return BPE_NO_CONNECTION_TO_FIRMWARE if failed to HECI initialation
 * @return BPE_INTERNAL_ERROR if receiver thread cannot be created or other internal failure
 */
DLL_EXPORT BH_RET BHP_Init(const BHP_TRANSPORT* transport, int do_vm_reset);


/**
 * Invoke this function before exiting.
 * If BHP_Init is not called, this function will do nothing.
 * If anything goes wrong, please call this function to release resources.
 *
 * do_vm_reset: if true performs a vm reset. otherwise doesn't perform a vm reset.
 *
 * @return BH_SUCCESS if success
 */
DLL_EXPORT BH_RET BHP_Deinit(int do_vm_reset);

/**
 * Send Reset command to SDM, Launcher and VM, to let them enter initial state.
 * This function will be blocked until receiving all the responses.
 *
 * @return BH_SUCCESS if success.
 */
DLL_EXPORT BH_RET BHP_Reset (void);

/**
 * Open SD(Security Domain) Session.
 *
 * @param SD_ID [IN] the SD ID(UUID) to be opened.
 * @param handle [OUT] the returned session handle.
 *
 * @return BH_SUCCESS if succuss
 *
 */
DLL_EXPORT BH_RET BHP_OpenSDSession(const char* SD_ID, SD_SESSION_HANDLE* pSession);

/**
 * close specified sd session.
 * @param handle [IN] the sd session handle.
 *
 * @return BH_SUCCESS if success
 */
DLL_EXPORT BH_RET BHP_CloseSDSession(const SD_SESSION_HANDLE handle);

/**
  * send admin cmd package to SD.
  * The admin cmd pkg could be for installsd, uninstallsd, install_javata,uninstall_javata,
  * install_nativeta, uninstall_nativeta.
  *
  * @param handle [IN] the sd session handle.
  * @param cmd_pkg [IN] admin cmd package.
  *        If it is Install Java/Native TA CMD, it shall include both ACP and the according Java/Native TA package.
  *        The two parts shall be concatenated continuously.
  *        The pkg_len parameter shall be the total length of these two parts.
  * @param pkg_len [IN] admin cmd package length.
  *
  * @return BH_SUCCESS if success
  */
DLL_EXPORT BH_RET BHP_SendAdminCmdPkg(const SD_SESSION_HANDLE handle, const char* cmd_pkg, unsigned int pkg_len);

/**
 * Open Session to specified Java TA.
 * The Firmware side might need to spawn the VM process, and create the TA instance.
 * This function will also connect to the VM process's heci address.
 * This function will block until VM replied the response.
 * Please call BHP_Deinit() to clean up when anything goes wrong.
 *
 * @param pSession [OUT] the ta session handle, which is used in the function BHP_SendAndRecv.
 * @param TA_Id [IN] the applet ID (UUID) to create session.
 * @param TA_pkg [in] TA binary package, i.e, .bpk data.
 * @param pkg_len [in] The length of TA binary package in bytes.
 * @param init_param [IN] the input buffer of the CreateSession command.
 * @param param_len [IN] the length of init_param in bytes
 *
 * @return BH_SUCCESS if success
 *
 */
DLL_EXPORT BH_RET BHP_OpenTASession (JAVATA_SESSION_HANDLE* pSession, const char *TA_Id,  const char* TA_pkg, unsigned int pkg_len,
                    const char* init_param, unsigned int param_len);

/**
 * Send a CloseTASession command to VM to close the specified Java TA session.
 * This function will be blocked until VM replies the response.
 *
 * @param pSession [IN] the java ta session handle to close.
 *
 * @return BH_SUCCESS if success
 *
 */
DLL_EXPORT BH_RET BHP_CloseTASession (const JAVATA_SESSION_HANDLE handle);

/** 
 * Send a ForceCloseSession command to VM. This function will be blocked until VM replies the response.
 * 
 * @param pSession [IN] the session handle to close.
 * 
 * @return BH_SUCCESS if success
 *
 */
DLL_EXPORT BH_RET BHP_ForceCloseTASession (const JAVATA_SESSION_HANDLE handle);

/**
 * Send a SendAndRecv command to VM. This function will be blocked until VM replies the response.
 *
 * @param pSession [IN] the java ta session handle.
 * @param nCommandId [IN] the command ID.
 * @param input [IN] the input buffer to be sent to TA.
 * @param length [IN] the length of input buffer.
 * @param output [OUT] the pointer to output buffer.
 * @param output_length [IN/OUT] the expected maximum length of output buffer / the actually length of output buffer.
 * @param pResponseCode [OUT] the command result, which is set by IntelApplet.setResponseCode()
 *
 * @return BH_SUCCESS if success
 *
 */
DLL_EXPORT BH_RET BHP_SendAndRecv (const JAVATA_SESSION_HANDLE handle, int nCommandId, const void* input, unsigned int length,
    void** output, unsigned int* output_length, int* pResponseCode);

/**
 * Send a SendAndRecvInternal command to VM. This function will be blocked until VM replies the response.
 *
 * @param pSession [IN] the java ta session handle.
 * @param what [IN] the message type.
 * @param nCommandId [IN] the command ID.
 * @param input [IN] the input buffer to be sent to TA.
 * @param length [IN] the length of input buffer.
 * @param output [OUT] the pointer to output buffer.
 * @param output_length [IN/OUT] the expected maximum length of output buffer / the actually length of output buffer.
 * @param pResponseCode [OUT] the command result, which is set by IntelApplet.setResponseCode()
 *
 * @return BH_SUCCESS if success
 *
 */
DLL_EXPORT BH_RET BHP_SendAndRecvInternal (const JAVATA_SESSION_HANDLE handle, int what, int nCommandId,
    const void* input, unsigned int length, void** output, unsigned int* output_length, int* pResponseCode);

/** 
 * Send a ListTAProperties command to VM. This function return all property names of the specific TA.
 * This function only works for the TA whose TA package is already downloaded to VM.
 * 
 * @param appId[IN]  the TA uuid string to query
 * @param count[OUT]  the number of TA properties
 * @param properties[OUT] the result list of property names
 * 
 * @return BH_SUCCESS if success
 *
 */
DLL_EXPORT BH_RET BHP_ListTAProperties (const char* appId, unsigned int *count, char*** properties);

/** 
 * Send a QueryTAProperty command to VM. This function will be blocked until VM replies the response.
 * This function only works for the TA whose TA package is already downloaded to VM.
 * 
 * @param appId [IN] the TA ID to Query.
 * @param property [IN] the property name to be queried.
 * @param length [IN] the length of property.
 * @param output [OUT] the pointer to output buffer, which is allocated by BHP and should be freed after used.
 * 
 * @return BH_SUCCESS if success
 *
 */
DLL_EXPORT BH_RET BHP_QueryTAProperty (const char *appId, const void* property, unsigned int length, char** output);

/** 
 * Send a ListTASessions comand to VM. This function return count and array of Session Handle if success.
 * 
 * @param appId [IN] Java TA ID to query
 * @param count [OUT] count of Sessions
 * @param pSessions [OUT] Array of Session Handles. Allocated by BHP,  and should be freed after used.
 * 
 * @return BH_SUCCESS if success
 *
 */
DLL_EXPORT BH_RET BHP_ListTASessions (const char* appId, unsigned int* count, JAVATA_SESSION_HANDLE** pSessions);

/**
 * Functions to get the list of downloaded TA Packages in Firmware for specific SD.
 *   It will list both JavaTA and Native TA.
 *
 * @param SD_ID [IN] the specific SD UUID string
 * @param count [OUT] number of packages
 * @param appIdStrs [OUT] String Array of TA UUID. Allocated by bhp, please free it after usage.
 *
 * @return BH_SUCCESS if success
 *
 */
DLL_EXPORT BH_RET BHP_ListDownloadedTAs (const char* SD_ID, unsigned int *count, char*** appIdStrs);

/**
 * List intalled SD information.
 * It will send listSD cmd to SDM process.
 *
 * @param handle [IN] SD Session Handle
 * @param count [OUT] count of the sd ids.
 * @param sdIdStrs [OUT] installed sd UUID strings. It is allocated inside this function,
 *        and caller should be responsible to free it using BHP_Free().
 *
 * @return BH_SUCCESS if success
 *
 */
DLL_EXPORT BH_RET BHP_ListInstalledSDs(const SD_SESSION_HANDLE handle, unsigned int* count, char*** sdIdStrs);

/**
 * List intalled TA in specified SD.
 * It will send listTA cmd to SDM process.
 *
 * @param handle [IN] SD Session Handle
 * @param SD_ID [IN] sd UUID string.
 * @param count [OUT] count of the ta UUIDs.
 * @param appIdStrs [OUT] installed ta UUID strings. It is allocated inside this function, 
 *        and caller should be responsible to free it using BHP_Free().
 *
 * @return BH_SUCCESS if success
 *
 */
DLL_EXPORT BH_RET BHP_ListInstalledTAs(const SD_SESSION_HANDLE handle, const char* SD_ID, unsigned int * count, char*** appIdStrs);

/**
 * Function to free memory allocated in BHP
 *
 * @param ptr of buffer to free
 */
DLL_EXPORT void BHP_Free (void* ptr);

/**
  * Run native TA.
  * It will download the TA package to firmware, then native TA will be spawned.
  *
  * @param TA_ID [IN] the TA UUID.
  * @param TA_pkg [IN] The native ta package.
  * @param pkg_len [IN] the length of ta package in bytes.
  *
  * @return BH_SUCCESS if success
  *
  */
DLL_EXPORT BH_RET BHP_RunNativeTA(const char* TA_ID, const char* TA_pkg, unsigned int pkg_len);

/**
  * Stop Native TA.
  * It will send StopNativeTA cmd to Launcher, which will kill the native TA process.
  *
  * @param TA_ID [IN]  the TA UUID.
  *
  * @return BH_SUCCESS if success
  *
  */
DLL_EXPORT BH_RET BHP_StopNativeTA(const char* TA_ID);

/**
  * Query Firmware TEE Metadata information.
  * It will send QueryTEEMetadata cmd to IVM to retrieve the meta information blob.
  *
  * @param metadata [OUT] the pointer to output buffer, which is allocated by BHP and should be freed after used.
  * @param length [OUT] the length of the returned output buffer.
  *
  * @return BH_SUCCESS if success
  *
  */
DLL_EXPORT BH_RET BHP_QueryTEEMetadata(unsigned char** metadata, unsigned int* length);
#ifdef __cplusplus
}
#endif

#endif
