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
**    @file CommandDispatcher.h
**
**    @brief  Contains implementation for commands dispatcher interface
**
**    @author Elad Dabool
**
********************************************************************************
*/
#ifndef _COMMANDDISPATCHER_H_
#define _COMMANDDISPATCHER_H_

#include "typedefs.h"
#include "CSTypedefs.h"
#include "ICommandDispatcher.h"
#include "Locker.h"
#include "jhi_service.h"

#include "GlobalsManager.h"
#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
#include "jhi_sdk.h"
#endif

namespace intel_dal
{
	
	class CommandDispatcher : public ICommandDispatcher
	{
	private:
		Locker  _jhiMutex;

		bool convertAppIDtoUpperCase(const char *pAppId,UINT8 ucConvertedAppId[LEN_APP_ID+1]);
		int verifyAppID(char *pAppId);

	public:
		CommandDispatcher();
		bool init();
		bool deinit();
		void processCommand(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize);

		void InvokeInit(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize);
		void InvokeGetVersionInfo(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize);		
		void InvokeInstall(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize);
		void InvokeUninstall(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize);
		void InvokeGetSessionsCount(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize);
		void InvokeCreateSession(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize);
		void InvokeCloseSession(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize);
		void InvokeSetSessionEventHandler(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize);
		void InvokeGetSessionInfo(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize);
		void InvokeGetSessionEventData(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize);
		void InvokeSendAndRecieve(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize);
		void InvokeGetAppletProperty(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize);
		void InvokeOpenSDSession(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize);
		void InvokeCloseSDSession(const uint8_t* inputData, uint32_t inputSize, uint8_t** outputData,uint32_t* outputSize);
		void InvokeListInstalledTAs(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize);
		void InvokeListInstalledSDs(const uint8_t* inputData, uint32_t inputSize, uint8_t** outputData, uint32_t* outputSize);
		void InvokeSendCmdPkg(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize);
		void InvokeQueryTeeMetadata(const uint8_t* inputData, uint32_t inputSize, uint8_t** outputData, uint32_t* outputSize);
#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
		void InvokeGetSessionDataTable(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize);
		void InvokeGetLoadedApplets(const uint8_t* inputData,uint32_t inputSize,uint8_t** outputData,uint32_t* outputSize);
#endif
	};

}

#endif