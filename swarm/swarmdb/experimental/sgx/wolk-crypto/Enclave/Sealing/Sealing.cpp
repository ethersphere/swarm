#include "sgx_trts.h"
#include "sgx_tseal.h"
#include "string.h"
#include "Enclave_t.h"

#include <stdio.h>

#include<string>
using namespace std;





/**
 * @brief      Seals the plaintext given into the sgx_sealed_data_t structure
 *             given.
 *
 * @details    The plaintext can be any data. uint8_t is used to represent a
 *             byte. The sealed size can be determined by computing
 *             sizeof(sgx_sealed_data_t) + plaintext_len, since it is using
 *             AES-GCM which preserves length of plaintext. The size needs to be
 *             specified, otherwise SGX will assume the size to be just
 *             sizeof(sgx_sealed_data_t), not taking into account the sealed
 *             payload.
 *
 * @param      plaintext      The data to be sealed
 * @param[in]  plaintext_len  The plaintext length
 * @param      sealed_data    The pointer to the sealed data structure
 * @param[in]  sealed_size    The size of the sealed data structure supplied
 *
 * @return     Truthy if seal successful, falsy otherwise.
 */
sgx_status_t seal(uint8_t* plaintext, size_t plaintext_len, sgx_sealed_data_t* sealed_data, size_t sealed_size) {
    sgx_status_t status = sgx_seal_data(0, NULL, plaintext_len, plaintext, sealed_size, sealed_data);
    return status;
}

/**
 * @brief      Unseal the sealed_data given into c-string
 *
 * @details    The resulting plaintext is of type uint8_t to represent a byte.
 *             The sizes/length of pointers need to be specified, otherwise SGX
 *             will assume a count of 1 for all pointers.
 *
 * @param      sealed_data        The sealed data
 * @param[in]  sealed_size        The size of the sealed data
 * @param      plaintext          A pointer to buffer to store the plaintext
 * @param[in]  plaintext_max_len  The size of buffer prepared to store the
 *                                plaintext
 *
 * @return     Truthy if unseal successful, falsy otherwise.
 */
sgx_status_t unseal(sgx_sealed_data_t* sealed_data, size_t sealed_size, uint8_t* plaintext, uint32_t plaintext_len) {
    sgx_status_t status = sgx_unseal_data(sealed_data, NULL, NULL, (uint8_t*)plaintext, &plaintext_len);
    return status;
}

sgx_status_t sgxGetSha256(uint8_t* src, size_t src_len, uint8_t* hash, size_t hash_len) {

    sgx_status_t sgx_ret = SGX_SUCCESS;
    sgx_sha_state_handle_t sha_context;
    sgx_sha256_hash_t sgx_hash;

    sgx_ret = sgx_sha256_init(&sha_context);
    if (sgx_ret != SGX_SUCCESS)
    {
        return sgx_ret;
    }

    sgx_ret = sgx_sha256_update((uint8_t*)src, src_len, sha_context);
    if (sgx_ret != SGX_SUCCESS)
    {
        sgx_sha256_close(sha_context);
        return sgx_ret;
    }

    sgx_ret = sgx_sha256_get_hash(sha_context, &sgx_hash);
    if (sgx_ret != SGX_SUCCESS)
    {
        sgx_sha256_close(sha_context);
        return sgx_ret;
    }

    memcpy(hash, sgx_hash, 32);

    sgx_ret = sgx_sha256_close(sha_context);

    return sgx_ret;
}

sgx_status_t sgxEcc256CreateKeyPair(sgx_ec256_private_t* p_private, sgx_ec256_public_t* p_public) {

    sgx_status_t sgx_ret = SGX_SUCCESS;
    sgx_ecc_state_handle_t ecc_handle;

    sgx_ret = sgx_ecc256_open_context(&ecc_handle);
    if (sgx_ret != SGX_SUCCESS) {
        switch (sgx_ret) {
            case SGX_ERROR_OUT_OF_MEMORY:
                //ocall_print("SGX_ERROR_OUT_OF_MEMORY");
                break;
            case SGX_ERROR_UNEXPECTED:
                //ocall_print("SGX_ERROR_UNEXPECTED");
                break;
        }
    }

    // create private, public key pair
    sgx_ret = sgx_ecc256_create_key_pair(p_private, p_public, ecc_handle);
    if (sgx_ret != SGX_SUCCESS)
    {
        return sgx_ret;
    }

    /*
    swarm.wolk.com/sgx/go-with-intel-sgx/Enclave/Enclave.cpp

    ocall_print("ecc private key");
    ocall_uint8_t_print(p_private.r, SGX_ECP256_KEY_SIZE);

    ocall_print("ecc public key.gx");
    ocall_uint8_t_print(p_public.gx, SGX_ECP256_KEY_SIZE);
    ocall_print("ecc public key.gy");
    ocall_uint8_t_print(p_public.gy, SGX_ECP256_KEY_SIZE);
    */

    sgx_ec256_signature_t p_signature;

    uint8_t sample_data[8]
        = {0x12, 0x13, 0x3f, 0x00,
           0x9a, 0x02, 0x10, 0x53};

     //12133f009a021053

    sgx_sha256_hash_t p_hash;
    sgx_sha256_msg(sample_data, 8, &p_hash);
    ocall_print("sha256");
    ocall_uint8_t_print(p_hash, SGX_SHA256_HASH_SIZE);



    // create digital signature used ecc
    sgx_ret = sgx_ecdsa_sign(p_hash,
                        32,
                        p_private,                                            //&p_private,
                        &p_signature,
                        ecc_handle);

    ocall_print("ecdsa signature x");
    ocall_uint32_t_print(p_signature.x, SGX_NISTP_ECP256_KEY_SIZE);
    ocall_print("ecdsa signature y");
    ocall_uint32_t_print(p_signature.y, SGX_NISTP_ECP256_KEY_SIZE);

    if (sgx_ret != SGX_SUCCESS) {
        ocall_print("ecdsa sign error");
    }










    // print p_signature
    uint8_t p_result;
    sgx_ret = sgx_ecdsa_verify(p_hash,
                    32,
                    p_public,                                      //&p_public,
                    &p_signature,
                    &p_result,
                    ecc_handle);
    ocall_print("verify result");
    ocall_uint8_t_print(&p_result, 1); // 0 on success, 1 on fail

    sgx_ret = sgx_ecc256_close_context(ecc_handle);
    if (sgx_ret != SGX_SUCCESS) {
        ocall_print("ecc256 close fails");
    }



    //sgx_sha256_hash_t p_hash;
    //sgx_sha256_msg(sample_data, 8, &p_hash);
    //ocall_print("sha256");
    //ocall_uint8_t_print(p_hash, SGX_SHA256_HASH_SIZE);

    return sgx_ret;
}


sgx_status_t sgxEcc256CreateKeyPair2(sgx_ec256_private_t* p_private, sgx_ec256_public_t* p_public) {

    sgx_status_t sgx_ret = SGX_SUCCESS;
    sgx_ecc_state_handle_t ecc_handle;

    sgx_ret = sgx_ecc256_open_context(&ecc_handle);
    if (sgx_ret != SGX_SUCCESS) {
        switch (sgx_ret) {
            case SGX_ERROR_OUT_OF_MEMORY:
                //ocall_print("SGX_ERROR_OUT_OF_MEMORY");
                break;
            case SGX_ERROR_UNEXPECTED:
                //ocall_print("SGX_ERROR_UNEXPECTED");
                break;
        }
    }

    // create private, public key pair
    sgx_ret = sgx_ecc256_create_key_pair(p_private, p_public, ecc_handle);
    if (sgx_ret != SGX_SUCCESS)
    {
        return sgx_ret;
    }

    /*
    swarm.wolk.com/sgx/go-with-intel-sgx/Enclave/Enclave.cpp

    ocall_print("ecc private key");
    ocall_uint8_t_print(p_private.r, SGX_ECP256_KEY_SIZE);

    ocall_print("ecc public key.gx");
    ocall_uint8_t_print(p_public.gx, SGX_ECP256_KEY_SIZE);
    ocall_print("ecc public key.gy");
    ocall_uint8_t_print(p_public.gy, SGX_ECP256_KEY_SIZE);
    */

    sgx_ec256_signature_t p_signature;


// ************************************************ test uint8_t array ****************************************************************************************
    //Short answer: a null terminated string is a char array with a null value (0x00) after the last valid character in the string.
    //uint8_t hash[32] = "\0"; // empty array 0000000000000000000000000000000000000000000000000000000000000000
    //It's important to remember that not every C and C++ compiler will initialize values for you
    //
    // https://www.rapidtables.com/code/text/ascii-table.html
    // Dec	Hex	Binary	    Character Description
    // 0	00	00000000	NUL	      null
/*
    uint8_t sample_data_test[10]
        = {0x41, 0x42, 0x43,   // A B C
           0x61, 0x62, 0x63,   // a b c
		   0x31, 0x32, 0x33, 0x00};  // 1 2 3 0
    ocall_uint8_t_print(sample_data_test, 10); //41424361626331323300
    ocall_print((char*)sample_data_test); //ABCabc123


    uint8_t sample_data_test2[10] = "ABCabc123";
    ocall_uint8_t_print(sample_data_test2, 10); //41424361626331323300
    ocall_print((char*)sample_data_test2); //ABCabc123
*/
// ************************************************ test uint8_t array ****************************************************************************************
/*
    uint8_t sample_data[8]
        = {0x12, 0x13, 0x3f, 0x00,
           0x9a, 0x02, 0x10, 0x53};
*/






// ************************************************ hex to string ****************************************************************************************
/*
string hex = "48656c6c6f";
int len = hex.length();
std::string newString;
for(int i=0; i< len; i+=2)
{
    string byte = hex.substr(i,2);
    char chr = (char) (int)strtol(byte.c_str(), NULL, 16);
    newString.push_back(chr);
}

ocall_print((char*)&newString);
*/
// ************************************************ hex to string ****************************************************************************************


    uint8_t sample_data[25] = "sAFcbjKkwBOCtyNJFroPxWqn";
    ocall_uint8_t_print(sample_data, 25); //73414663626A4B6B77424F4374794E4A46726F507857716E00
    ocall_print((char*)sample_data); //sAFcbjKkwBOCtyNJFroPxWqn

    sgx_sha256_hash_t p_hash;
    sgx_sha256_msg(sample_data, 24, &p_hash);
    ocall_print("sha256");
    ocall_uint8_t_print(p_hash, SGX_SHA256_HASH_SIZE);


    sgx_ec256_private_t p_private2;
    string hex = "ec558883af8d3c6783b3ad00fd17695492b42f172c001162ef29e21086562cfe";
    int len = hex.length();
    uint8_t r[SGX_ECP256_KEY_SIZE];
    int j=0;
    for(int i=0; i< len; i+=2)
    {
        string byte = hex.substr(i,2);
        p_private2.r[j] = (uint8_t) (int)strtol(byte.c_str(), NULL, 16);
        j=j+1;
    }


    ocall_uint8_t_print(p_private2.r, SGX_SHA256_HASH_SIZE);


    // create digital signature used ecc
    sgx_ret = sgx_ecdsa_sign(p_hash,
                        sizeof(sample_data) / sizeof(sample_data[0]),
                        &p_private2,                                            //&p_private,
                        &p_signature,
                        ecc_handle);

    ocall_print("ecdsa signature x");
    ocall_uint32_t_print(p_signature.x, SGX_NISTP_ECP256_KEY_SIZE);
    ocall_print("ecdsa signature y");
    ocall_uint32_t_print(p_signature.y, SGX_NISTP_ECP256_KEY_SIZE);

    if (sgx_ret != SGX_SUCCESS) {
        ocall_print("ecdsa sign error");
    }



    // print p_signature
    uint8_t p_result;
    sgx_ret = sgx_ecdsa_verify(sample_data,
    		        sizeof(sample_data) / sizeof(sample_data[0]),  //8
                    p_public,                                      //&p_public,
                    &p_signature,
                    &p_result,
                    ecc_handle);
    ocall_print("verify result");
    ocall_uint8_t_print(&p_result, 1); // 0 on success, 1 on fail

    sgx_ret = sgx_ecc256_close_context(ecc_handle);
    if (sgx_ret != SGX_SUCCESS) {
        ocall_print("ecc256 close fails");
    }



    //sgx_sha256_hash_t p_hash;
    //sgx_sha256_msg(sample_data, 8, &p_hash);
    //ocall_print("sha256");
    //ocall_uint8_t_print(p_hash, SGX_SHA256_HASH_SIZE);

    return sgx_ret;
}

sgx_status_t sgxEcdsaSign(uint8_t* sample_data, size_t sample_data_len, sgx_ec256_private_t* p_private, sgx_ec256_signature_t* p_signature) {

    sgx_status_t sgx_ret = SGX_SUCCESS;
    sgx_ecc_state_handle_t ecc_handle;

    sgx_ret = sgx_ecc256_open_context(&ecc_handle);
    if (sgx_ret != SGX_SUCCESS) {
        switch (sgx_ret) {
            case SGX_ERROR_OUT_OF_MEMORY:
                //ocall_print("SGX_ERROR_OUT_OF_MEMORY");
                break;
            case SGX_ERROR_UNEXPECTED:
                //ocall_print("SGX_ERROR_UNEXPECTED");
                break;
        }
    }

    // create digital signature used ecc
    sgx_ret = sgx_ecdsa_sign(sample_data,
                         sizeof(sample_data) / sizeof(sample_data[0]),
                         p_private,
                         p_signature,
                         ecc_handle);

    /*
    ocall_print("ecdsa signature x");
    ocall_uint32_t_print(p_signature.x, SGX_NISTP_ECP256_KEY_SIZE);
    ocall_print("ecdsa signature y");
    ocall_uint32_t_print(p_signature.y, SGX_NISTP_ECP256_KEY_SIZE);
    */

    if (sgx_ret != SGX_SUCCESS) {
       // ocall_print("ecdsa sign error");
    }

    return sgx_ret;
}









