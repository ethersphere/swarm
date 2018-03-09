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

#ifndef __REG_H__
#define __REG_H__

#include "typedefs.h"
//#include "misc.h"
#include "dbg.h"
//#include "jhi_i.h"

typedef UINT32 JHI_RET_I;

// Prototypes
#ifdef __cplusplus
extern "C" {
#endif

JHI_RET_I 
JhiQueryAppFileLocationFromRegistry(FILECHAR* outBuffer, uint32_t outBufferSize);

#ifndef _WIN32
JHI_RET_I
JhiQueryPluginLocationFromRegistry(FILECHAR* outBuffer, uint32_t outBufferSize);

JHI_RET_I
JhiQuerySpoolerLocationFromRegistry(FILECHAR* outBuffer, uint32_t outBufferSize);

JHI_RET_I
JhiQueryEventSocketsLocationFromRegistry(FILECHAR* outBuffer, uint32_t outBufferSize);

JHI_RET_I
RestartJhiService ();
#endif //!_WIN32

JHI_RET_I
JhiQueryServiceFileLocationFromRegistry(FILECHAR* outBuffer, uint32_t outBufferSize);

JHI_RET_I
JhiQueryServicePortFromRegistry(uint32_t* portNumber);

JHI_RET_I
JhiQueryAddressTypeFromRegistry(uint32_t* addressType);

JHI_RET_I
JhiQueryTransportTypeFromRegistry(uint32_t* transportType);

JHI_RET_I
JhiQuerySocketIpAddressFromRegistry(FILECHAR *ip);

JHI_RET_I
JhiQueryLogLevelFromRegistry(JHI_LOG_LEVEL *loglevel);

#ifdef __linux__
JHI_RET_I
JhiQueryDaemonSocketPathFromRegistry(char * path);
#endif

JHI_RET_I
JhiWritePortNumberToRegistry(uint32_t portNumber);

JHI_RET_I
JhiWriteAddressTypeToRegistry(uint32_t addressType);

#ifdef __cplusplus
}
#endif

#endif