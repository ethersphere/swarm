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

#include "XmlReaderWin32.h"
#include "dbg.h"
#include "jhi.h"
#include "misc.h"

namespace intel_dal
{

		XmlReaderWin32::XmlReaderWin32(string schemaString)  :
		docPtr(__uuidof(MSXML2::DOMDocument60), NULL, CLSCTX_INPROC_SERVER),
		schemaPtr(__uuidof(MSXML2::XMLSchemaCache60), NULL, CLSCTX_INPROC_SERVER),
		schemaXSD(__uuidof(MSXML2::DOMDocument60), NULL, CLSCTX_INPROC_SERVER)
	{
		_schemaString = schemaString;
		loaded = false;
		docPtr->setProperty(_TEXT("MaxElementDepth"),MAX_ELEMENT_DEPTH); // for security reasons, limit the max element depth
		docPtr->setProperty(_TEXT("MaxXMLSize"),MAX_XML_FILE_SIZE);	     // for security reasons, limit the size of file can read
		docPtr->setProperty(_TEXT("NewParser"),VARIANT_TRUE);			 // enable NewParser
	}


	bool XmlReaderWin32::LoadXml(wstring filePath)
	{
		VARIANT_BOOL ret;
		_filePath = filePath;
		ret = docPtr->load(_filePath.c_str());

		if (ret == VARIANT_FALSE)
		{
			return false;
		}

		loaded = true;
		return true;
	}

	bool XmlReaderWin32::GetNodeText(string xpath,string& value)
	{
		MSXML2::IXMLDOMNodePtr node = NULL;

		if (!loaded)
			return false;

		_bstr_t path(xpath.c_str());
		node = docPtr->selectSingleNode(path);

		if (node == NULL)
			return false;

		value = node->Gettext();

		return true;
	}

	bool XmlReaderWin32::GetNodeTextAsBase64(string xpath,uint8_t** value,long* blobSize)
	{
		MSXML2::IXMLDOMNodePtr node = NULL;
		long size=0;

		if (!loaded)
			return false;

		_bstr_t path(xpath.c_str());
		node = docPtr->selectSingleNode(path);

		if (node == NULL)
			return false;

		node->put_dataType(L"bin.base64");

		if (FAILED(SafeArrayGetUBound(node->nodeTypedValue.parray, 1, &size)))
		{
			TRACE0("failed reading applet blob from dalp file");
			return false;
		}

		// size of the array is UpperBound - LowerBound + 1 ==> size - 0 + 1
		size++; // we have a max limit of 30MB in the XML file - prevents integer overflow

		if (size > 1 && size <= MAX_APPLET_BLOB_SIZE) // a pack file size cannot be more than JHI_BUFFER_MAX (2MB)
		{
			*value = (uint8_t*) JHI_ALLOC(size);
			if (*value == NULL)
			{
				TRACE0("memory allocation failure");
				return false;
			}

			memcpy_s(*value,size,node->nodeTypedValue.parray->pvData,size);
			//value = (char*) node->nodeTypedValue.parray->pvData;

			*blobSize = size;

			return true;
		}

		return false;
	}

	int  XmlReaderWin32::GetNodeCount(string xpath)
	{
		MSXML2::IXMLDOMNodeListPtr nodeList = NULL;

		if (!loaded)
			return -1;

		_bstr_t path(xpath.c_str());
		nodeList = docPtr->selectNodes(path);

		if (!nodeList)
			return -1;

		return (int) nodeList->Getlength();
	}

	bool XmlReaderWin32::Validate()
	{
		IXMLDOMParseErrorPtr pError;
		long errorcode;

		if (!loaded)
			return false;

		//load the schema file   
		if (schemaXSD->loadXML(_schemaString.c_str()) == VARIANT_FALSE)
		{
			//BSTR errorString;
			//schemaXSD->GetparseError()->get_reason(&errorString);
			TRACE0("failed to load dalp schema");
			return false;
		}

		if (FAILED(schemaPtr->add("urn:dalp",schemaXSD.GetInterfacePtr())))
		{
			TRACE0("failed to load dalp schema");
			return false;
		}

		// Attaching the schema to the XML document.
		docPtr->schemas = schemaPtr.GetInterfacePtr();

		pError = docPtr->validate();

		if (!pError)
			return false;

		pError->get_errorCode(&errorcode);

		if(errorcode == S_OK)
			return true;

		//BSTR errorString;
		//pError->get_reason(&errorString);

		return false;
	}

	void XmlReaderWin32::Close()
	{
		loaded = false;
	}
}