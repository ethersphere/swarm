/*
 * Copyright 2010-2016 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package com.intel.security.dalinterface;
/**
 * DAL & JHI Error codes
 * and Exceptions encoding values
 */

public class DalConstants
{

	static	final String TAG 										= DalConstants.class.getSimpleName();
	final public static int  APP_ID_LEN 							= 32;									// applet id without \0 and separators

	/** 	JHI Native Error codes **/
	final public static int  JHI_FILE_MISSING_SRC					= 0x101;								// Source File not found in install/uninstall or unable to load
	// file in SendAndRecv
	final public static int  JHI_FILE_ERROR_AUTH					= 0x102;								// Attempted to load the file, but FW returned back
	// a manifest failure check and rejected it
	final public static int  JHI_FILE_ERROR_DELETE					= 0x104;								// Unable to remove file corresponding to the UUID in uninstall
	// Maybe permission issues
	final public static int  JHI_FILE_INVALID						= 0x105;								// Invalid file - bad characters or larger than 64K
	final public static int  JHI_FILE_ERROR_OPEN					= 0x106;								// Unable to open file. Maybe permission issues
	final public static int  JHI_FILE_UUID_MISMATCH					= 0x107;								// UUIDs dont match between applet file and function input
	final public static int  JHI_FILE_IDENTICAL						= 0x108;								// downloaded applet matches existing one in Jom

	final public static int  JHI_INVALID_COMMAND					= 0x202;								// invalid JHI interface command
	final public static int  JHI_ILLEGAL_VALUE						= 0x204;	  							// validation failed on input parameters

	final public static int  JHI_COMMS_ERROR						= 0x300;								// Communications error due to HECI timeouts
	// or ME auto resets or any other COMMS error
	final public static int  JHI_SERVICE_INVALID_GUID				= 0x302;								// Invalid COM guid (from DLL)

	final public static int  JHI_APPLET_TIMEOUT						= 0x401;								// This may be a result of a Java code in VM in an infinite loop.
	final public static int  JHI_APPID_NOT_EXIST					= 0x402;								// If appid is not present in app table
	final public static int  JHI_JOM_FATAL							= 0x403;								//JOM fatal error
	final public static int  JHI_JOM_OVERFLOW						= 0x404;								//exceeds max installed applets or active sessions in JOM
	final public static int  JHI_JOM_ERROR_DOWNLOAD					= 0x405;								//JOM download error
	final public static int  JHI_JOM_ERROR_UNLOAD					= 0x406;								//JOM unload error
	final public static int  JHI_ERROR_LOGGING						= 0x500;								// Error in logging
	final public static int  JHI_UNKNOWN_ERROR						= 0x600;								// Any other error

	//----------------------------------------------------------------------------------------------------------------
	//JHI Return codes
	//----------------------------------------------------------------------------------------------------------------

	//General JHI Return Code
	final public  static int  JHI_SUCCESS							= 0x00;									// general success response
	final public  static int  JHI_INVALID_HANDLE					= 0x201;								// invalid JHI handle
	final public  static int  JHI_INVALID_PARAMS					= 0x203;								// passed a null pointer to a required argument / illegal arguments passed to API function
	final public  static int  JHI_INVALID_APPLET_GUID				= JHI_ILLEGAL_VALUE;					// the applet UUID is invalid
	final public  static int  JHI_SERVICE_UNAVAILABLE				= 0x301;								// there is no connection to JHI service
	final public  static int  JHI_ERROR_REGISTRY					= 0x501;								// error for any registry based access or registry corruption
	final public  static int  JHI_ERROR_REPOSITORY_NOT_FOUND 		= 0x1000;								// when cannot find applets repository directory
	final public  static int  JHI_INTERNAL_ERROR					= 0x601;								// an unexpected internal error happened.
	final public  static int  JHI_INVALID_BUFFER_SIZE				= 0x1001;								// used a buffer that is larger than JHI_BUFFER_MAX
	final public  static int  JHI_INVALID_COMM_BUFFER				= 0x1002;								// JVM_COMM_BUFFER passed to function is invalid

	//Install errors
	final public  static int  JHI_INVALID_INSTALL_FILE				= 0x1003;								// the dalp file path is invalid
	final public  static int  JHI_READ_FROM_FILE_FAILED				= 0x1004;								// failed to read DALP file
	final public  static int  JHI_INVALID_PACKAGE_FORMAT			= 0x1005;								// dalp file format is not a valid
	final public  static int  JHI_FILE_ERROR_COPY					= 0x103;								// applet file could not be copied to repository
	final public  static int  JHI_INVALID_INIT_BUFFER				= 0x1006;								// passed an invalid init buffer to the function
	final public  static int  JHI_FILE_NOT_FOUND					= JHI_FILE_MISSING_SRC;					// could not find the specified dalp file
	final public  static int  JHI_INVALID_FILE_EXTENSION 			= 0x1007;								// applets package file must end with .dalp extension.
	final public  static int  JHI_MAX_INSTALLED_APPLETS_REACHED		= JHI_JOM_OVERFLOW;						// exceeds max applets allowed, need to uninstall an applet.
	final public  static int  JHI_INSTALL_FAILURE_SESSIONS_EXISTS 	= 0x1008;								// could not install because there are open sessions.
	final public  static int  JHI_INSTALL_FAILED					= 0x1009;								// no compatible applet was found in the DALP file
			
	//Uninstall errors
	final public  static int  JHI_DELETE_FROM_REPOSITORY_FAILURE	= JHI_FILE_ERROR_DELETE;				// unable to delete applet DALP file from repository
	final public  static int  JHI_UNINSTALL_FAILURE_SESSIONS_EXISTS	= 0x100A;								// for app uninstallation errors

	//Create Session errors
	final public  static int  JHI_APPLET_NOT_INSTALLED				= JHI_APPID_NOT_EXIST;					// trying to create a session of uninstalled applet
	final public  static int  JHI_MAX_SESSIONS_REACHED				= 0x100C;								// exceeds max sessions allowed, need to close a session.
	final public  static int  JHI_SHARED_SESSION_NOT_SUPPORTED		= 0x100D;								// the applet does not support shared sessions.
	final public  static int  JHI_MAX_SHARED_SESSION_REACHED		= 0x100E;								// failed to get session handle due to maximun handles limit.
			
	//Close Session errors
	final public  static int  JHI_INVALID_SESSION_HANDLE			= 0x100F;								// the session handle is not of an active session.

	//Send And Recieve errors
	final public  static int  JHI_INSUFFICIENT_BUFFER				= 0x200;								// buffer overflow - response greater than supplied Rx buffer
	final public  static int  JHI_APPLET_FATAL						= 0x400;								// This may be a result of uncaught exception or unusual applet
	// error that results in applet being terminated by  VM.
	//Register/Unregister session events
	final public  static int  JHI_SESSION_NOT_REGISTERED			= 0x1010;								// trying to unregister a session that is not registered for events.
	final public  static int  JHI_SESSION_ALREADY_REGSITERED		= 0x1011;								// Registration to an event is done only once.
	final public  static int  JHI_EVENTS_NOT_SUPPORTED				= 0x1012;								// events are not supported for this type of session

	//Get Applet Property errors:
	final public  static int  JHI_APPLET_PROPERTY_NOT_SUPPORTED		= 0x1013;								// Rerturned when calling GetAppletProperty with invalid property

	//Init errors
	final public  static int  JHI_SPOOLER_NOT_FOUND					= 0x1014;								// cannot find the spooler file
	final public  static int  JHI_INVALID_SPOOLER					= 0x1015;								// cannot download spooler / create an instance of the spooler
	final public  static int  JHI_NO_CONNECTION_TO_FIRMWARE			= JHI_COMMS_ERROR;						// JHI has no connection to the VM

	//DLL errors
	final public  static int  JHI_VM_DLL_FILE_NOT_FOUND				= 0x1016;								// VM DLL is missing from the exe path
	final public  static int  JHI_VM_DLL_VERIFY_FAILED				= 0x1017;								// DLL Signature or Publisher name are not valid.
	final public  static int  JHI_BAD_AAPLET_FORMAT					= 0x2001;
	//DAL_Service errors
	final public  static int  DAL_INVALID_APPLET_GUID_SIZE 			= JHI_INVALID_APPLET_GUID;
	final public  static int  DAL_INVALID_PARAMS 					= JHI_INVALID_PARAMS;
	final public  static int  DAL_INVALID_BUFFER_SIZE 				= JHI_INVALID_BUFFER_SIZE;
	final public  static int  DAL_SESSION_EXISTS 					= JHI_INSTALL_FAILURE_SESSIONS_EXISTS;
	final public  static int  DAL_INVALID_SESSION_HANDLE			= JHI_INVALID_SESSION_HANDLE;
	final public  static int  DAL_EVENTS_NOT_SUPPORTED				= JHI_EVENTS_NOT_SUPPORTED;
	final public  static int  DAL_SUCCESS							= JHI_SUCCESS;
	final public  static int  DAL_INTERNAL_ERROR					= JHI_INTERNAL_ERROR;
	final public  static int  DAL_REMOTE_EXCEPTION					= 0x3101;
	final public  static int  DAL_IO_EXCEPTION						= 0x3102;
	final public  static int  DAL_ILLEGAL_ACCESS_EXCEPTION			= 0x3103;
	final public  static int  DAL_INVOCATION_TARGET_EXCEPTION		= 0x3104;
	final public  static int  DAL_THROWABLE_EXCEPTION				= 0x3105;

	// Version Infostatic
	final public static int VERSION_BUFFER_SIZE						= 50;
	final public static int DAL_BUFFER_MAX							= 0x40000;
	final public static String[] JHI_COMMUNICATION_TYPE				= {"SOCKETS", "HECI", "INVALID_COM_TYPE"};
	final public static String[] JHI_PLATFROM_ID					= {"ME", "VLV", "INVALID_PLATFORM_ID"};

}
