#include "launch_enclave_u.h"
#include <errno.h>

typedef struct ms_le_get_launch_token_wrapper_t {
	int ms_retval;
	sgx_measurement_t* ms_mrenclave;
	sgx_measurement_t* ms_mrsigner;
	sgx_attributes_t* ms_se_attributes;
	token_t* ms_lictoken;
} ms_le_get_launch_token_wrapper_t;

typedef struct ms_le_init_white_list_wrapper_t {
	uint32_t ms_retval;
	uint8_t* ms_wl_cert_chain;
	uint32_t ms_wl_cert_chain_size;
} ms_le_init_white_list_wrapper_t;

static const struct {
	size_t nr_ocall;
	void * table[1];
} ocall_table_launch_enclave = {
	0,
	{ NULL },
};
sgx_status_t le_get_launch_token_wrapper(sgx_enclave_id_t eid, int* retval, const sgx_measurement_t* mrenclave, const sgx_measurement_t* mrsigner, const sgx_attributes_t* se_attributes, token_t* lictoken)
{
	sgx_status_t status;
	ms_le_get_launch_token_wrapper_t ms;
	ms.ms_mrenclave = (sgx_measurement_t*)mrenclave;
	ms.ms_mrsigner = (sgx_measurement_t*)mrsigner;
	ms.ms_se_attributes = (sgx_attributes_t*)se_attributes;
	ms.ms_lictoken = lictoken;
	status = sgx_ecall(eid, 0, &ocall_table_launch_enclave, &ms);
	if (status == SGX_SUCCESS && retval) *retval = ms.ms_retval;
	return status;
}

sgx_status_t le_init_white_list_wrapper(sgx_enclave_id_t eid, uint32_t* retval, const uint8_t* wl_cert_chain, uint32_t wl_cert_chain_size)
{
	sgx_status_t status;
	ms_le_init_white_list_wrapper_t ms;
	ms.ms_wl_cert_chain = (uint8_t*)wl_cert_chain;
	ms.ms_wl_cert_chain_size = wl_cert_chain_size;
	status = sgx_ecall(eid, 1, &ocall_table_launch_enclave, &ms);
	if (status == SGX_SUCCESS && retval) *retval = ms.ms_retval;
	return status;
}

