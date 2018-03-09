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

// Disable missing documentation warnings
#pragma warning disable 1591

using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace Intel.Dal
{
    /// <summary>
    /// JHI 8.0 return codes
    /// </summary>
    public enum JHI_ERROR_CODE : uint
    {
        // General JHI Return Code			
        /// <summary>general success response</summary>
        JHI_SUCCESS = 0x00,

        // error code for all the unknown error
        JHI_UNKOWN_ERROR_CODE = 0x01,

        /// <summary>invalid JHI handle</summary>
        JHI_INVALID_HANDLE = 0x201,

        /// <summary>passed a null pointer to a required argument / illegal arguments passed to API function</summary>
        JHI_INVALID_PARAMS = 0x203,

        /// <summary>the applet UUID is invalid</summary>
        JHI_INVALID_APPLET_GUID = 0x204,

        /// <summary>there is no connection to JHI service</summary>
        JHI_SERVICE_UNAVAILABLE = 0x301, 

        /// <summary>error for any registry based access or registry corruption</summary>
        JHI_ERROR_REGISTRY = 0x501,	

        /// <summary>cannot find applets repository directory</summary>
        JHI_ERROR_REPOSITORY_NOT_FOUND = 0x1000,

        /// <summary>an unexpected internal error happened</summary>
        JHI_INTERNAL_ERROR = 0x601,

        /// <summary>used a buffer that is larger than JHI_BUFFER_MAX</summary>
        JHI_INVALID_BUFFER_SIZE = 0x1001,

        /// <summary>JVM_COMM_BUFFER passed to function is invalid</summary>
        JHI_INVALID_COMM_BUFFER = 0x1002,


        // Install errors
        /// <summary>the dalp file path is invalid</summary>
        JHI_INVALID_INSTALL_FILE = 0x1003,

        /// <summary>failed to read DALP file</summary>
        JHI_READ_FROM_FILE_FAILED = 0x1004,

        /// <summary>dalp file format is not a valid</summary>
        JHI_INVALID_PACKAGE_FORMAT = 0x1005,

        /// <summary>applet file could not be copied to repository</summary>
        JHI_FILE_ERROR_COPY = 0x103,

        /// <summary>passed an invalid init buffer to the function</summary>
        JHI_INVALID_INIT_BUFFER = 0x1006,

        /// <summary>could not find the specified dalp file</summary>
        JHI_FILE_NOT_FOUND = 0x101,

        /// <summary>applets package file must end with .dalp extension</summary>
        JHI_INVALID_FILE_EXTENSION = 0x1007,

        /// <summary>exceeds max applets allowed, need to uninstall an applet</summary>
        JHI_MAX_INSTALLED_APPLETS_REACHED = 0x404,

        /// <summary>could not install because there are open sessions</summary>
        JHI_INSTALL_FAILURE_SESSIONS_EXISTS = 0x1008,

        /// <summary>no compatible applet was found in the DALP file</summary>
        JHI_INSTALL_FAILED = 0x1009, 


        // Uninstall errors
        /// <summary>unable to delete applet DALP file from repository</summary>
        JHI_DELETE_FROM_REPOSITORY_FAILURE = 0x104,

        /// <summary>for app uninstallation errors</summary>
        JHI_UNINSTALL_FAILURE_SESSIONS_EXISTS = 0x100A,


        // Create Session errors
        /// <summary>trying to create a session of uninstalled applet</summary>
        JHI_APPLET_NOT_INSTALLED = 0x402,

        /// <summary>trying to create a session with one JHI API while there are sessions of another JHI API</summary>
        JHI_INCOMPATIBLE_API_VERSION = 0x100B,

        /// <summary>exceeds max sessions allowed, need to close a session</summary>
        JHI_MAX_SESSIONS_REACHED = 0x100C,

        /// <summary>the applet does not support shared sessions</summary>
        JHI_SHARED_SESSION_NOT_SUPPORTED = 0x100D,

        /// <summary>failed to get session handle due to maximun handles limit</summary>
        JHI_MAX_SHARED_SESSION_REACHED = 0x100E,

        /// <summary>trying to use more than a single instance of an applet</summary>
        JHI_ONLY_SINGLE_INSTANCE_ALLOWED = 0x1019,


        // Close Session errors
        /// <summary>the session handle is not of an active session</summary>
        JHI_INVALID_SESSION_HANDLE = 0x100F,

        // Send And Recieve errors

        /// <summary>buffer overflow - response greater than supplied Rx buffer</summary>
        JHI_INSUFFICIENT_BUFFER = 0x200,

        /// <summary>This may be a result of uncaught exception or unusual applet error that results in applet being terminated by TL VM. </summary>
        JHI_APPLET_FATAL = 0x400, 
		

        // Register/Unregister session events
        /// <summary>trying to unregister a session that is not registered for events</summary>
        JHI_SESSION_NOT_REGISTERED = 0x1010,

        /// <summary>Registration to an event is done only once</summary>
        JHI_SESSION_ALREADY_REGSITERED = 0x1011,

        /// <summary>events are not supported for this type of session</summary>
        JHI_EVENTS_NOT_SUPPORTED = 0x1012,


        // Get Applet Property errors:			
        /// <summary>Rerturned when calling GetAppletProperty with invalid property</summary>
        JHI_APPLET_PROPERTY_NOT_SUPPORTED = 0x1013,


        // Init errors
        /// <summary>cannot find the spooler file</summary>
        JHI_SPOOLER_NOT_FOUND = 0x1014,

        /// <summary>cannot download spooler / create an instance of the spooler</summary>
        JHI_INVALID_SPOOLER = 0x1015,

        /// <summary>JHI has no connection to the VM</summary>
        JHI_NO_CONNECTION_TO_FIRMWARE = 0x300,

        // DLL errors
        /// <summary>VM DLL is missing from the exe path</summary>
        JHI_VM_DLL_FILE_NOT_FOUND = 0x1016,

        /// <summary>DLL Signature or Publisher name are not valid</summary>
        JHI_VM_DLL_VERIFY_FAILED = 0x1017,
        JHI_FIRMWARE_OUT_OF_RESOURCES = 0x1018,
        JHI_IAC_SERVER_SESSION_EXIST = 0x1020,			// May occur when trying to create two sessions on an IAC server applet
        JHI_IAC_SERVER_INTERNAL_SESSIONS_EXIST = 0x1021,					// May occur when trying to close an IAC server applet session that has internal sessions
        JHI_SVL_CHECK_FAIL					=	0x1040					// install failed due to an svl check 
    }

    public enum JHI_3_TEE_ERROR_CODES
    {
        //New JHI status: TEE_STATUS
        TEE_STATUS_INTERNAL_ERROR = 0x2001,
        TEE_STATUS_INVALID_PARAMS = 0x2002,
        TEE_STATUS_INVALID_HANDLE = 0x2003,
        TEE_STATUS_INVALID_UUID = 0x2004,
        TEE_STATUS_NO_FW_CONNECTION = 0x2005,
        TEE_STATUS_NOT_SUPPORTED = 0x2006,
        TEE_STATUS_UNSUPPORTED_PLATFORM = TEE_STATUS_NOT_SUPPORTED,

        // Service errors
        TEE_STATUS_SERVICE_UNAVAILABLE = 0x2100,
        TEE_STATUS_REGISTRY_ERROR = 0x2101,
        TEE_STATUS_REPOSITORY_ERROR = 0x2102,
        TEE_STATUS_SPOOLER_MISSING = 0x2103,
        TEE_STATUS_SPOOLER_INVALID = 0x2104,
        TEE_STATUS_MISSING_PLUGIN = 0x2105,
        TEE_STATUS_PLUGIN_VERIFY_FAILED = 0x2106,

        // Package errors
        TEE_STATUS_INVALID_PACKAGE = 0x2200,
        TEE_STATUS_INVALID_SIGNATURE = 0x2201,
        TEE_STATUS_MAX_SVLS_REACHED = 0x2202,

        // Install / uninstall TA errors:
        TEE_STATUS_CMD_FAILURE_SESSIONS_EXISTS = 0x2300,
        TEE_STATUS_CMD_FAILURE = 0x2301,
        TEE_STATUS_MAX_TAS_REACHED = 0x2302,
        TEE_STATUS_MISSING_ACCESS_CONTROL = 0x2303,
        TEE_STATUS_TA_DOES_NOT_EXIST = 0x2304,
        TEE_STATUS_SVL_CHECK_FAIL = 0x2305,
        TEE_STATUS_IDENTICAL_PACKAGE = 0x2306
    }

}

// Restore missing documentation warnings
#pragma warning restore 1591
