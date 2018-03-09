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
**    @file XmlReaderWin32.h
**
**    @brief  Contains win32 implemetation for reading an XML file
**
**    @author Elad Dabool
**
********************************************************************************
*/
#ifndef _IXMLREADERWIN32_H_
#define _IXMLREADERWIN32_H_

#include "IXmlReader.h"
#include "typedefs.h"
#include <tchar.h>
#include <windows.h>
#include <objbase.h>

#import <msxml6.dll> no_auto_exclude

namespace intel_dal
{
	using std::string;
	using std::wstring;

	#define MAX_ELEMENT_DEPTH 5
	#define MAX_XML_FILE_SIZE 30720 // in kilobytes (30MB)
	
	class XmlReaderWin32 : public IXmlReader
	{
	private:
		wstring _filePath;
		string _schemaString;

		MSXML2::IXMLDOMDocument2Ptr docPtr;
		MSXML2::IXMLDOMSchemaCollectionPtr schemaPtr;
		MSXML2::IXMLDOMDocumentPtr schemaXSD;

		bool loaded;

	public:

		XmlReaderWin32(string schemaString) ;
		bool LoadXml(wstring filePath);
		bool GetNodeText(string xpath,string& value);
		bool GetNodeTextAsBase64(string xpath,uint8_t** value,long* blobSize);
		int  GetNodeCount(string xpath);
		bool Validate();
		void Close();
	};

}
#endif