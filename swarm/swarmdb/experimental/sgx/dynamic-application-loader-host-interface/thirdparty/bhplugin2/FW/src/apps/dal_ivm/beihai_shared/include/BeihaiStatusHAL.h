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

#ifndef _BEIHAI_STATUS_HAL_H_
#define _BEIHAI_STATUS_HAL_H_

typedef enum _BH_STATUS
{
   HAL_SUCCESS                      = 0x00000000,

   HAL_TIMED_OUT                    = 0x00001001,
   HAL_FAILURE                      = 0x00001002,
   HAL_OUT_OF_RESOURCES             = 0x00001003,
   HAL_OUT_OF_MEMORY                = 0x00001004,
   HAL_BUFFER_TOO_SMALL             = 0x00001005,
   HAL_INVALID_HANDLE               = 0x00001006,
   HAL_NOT_INITIALIZED              = 0x00001007,
   HAL_INVALID_PARAMS               = 0x00001008,
   HAL_NOT_SUPPORTED                = 0x00001009,
   HAL_NO_EVENTS                    = 0x0000100A,
   HAL_NOT_READY                    = 0x0000100B,
   HAL_CONNECTION_CLOSED            = 0x0000100C,
   // ...etc

   HAL_INTERNAL_ERROR               = 0x00001100,
   HAL_ILLEGAL_FORMAT               = 0x00001101,
   HAL_LINKER_ERROR                 = 0x00001102,
   HAL_VERIFIER_ERROR               = 0x00001103,

   // User defined applet & session errors to be returned to the host (should be exposed also in the host DLL)
   HAL_FW_VERSION_MISMATCH          = 0x00002000,
   HAL_ILLEGAL_SIGNATURE            = 0x00002001,
   HAL_ILLEGAL_POLICY_SECTION       = 0x00002002,
   HAL_OUT_OF_STORAGE               = 0x00002003,
   HAL_UNSUPPORTED_PLATFORM_TYPE    = 0x00002004,
   HAL_UNSUPPORTED_CPU_TYPE         = 0x00002005,
   HAL_UNSUPPORTED_PCH_TYPE         = 0x00002006,
   HAL_UNSUPPORTED_FEATURE_SET      = 0x00002007,
   HAL_ILLEGAL_VERSION              = 0x00002008,
   HAL_ALREADY_INSTALLED            = 0x00002009,
   HAL_MISSING_POLICY               = 0x00002010,
   HAL_ILLEGAL_PLATFORM_ID          = 0x00002011,
   HAL_UNSUPPORTED_API_LEVEL        = 0x00002012,
   HAL_LIBRARY_VERSION_MISMATCH     = 0x00002013

   // ... etc

} BH_STATUS;

#endif		// _BEIHAI_STATUS_HAL_H_
