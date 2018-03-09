#include "App.h"

/* Global EID shared by multiple threads */
sgx_enclave_id_t global_eid = 0;

// OCall implementations
void ocall_print(const char* str) {
    printf("%s\n", str);
}
void ocall_uint32_t_print(uint32_t *arr, size_t len) {
    for (int i = 0; i < len; i++) {
        printf("%02X", arr[i]);
    }
    printf("\n");
}
void ocall_uint8_t_print(uint8_t *arr, size_t len) {
    for (int i = 0; i < len; i++) {
        printf("%02X", arr[i]);
    }
    printf("\n");
}

int testMain(void) {
    if (initialize_enclave(&global_eid, "enclave.token", "enclave.signed.so") < 0) {
        std::cout << "Fail to initialize enclave." << std::endl;
        return 1;
    }
    int ptr;
    sgx_status_t status = generate_random_number(global_eid, &ptr);
    std::cout << status << std::endl;
    if (status != SGX_SUCCESS) {
        std::cout << "noob" << std::endl;
    }
    printf("Random number: %d\n", ptr);

    // Seal the random number
    size_t sealed_size = sizeof(sgx_sealed_data_t) + sizeof(ptr);
    uint8_t* sealed_data = (uint8_t*)malloc(sealed_size);

    sgx_status_t ecall_status;
    status = seal(global_eid, &ecall_status,
            (uint8_t*)&ptr, sizeof(ptr),
            (sgx_sealed_data_t*)sealed_data, sealed_size);

    if (!is_ecall_successful(status, "Sealing failed :(", ecall_status)) {
        return 1;
    }

    int unsealed;
    status = unseal(global_eid, &ecall_status,
            (sgx_sealed_data_t*)sealed_data, sealed_size,
            (uint8_t*)&unsealed, sizeof(unsealed));

    if (!is_ecall_successful(status, "Unsealing failed :(", ecall_status)) {
        return 1;
    }

    std::cout << "Seal round trip success! Receive back " << unsealed << std::endl;

    std::cout << "test monotonic counter" << "\n";
    create_counter(global_eid, &ptr); 
    printf("return from tcc: %d\n", ptr);
    uint64_t ctr;
    uint64_t uptr;
    read_counter(global_eid, &uptr, &ctr);
    printf("read counter: %ld\n", ctr);

    increment_counter(global_eid, &ptr);
    printf("increment counter: %d\n", ptr);

    read_counter(global_eid, &uptr, &ctr);
    printf("read counter: %ld\n", ctr); 
    increment_counter(global_eid, &ptr);
    printf("increment counter: %d\n", ptr);

    read_counter(global_eid, &uptr, &ctr);
    printf("read counter: %ld\n", ctr); 
    increment_counter(global_eid, &ptr);
    printf("increment counter: %d\n", ptr);

    read_counter(global_eid, &uptr, &ctr);
    printf("read counter: %ld\n", ctr); 

    test_ecc(global_eid, &ptr);
    return 0;

}

int main(void) {
    testMain();
    return 0;
}

void old_functions(void) {

}
