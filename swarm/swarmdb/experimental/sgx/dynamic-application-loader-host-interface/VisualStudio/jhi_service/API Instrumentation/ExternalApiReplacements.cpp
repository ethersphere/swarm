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

#ifdef EXTERNAL_API_INSTRUMENTATION
#include "ExternalApiReplacements.h"

#include <fstream>
#include <sstream>
#include <tchar.h>
#include "dbg.h"
#include "reg.h"

int instrumentationCommand = 0;
void readCommandFromFile();
bool isCommandRead = false;

void init()
{
	if (isCommandRead)
	{
		return;
	}
	readCommandFromFile();
}

//getaddrinfo API instrumentation
INT
WSAAPI
getaddrinfo_instrumentation(
    __in_opt        PCSTR               pNodeName,
    __in_opt        PCSTR               pServiceName,
    __in_opt        const ADDRINFOA *   pHints,
    __deref_out     PADDRINFOA *        ppResult
    )
{
	init();
	if (instrumentationCommand == 1)
	{
		return EAI_NONAME;
	}
	if (instrumentationCommand == 2)
	{
		return EAI_BADFLAGS;
	}
	return getaddrinfo(pNodeName, pServiceName, pHints, ppResult);
}

//socket API instrumentation
__checkReturn
	SOCKET
	WSAAPI
	socket_instrumentation(
	__in int af,
	__in int type,
	__in int protocol
	)
{
	init();
	if (instrumentationCommand == 3)
	{
		return INVALID_SOCKET;
	}
	return socket(af, type, protocol);
}

int
	WSAAPI
	bind_instrumentation(
	__in SOCKET s,
	__in_bcount(namelen) const struct sockaddr FAR * name,
	__in int namelen
	)
{
	init();
	if (instrumentationCommand == 4)
	{
		return SOCKET_ERROR;
	}
	return bind(s, name, namelen);
}

void readCommandFromFile()
{
	isCommandRead = true; // attempting only once.
	LPWSTR filePath = new WCHAR[200];
	ZeroMemory(filePath, 200);
	LPWSTR fileName = L"/API Instrumentation.txt";

		//Read jhi service file location
	FILECHAR   jhiFileLocation[FILENAME_MAX+1]={0};
		
	if( JHI_SUCCESS != JhiQueryServiceFileLocationFromRegistry(
		jhiFileLocation,
		(FILENAME_MAX-1) * sizeof(FILECHAR)))
	{
		TRACE0( "unable to query file location from registry") ;
	}

	_tcscat(jhiFileLocation, fileName);

	//verify the file exist
	if (_waccess_s(jhiFileLocation,0) != 0)
	{
		TRACE0("Getting command file failed - the file (%S) does not exist", jhiFileLocation);
		return;
	}
	std::fstream infile(jhiFileLocation, std::ios_base::in);
	//std::ifstream infile(filePath);
	if ( ! infile ) 
	{
		TRACE0("Can't open the file named (%S).", jhiFileLocation);
	}
	TRACE0("Instrumentation command file loaded - (%S)", jhiFileLocation);
	int command;
	infile >> command;

	instrumentationCommand = command;
}


#endif //EXTERNAL_API_INSTRUMENTATION