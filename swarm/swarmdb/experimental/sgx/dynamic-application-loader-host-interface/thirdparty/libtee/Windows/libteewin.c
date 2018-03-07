/* Copyright 2014 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
#include <assert.h>
#include <windows.h>
#include <SetupAPI.h>
#include <initguid.h>
#include <tchar.h>
#include <libtee\helpers.h>
#include "Public.h"
#include <libtee\libtee.h>
#include <cfgmgr32.h>
#include <Objbase.h>
#include <Devpkey.h>
#include <Strsafe.h>

/**********************************************************************
 **                          TEE Lib Function                         *
 **********************************************************************/
TEESTATUS TEEAPI TeeInit(IN OUT PTEEHANDLE handle, IN const UUID *uuid, IN OPTIONAL const char *device)
{

	TEESTATUS Status = ERROR_SUCCESS;
	CONFIGRET cr = CR_SUCCESS;
	PWSTR deviceInterfaceList = NULL;
	ULONG deviceInterfaceListLength = 0;
	PWSTR nextInterface;
	HRESULT hr = E_FAIL;
	WCHAR DevicePath[256];
	size_t BufLen = 256;
	HANDLE   DeviceHandle = INVALID_HANDLE_VALUE;

	cr = CM_Get_Device_Interface_List_Size(
		&deviceInterfaceListLength,
		(LPGUID)&GUID_DEVINTERFACE_HECI,
		NULL,
		CM_GET_DEVICE_INTERFACE_LIST_PRESENT);
	if (cr != CR_SUCCESS) {
		printf("Error 0x%x retrieving device interface list size.\n", cr);
		Status = TEE_INTERNAL_ERROR;
		goto Cleanup;
	}

	if (deviceInterfaceListLength <= 1) {
		printf("Error: No active device interfaces found.\n"
			" Is the sample driver loaded?");
		Status = TEE_INTERNAL_ERROR;
		goto Cleanup;
	}

	deviceInterfaceList = (PWSTR)malloc(deviceInterfaceListLength * sizeof(WCHAR));
	if (deviceInterfaceList == NULL) {
		printf("Error allocating memory for device interface list.\n");
		Status = TEE_INTERNAL_ERROR;
		goto Cleanup;
	}
	ZeroMemory(deviceInterfaceList, deviceInterfaceListLength * sizeof(WCHAR));

	cr = CM_Get_Device_Interface_List(
		(LPGUID)&GUID_DEVINTERFACE_HECI,
		NULL,
		deviceInterfaceList,
		deviceInterfaceListLength,
		CM_GET_DEVICE_INTERFACE_LIST_PRESENT);
	if (cr != CR_SUCCESS) {
		printf("Error 0x%x retrieving device interface list.\n", cr);
		Status = TEE_INTERNAL_ERROR;
		goto Cleanup;
	}

	nextInterface = deviceInterfaceList + wcslen(deviceInterfaceList) + 1;
	if (*nextInterface != UNICODE_NULL) {
		printf("Warning: More than one device interface instance found. \n"
			"Selecting first matching device.\n\n");
	}

	hr = StringCchCopy(DevicePath, BufLen, deviceInterfaceList);
	if (FAILED(hr)) {
		Status = TEE_INTERNAL_ERROR;
		printf("Error: StringCchCopy failed with HRESULT 0x%x", hr);
		goto Cleanup;
	}


	DeviceHandle = CreateFile(DevicePath, GENERIC_READ | GENERIC_WRITE,
		0, 0, OPEN_EXISTING, FILE_FLAG_OVERLAPPED, 0);

	// if getting handle failed
	if (DeviceHandle == INVALID_HANDLE_VALUE)
	{
		Status = (TEESTATUS)GetLastError();
		goto Cleanup;
	}
Cleanup:
	if (deviceInterfaceList != NULL) {
		free(deviceInterfaceList);
	}

	if (Status == ERROR_SUCCESS) {

		handle->handle = DeviceHandle;
		memcpy(&handle->uuid, uuid, sizeof(UUID));
	}
	else {
		CloseHandle(DeviceHandle);
	}

	return Status;
}

TEESTATUS TEEAPI TeeConnect(OUT PTEEHANDLE handle)
{
	TEESTATUS       status        = INIT_STATUS;
	DWORD           bytesReturned = 0;
	FW_CLIENT       fwClient      = {0};


	FUNC_ENTRY();

	if (NULL == handle) {
		status = TEE_INVALID_PARAMETER;
		ERRPRINT("One of the parameters was illegal");
		goto Cleanup;
	}

	status = SendIOCTL(handle->handle, (DWORD)IOCTL_TEEDRIVER_CONNECT_CLIENT,
						(LPVOID)&handle->uuid, sizeof(GUID),
						&fwClient, sizeof(FW_CLIENT),
						&bytesReturned);
	if (status) {
		DWORD err = GetLastError();
		status = Win32ErrorToTee(err);
		ERRPRINT("Error in SendIOCTL, error: %d\n", err);
		goto Cleanup;
	}

	handle->maxMsgLen  = fwClient.MaxMessageLength;
	handle->protcolVer = fwClient.ProtocolVersion;

	status = TEE_SUCCESS;

Cleanup:

	FUNC_EXIT(status);

	return status;
}

TEESTATUS TEEAPI TeeRead(IN PTEEHANDLE handle, IN OUT void* buffer, IN size_t bufferSize,
			 OUT OPTIONAL size_t* pNumOfBytesRead)
{
	TEESTATUS       status = INIT_STATUS;
	EVENTHANDLE     evt    = NULL;

	FUNC_ENTRY();

	if (IS_HANDLE_INVALID(handle) || NULL == buffer || 0 == bufferSize) {
		status = TEE_INVALID_PARAMETER;
		ERRPRINT("One of the parameters was illegal");
		goto Cleanup;
	}

	status = BeginReadInternal(handle->handle, buffer, (ULONG)bufferSize, &evt);
	if (status) {
		ERRPRINT("Error in BeginReadInternal, error: %d\n", status);
		goto Cleanup;
	}

	handle->evt = evt;

	status = EndReadInternal(handle->handle, evt, INFINITE, (LPDWORD)pNumOfBytesRead);
	if (status) {
		ERRPRINT("Error in EndReadInternal, error: %d\n", status);
		goto Cleanup;
	}

	status = TEE_SUCCESS;

Cleanup:
	handle->evt = NULL;

	FUNC_EXIT(status);

	return status;
}

TEESTATUS TEEAPI TeeWrite(IN PTEEHANDLE handle, IN const void* buffer, IN size_t bufferSize,
			  OUT OPTIONAL size_t* numberOfBytesWritten)
{
	TEESTATUS       status = INIT_STATUS;
	EVENTHANDLE     evt    = NULL;

	FUNC_ENTRY();

	if (IS_HANDLE_INVALID(handle) || NULL == buffer || 0 == bufferSize) {
		status = TEE_INVALID_PARAMETER;
		ERRPRINT("One of the parameters was illegal");
		goto Cleanup;
	}

	status = BeginWriteInternal(handle->handle, (PVOID)buffer, (ULONG)bufferSize, &evt);
	if (status) {
		ERRPRINT("Error in BeginWrite, error: %d\n", status);
		goto Cleanup;
	}

	handle->evt = evt;

	status = EndWriteInternal(handle->handle, evt, INFINITE, (LPDWORD)numberOfBytesWritten);
	if (status) {
		ERRPRINT("Error in EndWrite, error: %d\n", status);
		goto Cleanup;
	}

	status = TEE_SUCCESS;

Cleanup:
	handle->evt = NULL;
	FUNC_EXIT(status);

	return status;
}

TEESTATUS TEEAPI TeeCancel(IN PTEEHANDLE handle)
{
	TEESTATUS status = INIT_STATUS;
	DWORD ret;

	FUNC_ENTRY();

	if (!CancelIo(handle->handle)) {
		status = (TEESTATUS)GetLastError();
		goto Cleanup;
	}

	ret = WaitForSingleObject(handle->evt, CANCEL_TIMEOUT);
	if (ret != WAIT_OBJECT_0) {
		ERRPRINT("Error in WaitForSingleObject, return: %d, error: %d\n", ret, GetLastError());
		status = TEE_INTERNAL_ERROR;
		goto Cleanup;
	}

	status = TEE_SUCCESS;

Cleanup:

	FUNC_EXIT(status);

	return status;
}

VOID TEEAPI TeeDisconnect(IN PTEEHANDLE handle)
{
	FUNC_ENTRY();
	if (handle && handle->handle) {
		CloseHandle(handle->handle);
		handle->handle = INVALID_HANDLE_VALUE;
	}
	FUNC_EXIT(0);
}
