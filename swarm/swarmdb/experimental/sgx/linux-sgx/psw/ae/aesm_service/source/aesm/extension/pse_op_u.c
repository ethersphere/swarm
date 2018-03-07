#include "pse_op_u.h"
#include <errno.h>

typedef struct ms_create_session_wrapper_t {
	ae_error_t ms_retval;
	uint64_t ms_tick;
	uint32_t* ms_id;
	pse_dh_msg1_t* ms_dh_msg1;
} ms_create_session_wrapper_t;

typedef struct ms_exchange_report_wrapper_t {
	ae_error_t ms_retval;
	uint64_t ms_tick;
	uint32_t ms_sid;
	sgx_dh_msg2_t* ms_dh_msg2;
	pse_dh_msg3_t* ms_dh_msg3;
} ms_exchange_report_wrapper_t;

typedef struct ms_close_session_wrapper_t {
	ae_error_t ms_retval;
	uint32_t ms_sid;
} ms_close_session_wrapper_t;

typedef struct ms_invoke_service_wrapper_t {
	ae_error_t ms_retval;
	uint64_t ms_tick;
	uint8_t* ms_req_msg;
	uint32_t ms_req_msg_size;
	uint8_t* ms_resp_msg;
	uint32_t ms_resp_msg_size;
} ms_invoke_service_wrapper_t;

typedef struct ms_initialize_sqlite_database_file_wrapper_t {
	ae_error_t ms_retval;
	bool ms_is_for_empty_db_creation;
} ms_initialize_sqlite_database_file_wrapper_t;

typedef struct ms_ephemeral_session_m2m3_wrapper_t {
	ae_error_t ms_retval;
	pairing_blob_t* ms_sealed_blob;
	pse_cse_msg2_t* ms_pse_cse_msg2;
	pse_cse_msg3_t* ms_pse_cse_msg3;
} ms_ephemeral_session_m2m3_wrapper_t;

typedef struct ms_ephemeral_session_m4_wrapper_t {
	ae_error_t ms_retval;
	pse_cse_msg4_t* ms_pse_cse_msg4;
} ms_ephemeral_session_m4_wrapper_t;

typedef struct ms_sqlite_db_init_hash_tree_table_t {
	pse_op_error_t ms_retval;
} ms_sqlite_db_init_hash_tree_table_t;

typedef struct ms_sqlite_read_db_t {
	pse_op_error_t ms_retval;
	uint32_t ms_leaf_id;
	pse_vmc_hash_tree_cache_t* ms_cache;
} ms_sqlite_read_db_t;

typedef struct ms_sqlite_write_db_t {
	pse_op_error_t ms_retval;
	pse_vmc_hash_tree_cache_t* ms_cache;
	uint8_t ms_is_for_update_flag;
	op_leafnode_flag_t* ms_op_flag_info;
} ms_sqlite_write_db_t;

typedef struct ms_sqlite_read_children_of_root_t {
	pse_op_error_t ms_retval;
	pse_vmc_children_of_root_t* ms_children;
} ms_sqlite_read_children_of_root_t;

typedef struct ms_sqlite_get_empty_leafnode_t {
	pse_op_error_t ms_retval;
	int* ms_leaf_node_id;
	sgx_measurement_t* ms_mr_signer;
} ms_sqlite_get_empty_leafnode_t;

typedef struct ms_psda_invoke_service_ocall_t {
	ae_error_t ms_retval;
	uint8_t* ms_psda_req_msg;
	uint32_t ms_psda_req_msg_size;
	uint8_t* ms_psda_resp_msg;
	uint32_t ms_psda_resp_msg_size;
} ms_psda_invoke_service_ocall_t;

typedef struct ms_sqlite_rollback_db_file_t {
	pse_op_error_t ms_retval;
} ms_sqlite_rollback_db_file_t;

static sgx_status_t SGX_CDECL pse_op_sqlite_db_init_hash_tree_table(void* pms)
{
	ms_sqlite_db_init_hash_tree_table_t* ms = SGX_CAST(ms_sqlite_db_init_hash_tree_table_t*, pms);
	ms->ms_retval = sqlite_db_init_hash_tree_table();

	return SGX_SUCCESS;
}

static sgx_status_t SGX_CDECL pse_op_sqlite_read_db(void* pms)
{
	ms_sqlite_read_db_t* ms = SGX_CAST(ms_sqlite_read_db_t*, pms);
	ms->ms_retval = sqlite_read_db(ms->ms_leaf_id, ms->ms_cache);

	return SGX_SUCCESS;
}

static sgx_status_t SGX_CDECL pse_op_sqlite_write_db(void* pms)
{
	ms_sqlite_write_db_t* ms = SGX_CAST(ms_sqlite_write_db_t*, pms);
	ms->ms_retval = sqlite_write_db(ms->ms_cache, ms->ms_is_for_update_flag, ms->ms_op_flag_info);

	return SGX_SUCCESS;
}

static sgx_status_t SGX_CDECL pse_op_sqlite_read_children_of_root(void* pms)
{
	ms_sqlite_read_children_of_root_t* ms = SGX_CAST(ms_sqlite_read_children_of_root_t*, pms);
	ms->ms_retval = sqlite_read_children_of_root(ms->ms_children);

	return SGX_SUCCESS;
}

static sgx_status_t SGX_CDECL pse_op_sqlite_get_empty_leafnode(void* pms)
{
	ms_sqlite_get_empty_leafnode_t* ms = SGX_CAST(ms_sqlite_get_empty_leafnode_t*, pms);
	ms->ms_retval = sqlite_get_empty_leafnode(ms->ms_leaf_node_id, ms->ms_mr_signer);

	return SGX_SUCCESS;
}

static sgx_status_t SGX_CDECL pse_op_psda_invoke_service_ocall(void* pms)
{
	ms_psda_invoke_service_ocall_t* ms = SGX_CAST(ms_psda_invoke_service_ocall_t*, pms);
	ms->ms_retval = psda_invoke_service_ocall(ms->ms_psda_req_msg, ms->ms_psda_req_msg_size, ms->ms_psda_resp_msg, ms->ms_psda_resp_msg_size);

	return SGX_SUCCESS;
}

static sgx_status_t SGX_CDECL pse_op_sqlite_rollback_db_file(void* pms)
{
	ms_sqlite_rollback_db_file_t* ms = SGX_CAST(ms_sqlite_rollback_db_file_t*, pms);
	ms->ms_retval = sqlite_rollback_db_file();

	return SGX_SUCCESS;
}

static const struct {
	size_t nr_ocall;
	void * table[7];
} ocall_table_pse_op = {
	7,
	{
		(void*)pse_op_sqlite_db_init_hash_tree_table,
		(void*)pse_op_sqlite_read_db,
		(void*)pse_op_sqlite_write_db,
		(void*)pse_op_sqlite_read_children_of_root,
		(void*)pse_op_sqlite_get_empty_leafnode,
		(void*)pse_op_psda_invoke_service_ocall,
		(void*)pse_op_sqlite_rollback_db_file,
	}
};
sgx_status_t create_session_wrapper(sgx_enclave_id_t eid, ae_error_t* retval, uint64_t tick, uint32_t* id, pse_dh_msg1_t* dh_msg1)
{
	sgx_status_t status;
	ms_create_session_wrapper_t ms;
	ms.ms_tick = tick;
	ms.ms_id = id;
	ms.ms_dh_msg1 = dh_msg1;
	status = sgx_ecall(eid, 0, &ocall_table_pse_op, &ms);
	if (status == SGX_SUCCESS && retval) *retval = ms.ms_retval;
	return status;
}

sgx_status_t exchange_report_wrapper(sgx_enclave_id_t eid, ae_error_t* retval, uint64_t tick, uint32_t sid, sgx_dh_msg2_t* dh_msg2, pse_dh_msg3_t* dh_msg3)
{
	sgx_status_t status;
	ms_exchange_report_wrapper_t ms;
	ms.ms_tick = tick;
	ms.ms_sid = sid;
	ms.ms_dh_msg2 = dh_msg2;
	ms.ms_dh_msg3 = dh_msg3;
	status = sgx_ecall(eid, 1, &ocall_table_pse_op, &ms);
	if (status == SGX_SUCCESS && retval) *retval = ms.ms_retval;
	return status;
}

sgx_status_t close_session_wrapper(sgx_enclave_id_t eid, ae_error_t* retval, uint32_t sid)
{
	sgx_status_t status;
	ms_close_session_wrapper_t ms;
	ms.ms_sid = sid;
	status = sgx_ecall(eid, 2, &ocall_table_pse_op, &ms);
	if (status == SGX_SUCCESS && retval) *retval = ms.ms_retval;
	return status;
}

sgx_status_t invoke_service_wrapper(sgx_enclave_id_t eid, ae_error_t* retval, uint64_t tick, uint8_t* req_msg, uint32_t req_msg_size, uint8_t* resp_msg, uint32_t resp_msg_size)
{
	sgx_status_t status;
	ms_invoke_service_wrapper_t ms;
	ms.ms_tick = tick;
	ms.ms_req_msg = req_msg;
	ms.ms_req_msg_size = req_msg_size;
	ms.ms_resp_msg = resp_msg;
	ms.ms_resp_msg_size = resp_msg_size;
	status = sgx_ecall(eid, 3, &ocall_table_pse_op, &ms);
	if (status == SGX_SUCCESS && retval) *retval = ms.ms_retval;
	return status;
}

sgx_status_t initialize_sqlite_database_file_wrapper(sgx_enclave_id_t eid, ae_error_t* retval, bool is_for_empty_db_creation)
{
	sgx_status_t status;
	ms_initialize_sqlite_database_file_wrapper_t ms;
	ms.ms_is_for_empty_db_creation = is_for_empty_db_creation;
	status = sgx_ecall(eid, 4, &ocall_table_pse_op, &ms);
	if (status == SGX_SUCCESS && retval) *retval = ms.ms_retval;
	return status;
}

sgx_status_t ephemeral_session_m2m3_wrapper(sgx_enclave_id_t eid, ae_error_t* retval, pairing_blob_t* sealed_blob, pse_cse_msg2_t* pse_cse_msg2, pse_cse_msg3_t* pse_cse_msg3)
{
	sgx_status_t status;
	ms_ephemeral_session_m2m3_wrapper_t ms;
	ms.ms_sealed_blob = sealed_blob;
	ms.ms_pse_cse_msg2 = pse_cse_msg2;
	ms.ms_pse_cse_msg3 = pse_cse_msg3;
	status = sgx_ecall(eid, 5, &ocall_table_pse_op, &ms);
	if (status == SGX_SUCCESS && retval) *retval = ms.ms_retval;
	return status;
}

sgx_status_t ephemeral_session_m4_wrapper(sgx_enclave_id_t eid, ae_error_t* retval, pse_cse_msg4_t* pse_cse_msg4)
{
	sgx_status_t status;
	ms_ephemeral_session_m4_wrapper_t ms;
	ms.ms_pse_cse_msg4 = pse_cse_msg4;
	status = sgx_ecall(eid, 6, &ocall_table_pse_op, &ms);
	if (status == SGX_SUCCESS && retval) *retval = ms.ms_retval;
	return status;
}

