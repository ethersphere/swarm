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

#include "teemanagement.h"
#include "CommandInvoker.h"
#include "ServiceManager.h"
#include "Locker.h"
using namespace intel_dal;

bool serviceStarted = false;		 // an indicator that says if the service was started once already.
Locker serviceStatusLocker;			 // a lock for syncronization of serviceStarted.


void checkServiceStatus()
{
	// Not relevant for Linux/Android. Services are managed differently.
#if defined(_WIN32)
	serviceStatusLocker.Lock();
	if (!serviceStarted)
	{
		startJHIService();
		serviceStarted = true;
	}
	serviceStatusLocker.UnLock();
#endif
}

TEE_STATUS TEE_OpenSDSession (
	IN 	const char* 		sdId, 
	OUT SD_SESSION_HANDLE* 	sdHandle
	)
{
	TEE_STATUS ret = TEE_STATUS_INTERNAL_ERROR;
	CommandInvoker cInvoker;

	if (!validateUuidChar(sdId))
	{
		return TEE_STATUS_INVALID_UUID;
	}

	checkServiceStatus();

	ret = cInvoker.JhisOpenSDSession(string(sdId), sdHandle);

	return ret;
}

TEE_STATUS TEE_CloseSDSession (IN OUT SD_SESSION_HANDLE* sdHandle)
{
	TEE_STATUS ret = TEE_STATUS_INTERNAL_ERROR;
	CommandInvoker cInvoker;

	checkServiceStatus();
	ret = cInvoker.JhisCloseSDSession(sdHandle);

	return ret;
}

TEE_STATUS TEE_SendAdminCmdPkg (
	IN const SD_SESSION_HANDLE 		sdHandle,
	IN const uint8_t*			package,
	IN uint32_t					packageSize
	)
{
	TEE_STATUS ret = TEE_STATUS_INTERNAL_ERROR;
	CommandInvoker cInvoker;

	if ( (sdHandle == NULL) || (package == NULL) || (packageSize == 0) )
	{
		return TEE_STATUS_INVALID_PARAMS;
	}

	checkServiceStatus();
	ret = cInvoker.JhisSendAdminCmdPkg(sdHandle, package, packageSize);

	return ret;
}

TEE_STATUS TEE_ListInstalledTAs (
	IN 	const SD_SESSION_HANDLE 	sdHandle, 
	OUT	UUID_LIST*					uuidList
	)
{
	TEE_STATUS ret = TEE_STATUS_INTERNAL_ERROR;
	CommandInvoker cInvoker;

	if ( (sdHandle == NULL) )
	{
		return TEE_STATUS_INVALID_PARAMS;
	}

	checkServiceStatus();
	ret = cInvoker.JhisListInstalledTAs(sdHandle, uuidList);

	return ret;
}

TEE_STATUS TEE_ListInstalledSDs(
	IN 	const SD_SESSION_HANDLE 	sdHandle,
	OUT	UUID_LIST*					uuidList
	)
{
	TEE_STATUS ret = TEE_STATUS_INTERNAL_ERROR;
	CommandInvoker cInvoker;

	if ((sdHandle == NULL))
	{
		return TEE_STATUS_INVALID_PARAMS;
	}

	checkServiceStatus();
	ret = cInvoker.JhisListInstalledSDs(sdHandle, uuidList);

	return ret;
}

TEE_STATUS TEE_QueryTEEMetadata (
	IN 	const SD_SESSION_HANDLE 	sdHandle,
	OUT dal_tee_metadata*           metadata
	)
{
	TEE_STATUS ret = TEE_STATUS_INTERNAL_ERROR;
	CommandInvoker cInvoker;

	checkServiceStatus();
	ret = cInvoker.JhisQueryTEEMetadata(metadata, sizeof(dal_tee_metadata));

	return ret;
}


void TEE_DEALLOC(void* handle)
{
	JHI_DEALLOC(handle);
}