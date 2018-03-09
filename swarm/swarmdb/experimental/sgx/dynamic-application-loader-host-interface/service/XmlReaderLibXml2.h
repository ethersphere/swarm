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
**    @file XmlReaderLibXml2.h
**
**    @brief  Contains libxml2 implemetation for reading an XML file
**
**    @author Alexander Usyskin
**
********************************************************************************
*/
#ifndef _XML_READER_LIBXML2_H_
#define _XML_READER_LIBXML2_H_

#include "IXmlReader.h"
#include "typedefs.h"

#include <libxml/parser.h>
#include <libxml/xpath.h>

namespace intel_dal
{
	using std::string;

#define MAX_APPLET_BLOB_SIZE  2097152 // applet blob cannot be more then 2MB

	class XmlReaderLibXml2 : public IXmlReader
	{
	private:
		string _filePath;
		string _schemaString;

		xmlDocPtr _doc;
		xmlXPathContextPtr _xpathCtx;

		bool _loaded;

	public:
		XmlReaderLibXml2(string schemaString);
        ~XmlReaderLibXml2();
		bool LoadXml(string filePath);
		bool GetNodeText(string xpath, string& value);
		bool GetNodeTextAsBase64(string xpath, uint8_t** value, long* blobSize);
		int  GetNodeCount(string xpath);
		bool Validate();
		void Close();
	};

}
#endif//_XML_READER_LIBXML2_H_
