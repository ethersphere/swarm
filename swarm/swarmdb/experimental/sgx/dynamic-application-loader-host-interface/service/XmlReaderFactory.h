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
**    @file XmlReaderFactory.h
**
**    @brief  Contains factory design pattern that creates IXmlReader instances
**
**    @author Alexander Usyskin
**
********************************************************************************
*/
#ifndef _XML_READER_FACTORY_H_
#define _XML_READER_FACTORY_H_

#include "IXmlReader.h"

#ifdef WIN32
#include "XmlReaderWin32.h"
#else
#include "XmlReaderLibXml2.h"
#endif//WIN32

namespace intel_dal
{
	class XmlReaderFactory
	{
	public:
		static IXmlReader* createInstance(std::string schemaString)
		{

#ifdef WIN32
			return new XmlReaderWin32(schemaString);
#else
			return new XmlReaderLibXml2(schemaString);
#endif
		}
	};
}

#endif//_XML_READER_FACTORY_H_

