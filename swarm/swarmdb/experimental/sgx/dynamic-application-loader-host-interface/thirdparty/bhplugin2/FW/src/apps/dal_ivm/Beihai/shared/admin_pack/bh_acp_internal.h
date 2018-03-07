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

#ifndef _BH_ACP_INTERNAL_H
#define _BH_ACP_INTERNAL_H

#ifdef __cplusplus
extern "C" {
#endif

#include "bh_shared_types.h"
#include "bh_shared_errcode.h"
#include "bh_acp_format.h"

// Intel CSS Header + CSS Cypto Block which prefixes each signed ACP pkg
#define BH_ACP_CSS_HEADER_LENGTH    (128 + 520) // CSS Header + CSS Crypto Block

/*PackReader hold a reference of raw pack and read items with alignment support*/
typedef struct {
    const char *cur;
    const char *head;
    unsigned total;
} PackReader;

BH_RET pr_init(const char *data, unsigned n, PackReader* out);

/*whether pack reader reaches the end of buffer, alignment considered*/
BH_RET pr_is_end(PackReader *pr);

BH_RET ACP_load_pack_head(PackReader *pr, ACPackHeader** head);

BH_RET ACP_load_prop(PackReader *pr, ACProp** props);

BH_RET ACP_load_ins_sd(PackReader *pr, ACInsSDPack* pack);

BH_RET ACP_load_uns_sd(PackReader *pr, ACUnsSDPack* pack);

BH_RET ACP_load_ins_jta(PackReader *pr, ACInsJTAPack* pack);

BH_RET ACP_load_ins_nta(PackReader *pr, ACInsNTAPack* pack);

BH_RET ACP_load_uns_ta(PackReader *pr, ACUnsTAPack* pack);
BH_RET ACP_load_ta_pack(PackReader *pr, char** ta_pack);
BH_RET ACP_load_ins_jta_prop(PackReader *pr, ACInsJTAProp* pack);
BH_RET ACP_load_update_svl(PackReader* pr, ACUpdateSVLPack* pack);
#ifdef __cplusplus
}
#endif

#endif
