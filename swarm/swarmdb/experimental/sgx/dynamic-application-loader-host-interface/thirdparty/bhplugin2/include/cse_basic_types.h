/*++
INTEL CONFIDENTIAL
Copyright 2013-2015 Intel Corporation All Rights Reserved.

The source code contained or described herein and all documents
related to the source code ("Material") are owned by Intel Corporation
or its suppliers or licensors. Title to the Material remains with
Intel Corporation or its suppliers and licensors. The Material
contains trade secrets and proprietary and confidential information of
Intel or its suppliers and licensors. The Material is protected by
worldwide copyright and trade secret laws and treaty provisions. No
part of the Material may be used, copied, reproduced, modified,
published, uploaded, posted, transmitted, distributed, or disclosed in
any way without Intel's prior express written permission.

No license under any patent, copyright, trade secret or other
intellectual property right is granted to or conferred upon you by
disclosure or delivery of the Materials, either expressly, by
implication, inducement, estoppel or otherwise. Any license under such
intellectual property rights must be express and approved by Intel in
writing.
--*/

/*
File Name:
   cse_basic_types.h
Abstract:
   Basic types for the CSE.
*/

#ifndef _CSE_BASIC_TYPES_H_
#define _CSE_BASIC_TYPES_H_

#include <cse_basic_def.h>

#ifndef ANDROID
#if !MANUF_TOOLS || __APPLE__
typedef unsigned char        uint8_t;
typedef unsigned short       uint16_t;
typedef unsigned int         uint32_t;
typedef unsigned long long   uint64_t;

C_ASSERT(sizeof(uint8_t) == 1);
C_ASSERT(sizeof(uint16_t) == 2);
C_ASSERT(sizeof(uint32_t) == 4);
C_ASSERT(sizeof(uint64_t) == 8);
#endif // !MANUF_TOOLS || __APPLE__
#endif // ANDROID

#ifndef UINT8_MAX
#define UINT8_MAX (255)
#endif

#ifndef UINT16_MAX
#define UINT16_MAX (65535U)
#endif

#ifndef UINT32_MAX
#define UINT32_MAX (4294967295UL)
#endif

#ifndef UINT64_MAX
#define UINT64_MAX (18446744073709551615ULL)
#endif

#endif //_CSE_BASIC_TYPES_H_
