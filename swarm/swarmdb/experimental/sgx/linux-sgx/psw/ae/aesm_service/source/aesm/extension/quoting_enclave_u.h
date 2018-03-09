#ifndef QUOTING_ENCLAVE_U_H__
#define QUOTING_ENCLAVE_U_H__

#include <stdint.h>
#include <wchar.h>
#include <stddef.h>
#include <string.h>
#include "sgx_edger8r.h" /* for sgx_satus_t etc. */

#include "sgx_report.h"
#include "sgx_quote.h"

#include <stdlib.h> /* for size_t */

#define SGX_CAST(type, item) ((type)(item))

#ifdef __cplusplus
extern "C" {
#endif


sgx_status_t verify_blob(sgx_enclave_id_t eid, uint32_t* retval, uint8_t* p_blob, uint32_t blob_size, uint8_t* p_is_resealed);
sgx_status_t get_quote(sgx_enclave_id_t eid, uint32_t* retval, uint8_t* p_blob, uint32_t blob_size, const sgx_report_t* p_report, sgx_quote_sign_type_t quote_type, const sgx_spid_t* p_spid, const sgx_quote_nonce_t* p_nonce, const uint8_t* p_sig_rl, uint32_t sig_rl_size, sgx_report_t* qe_report, uint8_t* p_quote, uint32_t quote_size, sgx_isv_svn_t pce_isvnsvn);

#ifdef __cplusplus
}
#endif /* __cplusplus */

#endif
