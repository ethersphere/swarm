#include "Enclave_t.h"
#include "sgx_tae_service.h"
#include "sgx.h"
#include "sgx_tcrypto.h"
#include <stdio.h>
#include "sgx_trts.h"
#include <cstring>

// monotonic counter
sgx_mc_uuid_t counter_uuid;
void check_sgx_status(sgx_status_t &ret);

uint32_t latest_counter;

int test_ecc(void) {
    int i;
    sgx_status_t ret;
    sgx_ecc_state_handle_t ecc_handle;
    sgx_ec256_private_t p_private;
    sgx_ec256_public_t p_public;
    sgx_ec256_signature_t p_signature;

    uint8_t sample_data[8]
        = {0x12, 0x13, 0x3f, 0x00,
           0x9a, 0x02, 0x10, 0x53};

    ret = sgx_ecc256_open_context(&ecc_handle);
    if (ret != SGX_SUCCESS) {
        switch (ret) {
            case SGX_ERROR_OUT_OF_MEMORY:
                ocall_print("SGX_ERROR_OUT_OF_MEMORY");
                break;
            case SGX_ERROR_UNEXPECTED:
                ocall_print("SGX_ERROR_UNEXPECTED");
                break;
        }
    }
    // create private, public key pair
    ret = sgx_ecc256_create_key_pair(&p_private, &p_public, ecc_handle);
    ocall_print("ecc private key");
    ocall_uint8_t_print(p_private.r, SGX_ECP256_KEY_SIZE);

    ocall_print("ecc public key.gx");
    ocall_uint8_t_print(p_public.gx, SGX_ECP256_KEY_SIZE);
    ocall_print("ecc public key.gy");
    ocall_uint8_t_print(p_public.gy, SGX_ECP256_KEY_SIZE);
    
    // create digital signature used ecc
    ret = sgx_ecdsa_sign(sample_data,
                        sizeof(sample_data) / sizeof(sample_data[0]),
                        &p_private,
                        &p_signature,
                        ecc_handle);

    ocall_print("ecdsa signature x");
    ocall_uint32_t_print(p_signature.x, SGX_NISTP_ECP256_KEY_SIZE);
    ocall_print("ecdsa signature y");
    ocall_uint32_t_print(p_signature.y, SGX_NISTP_ECP256_KEY_SIZE);

    if (ret != SGX_SUCCESS) {
        ocall_print("ecdsa sign error");
    }
    
    // print p_signature
    uint8_t p_result;
    ret = sgx_ecdsa_verify(sample_data, 
                    8,
                    &p_public,
                    &p_signature,
                    &p_result,
                    ecc_handle);
    ocall_print("verify result");
    ocall_uint8_t_print(&p_result, 1); // 0 on success, 1 on fail

    ret = sgx_ecc256_close_context(ecc_handle);
    if (ret != SGX_SUCCESS) {
        ocall_print("ecc256 close fails");
    }

    sgx_sha256_hash_t p_hash;
    sgx_sha256_msg(sample_data, 8, &p_hash);
    ocall_print("sha256");
    ocall_uint8_t_print(p_hash, SGX_SHA256_HASH_SIZE);
}

uint32_t create_counter(void) {
    sgx_status_t ret;
    int busy_retry_times = 2;

    do {
        ret = sgx_create_pse_session();
    } while (ret == SGX_ERROR_BUSY && busy_retry_times--);
    if (ret != SGX_SUCCESS)
        return ret;

    ret = sgx_create_monotonic_counter(&counter_uuid, &latest_counter);
    check_sgx_status(ret);
    sgx_close_pse_session();

    return latest_counter;
}

uint32_t increment_counter(void) {
    sgx_status_t ret;
    int busy_retry_times = 2;

    do {
        ret = sgx_create_pse_session();
    } while (ret == SGX_ERROR_BUSY && busy_retry_times--);
    if (ret != SGX_SUCCESS)
        return ret;

    ret = sgx_increment_monotonic_counter(&counter_uuid, &latest_counter);
    check_sgx_status(ret);
    sgx_close_pse_session();
    return latest_counter;
}

uint32_t read_counter(uint32_t *ctr) {
    sgx_status_t ret ;
    int busy_retry_times = 2;

    do {
        ret = sgx_create_pse_session();
    } while (ret == SGX_ERROR_BUSY && busy_retry_times--);
    if (ret != SGX_SUCCESS)
        return ret;

    ret = sgx_read_monotonic_counter(&counter_uuid, &latest_counter);
    check_sgx_status(ret);
    sgx_close_pse_session();
    *ctr = latest_counter;
    return latest_counter;
}

uint32_t destroy_counter(void) {
    sgx_status_t ret ;
    int busy_retry_times = 2;

    do {
        ret = sgx_create_pse_session();
    } while (ret == SGX_ERROR_BUSY && busy_retry_times--);
    if (ret != SGX_SUCCESS)
        return ret;

    ret = sgx_destroy_monotonic_counter(&counter_uuid);
    check_sgx_status(ret);
    sgx_close_pse_session();
    return latest_counter;
}

void check_sgx_status(sgx_status_t &ret) {
    if(ret != SGX_SUCCESS)
    {
        switch(ret)
        {
            case SGX_ERROR_SERVICE_UNAVAILABLE:
                /* Architecture Enclave Service Manager is not installed or not
                   working properly.*/
                break;
            case SGX_ERROR_SERVICE_TIMEOUT:
                /* retry the operation later*/
                break;
            case SGX_ERROR_BUSY:
                /* retry the operation later*/
                break;
            case SGX_ERROR_MC_OVER_QUOTA:
                /* SGX Platform Service enforces a quota scheme on the Monotonic
                   Counters a SGX app can maintain. the enclave has reached the
                   quota.*/
                break;
            case SGX_ERROR_MC_USED_UP:
                /* the Monotonic Counter has been used up and cannot create
                   Monotonic Counter anymore.*/
                break;
            default:
                /*other errors*/
                break;
        }
    }

}
