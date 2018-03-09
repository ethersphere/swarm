#include "TEE.h"

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
    uint32_t ptr;
    uint32_t ctr;

    std::cout << "*************************************************" << "\n";
    std::cout << "************* TEST MONOTONIC COUNTER ************" << "\n";
    std::cout << "*************************************************" << "\n";

    create_counter(global_eid, &ptr); 
    std::cout << "create monotonic counter: " << ptr << "\n";

    read_counter(global_eid, &ptr, &ctr);
    std::cout << "read monotonic counter: " << ptr  << "\n";
    // printf("read counter: %d\n", ctr);

    std::cout << "increment monotonic counter: ";
    increment_counter(global_eid, &ptr);
    std::cout << ptr << "\n";
    read_counter(global_eid, &ptr, &ctr);

    std::cout << "increment monotonic counter three times ";
    increment_counter(global_eid, &ptr);
    increment_counter(global_eid, &ptr);
    increment_counter(global_eid, &ptr);
    std::cout << ptr << "\n\n";

    std::cout << "*************************************************" << "\n";
    std::cout << "******************* TEST ECDSA ******************" << "\n";
    std::cout << "*************************************************" << "\n";
    int ecc_ptr;
    test_ecc(global_eid, &ecc_ptr);
    return 0;

}
