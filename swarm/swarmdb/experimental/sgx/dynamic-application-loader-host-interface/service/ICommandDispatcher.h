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
**    @file ICommandDispatcher.h
**
**    @brief  Contains interface for commands dispatcher
**
**    @author Elad Dabool
**
********************************************************************************
*/
#ifndef _ICOMMANDDISPATCHER_H_
#define _ICOMMANDDISPATCHER_H_

#include "typedefs.h"
#include "jhi_i.h"

namespace intel_dal
{
	class ICommandDispatcher
	{	
	public:
	   virtual bool init() = 0;
	   virtual bool deinit() = 0;
	   virtual void processCommand(IN const uint8_t* inputData,IN uint32_t inputSize,OUT uint8_t** outputData,OUT uint32_t* outputSize) = 0;
	   virtual ~ICommandDispatcher() { }
	};
}
#endif