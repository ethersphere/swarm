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

/*
 *
 * @file  bh_acp_util.h
 * @brief This file declared beihai admin package utility api
 * @author Wenlong Feng(wenlong.feng@intel.com)
 *
 */
#ifndef _BH_ACP_UTIL_H
#define _BH_ACP_UTIL_H

#ifdef __cplusplus
extern "C" {
#endif

#include "bh_shared_types.h"

/**
 * Convert a string with hex style into the hex value
 * Example str  "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1"
 *         uuid {0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 
 *               0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa,
 *               0xaa, 0xaa, 0xaa, 0xa1}
 * @param str   [in] A string with hex style
 * @param uuid  [out] The converted UUID on succ, unchanged on failure
 * @return      return 1 if succeeded, 0 otherwise
 */
BH_I32 string_to_uuid(const BH_I8* str, BH_I8* uuid);

/**
 * Convert a hex uuid into string style
 * Example uuid {0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 
 *               0xaa, 0xaa, 0xaa, 0xaa, 0xaa, 0xaa,
 *               0xaa, 0xaa, 0xaa, 0xa1}
 *         str  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1"
 * @param uuid  [in] The hex UUID
 * @param str   [out] A hex style string on succ
 */
void uuid_to_string(const BH_I8* uuid, BH_I8* str);

/**
 * Convert a string with variable lenghth with hex style into the hex value
 * Example str  "0123456abD"
 *         out {0x01, 0x23, 0x45, 0x6a, 0xbD}
 * @param str   [in] A string with hex style
 * @str_len the  [in] the length of the string
 * @param uuid  [out] The converted UUID on succ, unchanged on failure
 * @return      return 1 if succeeded, 0 otherwise
 */
BH_I32 hexstring_to_binary(const BH_I8* str, BH_U32 str_len, BH_I8* out);

#ifdef __cplusplus
}
#endif

#endif
