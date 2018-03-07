#ifndef PSE_OP_U_H__
#define PSE_OP_U_H__

#include <stdint.h>
#include <wchar.h>
#include <stddef.h>
#include <string.h>
#include "sgx_edger8r.h" /* for sgx_satus_t etc. */

#include "monotonic_counter_database_types.h"
#include "pse_types.h"
#include "sgx_dh.h"
#include "t_pairing_blob.h"

#include <stdlib.h> /* for size_t */

#define SGX_CAST(type, item) ((type)(item))

#ifdef __cplusplus
extern "C" {
#endif

pse_op_error_t SGX_UBRIDGE(SGX_NOCONVENTION, sqlite_db_init_hash_tree_table, ());
pse_op_error_t SGX_UBRIDGE(SGX_NOCONVENTION, sqlite_read_db, (uint32_t leaf_id, pse_vmc_hash_tree_cache_t* cache));
pse_op_error_t SGX_UBRIDGE(SGX_NOCONVENTION, sqlite_write_db, (pse_vmc_hash_tree_cache_t* cache, uint8_t is_for_update_flag, op_leafnode_flag_t* op_flag_info));
pse_op_error_t SGX_UBRIDGE(SGX_NOCONVENTION, sqlite_read_children_of_root, (pse_vmc_children_of_root_t* children));
pse_op_error_t SGX_UBRIDGE(SGX_NOCONVENTION, sqlite_get_empty_leafnode, (int* leaf_node_id, sgx_measurement_t* mr_signer));
ae_error_t SGX_UBRIDGE(SGX_NOCONVENTION, psda_invoke_service_ocall, (uint8_t* psda_req_msg, uint32_t psda_req_msg_size, uint8_t* psda_resp_msg, uint32_t psda_resp_msg_size));
pse_op_error_t SGX_UBRIDGE(SGX_NOCONVENTION, sqlite_rollback_db_file, ());

sgx_status_t create_session_wrapper(sgx_enclave_id_t eid, ae_error_t* retval, uint64_t tick, uint32_t* id, pse_dh_msg1_t* dh_msg1);
sgx_status_t exchange_report_wrapper(sgx_enclave_id_t eid, ae_error_t* retval, uint64_t tick, uint32_t sid, sgx_dh_msg2_t* dh_msg2, pse_dh_msg3_t* dh_msg3);
sgx_status_t close_session_wrapper(sgx_enclave_id_t eid, ae_error_t* retval, uint32_t sid);
sgx_status_t invoke_service_wrapper(sgx_enclave_id_t eid, ae_error_t* retval, uint64_t tick, uint8_t* req_msg, uint32_t req_msg_size, uint8_t* resp_msg, uint32_t resp_msg_size);
sgx_status_t initialize_sqlite_database_file_wrapper(sgx_enclave_id_t eid, ae_error_t* retval, bool is_for_empty_db_creation);
sgx_status_t ephemeral_session_m2m3_wrapper(sgx_enclave_id_t eid, ae_error_t* retval, pairing_blob_t* sealed_blob, pse_cse_msg2_t* pse_cse_msg2, pse_cse_msg3_t* pse_cse_msg3);
sgx_status_t ephemeral_session_m4_wrapper(sgx_enclave_id_t eid, ae_error_t* retval, pse_cse_msg4_t* pse_cse_msg4);

#ifdef __cplusplus
}
#endif /* __cplusplus */

#endif
