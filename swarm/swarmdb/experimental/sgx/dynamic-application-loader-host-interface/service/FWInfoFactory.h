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
**    @file FWInfoFactory.h
**
**    @brief  Contains factory design pattern that creates IFirmwareInfo instances
**
**    @author Elad Dabool
**
********************************************************************************
*/
#ifndef _FW_INFO_FACTORY_H_
#define _FW_INFO_FACTORY_H_

#include "IFirmwareInfo.h"
#include "misc.h"
#include "GlobalsManager.h"

#ifdef _WIN32

#include "FWInfoWin32Sockets.h"
#include "FWInfoWin32.h"

#endif

#ifdef __linux__
#include "FWInfoLinuxSockets.h"
#include "FWInfoLinux.h"
#endif

namespace intel_dal
{
	class FWInfoFactory
	{
	public:
		static IFirmwareInfo* createInstance()
		{
			TEE_TRANSPORT_TYPE transportType;
			transportType =	GlobalsManager::Instance().getTransportType();
#ifdef _WIN32
			if (transportType == TEE_TRANSPORT_TYPE_SOCKET)
			{
				return JHI_ALLOC_T(FWInfoWin32Sockets); // is there a problem with inheritance?
			}
			else
			{
				return JHI_ALLOC_T(FWInfoWin32);
			}
#else //!WIN32
	        if (transportType == TEE_TRANSPORT_TYPE_SOCKET)
			{
				return JHI_ALLOC_T(FWInfoLinuxSockets);
			}
			else
			{
				return JHI_ALLOC_T(FWInfoLinux);
			}
#endif //WIN32

		}
	};

}

#endif
