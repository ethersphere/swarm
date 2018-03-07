/*
   Copyright 2010-2016 Intel Corporation

   This software is licensed to you in accordance
   with the agreement between you and Intel Corporation.

   Alternatively, you can use this file in compliance
   with the Apache license, Version 2.


   Apache License, Version 2.0

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

#include "DLL_Loader.h"
#include "dbg.h"
#include "GlobalsManager.h"
#include <cstdlib>
#include <vector>
#ifdef _WIN32
// For signature checking
#include <windows.h>
#include <Softpub.h>
#include <wincrypt.h>
#include <wintrust.h>
#define ENCODING				(X509_ASN_ENCODING | PKCS_7_ASN_ENCODING)
#else
#include <dlfcn.h>
#endif // _WIN32
#include "string_s.h"

namespace intel_dal
{

#ifdef _WIN32
	//-------------------------------------------------------------------
	// Copyright (C) Microsoft.  All rights reserved.
	// Example of verifying the embedded signature of a PE file by using 
	// the WinVerifyTrust function.
	// Code taken from: 
	// http://msdn.microsoft.com/en-us/library/aa382384(v=vs.85).aspx
	bool DLL_Loader::VerifyFileSignature(FILESTRING dllFullPath)
	{
		LONG				lStatus;
		DWORD				dwLastError;
		WINTRUST_FILE_INFO	FileData;
		GUID				WVTPolicyGUID = WINTRUST_ACTION_GENERIC_VERIFY_V2;
		WINTRUST_DATA		WinTrustData;
		bool				bResult = false;

		T_TRACE1(_T("VerifyFileSignature: On file %s\n"), dllFullPath.c_str());

		// Initialize the WINTRUST_FILE_INFO structure.
		memset(&FileData, 0, sizeof(FileData));
		FileData.cbStruct = sizeof(WINTRUST_FILE_INFO);
		FileData.pcwszFilePath = dllFullPath.c_str();
		FileData.hFile = NULL;
		FileData.pgKnownSubject = NULL;

		/*
		WVTPolicyGUID specifies the policy to apply on the file
		WINTRUST_ACTION_GENERIC_VERIFY_V2 policy checks:

		1) The certificate used to sign the file chains up to a root 
		certificate located in the trusted root certificate store. This 
		implies that the identity of the publisher has been verified by 
		a certification authority.

		2) In cases where user interface is displayed (which this example
		does not do), WinVerifyTrust will check for whether the  
		end entity certificate is stored in the trusted publisher store,  
		implying that the user trusts content from this publisher.

		3) The end entity certificate has sufficient permission to sign 
		code, as indicated by the presence of a code signing EKU or no 
		EKU.
		*/

		// Initialize the WinVerifyTrust input data structure.
		// Default all fields to 0.
		memset(&WinTrustData, 0, sizeof(WinTrustData));

		WinTrustData.cbStruct = sizeof(WinTrustData);

		// Use default code signing EKU.
		WinTrustData.pPolicyCallbackData = NULL;

		// No data to pass to SIP.
		WinTrustData.pSIPClientData = NULL;

		// Disable WVT UI.
		WinTrustData.dwUIChoice = WTD_UI_NONE;

		// No revocation checking.
		WinTrustData.fdwRevocationChecks = WTD_REVOKE_NONE; 
		WinTrustData.dwProvFlags = WTD_CACHE_ONLY_URL_RETRIEVAL;

		// Verify an embedded signature on a file.
		WinTrustData.dwUnionChoice = WTD_CHOICE_FILE;

		// Default verification.
		WinTrustData.dwStateAction = 0;

		// Not applicable for default verification of embedded signature.
		WinTrustData.hWVTStateData = NULL;

		// Not used.
		WinTrustData.pwszURLReference = NULL;

		// This is not applicable if there is no UI because it changes 
		// the UI to accommodate running applications instead of 
		// installing applications.
		WinTrustData.dwUIContext = 0;

		// Set pFile.
		WinTrustData.pFile = &FileData;

		// WinVerifyTrust verifies signatures as specified by the GUID 
		// and Wintrust_Data.
		lStatus = WinVerifyTrust(
			NULL,
			&WVTPolicyGUID,
			&WinTrustData);

		switch (lStatus) 
		{
		case ERROR_SUCCESS:
			/*
			Signed file:
			- Hash that represents the subject is trusted.

			- Trusted publisher without any verification errors.

			- UI was disabled in dwUIChoice. No publisher or 
			time stamp chain errors.

			- UI was enabled in dwUIChoice and the user clicked 
			"Yes" when asked to install and run the signed 
			subject.
			*/
			TRACE0("VerifyFileSignature: File is signed, signature valid\n");
			bResult = true;
			break;

		case TRUST_E_NOSIGNATURE:
			// The file was not signed or had a signature 
			// that was not valid.

			// Get the reason for no signature.
			dwLastError = GetLastError();
			if (TRUST_E_NOSIGNATURE == dwLastError ||
				TRUST_E_SUBJECT_FORM_UNKNOWN == dwLastError ||
				TRUST_E_PROVIDER_UNKNOWN == dwLastError) 
			{
				// The file was not signed.
				TRACE0("VerifyFileSignature: File not signed \n");
			} 
			else 
			{
				// The signature was not valid or there was an error 
				// opening the file.
				TRACE0("VerifyFileSignature: Unknown error verifying file\n");
			}
			break;

		case TRUST_E_EXPLICIT_DISTRUST:
			// The hash that represents the subject or the publisher 
			// is not allowed by the admin or user.
			TRACE0("VerifyFileSignature: Signature present, disallowed \n");
			break;

		case TRUST_E_SUBJECT_NOT_TRUSTED:
			// The user clicked "No" when asked to install and run.
			TRACE0("VerifyFileSignature: Signature present, not trusted \n");
			break;

		case CRYPT_E_SECURITY_SETTINGS:
			/*
			The hash that represents the subject or the publisher 
			was not explicitly trusted by the admin and the 
			admin policy has disabled user trust. No signature, 
			publisher or time stamp errors.
			*/
			TRACE0("VerifyFileSignature: Admin policy error \n");
			break;

		default:
			// The UI was disabled in dwUIChoice or the admin policy 
			// has disabled user trust. lStatus contains the 
			// publisher or time stamp chain error.
			TRACE0("VerifyFileSignature: Unknown error\n");
			break;
		}

		return bResult;
	} // VerifyFileSignature

	//-------------------------------------------------------------------
	// The following function is based on sample code from:
	// http://support.microsoft.com/kb/323809
	bool DLL_Loader::VerifyFilePublisher(const FILECHAR* szDllFullFileName, LPTSTR *subjectFound)
	{
		HCERTSTORE hStore = NULL;
		HCRYPTMSG hMsg = NULL; 
		PCCERT_CONTEXT pCertContext = NULL;
		BOOL fResult;   
		DWORD dwEncoding, dwContentType, dwFormatType;
		PCMSG_SIGNER_INFO pSignerInfo = NULL;
		DWORD dwSignerInfo;
		CERT_INFO CertInfo;     
		LPTSTR szName = NULL;
		DWORD dwData;
		bool bResult = false;

		TRACE0("VerifyFilePublisher: Starting... \n");

		if (!szDllFullFileName)
			return bResult;

		T_TRACE1(_T("VerifyFilePublisher: On %s \n"), szDllFullFileName);

		__try
		{
			// Get message handle and store handle from the signed file.
			fResult = CryptQueryObject(CERT_QUERY_OBJECT_FILE,
				szDllFullFileName,
				CERT_QUERY_CONTENT_FLAG_PKCS7_SIGNED_EMBED,
				CERT_QUERY_FORMAT_FLAG_BINARY,
				0,
				&dwEncoding,
				&dwContentType,
				&dwFormatType,
				&hStore,
				&hMsg,
				NULL);
			if (!fResult)
			{
				TRACE0("VerifyFilePublisher: CryptQueryObject failed \n");
				__leave;
			}

			// Get signer information size.
			fResult = CryptMsgGetParam(hMsg, 
				CMSG_SIGNER_INFO_PARAM, 
				0, 
				NULL, 
				&dwSignerInfo);
			if (!fResult)
			{
				TRACE0("VerifyFilePublisher: CryptMsgGetParam failed \n");
				__leave;
			}
			// Allocate memory for signer information.
			pSignerInfo = (PCMSG_SIGNER_INFO)LocalAlloc(LPTR, dwSignerInfo);
			if (!pSignerInfo)
			{
				TRACE0("VerifyFilePublisher: Mem alloc for SignerInfo failed \n");
				__leave;
			}
			// Get Signer Information.
			fResult = CryptMsgGetParam(hMsg, 
				CMSG_SIGNER_INFO_PARAM, 
				0, 
				(PVOID)pSignerInfo, 
				&dwSignerInfo);
			if (!fResult)
			{
				TRACE0("VerifyFilePublisher: CryptMsgGetParam failed \n");
				__leave;
			}

			// Search for the signer certificate in the temporary 
			// certificate store.
			CertInfo.Issuer = pSignerInfo->Issuer;
			CertInfo.SerialNumber = pSignerInfo->SerialNumber;

			pCertContext = CertFindCertificateInStore(hStore,
				ENCODING,
				0,
				CERT_FIND_SUBJECT_CERT,
				(PVOID)&CertInfo,
				NULL);
			if (!pCertContext)
			{
				TRACE0("VerifyFilePublisher: CertFindCertificateInStore failed \n");
				__leave;
			}

			// Verify publisher (Subject Name in cert)
			// Get Subject name size.
			if (!(dwData = CertGetNameString(pCertContext, 
				CERT_NAME_SIMPLE_DISPLAY_TYPE,
				0,
				NULL,
				NULL,
				0)))
			{
				TRACE0("VerifyFilePublisher: CertGetNameString failed \n");
				__leave;
			}
			// Allocate memory for subject name.
			szName = (LPTSTR)LocalAlloc(LPTR, dwData * sizeof(TCHAR));
			if (!szName)
			{
				TRACE0("VerifyFilePublisher: Mem alloc for subject name failed \n");
				__leave;
			}
			// Get subject name.
			if (!(CertGetNameString(pCertContext, 
				CERT_NAME_SIMPLE_DISPLAY_TYPE,
				0,
				NULL,
				szName,
				dwData)))
			{
				TRACE0("VerifyFilePublisher: CertGetNameString failed \n");
				__leave;
			}

			T_TRACE1(_T("VerifyFilePublisher: Subject Name in cert is: %s \n"),
				szName);

			bResult = true;
		}
		__finally
		{
			if (pSignerInfo != NULL) 
				LocalFree(pSignerInfo);
			if (pCertContext != NULL) 
				CertFreeCertificateContext(pCertContext);
			if (hStore != NULL) 
				CertCloseStore(hStore, 0);
			if (hMsg != NULL) 
				CryptMsgClose(hMsg);
			*subjectFound = szName;
		}
		return bResult;
	} // VerifyFilePublisher
#endif //WIN32

	JHI_RET DLL_Loader::UnloadDll(HMODULE* loadedModule)
	{
		int error;
#ifdef _WIN32
		if (FALSE == FreeLibrary(*loadedModule))
		{
			error = GetLastError();
			TRACE1("Unable to unload module, error %d", error);
		}
#else
		error = dlclose(loadedModule);
		if (error)
		{
			TRACE1("Unable to unload module, error %d", error);
		}
#endif // _WIN32
		return JHI_SUCCESS;
	}


	JHI_RET DLL_Loader::LoadDll(FILESTRING path, FILESTRING dll_file_name, FILESTRING wsVendorName, bool verifySignatures, HMODULE* loadedModule)
	{
#ifdef _WIN32
		if ((path.length() > 0) && ( (path[path.length()-1] == FILEPREFIX('\\')) || ( path[path.length()-1] == FILEPREFIX('/') ) ) )
		{
			return LoadDll(path + dll_file_name, wsVendorName, verifySignatures, loadedModule);
		}
		else
		{
			return LoadDll(path + FILEPREFIX("/") + dll_file_name, wsVendorName, verifySignatures, loadedModule);
		}
#else
		return LoadDll(dll_file_name, wsVendorName, verifySignatures, loadedModule);
#endif
	}


	bool DLL_Loader::VerifyFile(FILESTRING dllFullPath, FILESTRING wsVendorName)
	{
#ifdef _WIN32
		// Check the signature on the DLL
		if (!VerifyFileSignature(dllFullPath))
		{
			// Signature not verified
			TRACE0 ("DLL signature NOT OK \n");
			return false;
		}
		// Signature OK
		TRACE0 ("DLL signature OK \n");
		LPTSTR subjectFound = NULL;
		// Verify the publisher (OEM)
		if (!VerifyFilePublisher(dllFullPath.c_str(), &subjectFound))
		{
			// Verification failed
			TRACE0 ("DLL publisher NOT OK \n");
			return false;
		}

		if (subjectFound == NULL) 
		{
			TRACE0("VerifyFilePublisher: Subject name does not match OEM \n");
			return false;
		}
		FILESTRING subjectString = FILESTRING(subjectFound);

		LocalFree(subjectFound);
		size_t foundLocation = subjectString.find(wsVendorName);
		if (foundLocation != FILESTRING::npos)
		{
			TRACE0("VerifyFilePublisher: Subject name matches OEM \n");
			// Verified OK
			TRACE0 ("DLL publisher OK \n");
			return true;
		}
		//else
		TRACE0("VerifyFilePublisher: Subject name does not match OEM \n");
		return false;
#else
		return true;
#endif // _WIN32
	}
	//-------------------------------------------------------------------
	// Main function 
	JHI_RET DLL_Loader::LoadDll(FILESTRING dllFullPath, FILESTRING wsVendorName, bool verifySignatures, HMODULE* loadedModule)
	{
		HMODULE g_hAppDll;
		JHI_RET  retCode = JHI_INTERNAL_ERROR;

		if (loadedModule == NULL)
		{
			goto LOADDLL_EXIT;
		}
#ifdef _WIN32
		// Check to see if the DLL is present in this path
		if (_waccess_s(dllFullPath.c_str(),0) != 0)
		{
			TRACE0("GetDllPath: Filename does not exist \n");
			retCode = JHI_VM_DLL_FILE_NOT_FOUND;
			goto LOADDLL_EXIT;
		}

		// DLL exists in current directory
		TRACE0 ("DLL exists in current directory\n");

		if(verifySignatures)
		{
			if (!VerifyFile(dllFullPath.c_str(), wsVendorName))
			{
				// Signature not verified
				TRACE0 ("DLL verify failed!\n");
				retCode = JHI_VM_DLL_VERIFY_FAILED;
				goto LOADDLL_EXIT;
			}
		}

		// Load the DLL now
		g_hAppDll = LoadLibrary(dllFullPath.c_str());
#else
		g_hAppDll = dlopen(dllFullPath.c_str(), RTLD_LAZY);
#endif //_WIN32
		if (!g_hAppDll)
		{	
			// DLL Load failed
			TRACE0 ("DLL load failed\n");
			retCode = JHI_INTERNAL_ERROR; // this shoudn't happen
			goto LOADDLL_EXIT;
		}
		// DLL Load OK
		TRACE0("DLL load OK\n");
		*loadedModule = g_hAppDll;
		retCode = JHI_SUCCESS;

LOADDLL_EXIT:

		return retCode;
	} // LoadDll
}