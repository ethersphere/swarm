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

/**
********************************************************************************
**
**    @file typedefs.h
**
**    @brief  Contains common type declarations used throughout the code (internal)
**
**    @author Alexander Usyskin
**
********************************************************************************
*/

#ifndef _TYPEDEFS_I_H_
#define _TYPEDEFS_I_H_

#ifdef __cplusplus
extern "C" {
#endif

#ifdef __linux__
#include <linux/uuid.h>

#ifdef __cplusplus
#include <cstdint>
#endif // __cplusplus

typedef void *                  HANDLE;
typedef void *                  HINSTANCE;
typedef void *                  HMODULE;
typedef struct{uint32_t dwHighDateTime;uint32_t dwLowDateTime;} FILETIME;

#if defined(__GNUC__)
# define C_ASSERT(e) extern void __C_ASSERT__(int [(e)?1:-1])
#endif // __GNUC__

typedef struct _GUID {
    uint32_t Data1;
    uint16_t Data2;
    uint16_t Data3;
    uint8_t  Data4[8];
} GUID, UUID;

/*
#define INFINITE            0xFFFFFFFF  // Infinite timeout

#define UUID_LE(a, b, c, d0, d1, d2, d3, d4, d5, d6, d7)                \
((uuid_le)                                                              \
{{ (a) & 0xff, ((a) >> 8) & 0xff, ((a) >> 16) & 0xff, ((a) >> 24) & 0xff, \
   (b) & 0xff, ((b) >> 8) & 0xff,                                       \
   (c) & 0xff, ((c) >> 8) & 0xff,                                       \
   (d0), (d1), (d2), (d3), (d4), (d5), (d6), (d7) }})
*/
#endif // __linux__

#ifdef __cplusplus
}
#endif

#endif // _TYPEDEFS_I_H

