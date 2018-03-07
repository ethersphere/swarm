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

#include <stdio.h>
#include <string.h>

#include "bh_acp_format.h"
#include "bh_acp_exp.h"

static void dump_sdid(const char *sdid)
{
    int i;
    printf("sdid:  ");
    for(i = 0; i < BH_GUID_LENGTH; i++) {
        printf("%4d", (unsigned char)sdid[i]);
    }
    printf("\n");
}

static void dump_taid(BH_TAID taid)
{
    int i;
    printf("taid:  ");
    for(i = 0; i < BH_GUID_LENGTH; i++) {
        printf("%4d", (unsigned char)taid.data[i]);
    }
    printf("\n");
}

static void dump_pack_head(const ACPackHeader* head)
{
    printf("magic %c %c %c %c\n", head->magic[0], head->magic[1], head->magic[2], head->magic[3]);
    printf("version %u\n", (unsigned) head->version);
    printf("little_endian %u\n", (unsigned) head->little_endian);
    printf("reserved %u\n", head->reserved);
    printf("size %u\n", head->size);
    printf("cmd_id %u\n", head->cmd_id);
    printf("svn %u\n", head->svn);
    printf("idx_num %u\n", head->idx_num);
    printf("idx_condition %u\n", head->idx_condition);
    printf("idx_data %u\n", head->idx_data);
}

static void dump_ta_pack(char* ta_pack)
{
#ifdef BH_TEST
    unsigned i;
    printf("Just dumping first 10 char of raw ta_pack");
    for (i = 0; i < 10; i++) {
        printf ("%4d", ta_pack[i]);
    }
    printf("\n");
#endif
}

static void dump_metadata(const ACInsMetadata* meta)
{
    unsigned i;
    printf("Metadata len %u", meta->len);
    printf("data ");
    for (i = 0; i < meta->len; ++i) {
        printf("%4u", meta->data[i]);
    }
    printf("\n");
}

static void dump_reasons(const char* tag, const ACInsReasons* reasons)
{
    unsigned i;
    printf("%s\n", tag);
    printf("Reasons len %u", reasons->len);
    printf("data ");
    for (i = 0; i < reasons->len; ++i) {
        printf("%4u", reasons->data[i]);
    }
    printf("\n");
}

static void dump_taid_list(const ACTAIDList *taid_list)
{
    unsigned i, j;
    printf("taid_list num: %d\n", taid_list->num);
    for (i=0; i<taid_list->num; i++) {
        for (j = 0; j < sizeof(BH_TAID); j++) {
            printf("%4d", taid_list->list[i].data[j]);
        }
        printf("\n");
    }
}

static void dump_svl(const ACSVList *sv_list)
{
    unsigned i, j;
    printf("sv list num: %d\n", sv_list->num);
    for (i=0; i<sv_list->num; i++) {
        printf("svn:%d ", sv_list[i].data[i].ta_svn);
        printf("taid:");
        for (j = 0; j < sizeof(BH_TAID); j++) {
            printf("%4d", sv_list->data[i].ta_id.data[j]);
        }
        printf("\n");
    }
}

static void dump_prop(const ACProp *prop)
{
    char *data = (char*) prop->data;
    unsigned i;
    printf ("prop num: %d, len: %d\n", prop->num, prop->len);
    for (i=0; i<prop->num; i++) {
        printf ("|type<%s>|", data);
        data += strlen(data) + 1;
        printf ("key<%s>", data);
        data += strlen(data) + 1;
        printf ("|value<%s>|\n", data);
        data += strlen(data) + 1;
    }
}

static void dump_ins_sd_head(const ACInsSDHeader *head)
{
    printf("Ins sd head:\n");
    printf("sd_id");
    dump_sdid((char*)(&head->sd_id));
    printf("sd_svn %u\n", head->sd_svn);
    printf("ssd_num %u\n", head->ssd_num);
    printf("ta_type %u\n", head->ta_type);
    printf("reserved %u\n", head->reserved);
    printf("max_ta_installed %u\n", head->max_ta_can_install);
    printf("max_ta_running %u\n", head->max_ta_can_run);
    printf("flash quota %u\n", head->flash_quota);
    printf("groups %016llX\n", head->ac_groups);
    printf("sd_name %s\n", head->sd_name);
}

static void dump_ins_sd_sig(const ACInsSDSigKey *sig)
{
    int i;
    int data_size = AC_SIG_KEY_LEN;
    printf("InsSDSig sig_alg %4u\n", sig->sig_alg);
    printf("InsSDSig sig_key_type %4u\n", sig->sig_key_type);

    printf("sig_key (sig_manifest)");
    for(i = 0; i < data_size; i++) {
        printf("%4d", sig->sig_key[i]);
    }
    printf("\n");
}

static void dump_ins_sd(const ACInsSDPack *pack)
{
    printf("INS_SD\n");
    dump_prop(pack->ins_cond);
    dump_ins_sd_head(pack->head);
    dump_ins_sd_sig(pack->sig_key);
}

static void dump_uns_sd(const ACUnsSDPack *pack)
{
    printf("UNS_SD\n");
    dump_sdid((char*)(pack->p_sdid));
}

static void dump_ins_nta_head(const ACInsNTAHeader *head)
{
    dump_taid(head->ta_id);
}

void dump_binary(const char* name, unsigned len, const BH_U8* buf)
{
    int i;
    printf("%s:\n", name);
    for (i = 0; i < len; i++)
        printf(" %x ", buf[i]);
    printf("\n");
}

static void dump_ins_jta_head(const ACInsJTAHeader *head)
{
    printf("INS_JTA\n");
    dump_taid(head->ta_id);
    printf("ta_svn:%d\n", head->ta_svn);
    printf("hash_alg_type %u\n", head->hash_alg_type);
    dump_binary("hash", 32, head->hash.data);
}

static void dump_ins_jta(const ACInsJTAPack *pack)
{
    dump_prop(pack->ins_cond);
    dump_ins_jta_head(pack->head);
}

static void dump_ins_jta_prop_head(const ACInsJTAPropHeader *head)
{
    printf("mem_quota:%d\n",head->mem_quota);
    printf("ta_encrypted:%d\n",head->ta_encrypted);
    printf("groups:%ld\n",head->ac_groups);
    printf("timeout:%d\n",head->timeout);
    printf("allowed_inter_session_num:%d\n",head->allowed_inter_session_num);
}

void dump_ins_jta_prop(const ACInsJTAProp *pack)
{
    printf("INS_JTA_PROP\n");
    dump_ins_jta_prop_head(pack->head);
    dump_reasons("post_reasons", pack->post_reasons);
    dump_reasons("reg_reasons", pack->reg_reasons);
    dump_prop(pack->prop);
    dump_taid_list(pack->used_service_list);
}

static void dump_ins_nta(const ACInsNTAPack *pack)
{
    printf("INS_NTA\n");
    dump_prop(pack->ins_cond);
    dump_ins_nta_head(pack->head);
    dump_metadata(pack->mdata);
}

static void dump_uns_ta(const ACUnsTAPack *pack)
{
    printf("UNS_TA\n");
    dump_taid(*pack->p_taid);
}

static void dump_upt_svl(const ACUpdateSVLPack* pack)
{
    printf("UpdateSVL\n");
    dump_prop(pack->ins_cond);
    dump_svl(pack->sv_list);
}

void dump_pack(ACPack *pack)
{
    printf("---------------------------\n");
    dump_pack_head(pack->head);
    switch(pack->head->cmd_id) {
    case AC_INSTALL_SD:
        dump_ins_sd((ACInsSDPack*)(&(pack->data)));
        break;
    case AC_UNINSTALL_SD:
        dump_uns_sd((ACUnsSDPack*)(&(pack->data)));
        break;
    case AC_INSTALL_JTA:
        dump_ins_jta((ACInsJTAPack*)(&(pack->data)));
        dump_binary("jta_pack_binary", 4, (((ACInsJTAPackExt*)pack)->ta_pack));
        break;
    case AC_INSTALL_NTA:
        dump_ins_nta((ACInsNTAPack*)(&(pack->data)));
        dump_binary("nta_binary", 4, (((ACInsNTAPackExt*)pack)->ta_pack));
        break;
    case AC_UNINSTALL_NTA:
    case AC_UNINSTALL_JTA:
        dump_uns_ta((ACUnsTAPack*)(&(pack->data)));
        break;
    case AC_UPDATE_SVL:
        dump_upt_svl((ACUpdateSVLPack*)(&(pack->data)));
    default:
        break;
    }
}
