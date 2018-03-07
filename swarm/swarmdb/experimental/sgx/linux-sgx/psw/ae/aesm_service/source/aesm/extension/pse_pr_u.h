#ifndef PSE_PR_U_H__
#define PSE_PR_U_H__

#include <stdint.h>
#include <wchar.h>
#include <stddef.h>
#include <string.h>
#include "sgx_edger8r.h" /* for sgx_satus_t etc. */

#include "aeerror.h"
#include "Epid11_rl.h"
#include "sgx_report.h"
#include "pairing_blob.h"
#include "pse_pr_sigma_1_1_defs.h"

#include <stdlib.h> /* for size_t */

#define SGX_CAST(type, item) ((type)(item))

#ifdef __cplusplus
extern "C" {
#endif


sgx_status_t ecall_tPrepareForCertificateProvisioning(sgx_enclave_id_t eid, ae_error_t* retval, uint64_t nonce64, const sgx_target_info_t* pTargetInfo, uint16_t nMaxLen_CSR_pse, uint8_t* pCSR_pse, uint16_t* pnTotalLen_CSR_pse, sgx_report_t* pREPORT, pairing_blob_t* pPairingBlob);
sgx_status_t ecall_tGenM7(sgx_enclave_id_t eid, ae_error_t* retval, const SIGMA_S1_MESSAGE* pS1, const EPID11_SIG_RL* pSigRL, const uint8_t* pOcspResp, uint32_t nTotalLen_OcspResp, const uint8_t* pVerifierCert, uint32_t nTotalLen_VerifierCert, const pairing_blob_t* pPairingBlob, uint32_t nMaxLen_S2, SIGMA_S2_MESSAGE* pS2, uint32_t* pnTotalLen_S2);
sgx_status_t ecall_tVerifyM8(sgx_enclave_id_t eid, ae_error_t* retval, const SIGMA_S3_MESSAGE* pS3, uint32_t nTotalLen_S3, const EPID11_PRIV_RL* pPrivRL, pairing_blob_t* pPairingBlob, uint8_t* bNewPairing);

#ifdef __cplusplus
}
#endif /* __cplusplus */

#endif
