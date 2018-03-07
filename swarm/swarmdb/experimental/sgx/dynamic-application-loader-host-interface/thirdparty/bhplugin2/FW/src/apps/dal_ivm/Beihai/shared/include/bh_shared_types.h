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
 * @file  bh_shared_types.h
 * @brief This file declares the shared type definition across Beihai different
 *        components(host and firmware), so it should be cross-platform.
 * @author
 * @version
 *
 */
 
#ifndef __BH_SHARED_TYPES_H
#define __BH_SHARED_TYPES_H
#ifdef __cplusplus
extern "C" {
#endif

typedef char BH_I8;
typedef unsigned char BH_U8;
typedef short BH_I16;
typedef unsigned short BH_U16;
typedef int BH_I32;
typedef unsigned int BH_U32;
typedef long long BH_I64;
typedef unsigned long long BH_U64;
typedef BH_U64 BH_GROUP;
#define BH_GUID_LENGTH 16
#define BH_MAX_PACK_HASH_LEN 32

typedef struct {
    BH_U8 data[BH_MAX_PACK_HASH_LEN];
} BH_PACK_HASH;

typedef struct {
    BH_U8 data[BH_GUID_LENGTH];
} BH_TAID;
#define BH_TAID_LEN sizeof(BH_TAID)

typedef struct {
    BH_U8 data[BH_GUID_LENGTH];
} BH_SDID;

#define BH_SDID_LEN sizeof(BH_SDID)

#ifdef _WIN32
#pragma warning (disable:4200)
#endif

/*
install_condition is like properties, and formatted as "type\0key\0value\0".
Example: "string\0name\0Tom\0int\0Age\013\0"
*/
struct _bh_ta_install_condition_list_t { //same as BH_PROP_LIST, and needs discussion with CSG
    BH_U32 num; //number of properties
    BH_U32 len; //the size of data in byte
    BH_I8 data[0];
};

#ifdef __cplusplus
}
#endif

#endif

