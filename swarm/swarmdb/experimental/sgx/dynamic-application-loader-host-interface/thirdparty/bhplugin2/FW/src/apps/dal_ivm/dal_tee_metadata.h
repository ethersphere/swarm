/*++
INTEL CONFIDENTIAL
Copyright 2010-2016 Intel Corporation All Rights Reserved.

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

#ifndef _DAL_TEE_METADATA_H_
#define _DAL_TEE_METADATA_H_

#include <stdint.h>

#define DAL_MAX_PLATFORM_TYPE_LEN       (8)
#define DAL_MAX_VM_TYPE_LEN             (16)
#define DAL_MAX_VM_VERSION_LEN          (12)
#define DAL_RESERVED_DWORDS             (16)

#define DAL_PRODUCTION_KEY_HASH_LEN     (32)


#pragma pack(1)

/**
 *  The DAL TEE Metadata definition which is provided to the host.
 */
typedef union _dal_access_control_groups
{
    uint64_t  groups;

    struct
    {
        uint64_t  internal        : 1;
        uint64_t  cryptography    : 1;
        uint64_t  utils           : 1;
        uint64_t  secure_time     : 1;
        uint64_t  debug           : 1;
        uint64_t  storage         : 1;
        uint64_t  key_exchange    : 1;
        uint64_t  trusted_output  : 1;
        uint64_t  SSL             : 1;
        uint64_t  sensors         : 1;
        uint64_t  NFC             : 1;
        uint64_t  IAC             : 1;
        uint64_t  platform        : 1;
        uint64_t  secure_enclave  : 1;
        uint64_t  AMT             : 1;

        uint64_t  reserved        : 49;
    };

} dal_access_control_groups;


typedef union _dal_feature_set_values
{
    uint32_t  values;

    struct
    {
        uint32_t  cryptography    : 1;
        uint32_t  utils           : 1;
        uint32_t  secure_time     : 1;
        uint32_t  debug           : 1;
        uint32_t  storage         : 1;
        uint32_t  key_exchange    : 1;
        uint32_t  trusted_output  : 1;
        uint32_t  SSL             : 1;
        uint32_t  sensors         : 1;
        uint32_t  NFC             : 1;
        uint32_t  IAC             : 1;
        uint32_t  platform        : 1;
        uint32_t  secure_enclave  : 1;
        uint32_t  AMT             : 1;
        uint32_t  VTEE            : 1;
        uint32_t  reserved        : 17;
    };

} dal_feature_set_values;


typedef struct _dal_fw_version
{
    uint16_t        major;
    uint16_t        minor;
    uint16_t        hotfix;
    uint16_t        build;

} dal_fw_version;


typedef struct _dal_tee_metadata
{
    uint32_t    api_level; // the API level of the DAL Java Class Library, unsigned integer
    uint32_t    library_version; // the version of the DAL Java Class Library for this platform, unsigned integer
    uint8_t     platform_type[DAL_MAX_PLATFORM_TYPE_LEN]; // the underlying security engine on the platform, char string
    uint8_t     dal_key_hash[DAL_PRODUCTION_KEY_HASH_LEN]; // SHA256 hash of the DAL Sign Once public key embedded in the firmware, byte array
    uint32_t    feature_set; // a bitmask of the features the platform support (SSL, NFC and etc),
                             // unsigned integer bitmask vlaues in dal_feature_set_values
    uint8_t     vm_type[DAL_MAX_VM_TYPE_LEN]; // the Beihai VM type in DAL, char string
    uint8_t     vm_version[DAL_MAX_VM_VERSION_LEN]; // the Beihai drop version integrated into the DAL, char string
    uint64_t    access_control_groups; // a bitmask of the access control groups defined in the Java Class Library on this platform,
                                       // unsigned integer bitmask values in dal_access_control_groups
    dal_fw_version fw_version; // the version of the firmware image on this platform
    uint32_t    reserved[DAL_RESERVED_DWORDS]; // reserved DWORDS for future use

} dal_tee_metadata;

#pragma pack()

#ifdef _WIN32
C_ASSERT((sizeof(dal_tee_metadata) % 4) == 0);
#endif

#endif      // _DAL_TEE_METADATA_H_

