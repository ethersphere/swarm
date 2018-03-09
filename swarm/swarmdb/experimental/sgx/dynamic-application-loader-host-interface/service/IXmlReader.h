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
**    @file IXmlReader.h
**
**    @brief  Contains interface for reading an XML file
**
**    @author Elad Dabool
**
********************************************************************************
*/
#ifndef _IXMLREADER_H_
#define _IXMLREADER_H_

#include <string>
#include "typedefs.h"
#include "jhi_i.h"

namespace intel_dal
{
	using std::string;

	class IXmlReader
	{
	public:
	   virtual bool LoadXml(FILESTRING filePath) = 0;
	   virtual bool GetNodeText(string xpath,string& value) = 0;
	   virtual bool GetNodeTextAsBase64(string xpath,uint8_t** value,long* blobSize) = 0;
	   virtual int  GetNodeCount(string xpath) = 0;
	   virtual bool Validate() = 0;
	   virtual void Close() = 0;

	   virtual ~IXmlReader() { }
	};

}
#endif