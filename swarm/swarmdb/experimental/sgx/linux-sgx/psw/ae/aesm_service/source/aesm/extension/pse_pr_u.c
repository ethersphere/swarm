#include "pse_pr_u.h"
#include <errno.h>

typedef struct ms_ecall_tPrepareForCertificateProvisioning_t {
	ae_error_t ms_retval;
	uint64_t ms_nonce64;
	sgx_target_info_t* ms_pTargetInfo;
	uint16_t ms_nMaxLen_CSR_pse;
	uint8_t* ms_pCSR_pse;
	uint16_t* ms_pnTotalLen_CSR_pse;
	sgx_report_t* ms_pREPORT;
	pairing_blob_t* ms_pPairingBlob;
} ms_ecall_tPrepareForCertificateProvisioning_t;

typedef struct ms_ecall_tGenM7_t {
	ae_error_t ms_retval;
	SIGMA_S1_MESSAGE* ms_pS1;
	EPID11_SIG_RL* ms_pSigRL;
	uint8_t* ms_pOcspResp;
	uint32_t ms_nTotalLen_OcspResp;
	uint8_t* ms_pVerifierCert;
	uint32_t ms_nTotalLen_VerifierCert;
	pairing_blob_t* ms_pPairingBlob;
	uint32_t ms_nMaxLen_S2;
	SIGMA_S2_MESSAGE* ms_pS2;
	uint32_t* ms_pnTotalLen_S2;
} ms_ecall_tGenM7_t;

typedef struct ms_ecall_tVerifyM8_t {
	ae_error_t ms_retval;
	SIGMA_S3_MESSAGE* ms_pS3;
	uint32_t ms_nTotalLen_S3;
	EPID11_PRIV_RL* ms_pPrivRL;
	pairing_blob_t* ms_pPairingBlob;
	uint8_t* ms_bNewPairing;
} ms_ecall_tVerifyM8_t;

static const struct {
	size_t nr_ocall;
	void * table[1];
} ocall_table_pse_pr = {
	0,
	{ NULL },
};
sgx_status_t ecall_tPrepareForCertificateProvisioning(sgx_enclave_id_t eid, ae_error_t* retval, uint64_t nonce64, const sgx_target_info_t* pTargetInfo, uint16_t nMaxLen_CSR_pse, uint8_t* pCSR_pse, uint16_t* pnTotalLen_CSR_pse, sgx_report_t* pREPORT, pairing_blob_t* pPairingBlob)
{
	sgx_status_t status;
	ms_ecall_tPrepareForCertificateProvisioning_t ms;
	ms.ms_nonce64 = nonce64;
	ms.ms_pTargetInfo = (sgx_target_info_t*)pTargetInfo;
	ms.ms_nMaxLen_CSR_pse = nMaxLen_CSR_pse;
	ms.ms_pCSR_pse = pCSR_pse;
	ms.ms_pnTotalLen_CSR_pse = pnTotalLen_CSR_pse;
	ms.ms_pREPORT = pREPORT;
	ms.ms_pPairingBlob = pPairingBlob;
	status = sgx_ecall(eid, 0, &ocall_table_pse_pr, &ms);
	if (status == SGX_SUCCESS && retval) *retval = ms.ms_retval;
	return status;
}

sgx_status_t ecall_tGenM7(sgx_enclave_id_t eid, ae_error_t* retval, const SIGMA_S1_MESSAGE* pS1, const EPID11_SIG_RL* pSigRL, const uint8_t* pOcspResp, uint32_t nTotalLen_OcspResp, const uint8_t* pVerifierCert, uint32_t nTotalLen_VerifierCert, const pairing_blob_t* pPairingBlob, uint32_t nMaxLen_S2, SIGMA_S2_MESSAGE* pS2, uint32_t* pnTotalLen_S2)
{
	sgx_status_t status;
	ms_ecall_tGenM7_t ms;
	ms.ms_pS1 = (SIGMA_S1_MESSAGE*)pS1;
	ms.ms_pSigRL = (EPID11_SIG_RL*)pSigRL;
	ms.ms_pOcspResp = (uint8_t*)pOcspResp;
	ms.ms_nTotalLen_OcspResp = nTotalLen_OcspResp;
	ms.ms_pVerifierCert = (uint8_t*)pVerifierCert;
	ms.ms_nTotalLen_VerifierCert = nTotalLen_VerifierCert;
	ms.ms_pPairingBlob = (pairing_blob_t*)pPairingBlob;
	ms.ms_nMaxLen_S2 = nMaxLen_S2;
	ms.ms_pS2 = pS2;
	ms.ms_pnTotalLen_S2 = pnTotalLen_S2;
	status = sgx_ecall(eid, 1, &ocall_table_pse_pr, &ms);
	if (status == SGX_SUCCESS && retval) *retval = ms.ms_retval;
	return status;
}

sgx_status_t ecall_tVerifyM8(sgx_enclave_id_t eid, ae_error_t* retval, const SIGMA_S3_MESSAGE* pS3, uint32_t nTotalLen_S3, const EPID11_PRIV_RL* pPrivRL, pairing_blob_t* pPairingBlob, uint8_t* bNewPairing)
{
	sgx_status_t status;
	ms_ecall_tVerifyM8_t ms;
	ms.ms_pS3 = (SIGMA_S3_MESSAGE*)pS3;
	ms.ms_nTotalLen_S3 = nTotalLen_S3;
	ms.ms_pPrivRL = (EPID11_PRIV_RL*)pPrivRL;
	ms.ms_pPairingBlob = pPairingBlob;
	ms.ms_bNewPairing = bNewPairing;
	status = sgx_ecall(eid, 2, &ocall_table_pse_pr, &ms);
	if (status == SGX_SUCCESS && retval) *retval = ms.ms_retval;
	return status;
}

