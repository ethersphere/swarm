#ifndef PCE_U_H__
#define PCE_U_H__

#include <stdint.h>
#include <wchar.h>
#include <stddef.h>
#include <string.h>
#include "sgx_edger8r.h" /* for sgx_satus_t etc. */

#include "pce_cert.h"
#include "sgx_report.h"

#include <stdlib.h> /* for size_t */

#define SGX_CAST(type, item) ((type)(item))

#ifdef __cplusplus
extern "C" {
#endif


sgx_status_t get_pc_info(sgx_enclave_id_t eid, uint32_t* retval, const sgx_report_t* report, const uint8_t* public_key, uint32_t key_size, uint8_t crypto_suite, uint8_t* encrypted_ppid, uint32_t encrypted_ppid_buf_size, uint32_t* encrypted_ppid_out_size, pce_info_t* pce_info, uint8_t* signature_scheme);
sgx_status_t certify_enclave(sgx_enclave_id_t eid, uint32_t* retval, const psvn_t* cert_psvn, const sgx_report_t* report, uint8_t* signature, uint32_t signature_buf_size, uint32_t* signature_out_size);

#ifdef __cplusplus
}
#endif /* __cplusplus */

#endif
