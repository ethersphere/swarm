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
   cse_basic_def.h
Abstract:
   Basic definitions for the CSE.
*/

#ifndef _CSE_BASIC_DEF_H_
#define _CSE_BASIC_DEF_H_

#if !defined(EFIX64) && !defined(WIN64) && !MANUF_TOOLS && !__APPLE__
        typedef unsigned int size_t;
#elif __APPLE__
        typedef unsigned long size_t;
#endif

// offset_of macros is a replacement to "offsetof" defined in <stddef.h>
#if defined(__GNUC__) || defined(WIN32_EMU)// GCC
    #ifndef NULL
        #define NULL (void*)0
    #endif
#else
#include <stddef.h>
#endif // __GNUC__

#if defined(__GNUC__)
    #ifndef offsetof
        #define offsetof(type, memeber) __builtin_offsetof(type, memeber)
    #endif
#elif defined(WIN32_EMU)
#define offsetof(s,m)   ((size_t)&((((s *)0)->m)))
#endif

// C_ASSERT macros isused for compile-time validations (typically size of the objects)
#ifndef C_ASSERT
#if defined(_MSC_VER)
#define C_ASSERT(e) typedef char __C_ASSERT__[(e)?1:-1]
#elif defined(__GNUC__)
#define C_ASSERT_CONCAT(x, y)    C_ASSERT_XCONCAT(x, y)
#define C_ASSERT_XCONCAT(x, y)   x ## y
#define C_ASSERT(e) extern char C_ASSERT_CONCAT(__C_ASSERT__, __COUNTER__)[(e)?1:-1] __attribute__((unused))
#else
#define C_ASSERT(e)
#endif
#endif // C_ASSERT

// force underscore into symbol that must match ld script.
#define LD_SYMBOL(name) name

#ifdef _MSC_VER

#define COMPILER_MESSAGE_STRINGIZE_HELPER(x) #x
#define COMPILER_MESSAGE_STRINGIZE(x) COMPILER_MESSAGE_STRINGIZE_HELPER(x)

#define COMPILER_MESSAGE_INTERNAL(desc) message(__FILE__ "(" COMPILER_MESSAGE_STRINGIZE(__LINE__) ") : [Developer Message] " #desc)
#define COMPILER_MESSAGE(x) __pragma(COMPILER_MESSAGE_INTERNAL(x))

#elif __GNUC__

#define COMPILER_MESSAGE_DO_PRAGMA(x) _Pragma (#x)
#define COMPILER_MESSAGE(x) COMPILER_MESSAGE_DO_PRAGMA(message ("[Developer Message] " #x))

#endif

// NUM_ELEMENTS macros provides number of elements in an array (instead of using constant explicitely)
#define NUM_ELEMENTS(array)              (sizeof(array)/sizeof((array)[0]))

#define ONE_KILO                       1024
#define ONE_MEGA                       (ONE_KILO * ONE_KILO)
#define ONE_GIGA                       (ONE_KILO * ONE_MEGA)

#define BITS_PER_BYTE                  8
#define BITS_PER_DWORD                 32

#endif //_CSE_BASIC_DEF_H_
