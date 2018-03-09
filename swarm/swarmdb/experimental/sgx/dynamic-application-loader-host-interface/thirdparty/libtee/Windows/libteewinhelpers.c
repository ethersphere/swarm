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
#include "Public.h"


/*********************************************************************
**                       Windows Helper Functions                   **
**********************************************************************/

/*
**	Start Overlapped Operation
**
**	Parameters:
**
**	Return:
**		TEE_INVALID_PARAMETER
**		TEE_INTERNAL_ERROR
*/
TEESTATUS TEEAPI BeginOverlappedInternal(IN TEE_OPERATION operation, IN HANDLE handle, IN PVOID buffer, IN ULONG bufferSize, OUT PEVENTHANDLE evt)
{
	TEESTATUS       status          = INIT_STATUS;
	EVENTHANDLE     pOverlapped     = NULL;
	DWORD           bytesTransferred= 0;
	BOOLEAN         optSuccesed     = FALSE;

	FUNC_ENTRY();

	if (INVALID_HANDLE_VALUE == handle || NULL == buffer || 0 == bufferSize || NULL == evt) {
		status = TEE_INVALID_PARAMETER;
		ERRPRINT("One of the parameters was illegal");
		goto Cleanup;
	}

	// allocate overlapped struct
	pOverlapped = (EVENTHANDLE)MALLOC(sizeof(OVERLAPPED));
	if (NULL == pOverlapped) {
		status = TEE_INTERNAL_ERROR;
		ERRPRINT("Error in MALLOC, error: %d\n", GetLastError());
		goto Cleanup;
	}

	pOverlapped->hEvent = CreateEvent(NULL, TRUE, FALSE, NULL);
	if (NULL == pOverlapped->hEvent) {
		status = TEE_INTERNAL_ERROR;
		ERRPRINT("Error in CreateEvent, error: %d\n", GetLastError());
		goto Cleanup;
	}


	if (operation == ReadOperation) {
		if (ReadFile(handle, buffer, bufferSize, &bytesTransferred, (LPOVERLAPPED)pOverlapped)) {		
			optSuccesed = TRUE;
		}
		else if (ERROR_IO_PENDING == GetLastError())
		{
			ERRPRINT(" ReadFile ERROR_IO_PENDING \n");
			status = TEE_SUCCESS;
		}
	}
	else if (operation == WriteOperation) {
		if (WriteFile(handle, buffer, bufferSize, &bytesTransferred, (LPOVERLAPPED)pOverlapped)) {
			optSuccesed = TRUE;
		}
		else if (ERROR_IO_PENDING == GetLastError())
		{
			ERRPRINT(" WriteFile ERROR_IO_PENDING \n");
			status = TEE_SUCCESS;
		}
	}

	if (optSuccesed == FALSE) {
		DWORD err = GetLastError();

		if (ERROR_IO_PENDING != err) {
			status = Win32ErrorToTee(err);
			ERRPRINT("Error in ReadFile/Write, error: %d\n", err);
		}
		else {
			ERRPRINT("Pending in ReadFile/Write");
			status = TEE_SUCCESS;
		}
	}
	else {
		status = TEE_SUCCESS;
	}

Cleanup:
	if (TEE_SUCCESS != status) {
		if (pOverlapped) {
			if (pOverlapped->hEvent)
				CloseHandle(pOverlapped->hEvent);
			FREE(pOverlapped);
		}
	}
	else {
		*evt = (EVENTHANDLE)pOverlapped;
	}

	FUNC_EXIT(status);

	return status;

}

TEESTATUS TEEAPI EndOverlapped(IN HANDLE handle, IN EVENTHANDLE evt, IN DWORD miliseconds, OUT OPTIONAL LPDWORD pNumberOfBytesTransferred)
{
	TEESTATUS       status                  = INIT_STATUS;
	DWORD           err                     = ERROR_INTERNAL_ERROR;
	EVENTHANDLE     pOverlapped             = evt;
	DWORD           bytesTransferred        = 0;
	LPDWORD         pBytesTransferred       = NULL;

	FUNC_ENTRY();

	if (INVALID_HANDLE_VALUE == handle || NULL == evt) {
		status = TEE_INVALID_PARAMETER;
		ERRPRINT("One of the parameters was illegal\n");
		goto Cleanup;
	}

	pBytesTransferred = pNumberOfBytesTransferred ? pNumberOfBytesTransferred : &bytesTransferred;

	// wait for the answer
	err = WaitForSingleObject(evt->hEvent, miliseconds);
	if (err == WAIT_TIMEOUT) {
		status = TEE_TIMEOUT;
		ERRPRINT("WaitForSingleObject timed out!\n");
		goto Cleanup;
	}

	if (err != WAIT_OBJECT_0) {
		assert(WAIT_FAILED == err);
		err = GetLastError();
		status = Win32ErrorToTee(err);

		ERRPRINT("WaitForSingleObject reported error: %d\n", err);
		goto Cleanup;
	}

	 // last parameter is true b/c if we're here the operation has been completed)
	if (!GetOverlappedResult(handle, (LPOVERLAPPED)pOverlapped, pBytesTransferred, TRUE)) {
		err = GetLastError();
		status = Win32ErrorToTee(err);
		ERRPRINT("Error in GetOverlappedResult, error: %d\n", err);
		goto Cleanup;
	}

	status = TEE_SUCCESS; //not really needed, but for completeness...

Cleanup:
	if (pOverlapped) {
		if (pOverlapped->hEvent)
			CloseHandle(pOverlapped->hEvent);
		FREE(pOverlapped);
	}
	FUNC_EXIT(status);

	return status;
}

DWORD WINAPI WaitForOperationEnd(LPVOID lpThreadParameter)
{
	TEESTATUS status = INIT_STATUS;
	DWORD bytesTransferred = 0;
	POPERATION_CONTEXT pOpContext = (POPERATION_CONTEXT)lpThreadParameter;

	FUNC_ENTRY();
	if (pOpContext == NULL) {
		status = ERROR_INVALID_PARAMETER;
		goto exit;
	}	

	status = EndOverlapped(pOpContext->handle, pOpContext->pOverlapped, INFINITE, &bytesTransferred);

	if (pOpContext->completionRoutine)
		pOpContext->completionRoutine(status, (size_t)bytesTransferred);

	//EndOverlapped already freed the overlapped structure
	FREE(pOpContext);
exit:
	FUNC_EXIT(status);
	return status;
}

TEESTATUS TEEAPI EndReadInternal(IN HANDLE handle, IN EVENTHANDLE evt, DWORD miliseconds, OUT OPTIONAL LPDWORD pNumberOfBytesRead)

{
	TEESTATUS status = INIT_STATUS;

	FUNC_ENTRY();

	status = EndOverlapped(handle, evt, miliseconds, pNumberOfBytesRead);

	FUNC_EXIT(status);

	return status;
}

TEESTATUS TEEAPI BeginReadInternal(IN HANDLE handle, IN PVOID buffer, IN ULONG bufferSize, OUT PEVENTHANDLE evt)

{
	TEESTATUS status = INIT_STATUS;

	FUNC_ENTRY();

	status = BeginOverlappedInternal(ReadOperation ,handle, buffer, bufferSize, evt);

	FUNC_EXIT(status);

	return status;
}

TEESTATUS TEEAPI BeginWriteInternal(IN HANDLE handle, IN const PVOID buffer, IN ULONG bufferSize, OUT PEVENTHANDLE evt)
{
	TEESTATUS status = INIT_STATUS;

	FUNC_ENTRY();

	status = BeginOverlappedInternal(WriteOperation ,handle, buffer, bufferSize, evt);

	FUNC_EXIT(status);

	return status;
}

TEESTATUS TEEAPI EndWriteInternal(IN HANDLE handle, IN EVENTHANDLE evt, DWORD miliseconds, OUT OPTIONAL LPDWORD pNumberOfBytesWritten)
{
	TEESTATUS status = INIT_STATUS;

	FUNC_ENTRY();

	status = EndOverlapped(handle, evt, miliseconds, pNumberOfBytesWritten);

	FUNC_EXIT(status);

	return status;
}

TEESTATUS TEEAPI BeginOverlapped(IN TEE_OPERATION operation, IN PTEEHANDLE handle, IN PVOID buffer, IN ULONG bufferSize, IN LPTEE_COMPLETION_ROUTINE completionRoutine)
{
	TEESTATUS               status                  = INIT_STATUS;
	TEESTATUS               tempStatus              = INIT_STATUS;
	EVENTHANDLE             pOverlapped             = NULL;
	DWORD                   bytesTransferred        = 0;
	POPERATION_CONTEXT      pOpContext              = NULL;
	HANDLE                  tHandle                 = INVALID_HANDLE_VALUE;
	DWORD                   tID                     = 0;


	FUNC_ENTRY();

	if (IS_HANDLE_INVALID(handle) || NULL == buffer || 0 == bufferSize || NULL == completionRoutine) {
		status = ERROR_INVALID_PARAMETER;
		ERRPRINT("One of the parameters was illegal");
		goto Cleanup;
	}

	// allocate overlapped struct
	pOverlapped = (EVENTHANDLE)MALLOC(sizeof(OVERLAPPED));
	if (NULL == pOverlapped) {
		status = (TEESTATUS)GetLastError();
		ERRPRINT("Error in MALLOC, error: %d\n", status);
		goto Cleanup;
	}

	pOverlapped->hEvent = CreateEvent(NULL, TRUE, FALSE, NULL);
	if (NULL == pOverlapped->hEvent) {
		status = (TEESTATUS)GetLastError();
		ERRPRINT("Error in MALLOC, error: %d\n", status);
		goto Cleanup;
	}

	if (operation == ReadOperation) {
		if (!ReadFile(handle->handle, buffer, bufferSize, &bytesTransferred, (LPOVERLAPPED)pOverlapped)) {
			status = (TEESTATUS)GetLastError();
			ERRPRINT("Error in ReadFile, error: %d\n", status);
			goto Cleanup;
		}
		else {
			status = 0;
		}
	}
	else if (operation == WriteOperation) {
		if (!WriteFile(handle->handle, buffer, bufferSize, &bytesTransferred, (LPOVERLAPPED)pOverlapped)) {
			status = (TEESTATUS)GetLastError();
			ERRPRINT("Error in WriteFile, error: %d\n", status);
			goto Cleanup;
		}
		else {
			status = 0;
		}
	}

	if (!status) {
		tempStatus = (TEESTATUS)GetLastError();  //we don't want to change the main status b/c IO_PENDING us OK

		// it's ok to get an error here, because it's overlapped
		if (ERROR_IO_PENDING != tempStatus) {
			status = tempStatus;
			ERRPRINT("Error in ReadFile, error: %d\n", status);
			goto Cleanup;
		}
	}

	//Create the operation context
	pOpContext = (POPERATION_CONTEXT)MALLOC(sizeof(OPERATION_CONTEXT));
	if (NULL == pOpContext) {
		status = (TEESTATUS)GetLastError();
		ERRPRINT("Error in MALLOC, error: %d\n", status);
		goto Cleanup;
	}

	pOpContext->handle = handle->handle;
	pOpContext->pOverlapped = pOverlapped;
	pOpContext->completionRoutine = completionRoutine;
	tHandle = CreateThread(NULL, 0, WaitForOperationEnd, pOpContext, 0, &tID);
	if (NULL == tHandle) {
		status = (TEESTATUS)GetLastError();
		ERRPRINT("Error in CreateThread , error: %d\n", status);
		goto Cleanup;
	}

	ERRPRINT("Created thread with tid %d\n", tID);

	status = TEE_SUCCESS;

Cleanup:

	if (TEE_SUCCESS != status) {
		if (pOverlapped) {
			if (pOverlapped->hEvent)
				CloseHandle(pOverlapped->hEvent);
			FREE(pOverlapped);
		}

		if (pOpContext)
			FREE(pOpContext);
	}

	FUNC_EXIT(status);

	return status;

}

/*
**	Return the given Device Path according to it's device GUID
**
**	Parameters:
**		InterfaceGuid - Device GUID
**		path - Device path buffer
**		pathSize - Device Path buffer size
**
**	Return:
**		TEE_DEVICE_NOT_FOUND
**		TEE_INVALID_PARAMETER
**		TEE_INTERNAL_ERROR
*/

TEESTATUS SendIOCTL(IN HANDLE handle, IN DWORD ioControlCode, IN LPVOID pInBuffer, IN DWORD inBufferSize, IN LPVOID pOutBuffer, IN DWORD outBufferSize, OUT LPDWORD pBytesRetuned)
{
	OVERLAPPED      overlapped = {0}; // it's OK to put the overlapped in the stack here
	TEESTATUS       status     = INIT_STATUS;

	FUNC_ENTRY();

	if (INVALID_HANDLE_VALUE == handle || NULL == pBytesRetuned) {
		status = ERROR_INVALID_PARAMETER;
		ERRPRINT("One of the parameters was illegal");
		goto Cleanup;
	}

	overlapped.hEvent = CreateEvent(NULL, TRUE, FALSE, NULL);
	if (INVALID_HANDLE_VALUE == overlapped.hEvent) {
		status = (TEESTATUS)GetLastError();
		ERRPRINT("Error in CreateEvent, error: %d\n", status);
		goto Cleanup;
	}

	if (!DeviceIoControl(handle, ioControlCode,
			     pInBuffer, inBufferSize,
			     pOutBuffer, outBufferSize,
			     pBytesRetuned, &overlapped)) {

		TEESTATUS tempStatus = (TEESTATUS)GetLastError();
		// it's ok to get an error here, because it's overlapped
		if (ERROR_IO_PENDING != tempStatus) {
			ERRPRINT("Error in DeviceIoControl, error: %d\n", tempStatus);
			status = tempStatus;
			goto Cleanup;
		}
	}


	if (!GetOverlappedResult(handle, &overlapped, pBytesRetuned, TRUE)) {
		status = (TEESTATUS)GetLastError();
		ERRPRINT("Error in GetOverlappedResult, error: %d\n", status);
		goto Cleanup;
	}

	status = TEE_SUCCESS;

Cleanup:
	if (overlapped.hEvent)
		CloseHandle(overlapped.hEvent);

	FUNC_EXIT(status);

	return status;
}

TEESTATUS Win32ErrorToTee(_In_ DWORD win32Error)
{
	switch(win32Error) {
	case ERROR_INVALID_HANDLE:
	case ERROR_INSUFFICIENT_BUFFER:
		return TEE_INVALID_PARAMETER;
	case ERROR_GEN_FAILURE:
		return TEE_UNABLE_TO_COMPLETE_OPERTAION;
	case ERROR_DEVICE_NOT_CONNECTED:
		return TEE_DEVICE_NOT_READY;

	default:
		return TEE_INTERNAL_ERROR;
	}
}
