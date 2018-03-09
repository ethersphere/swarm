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

/**                                                                            
********************************************************************************
**
**    @file IFirmwareInfo.h
**
**    @brief  Contains interface for retrieving information from FW
**
**    @author Elad Dabool
**
********************************************************************************
*/
#ifndef _IFIRMWARE_INFO_H_
#define _IFIRMWARE_INFO_H_

#include "typedefs.h"
#include "jhi_version.h"

namespace intel_dal
{

	typedef union _ME_PLATFORM_TYPE
	{
	   UINT32    Data;
	   struct
	   {
		  UINT32   Mobile:   1;
		  UINT32   Desktop:  1; 
		  UINT32   Server:   1;
		  UINT32   WorkStn:  1;
		  UINT32   Corporate:1;
		  UINT32   Consumer: 1;
		  UINT32   SuperSKU: 1;
		  UINT32   IS_SEC:	 1;
		  UINT32   ImageType:4;
		  UINT32   Brand:    4;
		  UINT32   CpuType: 4;
		  UINT32   Chipset: 4;
		  UINT32   CpuBrandClass:    4;
		  UINT32   PchNetInfraFuses :3;
		  UINT32   Rsvd1:  1;
	   }Fields;
	}ME_PLATFORM_TYPE;

	class IFirmwareInfo
	{
	public:
	   virtual bool Connect() = 0;
	   virtual bool Disconnect() = 0;
	   virtual bool GetFwVersion(VERSION* fw_version) = 0;
	   virtual ~IFirmwareInfo() { }
	};
}
#endif