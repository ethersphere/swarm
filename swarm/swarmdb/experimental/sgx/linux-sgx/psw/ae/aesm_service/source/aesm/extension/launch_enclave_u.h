#ifndef LAUNCH_ENCLAVE_U_H__
#define LAUNCH_ENCLAVE_U_H__

#include <stdint.h>
#include <wchar.h>
#include <stddef.h>
#include <string.h>
#include "sgx_edger8r.h" /* for sgx_satus_t etc. */

#include "arch.h"
#include "sgx_report.h"

#include <stdlib.h> /* for size_t */

#define SGX_CAST(type, item) ((type)(item))

#ifdef __cplusplus
extern "C" {
#endif


sgx_status_t le_get_launch_token_wrapper(sgx_enclave_id_t eid, int* retval, const sgx_measurement_t* mrenclave, const sgx_measurement_t* mrsigner, const sgx_attributes_t* se_attributes, token_t* lictoken);
sgx_status_t le_init_white_list_wrapper(sgx_enclave_id_t eid, uint32_t* retval, const uint8_t* wl_cert_chain, uint32_t wl_cert_chain_size);

#ifdef __cplusplus
}
#endif /* __cplusplus */

#endif
