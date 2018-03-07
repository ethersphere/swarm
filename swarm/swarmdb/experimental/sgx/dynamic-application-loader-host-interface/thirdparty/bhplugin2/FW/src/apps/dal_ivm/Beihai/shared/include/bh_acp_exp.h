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
 * @file   bh_acp_exp.h
 * @author Wenlong <wenlong.feng@intel.com>
 * @date   Fri Oct 11 11:30:33 2013
 *
 * @brief  This file defined external data structure and api for application layer
 *         pload is short for "Parse and Load"
 *         These api are used to parse the format of the raw_data and assign each item pointer with correct address
 *         caller should be responsible to ensure raw_data is not released when using parsed pack struct
 */
#ifndef __BH_ACP_EXP_H
#define __BH_ACP_EXP_H

#ifdef __cplusplus
extern "C" {
#endif

#include "bh_shared_errcode.h"
#include "bh_acp_format.h"

// ALEX: padding issue
#pragma pack(1)

typedef struct {
    ACPackHeader* head;
    ACInsSDPack cmd_pack;
} ACInsSDPackExt;

typedef struct {
    ACPackHeader* head;
    ACUnsSDPack cmd_pack;
} ACUnsSDPackExt;

typedef struct {
    ACPackHeader* head;
    ACInsJTAPack cmd_pack;
    char* ta_pack;
} ACInsJTAPackExt;

typedef struct {
    ACPackHeader* head;
    ACInsNTAPack cmd_pack;
    char* ta_pack;
} ACInsNTAPackExt;

typedef struct {
    ACPackHeader *head;
    ACUnsTAPack cmd_pack;
} ACUnsTAPackExt;

typedef struct {
    ACInsJTAProp cmd_pack;
    char* jeff_pack;
} ACInsJTAPropExt;

typedef struct {
    ACPackHeader* head;
    ACUpdateSVLPack cmd_pack;
} ACUpdateSVLPackExt;

typedef struct {
    ACProp *props;
} ACTAProps;

#pragma pack()

BH_RET ACP_pload_ins_sd(const void* raw_data, unsigned size, ACInsSDPackExt *pack);
BH_RET ACP_pload_uns_sd(const void* raw_data, unsigned size, ACUnsSDPackExt *pack);
BH_RET ACP_pload_ins_jta(const void* raw_data, unsigned size, ACInsJTAPackExt *pack);
BH_RET ACP_pload_ins_nta(const void* raw_data, unsigned size, ACInsNTAPackExt *pack);
BH_RET ACP_pload_uns_jta(const void* raw_data, unsigned size, ACUnsTAPackExt *pack);
BH_RET ACP_pload_uns_nta(const void* raw_data, unsigned size, ACUnsTAPackExt *pack);
BH_RET ACP_pload_ins_jta_prop(const void* raw_data, unsigned size, ACInsJTAPropExt *pack);
BH_RET ACP_get_cmd_id(const void* raw_data, unsigned size, int* cmd_id);
BH_RET ACP_pload_update_svl(const void *raw_data, unsigned size, ACUpdateSVLPackExt*pack);

#ifdef __cplusplus
}
#endif

#endif
