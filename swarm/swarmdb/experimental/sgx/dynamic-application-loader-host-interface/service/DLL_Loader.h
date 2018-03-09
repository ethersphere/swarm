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

#ifndef __DLL_VALIDATION_H
#define __DLL_VALIDATION_H

#include <string>

#include "jhi.h"
#include "typedefs_i.h"
#include "jhi_i.h"

namespace intel_dal
{
	class DLL_Loader
	{	
	private:
#ifdef _WIN32
		static bool VerifyFileSignature(FILESTRING dllFullPath);
		static bool VerifyFilePublisher(const FILECHAR* szFileName, LPTSTR* subjectFound);
#endif
		static bool VerifyFile(FILESTRING dllFullPath, FILESTRING wsVendorName);

	public:
		static JHI_RET UnloadDll(HMODULE* loadedModule);
		static JHI_RET LoadDll(FILESTRING path, FILESTRING dll_file_name, FILESTRING wsVendorName, bool verifySignatures, HMODULE* loadedModule);
		static JHI_RET LoadDll(FILESTRING dll_full_path, FILESTRING wsVendorName, bool verifySignatures, HMODULE* loadedModule);
	};

}


#endif 

