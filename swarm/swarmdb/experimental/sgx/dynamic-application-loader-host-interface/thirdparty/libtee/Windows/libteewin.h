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
#ifndef __TEELIBWIN_H
#define __TEELIBWIN_H

#include <Windows.h>
#include "LibTee.h"
#if (_MSC_PLATFORM_TOOLSET < 140)
#ifndef _Out_writes_

#define _Out_writes_(x)

#endif
#endif
#define CANCEL_TIMEOUT 5000

/*********************************************************************
**                       Windows Helper Types                       **
**********************************************************************/
typedef LPOVERLAPPED EVENTHANDLE, *PEVENTHANDLE;

typedef enum _TEE_OPERATION
{
	ReadOperation,
	WriteOperation
} TEE_OPERATION, *PTEE_OPERATION;

/*
	This callback function is called when an asynchronous TEE operation is completed.

	Parameters:
		status	- The operation status. This parameter is 0 is the operation was successful.
				Otherwise it returns a Win32 error value.
		numberOfBytesTransfered - The number of bytes transferred.
				If an error occurs, this parameter is zero.
*/
typedef
void
(TEEAPI *LPTEE_COMPLETION_ROUTINE)(
	IN    TEESTATUS status,
	IN    size_t numberOfBytesTransfered
	);

typedef struct _OPERATION_CONTEXT
{
	HANDLE                          handle;
	LPOVERLAPPED                    pOverlapped;
	LPTEE_COMPLETION_ROUTINE        completionRoutine;
} OPERATION_CONTEXT, *POPERATION_CONTEXT;


/*********************************************************************
**					Windows Helper Functions 						**
**********************************************************************/
TEESTATUS TEEAPI BeginOverlappedInternal(IN TEE_OPERATION operation, IN HANDLE handle, IN PVOID buffer, IN ULONG bufferSize, OUT PEVENTHANDLE evt);
TEESTATUS TEEAPI EndOverlapped(IN HANDLE handle, IN EVENTHANDLE evt, IN DWORD miliseconds, OUT OPTIONAL LPDWORD pNumberOfBytesTransferred);
DWORD WINAPI WaitForOperationEnd(LPVOID lpThreadParameter);
TEESTATUS TEEAPI EndReadInternal(IN HANDLE handle, IN EVENTHANDLE evt, DWORD miliseconds, OUT OPTIONAL LPDWORD pNumberOfBytesRead);
TEESTATUS TEEAPI BeginReadInternal(IN HANDLE handle, IN PVOID buffer, IN ULONG bufferSize, OUT PEVENTHANDLE evt);
TEESTATUS TEEAPI BeginWriteInternal(IN HANDLE handle, IN const PVOID buffer, IN ULONG bufferSize, OUT PEVENTHANDLE evt);
TEESTATUS TEEAPI EndWriteInternal(IN HANDLE handle, IN EVENTHANDLE evt, DWORD miliseconds, OUT OPTIONAL LPDWORD pNumberOfBytesWritten);
TEESTATUS TEEAPI BeginOverlapped(IN TEE_OPERATION operation, IN PTEEHANDLE handle, IN PVOID buffer, IN ULONG bufferSize, IN LPTEE_COMPLETION_ROUTINE completionRoutine);
TEESTATUS GetDevicePath(_In_ LPCGUID InterfaceGuid, _Out_writes_(pathSize) PTCHAR path, _In_ SIZE_T pathSize);
TEESTATUS SendIOCTL(IN HANDLE handle, IN DWORD ioControlCode, IN LPVOID pInBuffer, IN DWORD inBufferSize, IN LPVOID pOutBuffer, IN DWORD outBufferSize, OUT LPDWORD pBytesRetuned);
TEESTATUS Win32ErrorToTee(_In_ DWORD win32Error);

#endif
