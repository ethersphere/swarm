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

#include <string.h>
#include "XmlReaderLibXml2.h"
#include "dbg.h"
#include "jhi.h"
#include "misc.h"

#include <libxml/tree.h>
#include <libxml/xmlschemas.h>
#include <libxml/xpathInternals.h>

extern "C"
{
#include <b64/cdecode.h>//Base64
}

namespace intel_dal {

XmlReaderLibXml2::XmlReaderLibXml2(string schemaString)
	:_doc(nullptr), _xpathCtx(nullptr)
{
	_schemaString = schemaString;
	_loaded = false;
	xmlInitParser();
}

XmlReaderLibXml2::~XmlReaderLibXml2()
{
    Close();
}

bool XmlReaderLibXml2::LoadXml(string filePath)
{
	_filePath = filePath;

	_doc = xmlParseFile(_filePath.c_str());
	if (NULL == _doc)
	{
		TRACE0("failed parse dalp file");
		return false;
	}
	_xpathCtx = xmlXPathNewContext(_doc);
	if (NULL == _xpathCtx)
	{
		TRACE0("failed create context dalp file");
		return false;
	}
	_loaded = true;

	return true;
}

bool XmlReaderLibXml2::GetNodeText(string xpath,string& value)
{
	if (!_loaded)
		return false;

	xmlXPathObjectPtr xpathObj;
	xpathObj = xmlXPathEvalExpression(reinterpret_cast<const xmlChar *>(xpath.c_str()), _xpathCtx);
	if (NULL == xpathObj)
	{
		TRACE0("failed to eval xpath");
		return false;
	}

	bool res = (xpathObj->nodesetval)? xpathObj->nodesetval->nodeNr: 0;
	if (1 != res)
	{
		TRACE0("not exactly one line received");
		xmlXPathFreeObject(xpathObj);
		return false;
	}

	xmlChar* data = xmlNodeGetContent(xpathObj->nodesetval->nodeTab[0]);
	if (NULL == data)
	{
		TRACE0("no datat received");
		xmlXPathFreeObject(xpathObj);
		return false;
	}

	value = reinterpret_cast<const char *>(data);
	xmlFree(data);

	xmlXPathFreeObject(xpathObj);
	return true;
}

bool XmlReaderLibXml2::GetNodeTextAsBase64(string xpath,uint8_t** value,long* blobSize)
{
	if (!_loaded)
		return false;

	xmlXPathObjectPtr xpathObj;
	xpathObj = xmlXPathEvalExpression(reinterpret_cast<const xmlChar *>(xpath.c_str()), _xpathCtx);
	if (NULL == xpathObj)
	{
		TRACE0("failed to eval xpath");
		return false;
	}

	bool res = (xpathObj->nodesetval)? xpathObj->nodesetval->nodeNr: 0;
	if (1 != res)
	{
		TRACE0("not exactly one line received");
		xmlXPathFreeObject(xpathObj);
		return false;
	}

	xmlChar* data = xmlNodeGetContent(xpathObj->nodesetval->nodeTab[0]);
	if (NULL == data)
	{
		TRACE0("no data received");
		xmlXPathFreeObject(xpathObj);
		return false;
	}

	size_t size = strlen(reinterpret_cast<const char*>(data));
	if (size < 1 || size >= MAX_APPLET_BLOB_SIZE) // a pack file size cannot be more than JHI_BUFFER_MAX (2MB)
	{
		TRACE0("size is wrong");
		xmlFree(data);
		xmlXPathFreeObject(xpathObj);
		return false;
	}

	*value = (uint8_t*)JHI_ALLOC(size);
	if (NULL == *value)
	{
		TRACE0("memory allocation failure");
		xmlFree(data);
		xmlXPathFreeObject(xpathObj);
		return false;
	}

	base64_decodestate state;
	base64_init_decodestate(&state);
	*blobSize = base64_decode_block(reinterpret_cast<const char*>(data), size, reinterpret_cast<char*>(*value), &state);
	base64_init_decodestate(&state);

	xmlFree(data);
	xmlXPathFreeObject(xpathObj);
	return true;
}

int  XmlReaderLibXml2::GetNodeCount(string xpath)
{
	if (!_loaded)
	{
		return -1;
	}

	xmlXPathObjectPtr xpathObj;
	xpathObj = xmlXPathEvalExpression(reinterpret_cast<const xmlChar *>(xpath.c_str()), _xpathCtx);
	if (NULL == xpathObj)
	{
		TRACE0("failed to eval xpath");
		return -1;
	}

	int res = (xpathObj->nodesetval)? xpathObj->nodesetval->nodeNr: 0;
	xmlXPathFreeObject(xpathObj);
	return res;
}

bool XmlReaderLibXml2::Validate()
{
	if (!_loaded)
		return false;

	xmlDocPtr doc;
	doc = xmlReadMemory(_schemaString.c_str(), _schemaString.length(), "noname.xml", NULL, XML_PARSE_NONET);
	if (NULL == doc)
	{
		TRACE0("failed to load dalp schema");
		return false;
	}

	xmlSchemaParserCtxtPtr parser_ctxt = xmlSchemaNewDocParserCtxt(doc);
	if (NULL == parser_ctxt)
	{
		TRACE0("failed to init dalp schema");
		xmlFreeDoc(doc);
		return false;
	}

	xmlSchemaPtr schema = xmlSchemaParse(parser_ctxt);
	if (NULL == schema)
	{
		TRACE0("failed to parse dalp schema");
		xmlSchemaFreeParserCtxt(parser_ctxt);
		xmlFreeDoc(doc);
		return false;
	}

	xmlSchemaValidCtxtPtr valid_ctxt = xmlSchemaNewValidCtxt(schema);
	if (NULL == valid_ctxt)
	{
		TRACE0("failed to context dalp schema");
		xmlSchemaFree(schema);
		xmlSchemaFreeParserCtxt(parser_ctxt);
		xmlFreeDoc(doc);
		return false;
	}

	bool is_valid = (xmlSchemaValidateDoc(valid_ctxt, _doc) == 0);
	xmlSchemaFreeValidCtxt(valid_ctxt);
	xmlSchemaFree(schema);
	xmlSchemaFreeParserCtxt(parser_ctxt);
	xmlFreeDoc(doc);
	return is_valid;
}

void XmlReaderLibXml2::Close()
{
	_loaded = false;
	if(_xpathCtx)
		xmlXPathFreeContext(_xpathCtx);
	if(_doc)
		xmlFreeDoc(_doc);
}

}//namespace intel_dal
