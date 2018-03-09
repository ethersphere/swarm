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

typedef struct ms_sgxGetSha256_t {
	sgx_status_t ms_retval;
	uint8_t* ms_src;
	size_t ms_src_len;
	uint8_t* ms_hash;
	size_t ms_hash_len;
} ms_sgxGetSha256_t;

typedef struct ms_sgxEcc256CreateKeyPair_t {
	sgx_status_t ms_retval;
	sgx_ec256_private_t* ms_p_private;
	sgx_ec256_public_t* ms_p_public;
} ms_sgxEcc256CreateKeyPair_t;

typedef struct ms_sgxEcdsaSign_t {
	sgx_status_t ms_retval;
	uint8_t* ms_sample_data;
	size_t ms_sample_data_len;
	sgx_ec256_private_t* ms_p_private;
	sgx_ec256_signature_t* ms_p_signature;
} ms_sgxEcdsaSign_t;

typedef struct ms_ocall_print_t {
	char* ms_str;
} ms_ocall_print_t;

typedef struct ms_ocall_uint8_t_print_t {
	uint8_t* ms_arr;
	size_t ms_len;
} ms_ocall_uint8_t_print_t;

typedef struct ms_ocall_uint32_t_print_t {
	uint32_t* ms_arr;
	size_t ms_len;
} ms_ocall_uint32_t_print_t;

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

static sgx_status_t SGX_CDECL sgx_sgxGetSha256(void* pms)
{
	CHECK_REF_POINTER(pms, sizeof(ms_sgxGetSha256_t));
	ms_sgxGetSha256_t* ms = SGX_CAST(ms_sgxGetSha256_t*, pms);
	sgx_status_t status = SGX_SUCCESS;
	uint8_t* _tmp_src = ms->ms_src;
	uint8_t* _tmp_hash = ms->ms_hash;


	ms->ms_retval = sgxGetSha256(_tmp_src, ms->ms_src_len, _tmp_hash, ms->ms_hash_len);


	return status;
}

static sgx_status_t SGX_CDECL sgx_sgxEcc256CreateKeyPair(void* pms)
{
	CHECK_REF_POINTER(pms, sizeof(ms_sgxEcc256CreateKeyPair_t));
	ms_sgxEcc256CreateKeyPair_t* ms = SGX_CAST(ms_sgxEcc256CreateKeyPair_t*, pms);
	sgx_status_t status = SGX_SUCCESS;
	sgx_ec256_private_t* _tmp_p_private = ms->ms_p_private;
	sgx_ec256_public_t* _tmp_p_public = ms->ms_p_public;


	ms->ms_retval = sgxEcc256CreateKeyPair(_tmp_p_private, _tmp_p_public);


	return status;
}

static sgx_status_t SGX_CDECL sgx_sgxEcdsaSign(void* pms)
{
	CHECK_REF_POINTER(pms, sizeof(ms_sgxEcdsaSign_t));
	ms_sgxEcdsaSign_t* ms = SGX_CAST(ms_sgxEcdsaSign_t*, pms);
	sgx_status_t status = SGX_SUCCESS;
	uint8_t* _tmp_sample_data = ms->ms_sample_data;
	sgx_ec256_private_t* _tmp_p_private = ms->ms_p_private;
	sgx_ec256_signature_t* _tmp_p_signature = ms->ms_p_signature;


	ms->ms_retval = sgxEcdsaSign(_tmp_sample_data, ms->ms_sample_data_len, _tmp_p_private, _tmp_p_signature);


	return status;
}

SGX_EXTERNC const struct {
	size_t nr_ecall;
	struct {void* ecall_addr; uint8_t is_priv;} ecall_table[5];
} g_ecall_table = {
	5,
	{
		{(void*)(uintptr_t)sgx_seal, 0},
		{(void*)(uintptr_t)sgx_unseal, 0},
		{(void*)(uintptr_t)sgx_sgxGetSha256, 0},
		{(void*)(uintptr_t)sgx_sgxEcc256CreateKeyPair, 0},
		{(void*)(uintptr_t)sgx_sgxEcdsaSign, 0},
	}
};

SGX_EXTERNC const struct {
	size_t nr_ocall;
	uint8_t entry_table[3][5];
} g_dyn_entry_table = {
	3,
	{
		{0, 0, 0, 0, 0, },
		{0, 0, 0, 0, 0, },
		{0, 0, 0, 0, 0, },
	}
};


sgx_status_t SGX_CDECL ocall_print(const char* str)
{
	sgx_status_t status = SGX_SUCCESS;
	size_t _len_str = str ? strlen(str) + 1 : 0;

	ms_ocall_print_t* ms = NULL;
	size_t ocalloc_size = sizeof(ms_ocall_print_t);
	void *__tmp = NULL;

	ocalloc_size += (str != NULL && sgx_is_within_enclave(str, _len_str)) ? _len_str : 0;

	__tmp = sgx_ocalloc(ocalloc_size);
	if (__tmp == NULL) {
		sgx_ocfree();
		return SGX_ERROR_UNEXPECTED;
	}
	ms = (ms_ocall_print_t*)__tmp;
	__tmp = (void *)((size_t)__tmp + sizeof(ms_ocall_print_t));

	if (str != NULL && sgx_is_within_enclave(str, _len_str)) {
		ms->ms_str = (char*)__tmp;
		__tmp = (void *)((size_t)__tmp + _len_str);
		memcpy((void*)ms->ms_str, str, _len_str);
	} else if (str == NULL) {
		ms->ms_str = NULL;
	} else {
		sgx_ocfree();
		return SGX_ERROR_INVALID_PARAMETER;
	}
	
	status = sgx_ocall(0, ms);


	sgx_ocfree();
	return status;
}

sgx_status_t SGX_CDECL ocall_uint8_t_print(uint8_t* arr, size_t len)
{
	sgx_status_t status = SGX_SUCCESS;
	size_t _len_arr = len;

	ms_ocall_uint8_t_print_t* ms = NULL;
	size_t ocalloc_size = sizeof(ms_ocall_uint8_t_print_t);
	void *__tmp = NULL;

	ocalloc_size += (arr != NULL && sgx_is_within_enclave(arr, _len_arr)) ? _len_arr : 0;

	__tmp = sgx_ocalloc(ocalloc_size);
	if (__tmp == NULL) {
		sgx_ocfree();
		return SGX_ERROR_UNEXPECTED;
	}
	ms = (ms_ocall_uint8_t_print_t*)__tmp;
	__tmp = (void *)((size_t)__tmp + sizeof(ms_ocall_uint8_t_print_t));

	if (arr != NULL && sgx_is_within_enclave(arr, _len_arr)) {
		ms->ms_arr = (uint8_t*)__tmp;
		__tmp = (void *)((size_t)__tmp + _len_arr);
		memcpy(ms->ms_arr, arr, _len_arr);
	} else if (arr == NULL) {
		ms->ms_arr = NULL;
	} else {
		sgx_ocfree();
		return SGX_ERROR_INVALID_PARAMETER;
	}
	
	ms->ms_len = len;
	status = sgx_ocall(1, ms);


	sgx_ocfree();
	return status;
}

sgx_status_t SGX_CDECL ocall_uint32_t_print(uint32_t* arr, size_t len)
{
	sgx_status_t status = SGX_SUCCESS;
	size_t _len_arr = len;

	ms_ocall_uint32_t_print_t* ms = NULL;
	size_t ocalloc_size = sizeof(ms_ocall_uint32_t_print_t);
	void *__tmp = NULL;

	ocalloc_size += (arr != NULL && sgx_is_within_enclave(arr, _len_arr)) ? _len_arr : 0;

	__tmp = sgx_ocalloc(ocalloc_size);
	if (__tmp == NULL) {
		sgx_ocfree();
		return SGX_ERROR_UNEXPECTED;
	}
	ms = (ms_ocall_uint32_t_print_t*)__tmp;
	__tmp = (void *)((size_t)__tmp + sizeof(ms_ocall_uint32_t_print_t));

	if (arr != NULL && sgx_is_within_enclave(arr, _len_arr)) {
		ms->ms_arr = (uint32_t*)__tmp;
		__tmp = (void *)((size_t)__tmp + _len_arr);
		memcpy(ms->ms_arr, arr, _len_arr);
	} else if (arr == NULL) {
		ms->ms_arr = NULL;
	} else {
		sgx_ocfree();
		return SGX_ERROR_INVALID_PARAMETER;
	}
	
	ms->ms_len = len;
	status = sgx_ocall(2, ms);


	sgx_ocfree();
	return status;
}

