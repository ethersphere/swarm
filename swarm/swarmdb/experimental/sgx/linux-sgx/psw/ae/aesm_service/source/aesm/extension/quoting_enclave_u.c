#include "quoting_enclave_u.h"
#include <errno.h>

typedef struct ms_verify_blob_t {
	uint32_t ms_retval;
	uint8_t* ms_p_blob;
	uint32_t ms_blob_size;
	uint8_t* ms_p_is_resealed;
} ms_verify_blob_t;

typedef struct ms_get_quote_t {
	uint32_t ms_retval;
	uint8_t* ms_p_blob;
	uint32_t ms_blob_size;
	sgx_report_t* ms_p_report;
	sgx_quote_sign_type_t ms_quote_type;
	sgx_spid_t* ms_p_spid;
	sgx_quote_nonce_t* ms_p_nonce;
	uint8_t* ms_p_sig_rl;
	uint32_t ms_sig_rl_size;
	sgx_report_t* ms_qe_report;
	uint8_t* ms_p_quote;
	uint32_t ms_quote_size;
	sgx_isv_svn_t ms_pce_isvnsvn;
} ms_get_quote_t;

static const struct {
	size_t nr_ocall;
	void * table[1];
} ocall_table_quoting_enclave = {
	0,
	{ NULL },
};
sgx_status_t verify_blob(sgx_enclave_id_t eid, uint32_t* retval, uint8_t* p_blob, uint32_t blob_size, uint8_t* p_is_resealed)
{
	sgx_status_t status;
	ms_verify_blob_t ms;
	ms.ms_p_blob = p_blob;
	ms.ms_blob_size = blob_size;
	ms.ms_p_is_resealed = p_is_resealed;
	status = sgx_ecall(eid, 0, &ocall_table_quoting_enclave, &ms);
	if (status == SGX_SUCCESS && retval) *retval = ms.ms_retval;
	return status;
}

sgx_status_t get_quote(sgx_enclave_id_t eid, uint32_t* retval, uint8_t* p_blob, uint32_t blob_size, const sgx_report_t* p_report, sgx_quote_sign_type_t quote_type, const sgx_spid_t* p_spid, const sgx_quote_nonce_t* p_nonce, const uint8_t* p_sig_rl, uint32_t sig_rl_size, sgx_report_t* qe_report, uint8_t* p_quote, uint32_t quote_size, sgx_isv_svn_t pce_isvnsvn)
{
	sgx_status_t status;
	ms_get_quote_t ms;
	ms.ms_p_blob = p_blob;
	ms.ms_blob_size = blob_size;
	ms.ms_p_report = (sgx_report_t*)p_report;
	ms.ms_quote_type = quote_type;
	ms.ms_p_spid = (sgx_spid_t*)p_spid;
	ms.ms_p_nonce = (sgx_quote_nonce_t*)p_nonce;
	ms.ms_p_sig_rl = (uint8_t*)p_sig_rl;
	ms.ms_sig_rl_size = sig_rl_size;
	ms.ms_qe_report = qe_report;
	ms.ms_p_quote = p_quote;
	ms.ms_quote_size = quote_size;
	ms.ms_pce_isvnsvn = pce_isvnsvn;
	status = sgx_ecall(eid, 1, &ocall_table_quoting_enclave, &ms);
	if (status == SGX_SUCCESS && retval) *retval = ms.ms_retval;
	return status;
}

