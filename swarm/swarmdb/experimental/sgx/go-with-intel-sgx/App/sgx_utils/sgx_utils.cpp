#include <cstdio>
#include <cstring>
#include "sgx_urts.h"
#include "sgx_utils.h"

#ifndef TRUE
# define TRUE 1
#endif

#ifndef FALSE
# define FALSE 0
#endif

/* Check error conditions for loading enclave */
void print_error_message(sgx_status_t ret) {
    printf("SGX error code: %d\n", ret);
}

/* Initialize the enclave:
 *   Step 1: try to retrieve the launch token saved by last transaction
 *   Step 2: call sgx_create_enclave to initialize an enclave instance
 *   Step 3: save the launch token if it is updated
 */
// if (initialize_enclave(&global_eid, "enclave.token", "enclave.signed.so") < 0) 
int initialize_enclave(sgx_enclave_id_t* eid, const std::string& launch_token_path, const std::string& enclave_name) {
    const char* token_path = launch_token_path.c_str();
    sgx_launch_token_t token = {0};
    sgx_status_t ret = SGX_ERROR_UNEXPECTED;
    int updated = 0;

    /* Step 1: try to retrieve the launch token saved by last transaction
     *         if there is no token, then create a new one.
     */
    /* try to get the token saved in $HOME */
    FILE* fp = fopen(token_path, "rb");
    if (fp == NULL && (fp = fopen(token_path, "wb")) == NULL) {
        printf("Warning: Failed to create/open the launch token file \"%s\".\n", token_path);
    }

    if (fp != NULL) {
        /* read the token from saved file */
        size_t read_num = fread(token, 1, sizeof(sgx_launch_token_t), fp);
        if (read_num != 0 && read_num != sizeof(sgx_launch_token_t)) {
            /* if token is invalid, clear the buffer */
            memset(&token, 0x0, sizeof(sgx_launch_token_t));
            printf("Warning: Invalid launch token read from \"%s\".\n", token_path);
        }
    }
    /* Step 2: call sgx_create_enclave to initialize an enclave instance */
    /* Debug Support: set 2nd parameter to 1 */
    ret = sgx_create_enclave(enclave_name.c_str(), SGX_DEBUG_FLAG, &token, &updated, eid, NULL);
    if (ret != SGX_SUCCESS) {
        print_error_message(ret);
        if (fp != NULL) fclose(fp);
        return -1;
    }

    /* Step 3: save the launch token if it is updated */
    if (updated == FALSE || fp == NULL) {
        /* if the token is not updated, or file handler is invalid, do not perform saving */
        if (fp != NULL) fclose(fp);
        return 0;
    }

    /* reopen the file with write capablity */
    fp = freopen(token_path, "wb", fp);
    if (fp == NULL) return 0;
    size_t write_num = fwrite(token, 1, sizeof(sgx_launch_token_t), fp);
    if (write_num != sizeof(sgx_launch_token_t))
        printf("Warning: Failed to save launch token to \"%s\".\n", token_path);
    fclose(fp);
    return 0;
}

bool is_ecall_successful(sgx_status_t sgx_status, const std::string& err_msg,
        sgx_status_t ecall_return_value) {
    if (sgx_status != SGX_SUCCESS || ecall_return_value != SGX_SUCCESS) {
        printf("%s\n", err_msg.c_str());
        print_error_message(sgx_status);
        print_error_message(ecall_return_value);
        return false;
    }
    return true;
}
