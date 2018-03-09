#ifndef ENCLAVE_T_H__
#define ENCLAVE_T_H__

#include <stdint.h>
#include <wchar.h>
#include <stddef.h>
#include "sgx_edger8r.h" /* for sgx_ocall etc. */

#include "sgx_tseal.h"

#include <stdlib.h> /* for size_t */

#define SGX_CAST(type, item) ((type)(item))

#ifdef __cplusplus
extern "C" {
#endif


sgx_status_t seal(uint8_t* plaintext, size_t plaintext_len, sgx_sealed_data_t* sealed_data, size_t sealed_size);
sgx_status_t unseal(sgx_sealed_data_t* sealed_data, size_t sealed_size, uint8_t* plaintext, uint32_t plaintext_len);
sgx_status_t sgxGetSha256(uint8_t* src, size_t src_len, uint8_t* hash, size_t hash_len);
sgx_status_t sgxEcc256CreateKeyPair(sgx_ec256_private_t* p_private, sgx_ec256_public_t* p_public);
sgx_status_t sgxEcdsaSign(uint8_t* sample_data, size_t sample_data_len, sgx_ec256_private_t* p_private, sgx_ec256_signature_t* p_signature);

sgx_status_t SGX_CDECL ocall_print(const char* str);
sgx_status_t SGX_CDECL ocall_uint8_t_print(uint8_t* arr, size_t len);
sgx_status_t SGX_CDECL ocall_uint32_t_print(uint32_t* arr, size_t len);

#ifdef __cplusplus
}
#endif /* __cplusplus */

#endif
