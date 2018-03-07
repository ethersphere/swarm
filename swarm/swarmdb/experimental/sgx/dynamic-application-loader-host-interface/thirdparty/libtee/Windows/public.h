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
/*++

Module Name:

    public.h

Abstract:

    This module contains the common declarations shared by driver
    and user applications.

Environment:

    user and kernel

--*/

#ifndef __PUBLIC_H
#define __PUBLIC_H

//
// Define an Interface Guid so that app can find the device and talk to it.
//

DEFINE_GUID(GUID_DEVINTERFACE_HECI, 0xE2D1FF34, 0x3458, 0x49A9,
  0x88, 0xDA, 0x8E, 0x69, 0x15, 0xCE, 0x9B, 0xE5);
// {1b6cc5ff-1bba-4a0d-9899-13427aa05156}

#define FILE_DEVICE_HECI  0x8000

// Define Interface reference/dereference routines for
// Interfaces exported by IRP_MN_QUERY_INTERFACE

#define IOCTL_TEEDRIVER_GET_VERSION \
    CTL_CODE(FILE_DEVICE_HECI, 0x800, METHOD_BUFFERED, FILE_READ_ACCESS|FILE_WRITE_ACCESS)
#define IOCTL_TEEDRIVER_CONNECT_CLIENT \
    CTL_CODE(FILE_DEVICE_HECI, 0x801, METHOD_BUFFERED, FILE_READ_ACCESS|FILE_WRITE_ACCESS)
#define IOCTL_TEEDRIVER_WD \
    CTL_CODE(FILE_DEVICE_HECI, 0x802, METHOD_BUFFERED, FILE_READ_ACCESS|FILE_WRITE_ACCESS)
#define IOCTL_TEEDRIVER_GET_FW_STS \
    CTL_CODE(FILE_DEVICE_HECI, 0x803, METHOD_BUFFERED, FILE_READ_ACCESS|FILE_WRITE_ACCESS)

#define IOCTL_TEEDRIVER_ENTER_PG \
    CTL_CODE(FILE_DEVICE_HECI, 0x810, METHOD_BUFFERED, FILE_READ_ACCESS|FILE_WRITE_ACCESS)

#define IOCTL_TEEDRIVER_EXIT_PG \
    CTL_CODE(FILE_DEVICE_HECI, 0x811, METHOD_BUFFERED, FILE_READ_ACCESS|FILE_WRITE_ACCESS)

#define IOCTL_HECI_GET_VERSION          IOCTL_TEEDRIVER_GET_VERSION
#define IOCTL_HECI_CONNECT_CLIENT       IOCTL_TEEDRIVER_CONNECT_CLIENT
#define IOCTL_HECI_WD                   IOCTL_TEEDRIVER_WD
#define IOCTL_HECI_GET_FW_STS           IOCTL_TEEDRIVER_GET_FW_STS

#if DEBUG_IOCTLS
#if VLV
#define IOCTL_TEEDRIVER_TXEI_READ_SEC_REGISTER \
    CTL_CODE(FILE_DEVICE_HECI, 0x891, METHOD_BUFFERED, FILE_READ_ACCESS|FILE_WRITE_ACCESS)
#define IOCTL_TEEDRIVER_TXEI_READ_BRIDGE_REGISTER \
    CTL_CODE(FILE_DEVICE_HECI, 0x892, METHOD_BUFFERED, FILE_READ_ACCESS|FILE_WRITE_ACCESS)
#endif //VLV
#endif //DEBUG_IOCTLS


#pragma pack(1)
typedef struct _HECI_VERSION
{
	UINT8 major;
	UINT8 minor;
	UINT8 hotfix;
	UINT16 build;
} HECI_VERSION, TEE_VERSION;

typedef struct _FW_CLIENT
{
	UINT32 MaxMessageLength;
	UINT8  ProtocolVersion;
} FW_CLIENT, HECI_CLIENT, TEE_FW_CLIENT_PROPERTIES;
#pragma pack( )


#endif
