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

#include "bh_acp_exp.h"
#include "bh_acp_internal.h"

#ifndef NULL
#define NULL (void*)0
#endif


#define BREAKIF(ret)  {if (ret != BH_SUCCESS) break;}

static BH_RET ACP_load_pack(const char *raw_pack,  unsigned size, int cmd_id, ACPack *pack)
{
    BH_RET ret = BHE_FAILED;
    PackReader pr = {0};

    if (BH_SUCCESS != pr_init(raw_pack, size, &pr))
        return BHE_INVALID_BPK_FILE;

    if ((cmd_id != AC_INSTALL_JTA_PROP) &&
        BH_SUCCESS != (ret = ACP_load_pack_head(&pr, &(pack->head))))
        return ret;

    if ((cmd_id != AC_INSTALL_JTA_PROP) && (cmd_id != pack->head->cmd_id))
        return BHE_BAD_PARAMETER;

    switch(cmd_id) {
    case AC_INSTALL_SD:
        ret = ACP_load_ins_sd(&pr, &(((ACInsSDPackExt*)pack)->cmd_pack));
        BREAKIF(ret);
        break;
    case AC_UNINSTALL_SD:
        ret = ACP_load_uns_sd(&pr, &(((ACUnsSDPackExt*)pack)->cmd_pack));
        BREAKIF(ret);
        break;
    case AC_INSTALL_JTA:
        ret = ACP_load_ins_jta(&pr, &(((ACInsJTAPackExt*)pack)->cmd_pack));
        BREAKIF(ret);
        ret = ACP_load_ta_pack(&pr, &(((ACInsJTAPackExt*)pack)->ta_pack));
        break;
    case AC_INSTALL_NTA:
        ret = ACP_load_ins_nta(&pr, &(((ACInsNTAPackExt*)pack)->cmd_pack));
        BREAKIF(ret);
        ret = ACP_load_ta_pack(&pr, &(((ACInsNTAPackExt*)pack)->ta_pack));
        break;
    case AC_UNINSTALL_JTA:
    case AC_UNINSTALL_NTA:
        ret = ACP_load_uns_ta(&pr, &(((ACUnsTAPackExt*)pack)->cmd_pack));
        BREAKIF(ret);
        break;
    case AC_INSTALL_JTA_PROP:
        ret = ACP_load_ins_jta_prop(&pr, &(((ACInsJTAPropExt*)pack)->cmd_pack));
        BREAKIF(ret);
        //Note: the next section is JEFF file, and not ta_pack(JTA_properties+JEFF file),
        //  but we could reuse the ACP_load_ta_pack() here.
        ret = ACP_load_ta_pack(&pr, &(((ACInsJTAPropExt*)pack)->jeff_pack));
        break;
    case AC_UPDATE_SVL:
        ret = ACP_load_update_svl(&pr, &(((ACUpdateSVLPackExt*)pack)->cmd_pack));
        BREAKIF(ret);
        break;
    default:
        return BHE_BAD_PARAMETER;
    }
    if (BH_SUCCESS != pr_is_end(&pr))
        return BHE_INVALID_BPK_FILE;
    return ret;
}

BH_RET ACP_pload_ins_sd(const void *raw_data, unsigned size, ACInsSDPackExt *pack)
{
    if (NULL == raw_data || size <= BH_ACP_CSS_HEADER_LENGTH || NULL == pack)
        return BHE_BAD_PARAMETER;
    return ACP_load_pack((char*)raw_data + BH_ACP_CSS_HEADER_LENGTH,
                         size - BH_ACP_CSS_HEADER_LENGTH, AC_INSTALL_SD, (ACPack*)pack);
}

BH_RET ACP_pload_uns_sd(const void *raw_data, unsigned size, ACUnsSDPackExt *pack)
{
    if (NULL == raw_data || size <= BH_ACP_CSS_HEADER_LENGTH || NULL == pack)
        return BHE_BAD_PARAMETER;
    return ACP_load_pack((char*)raw_data + BH_ACP_CSS_HEADER_LENGTH,
                         size - BH_ACP_CSS_HEADER_LENGTH, AC_UNINSTALL_SD, (ACPack*)pack);
}

BH_RET ACP_pload_ins_jta(const void *raw_data, unsigned size, ACInsJTAPackExt *pack)
{
    if (NULL == raw_data || size <= BH_ACP_CSS_HEADER_LENGTH || NULL == pack)
        return BHE_BAD_PARAMETER;
    return ACP_load_pack((char*)raw_data + BH_ACP_CSS_HEADER_LENGTH,
                         size - BH_ACP_CSS_HEADER_LENGTH, AC_INSTALL_JTA, (ACPack*)pack);
}

BH_RET ACP_pload_ins_nta(const void *raw_data, unsigned size, ACInsNTAPackExt *pack)
{
    if (NULL == raw_data || size <= BH_ACP_CSS_HEADER_LENGTH || NULL == pack)
        return BHE_BAD_PARAMETER;
    return ACP_load_pack((char*)raw_data + BH_ACP_CSS_HEADER_LENGTH,
                         size - BH_ACP_CSS_HEADER_LENGTH, AC_INSTALL_NTA, (ACPack*)pack);
}

BH_RET ACP_pload_uns_jta(const void *raw_data, unsigned size, ACUnsTAPackExt *pack)
{
    if (NULL == raw_data || size <= BH_ACP_CSS_HEADER_LENGTH || NULL == pack)
        return BHE_BAD_PARAMETER;
    return ACP_load_pack((char*)raw_data + BH_ACP_CSS_HEADER_LENGTH,
                         size - BH_ACP_CSS_HEADER_LENGTH, AC_UNINSTALL_JTA, (ACPack*)pack);
}

BH_RET ACP_pload_uns_nta(const void *raw_data, unsigned size, ACUnsTAPackExt *pack)
{
    if (NULL == raw_data || size <= BH_ACP_CSS_HEADER_LENGTH || NULL == pack)
        return BHE_BAD_PARAMETER;
    return ACP_load_pack((char*)raw_data + BH_ACP_CSS_HEADER_LENGTH,
                         size - BH_ACP_CSS_HEADER_LENGTH, AC_UNINSTALL_NTA, (ACPack*)pack);
}

BH_RET ACP_pload_ins_jta_prop(const void *raw_data, unsigned size, ACInsJTAPropExt *pack)
{
    if (NULL == raw_data || NULL == pack)
        return BHE_BAD_PARAMETER;
    return ACP_load_pack((char*)raw_data, size, AC_INSTALL_JTA_PROP, (ACPack*)pack);
}

BH_RET ACP_pload_update_svl(const void *raw_data, unsigned size, ACUpdateSVLPackExt *pack)
{
    if (NULL == raw_data || size <= BH_ACP_CSS_HEADER_LENGTH || NULL == pack)
        return BHE_BAD_PARAMETER;
    return ACP_load_pack((char*)raw_data + BH_ACP_CSS_HEADER_LENGTH,
                         size - BH_ACP_CSS_HEADER_LENGTH, AC_UPDATE_SVL, (ACPack*)pack);
}

BH_RET ACP_get_cmd_id(const void *raw_data, unsigned size, int* cmd_id)
{
    BH_RET ret = BH_SUCCESS;
    PackReader pr = {0};
    ACPackHeader *ph = 0;

    if (NULL == raw_data || size <= BH_ACP_CSS_HEADER_LENGTH || NULL == cmd_id)
        return BHE_BAD_PARAMETER;

    *cmd_id = AC_CMD_INVALID;
    if (BH_SUCCESS != pr_init((char*)raw_data+BH_ACP_CSS_HEADER_LENGTH, size-BH_ACP_CSS_HEADER_LENGTH, &pr))
        return BHE_INVALID_BPK_FILE;
    if (BH_SUCCESS != (ret = ACP_load_pack_head(&pr, &ph)))
        return ret;
    *cmd_id = (*ph).cmd_id;
    return BH_SUCCESS;
}
