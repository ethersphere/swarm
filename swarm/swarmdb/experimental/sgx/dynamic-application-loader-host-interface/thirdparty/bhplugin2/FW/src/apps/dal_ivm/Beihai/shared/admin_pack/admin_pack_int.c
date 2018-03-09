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
 * @file  admin_pack_int.cpp
 * @brief This file implements internal atomic api of admin command parsing
 *        The counter part which generate admin package is BPKT
 * @author Wenlong Feng(wenlong.feng@intel.com)
 */
 
#include <stddef.h>
#include <limits.h>
#include <stdint.h>

#include "bh_shared_errcode.h"
#include "bh_acp_format.h"
#include "bh_acp_internal.h"
#include "bh_acp_exp.h"

#define PR_ALIGN 4
#define PR_ALIGN_MASK (PR_ALIGN - 1)

BH_RET pr_init(const char *data, unsigned n, PackReader* out)
{
    //check integer overflow
    if ((size_t) data > UINTPTR_MAX - n)
        return BHE_INVALID_BPK_FILE;

    out->cur = out->head = data;
    out->total = n;
    return BH_SUCCESS;
}

static BH_RET pr_8b_align_move(PackReader* pr, size_t n_move)
{
    unsigned offset = 0;
    size_t new_cur = (size_t)pr->cur + n_move;
    size_t len_from_head = new_cur - (size_t)pr->head;

    if ((size_t)pr->cur > UINTPTR_MAX - n_move || new_cur < (size_t)pr->head)
        return BHE_INVALID_BPK_FILE;
    offset = ((8 - (len_from_head & 7)) & 7);
    if (new_cur > UINTPTR_MAX - offset) return BHE_INVALID_BPK_FILE;
    new_cur = new_cur + offset;
    if (new_cur > (size_t)pr->head + pr->total)
        return BHE_INVALID_BPK_FILE;
    pr->cur = (char*) new_cur;
    return BH_SUCCESS;
}

static BH_RET pr_align_move(PackReader* pr, size_t n_move)
{
    size_t new_cur = (size_t) pr->cur + n_move;
    size_t len_from_head = new_cur - (size_t)pr->head;
    size_t offset = 0;

    if ((size_t)pr->cur > UINTPTR_MAX - n_move || new_cur < (size_t)pr->head)
        return BHE_INVALID_BPK_FILE;
    offset = ((PR_ALIGN - (len_from_head & PR_ALIGN_MASK)) & PR_ALIGN_MASK);
    if (new_cur > UINTPTR_MAX - offset) return BHE_INVALID_BPK_FILE;
    new_cur = new_cur + offset;
    if (new_cur > (size_t)pr->head + pr->total)
        return BHE_INVALID_BPK_FILE;
    pr->cur = (char*) new_cur;
    return BH_SUCCESS;
}

static BH_RET pr_move(PackReader* pr, size_t n_move)
{
    size_t new_cur = (size_t) pr->cur + n_move;

    if ((size_t)pr->cur > UINTPTR_MAX - n_move ||
        new_cur > (size_t)pr->head + pr->total)  //integer overflow or out of acp pkg size
        return BHE_INVALID_BPK_FILE;
    pr->cur = (char*) new_cur;
    return BH_SUCCESS;
}

static int pr_is_safe_to_read(const PackReader* pr, size_t n_move)
{
    if ((size_t)pr->cur > UINTPTR_MAX - n_move) //integer overflow
        return BHE_INVALID_BPK_FILE;

    if ((size_t)pr->cur + n_move > (size_t)pr->head + pr->total)
        return BHE_INVALID_BPK_FILE;
    return BH_SUCCESS;
}

BH_RET pr_is_end(PackReader* pr)
{
    if ((size_t)pr->cur == (size_t)pr->head + pr->total)
        return BH_SUCCESS;
    else
        return BHE_INVALID_BPK_FILE;
}

static BH_RET ACP_load_ins_sd_head(PackReader* pr, ACInsSDHeader **head)
{
    if (BH_SUCCESS == pr_is_safe_to_read(pr, sizeof(ACInsSDHeader))) {
        *head = (ACInsSDHeader*)(pr->cur);
        return pr_align_move(pr, sizeof(ACInsSDHeader));
    }
    return BHE_INVALID_BPK_FILE;
}

static BH_RET ACP_load_ins_sd_sig(PackReader* pr, ACInsSDSigKey **sig)
{
    /*check buffer border before read the value to avoid access violation*/
    if (BH_SUCCESS == pr_is_safe_to_read(pr, sizeof(ACInsSDSigKey))) {
            *sig = (ACInsSDSigKey*)pr->cur;
            return pr_align_move(pr, sizeof(ACInsSDSigKey));
    }
    return BHE_INVALID_BPK_FILE;
}

static BH_RET ACP_load_groups(PackReader* pr, BH_U64 **groups)
{

    if (BH_SUCCESS == pr_is_safe_to_read(pr, sizeof(BH_U64))) {
        *groups = (BH_U64*)(pr->cur);
        return pr_align_move(pr, sizeof(BH_U64));
    }
    return BHE_INVALID_BPK_FILE;
}

static BH_RET ACP_load_hash(PackReader* pr, ACInsHash **hash)
{
    size_t len = 0;

    if (BH_SUCCESS == pr_is_safe_to_read(pr, sizeof(ACInsHash))) {
        *hash = (ACInsHash*)(pr->cur);
        if ((*hash)->len > BH_MAX_PACK_HASH_LEN)
            return BHE_INVALID_BPK_FILE;

        len = sizeof(ACInsHash) + (*hash)->len * sizeof((*hash)->data[0]);

        if (BH_SUCCESS == pr_is_safe_to_read(pr, len)) {
            return pr_align_move(pr, len);
        }
    }
    return BHE_INVALID_BPK_FILE;
}

static BH_RET ACP_load_sdid(PackReader* pr, BH_SDID** pp_sdid)
{
    if (BH_SUCCESS == pr_is_safe_to_read(pr, BH_SDID_LEN)) {
        *pp_sdid = (BH_SDID*) pr->cur;
        return pr_align_move(pr, BH_SDID_LEN);
    }
    return BHE_INVALID_BPK_FILE;
}

static BH_RET ACP_load_taid(PackReader* pr, BH_TAID** pp_taid)
{
    if (BH_SUCCESS == pr_is_safe_to_read(pr, BH_TAID_LEN)) {
        *pp_taid = (BH_TAID*) pr->cur;
        return pr_align_move(pr, BH_TAID_LEN);
    }
    return BHE_INVALID_BPK_FILE;
}

static BH_RET ACP_load_metadata(PackReader* pr, ACInsMetadata **metadata)
{
    size_t len = 0;

    if (BH_SUCCESS == pr_is_safe_to_read(pr, sizeof(ACInsMetadata))) {
        *metadata = (ACInsMetadata*)(pr->cur);
        if ((*metadata)->len > BH_MAX_ACP_NTA_METADATA_LENGTH)
            return BHE_INVALID_BPK_FILE;

        len = sizeof(ACInsMetadata) + (*metadata)->len * sizeof((*metadata)->data[0]);

        if (BH_SUCCESS == pr_is_safe_to_read(pr, len)) {
            return pr_align_move(pr, len);
        }
    }
    return BHE_INVALID_BPK_FILE;
}

static BH_RET ACP_load_reasons(PackReader* pr, ACInsReasons **reasons)
{
    size_t len = 0;

    if (BH_SUCCESS == pr_is_safe_to_read(pr, sizeof(ACInsReasons))) {
        *reasons = (ACInsReasons*)(pr->cur);
        if ((*reasons)->len > BH_MAX_ACP_INS_REASONS_LENGTH)
            return BHE_INVALID_BPK_FILE;
        len = sizeof(ACInsReasons) + (*reasons)->len * sizeof((*reasons)->data[0]);
        if (BH_SUCCESS == pr_is_safe_to_read(pr, len)) {
            return pr_align_move(pr, len);
        }
    }
    return BHE_INVALID_BPK_FILE;
}

BH_RET ACP_load_taid_list(PackReader* pr, ACTAIDList **taid_list)
{
    size_t len = 0;

    if (BH_SUCCESS == pr_is_safe_to_read(pr, sizeof(ACTAIDList))) {
        *taid_list = (ACTAIDList*)(pr->cur);
        if ((*taid_list)->num > BH_MAX_ACP_USED_SERVICES)
            return BHE_INVALID_BPK_FILE;

        len = sizeof(ACTAIDList) + (*taid_list)->num * sizeof((*taid_list)->list[0]);

        if (BH_SUCCESS == pr_is_safe_to_read(pr, len)) {
            return pr_align_move(pr, len);
        }
    }
    return BHE_INVALID_BPK_FILE;
}

BH_RET ACP_load_svl(PackReader* pr, ACSVList **svl)
{
    size_t len = 0;

    if (BH_SUCCESS == pr_is_safe_to_read(pr, sizeof(ACSVList))) {
        *svl = (ACSVList*)(pr->cur);
        if ((*svl)->num > BH_MAX_ACP_SVL_RECORDS)
            return BHE_INVALID_BPK_FILE;

        len = sizeof(ACSVList) + (*svl)->num * sizeof((*svl)->data[0]);

        if (BH_SUCCESS == pr_is_safe_to_read(pr, len)) {
            return pr_align_move(pr, len);
        }
    }
    return BHE_INVALID_BPK_FILE;
}

BH_RET ACP_load_prop(PackReader* pr, ACProp **prop)
{
    size_t len = 0;

    if (BH_SUCCESS == pr_is_safe_to_read(pr, sizeof(ACProp))) {
        *prop = (ACProp*)(pr->cur);
        if ((*prop)->len > BH_MAX_ACP_PORPS_LENGTH)
            return BHE_INVALID_BPK_FILE;

        len = sizeof(ACProp) + (*prop)->len * sizeof((*prop)->data[0]);

        if (BH_SUCCESS == pr_is_safe_to_read(pr, len)) {
            return pr_align_move(pr, len);
        }
    }
    return BHE_INVALID_BPK_FILE;
}

BH_RET ACP_load_ta_pack(PackReader* pr, char **ta_pack)
{
    size_t len = 0;

    /*8 byte align to obey jeff rule*/
    if (BH_SUCCESS == pr_8b_align_move(pr, 0)) {
        *ta_pack = (char*)(pr->cur);

        /*assume ta pack is the last item of one package,
           move cursor to the end directly*/
        if ((size_t)pr->cur > (size_t)pr->head + pr->total)
            return BHE_INVALID_BPK_FILE;
        len = (size_t)pr->head + pr->total - (size_t)pr->cur;
        if (BH_SUCCESS == pr_is_safe_to_read(pr, len)) {
            return pr_move(pr, len);
        }
    }
    return BHE_INVALID_BPK_FILE;
}

BH_RET ACP_load_ins_sd(PackReader* pr, ACInsSDPack* pack)
{
    BH_RET ret = BHE_INVALID_BPK_FILE;
    if((BH_SUCCESS != (ret = ACP_load_prop(pr, &(pack->ins_cond))))
       || (BH_SUCCESS != (ret = ACP_load_ins_sd_head(pr, &(pack->head))))
       || (BH_SUCCESS != (ret = ACP_load_ins_sd_sig(pr, &(pack->sig_key)))))
        return ret;
    return BH_SUCCESS;
}

BH_RET ACP_load_uns_sd(PackReader* pr, ACUnsSDPack* pack)
{
    return ACP_load_sdid(pr, &(pack->p_sdid));
}

static BH_RET ACP_load_ins_jta_prop_head(PackReader* pr, ACInsJTAPropHeader **head)
{
    if (BH_SUCCESS == pr_is_safe_to_read(pr, sizeof(ACInsJTAPropHeader))) {
        *head = (ACInsJTAPropHeader*)(pr->cur);
        return pr_align_move(pr, sizeof(ACInsJTAPropHeader));
    }
    return BHE_INVALID_BPK_FILE;
}

BH_RET ACP_load_ins_jta_prop(PackReader* pr, ACInsJTAProp* pack)
{
    BH_RET ret = BHE_INVALID_BPK_FILE;
    if((BH_SUCCESS != (ret = ACP_load_ins_jta_prop_head(pr, &(pack->head))))
        || (BH_SUCCESS != (ret = ACP_load_reasons(pr, &(pack->post_reasons))))
        || (BH_SUCCESS != (ret = ACP_load_reasons(pr, &(pack->reg_reasons))))
        || (BH_SUCCESS != (ret = ACP_load_prop(pr, &(pack->prop)))
        || (BH_SUCCESS != (ret = ACP_load_taid_list(pr, &(pack->used_service_list)))))
        )
        return ret;
    return BH_SUCCESS;
}

static BH_RET ACP_load_ins_jta_head(PackReader* pr, ACInsJTAHeader **head)
{
    if (BH_SUCCESS == pr_is_safe_to_read(pr, sizeof(ACInsJTAHeader))) {
        *head = (ACInsJTAHeader*)(pr->cur);
        return pr_align_move(pr, sizeof(ACInsJTAHeader));
    }
    return BHE_INVALID_BPK_FILE;
}

BH_RET ACP_load_ins_jta(PackReader* pr, ACInsJTAPack* pack)
{
    BH_RET ret = BHE_INVALID_BPK_FILE;
    if((BH_SUCCESS != (ret = ACP_load_prop(pr, &(pack->ins_cond))))
       || (BH_SUCCESS != (ret = ACP_load_ins_jta_head(pr, &(pack->head)))))
        return ret;
    return BH_SUCCESS;
}

static BH_RET ACP_load_ins_nta_head(PackReader* pr, ACInsNTAHeader **head)
{
    if (BH_SUCCESS == pr_is_safe_to_read(pr, sizeof(ACInsNTAHeader))) {
        *head = (ACInsNTAHeader*)(pr->cur);
        return pr_align_move(pr, sizeof(ACInsNTAHeader));
    }
    return BHE_INVALID_BPK_FILE;
}

BH_RET ACP_load_ins_nta(PackReader* pr, ACInsNTAPack* pack)
{
    BH_RET ret = BHE_INVALID_BPK_FILE;
    if((BH_SUCCESS != (ret = ACP_load_prop(pr, &(pack->ins_cond))))
       || (BH_SUCCESS != (ret = ACP_load_ins_nta_head(pr, &(pack->head))))
       || (BH_SUCCESS != (ret = ACP_load_metadata(pr, &(pack->mdata)))))
        return ret;
    return BH_SUCCESS;
}

BH_RET ACP_load_uns_ta(PackReader* pr, ACUnsTAPack* pack)
{
    return ACP_load_taid(pr, &(pack->p_taid));
}

BH_RET ACP_load_update_svl(PackReader* pr, ACUpdateSVLPack* pack)
{
    BH_RET ret = BHE_INVALID_BPK_FILE;
    if((BH_SUCCESS != (ret = ACP_load_prop(pr, &(pack->ins_cond))))
        || (BH_SUCCESS != (ret = ACP_load_svl(pr, &(pack->sv_list)))))
        return ret;
    return BH_SUCCESS;
}

BH_RET ACP_load_pack_head(PackReader* pr, ACPackHeader **head)
{
    if (BH_SUCCESS == pr_is_safe_to_read(pr, sizeof(ACPackHeader))) {
        *head = (ACPackHeader*)(pr->cur);
        return pr_align_move(pr, sizeof(ACPackHeader));
    }
    return BHE_INVALID_BPK_FILE;
}

/*
  static BH_RET ACP_load_pack_enc(PackReader* pr, ACEncryption **enc)
  {
  *enc = (ACEncryption*)(pr->cur);
  pr_align_move(pr, sizeof(ACEncryption) + (*enc)->len);
  return BH_SUCCESS;
  }
*/
#ifdef BPKT_UNIT_TEST
/*for debugging purpose*/
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

BH_RET ACP_load_pack(char *raw_pack,  unsigned size, int cmd_id, ACPack *pack);
void dump_pack(ACPack *pack);

void load_and_dump(const char *fname)
{
    FILE *fp = fopen(fname, "rb");
    unsigned n_read;
    long sz;
    char *buf;
    ACInsSDPackExt p1;
    ACUnsSDPackExt p2;
    ACInsJTAPackExt p3;
    ACUnsTAPackExt p4;
    ACInsNTAPackExt p5;
    ACUnsTAPackExt p6;
    ACInsJTAPropExt p7;
    ACUpdateSVLPackExt p8;
    BH_RET ret;
    int cmd_id = fname[0] - '0';

    fseek(fp, 0L, SEEK_END);
    sz = ftell(fp);
    fseek(fp, 0L, SEEK_SET);
    buf = malloc(sz);
    n_read = fread(buf, 1, sz, fp);
    switch(cmd_id) {
    case AC_INSTALL_SD: {
        ret = ACP_pload_ins_sd(buf, n_read, &p1);
        dump_pack((ACPack*)&p1);
        break;
    }
    case AC_UNINSTALL_SD: {
        ret = ACP_pload_uns_sd(buf, n_read, &p2);
        dump_pack((ACPack*)&p2);
        break;
    }
    case AC_INSTALL_JTA: {
        ret = ACP_pload_ins_jta(buf, n_read, &p3);
        dump_pack((ACPack*)&p3);
        break;
    }
    case AC_UNINSTALL_JTA: {
        ret = ACP_pload_uns_jta(buf, n_read, &p4);
        dump_pack((ACPack*)&p4);
        break;
    }
    case AC_INSTALL_NTA: {
        ret = ACP_pload_ins_nta(buf, n_read, &p5);
        dump_pack((ACPack*)&p5);
        break;
    }
    case AC_UNINSTALL_NTA: {
        ret = ACP_pload_uns_nta(buf, n_read, &p6);
        dump_pack((ACPack*)&p6);
        break;
    }
    case AC_INSTALL_JTA_PROP: {
        ret = ACP_pload_ins_jta_prop(buf, n_read, &p7);
        printf("---------------------------\n");
        dump_ins_jta_prop(&(p7.cmd_pack));
        dump_binary("jeff_binary", 4, p7.jeff_pack);
        break;
    }
    case AC_UPDATE_SVL: {
        ret = ACP_pload_update_svl(buf, n_read, &p8);
        printf("---------------------------\n");
        dump_pack((ACPack*) &p8);
        break;
    }
    default: {
        printf("illegal cmd id %d\n", cmd_id);
        ret = BHE_BAD_PARAMETER;
        break;
    }
    }
    fclose(fp);
    if (ret != BH_SUCCESS)
        abort();
}
#ifdef BH_TEST
int main (int argc, const char *argv[])
{
    char ch;

    system("del 1.out 2.out 3.out 4.out 5.out 6.out 7.out 8.out");
    system("bpkt_exe.exe 1 TEMPLATE_AC_INSTALL_SD.xml 1.out 00000000-0000-0000-0000-000000000001");
    system("bpkt_exe.exe 2 TEMPLATE_AC_UNINSTALL_SD.xml 2.out 00000000-0000-0000-0000-000000000002");
    system("bpkt_exe.exe 4 TEMPLATE_AC_UNINSTALL_JTA.xml 4.out 00000000-0000-0000-0000-000000000003");
    system("bpkt_exe.exe 5 TEMPLATE_AC_INSTALL_NTA.xml  5.out 00000000-0000-0000-0000-000000000003 a.out a.met");
    system("bpkt_exe.exe 6 TEMPLATE_AC_UNINSTALL_NTA.xml 6.out 00000000-0000-0000-0000-000000000003");
    system("bpkt_exe.exe 8 TEMPLATE_AC_INSTALL_JTA_PROP.xml 8.out a.jeff");
    system("bpkt_exe.exe 3 TEMPLATE_AC_INSTALL_JTA.xml 3.out 00000000-0000-0000-0000-000000000004 8.out");
    system("bpkt_exe.exe 7 TEMPLATE_AC_UPDATE_SVL.xml 7.out 00000000-0000-0000-0000-000000000003");

    load_and_dump("1.out");
    load_and_dump("2.out");
    load_and_dump("3.out");
    load_and_dump("4.out");
    load_and_dump("5.out");
    load_and_dump("6.out");
    load_and_dump("7.out");
    load_and_dump("8.out");
    printf("-------------\nSucc\n");

    scanf("%c", &ch);
    return 0;
}
#else
int main (int argc, const char *argv[])
{
    char ch;
    system("del 1.out 2.out 3.out 4.out 5.out 6.out 7.out 8.out");
    system("bpkt_exe.exe 1 TEMPLATE_AC_INSTALL_SD.xml 1.out");
    system("bpkt_exe.exe 2 TEMPLATE_AC_UNINSTALL_SD.xml 2.out");
    system("bpkt_exe.exe 4 TEMPLATE_AC_UNINSTALL_JTA.xml 4.out");
    system("bpkt_exe.exe 5 TEMPLATE_AC_INSTALL_NTA.xml  5.out a.met");
    system("bpkt_exe.exe 6 TEMPLATE_AC_UNINSTALL_NTA.xml 6.out");
    system("bpkt_exe.exe 8 TEMPLATE_AC_INSTALL_JTA_PROP.xml 8.out");
    system("bpkt_exe.exe 7 TEMPLATE_AC_UPDATE_SVL.xml 7.out");
    system("bpkt_exe.exe 3 TEMPLATE_AC_INSTALL_JTA.xml 3.out");

    load_and_dump("1.out");
    load_and_dump("2.out");
    load_and_dump("3.out");
    load_and_dump("4.out");
    load_and_dump("5.out");
    load_and_dump("6.out");
    load_and_dump("7.out");
    load_and_dump("8.out");
    printf("-------------\nSucc\n");

    scanf("%c", &ch);
    return 0;
}
#endif

#endif
