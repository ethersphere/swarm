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
/*! \file libtee.h
	\brief Tee library API
 */
#ifndef __LIBTEE_H
#define __LIBTEE_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>

#ifdef _WIN32
	#include <Windows.h>
	#include <initguid.h>

	#define TEEAPI      __stdcall
	#define UUID		GUID
#else	/* _WIN32 */
	#include <linux/uuid.h>
	#define TEEAPI
	#define UUID	uuid_le
	#ifndef IN
	   #define IN
	#endif

	#ifndef OUT
	   #define OUT
	#endif

	#ifndef OPTIONAL
	   #define OPTIONAL
	#endif
#endif /* _WIN32 */

/*! Structure to store connection data
 */
typedef struct _TEEHANDLE {

#ifdef _WIN32
	HANDLE handle;            /**< file descriptor - Handle to the Device File */
	UUID   uuid;              /**< fw client uuid */
	LPOVERLAPPED evt;
#else
	void *handle;             /**< Handle to the internal structure */
#endif
	size_t	maxMsgLen;        /**< FW Client Max Message Length */
	uint8_t protcolVer;       /**< FW Client Protocol FW */
} TEEHANDLE, *PTEEHANDLE;

#define TEEHANDLE_ZERO {0}
#define TEE_INIT_HANDLE(handle)		memset(&handle, 0, sizeof(TEEHANDLE))

/*** STATUS ***/
//TODO: Add Error defines
typedef uint16_t TEESTATUS; /**< return status for API functions */
#define TEE_ERROR_BASE                           0x0000U
#define TEE_SUCCESS                             (TEE_ERROR_BASE + 0)
#define TEE_INTERNAL_ERROR                      (TEE_ERROR_BASE + 1)
#define TEE_DEVICE_NOT_FOUND                    (TEE_ERROR_BASE + 2)
#define TEE_DEVICE_NOT_READY                    (TEE_ERROR_BASE + 3)
#define TEE_INVALID_PARAMETER                   (TEE_ERROR_BASE + 4)
#define TEE_UNABLE_TO_COMPLETE_OPERTAION        (TEE_ERROR_BASE + 5)
#define TEE_TIMEOUT                             (TEE_ERROR_BASE + 6)
#define TEE_NOTSUPPORTED                        (TEE_ERROR_BASE + 7)
#define TEE_CLIENT_NOT_FOUND                    (TEE_ERROR_BASE + 8)
#define TEE_BUSY                                (TEE_ERROR_BASE + 9)
#define TEE_DISCONNECTED                        (TEE_ERROR_BASE + 10)

#define TEE_IS_SUCCESS(Status) (((TEESTATUS)(Status)) == TEE_SUCCESS)


/*! Initializes a TEE connection
 *  \param handle A handle to the TEE device. All subsequent calls to the lib's functions
 *         must be with this handle
 *  \param uuid UUID/GUID of the FW client that want to start a session
 *  \param device optional device path, set NULL to use default
 *  \return 0 if successful, otherwise error code
 */
TEESTATUS TEEAPI TeeInit(IN OUT PTEEHANDLE handle, IN const UUID *uuid, IN OPTIONAL const char *device);

/*! Connects to the TEE driver and starts a session
 *  \param handle A handle to the TEE device
 *  \return 0 if successful, otherwise error code
 */
TEESTATUS TEEAPI TeeConnect(OUT PTEEHANDLE handle);

/*!	Read data from the TEE device synchronously.
 *  \param handle The handle of the session to read from.
 *  \param buffer A pointer to a buffer that receives the data read from the TEE device.
 *  \param bufferSize The number of bytes to be read.
 *  \param pNumOfBytesRead A pointer to the variable that receives the number of bytes read, ignored if set to NULL.
 *  \return 0 if successful, otherwise error code
 */
TEESTATUS TEEAPI TeeRead(IN PTEEHANDLE handle, IN OUT void *buffer, IN size_t bufferSize, OUT OPTIONAL size_t *pNumOfBytesRead);

/*! Writes the specified buffer to the TEE device synchronously.
 *  \param handle The handle of the session to write to.
 *  \param buffer A pointer to the buffer containing the data to be written to the TEE device.
 *  \param bufferSize The number of bytes to be written.
 *  \param numberOfBytesWritten A pointer to the variable that receives the number of bytes written, ignored if set to NULL.
 *  \return 0 if successful, otherwise error code
 */
TEESTATUS TEEAPI TeeWrite(IN PTEEHANDLE handle, IN const void *buffer, IN size_t bufferSize, OUT OPTIONAL size_t *numberOfBytesWritten);

/*!	Cancels an async operation that was started on the calling thread.
 *  This will cause the driver to disconnect the client and all subsequent requests will fail.
 *  \param handle The handle of the session that the request was sent on.
 *  \return 0 if successful, otherwise error code
 */
TEESTATUS TEEAPI TeeCancel(IN PTEEHANDLE handle);

/*! Closes the session to TEE driver
 *  Make sure that you call this function as soon as you are done with the device,
 *  as other clients might be blocked until the session is closed.
 *  \param handle The handle of the session to close.
 */
void TEEAPI TeeDisconnect(IN PTEEHANDLE handle);

#ifdef __cplusplus
}
#endif

#endif /* __LIBTEE_H */
