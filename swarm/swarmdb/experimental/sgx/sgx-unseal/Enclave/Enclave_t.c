#include "Enclave_t.h"

#include "sgx_trts.h" /* for sgx_ocalloc, sgx_is_outside_enclave */

#include <errno.h>
#include <string.h> /* for memcpy etc */
#include <stdlib.h> /* for malloc/free etc */

#define CHECK_REF_POINTER(ptr, siz) do {	\
	if (!(ptr) || ! sgx_is_outside_enclave((ptr), (siz)))	\
		return SGX_ERROR_INVALID_PARAMETER;\
} while (0)

#define CHECK_UNIQUE_POINTER(ptr, siz) do {	\
	if ((ptr) && ! sgx_is_outside_enclave((ptr), (siz)))	\
		return SGX_ERROR_INVALID_PARAMETER;\
} while (0)


typedef struct ms_seal_t {
	sgx_status_t ms_retval;
	uint8_t* ms_plaintext;
	size_t ms_plaintext_len;
	sgx_sealed_data_t* ms_sealed_data;
	size_t ms_sealed_size;
} ms_seal_t;

typedef struct ms_unseal_t {
	sgx_status_t ms_retval;
	sgx_sealed_data_t* ms_sealed_data;
	size_t ms_sealed_size;
	uint8_t* ms_plaintext;
	uint32_t ms_plaintext_len;
} ms_unseal_t;

static sgx_status_t SGX_CDECL sgx_seal(void* pms)
{
	CHECK_REF_POINTER(pms, sizeof(ms_seal_t));
	ms_seal_t* ms = SGX_CAST(ms_seal_t*, pms);
	sgx_status_t status = SGX_SUCCESS;
	uint8_t* _tmp_plaintext = ms->ms_plaintext;
	size_t _tmp_plaintext_len = ms->ms_plaintext_len;
	size_t _len_plaintext = _tmp_plaintext_len;
	uint8_t* _in_plaintext = NULL;
	sgx_sealed_data_t* _tmp_sealed_data = ms->ms_sealed_data;
	size_t _tmp_sealed_size = ms->ms_sealed_size;
	size_t _len_sealed_data = _tmp_sealed_size;
	sgx_sealed_data_t* _in_sealed_data = NULL;

	CHECK_UNIQUE_POINTER(_tmp_plaintext, _len_plaintext);
	CHECK_UNIQUE_POINTER(_tmp_sealed_data, _len_sealed_data);

	if (_tmp_plaintext != NULL) {
		_in_plaintext = (uint8_t*)malloc(_len_plaintext);
		if (_in_plaintext == NULL) {
			status = SGX_ERROR_OUT_OF_MEMORY;
			goto err;
		}

		memcpy(_in_plaintext, _tmp_plaintext, _len_plaintext);
	}
	if (_tmp_sealed_data != NULL) {
		if ((_in_sealed_data = (sgx_sealed_data_t*)malloc(_len_sealed_data)) == NULL) {
			status = SGX_ERROR_OUT_OF_MEMORY;
			goto err;
		}

		memset((void*)_in_sealed_data, 0, _len_sealed_data);
	}
	ms->ms_retval = seal(_in_plaintext, _tmp_plaintext_len, _in_sealed_data, _tmp_sealed_size);
err:
	if (_in_plaintext) free(_in_plaintext);
	if (_in_sealed_data) {
		memcpy(_tmp_sealed_data, _in_sealed_data, _len_sealed_data);
		free(_in_sealed_data);
	}

	return status;
}

static sgx_status_t SGX_CDECL sgx_unseal(void* pms)
{
	CHECK_REF_POINTER(pms, sizeof(ms_unseal_t));
	ms_unseal_t* ms = SGX_CAST(ms_unseal_t*, pms);
	sgx_status_t status = SGX_SUCCESS;
	sgx_sealed_data_t* _tmp_sealed_data = ms->ms_sealed_data;
	size_t _tmp_sealed_size = ms->ms_sealed_size;
	size_t _len_sealed_data = _tmp_sealed_size;
	sgx_sealed_data_t* _in_sealed_data = NULL;
	uint8_t* _tmp_plaintext = ms->ms_plaintext;
	uint32_t _tmp_plaintext_len = ms->ms_plaintext_len;
	size_t _len_plaintext = _tmp_plaintext_len;
	uint8_t* _in_plaintext = NULL;

	CHECK_UNIQUE_POINTER(_tmp_sealed_data, _len_sealed_data);
	CHECK_UNIQUE_POINTER(_tmp_plaintext, _len_plaintext);

	if (_tmp_sealed_data != NULL) {
		_in_sealed_data = (sgx_sealed_data_t*)malloc(_len_sealed_data);
		if (_in_sealed_data == NULL) {
			status = SGX_ERROR_OUT_OF_MEMORY;
			goto err;
		}

		memcpy(_in_sealed_data, _tmp_sealed_data, _len_sealed_data);
	}
	if (_tmp_plaintext != NULL) {
		if ((_in_plaintext = (uint8_t*)malloc(_len_plaintext)) == NULL) {
			status = SGX_ERROR_OUT_OF_MEMORY;
			goto err;
		}

		memset((void*)_in_plaintext, 0, _len_plaintext);
	}
	ms->ms_retval = unseal(_in_sealed_data, _tmp_sealed_size, _in_plaintext, _tmp_plaintext_len);
err:
	if (_in_sealed_data) free(_in_sealed_data);
	if (_in_plaintext) {
		memcpy(_tmp_plaintext, _in_plaintext, _len_plaintext);
		free(_in_plaintext);
	}

	return status;
}

SGX_EXTERNC const struct {
	size_t nr_ecall;
	struct {void* ecall_addr; uint8_t is_priv;} ecall_table[2];
} g_ecall_table = {
	2,
	{
		{(void*)(uintptr_t)sgx_seal, 0},
		{(void*)(uintptr_t)sgx_unseal, 0},
	}
};

SGX_EXTERNC const struct {
	size_t nr_ocall;
} g_dyn_entry_table = {
	0,
};


