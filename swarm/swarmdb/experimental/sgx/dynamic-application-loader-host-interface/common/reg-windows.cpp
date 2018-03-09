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
**    @file reg.c
**
**    @brief  Defines registry related functions
**
**    @author Niveditha Sundaram
**
********************************************************************************
*/

#include <stdio.h>
#include "jhi.h"
#include "reg.h"

#ifdef SCHANNEL_OVER_SOCKET // emulation mode

#define REGISTRY_PATH  "Software\\Intel\\Services\\DAL_EMULATION"

#else

#define REGISTRY_PATH  "Software\\Intel\\Services\\DAL"

#endif

#define REGISTRY_BASE  HKEY_LOCAL_MACHINE

// jhi registry keys
#define KEY_JHI_FILES_PATH L"FILELOCALE"
#define KEY_JHI_APPLETS_REPOSITORY_PATH L"APPLETSLOCALE"
#define KEY_JHI_SERVICE_PORT L"JHI_SERVICE_PORT"
#define KEY_JHI_ADDRESS_TYPE L"JHI_ADDRESS_TYPE"
#define KEY_JHI_TRANSPORT_TYPE L"JHI_TRANSPORT_TYPE"
#define KEY_JHI_FW_VERSION L"FW_VERSION"
#define KEY_JHI_LOG_FLAG L"JHI_LOG"


bool readStringFromRegistry(const wchar_t* key, wchar_t* outBuffer, uint32_t outBufferSize)
{
	HKEY hKey;
	DWORD dwType = REG_SZ;
	int maxElementSize = -1;

	if (key == NULL || outBuffer == NULL)
		return false;

	// Check if Module is a valid number
	if ( RegOpenKeyEx( REGISTRY_BASE,
		TEXT(REGISTRY_PATH),
		0,
		KEY_READ | KEY_WOW64_64KEY,
		&hKey) != ERROR_SUCCESS )
	{
		TRACE1( "Unable to open Registry [0x%x]\n", GetLastError());
		return false; 
	}

	// Check for the actual value
	if( RegQueryValueEx(hKey,key,0, &dwType, (LPBYTE)outBuffer, (LPDWORD)&outBufferSize) != ERROR_SUCCESS)
	{
		TRACE1("Registry read failure for %s\n",key);
		RegCloseKey(hKey);
		return false;
	}

	maxElementSize =  outBufferSize / sizeof(wchar_t);

	if (outBuffer[maxElementSize] != '\0') // RegQueryValueEx does not guarantee that the string returned is null terminated
	{
		TRACE1("Registry read failure for %s, string is not NULL terminated\n",key);
		outBuffer[maxElementSize] = '\0';
		RegCloseKey(hKey);
		return false;
	}

	TRACE1("Registry read success for %s\n",key);
	RegCloseKey(hKey);
	return true;
}

bool readIntegerFromRegistry(const wchar_t* key,uint32_t* value)
{
	HKEY hKey;
	DWORD dwType = REG_DWORD;
	DWORD size = sizeof(DWORD);

	if (key == NULL || value == NULL)
		return false;

	// Check if Module is a valid number
	if ( RegOpenKeyEx( REGISTRY_BASE,
		TEXT(REGISTRY_PATH),
		0,
		KEY_READ | KEY_WOW64_64KEY,
		&hKey) != ERROR_SUCCESS )
	{
		TRACE1( "Unable to open Registry [0x%x]\n", GetLastError());
		return false; 
	}

	// Check for the actual value
	if (RegQueryValueEx(hKey,key,0, &dwType, (LPBYTE)value, (LPDWORD)&size) != ERROR_SUCCESS)
	{
		TRACE1("Registry read integer key '%s' failed.\n",key);
		RegCloseKey(hKey);
		return false;
	}

	//TRACE1("Registry read integer key '%s' success\n",key);
	RegCloseKey(hKey);
	return true;
}

JHI_RET_I
JhiQueryAppFileLocationFromRegistry (wchar_t* outBuffer, uint32_t outBufferSize)
{
	if (!readStringFromRegistry(KEY_JHI_APPLETS_REPOSITORY_PATH,outBuffer,outBufferSize))
		return JHI_ERROR_REGISTRY;

	return JHI_SUCCESS;
}

JHI_RET_I
JhiQueryServiceFileLocationFromRegistry (wchar_t* outBuffer, uint32_t outBufferSize)
{
	if (!readStringFromRegistry(KEY_JHI_FILES_PATH,outBuffer,outBufferSize))
		return JHI_ERROR_REGISTRY;

	return JHI_SUCCESS;
}

JHI_RET_I
JhiQueryServicePortFromRegistry(uint32_t* portNumber)
{
	if (!readIntegerFromRegistry(KEY_JHI_SERVICE_PORT,portNumber))
		return JHI_ERROR_REGISTRY;

	return JHI_SUCCESS;
}

JHI_RET_I
JhiQueryTransportTypeFromRegistry(uint32_t* transportType)
{
	if (!readIntegerFromRegistry(KEY_JHI_TRANSPORT_TYPE,transportType))
		return JHI_ERROR_REGISTRY;

	return JHI_SUCCESS;
}

JHI_RET_I
JhiQueryAddressTypeFromRegistry(uint32_t* addressType)
{
	if (!readIntegerFromRegistry(KEY_JHI_ADDRESS_TYPE,addressType))
		return JHI_ERROR_REGISTRY;

	return JHI_SUCCESS;
}

JHI_RET_I
JhiQueryLogLevelFromRegistry(JHI_LOG_LEVEL *loglevel)
{
	uint32_t i_loglevel = 1; // Release logs by default
	*loglevel = JHI_LOG_LEVEL_RELEASE;

	if (!readIntegerFromRegistry(KEY_JHI_LOG_FLAG, &i_loglevel))
	{
		LOG0("LogLevel setting not found. Setting to release prints only.");
	}
	else
	{
		switch (i_loglevel)
		{
		case 0:
			*loglevel = JHI_LOG_LEVEL_OFF;
			break;
		case 1:
			*loglevel = JHI_LOG_LEVEL_RELEASE;
			break;
		case 2:
			*loglevel = JHI_LOG_LEVEL_DEBUG;
			break;
		default:
			*loglevel = JHI_LOG_LEVEL_RELEASE;
			break;
		}
	}
	
	return JHI_SUCCESS;
}

bool WriteStringToRegistry(const wchar_t* key,wchar_t* value, uint32_t value_size)
{
	HKEY hKey;

	// Check if Module is a valid number
	if ( RegOpenKeyEx( REGISTRY_BASE,
		TEXT(REGISTRY_PATH),
		0,
		KEY_WRITE  | KEY_WOW64_64KEY,
		&hKey) != ERROR_SUCCESS )
	{
		TRACE1( "Unable to open Registry [0x%x]\n", GetLastError());
		return false; 
	}

	if(RegSetValueEx(hKey, key ,0L, REG_SZ, (CONST BYTE*) value, value_size) != ERROR_SUCCESS)
	{
		TRACE2("write key: '%s' value: '%s' to registry falied.\n",key,value);
		RegCloseKey(hKey);
		return false;
	}
	
	TRACE2("write key: '%s' value: '%s' to registry succeeded.\n",key,value);
	RegCloseKey(hKey);
	return true;
}

JHI_RET_I
JhiWritePortNumberToRegistry(uint32_t portNumber)
{
	HKEY hKey;

	// Check if Module is a valid number
	if ( RegOpenKeyEx( REGISTRY_BASE,
		TEXT(REGISTRY_PATH),
		0,
		KEY_WRITE  | KEY_WOW64_64KEY,
		&hKey) != ERROR_SUCCESS )
	{
		TRACE1( "Unable to open Registry [0x%x]\n", GetLastError());
		return JHI_ERROR_REGISTRY; 
	}

	if( RegSetValueEx(hKey, KEY_JHI_SERVICE_PORT ,0L, REG_DWORD, (CONST BYTE*) &portNumber, sizeof(DWORD)) == ERROR_SUCCESS)
	{
		TRACE0("write port number to registry success\n");
	}
	else 
	{
		TRACE0("write port number to registry failed\n");
		RegCloseKey(hKey);
		return JHI_ERROR_REGISTRY;
	}

	RegCloseKey(hKey);
	return JHI_SUCCESS;
}

JHI_RET_I
JhiWriteAddressTypeToRegistry(uint32_t addressType)
{
	HKEY hKey;

	// Check if Module is a valid number
	if ( RegOpenKeyEx( REGISTRY_BASE,
		TEXT(REGISTRY_PATH),
		0,
		KEY_WRITE  | KEY_WOW64_64KEY,
		&hKey) != ERROR_SUCCESS )
	{
		TRACE1( "Unable to open Registry [0x%x]\n", GetLastError());
		return JHI_ERROR_REGISTRY; 
	}

	if( RegSetValueEx(hKey, KEY_JHI_ADDRESS_TYPE ,0L, REG_DWORD, (CONST BYTE*) &addressType, sizeof(DWORD)) == ERROR_SUCCESS)
	{
		TRACE0("write address to registry success\n");
	}
	else 
	{
		TRACE0("write address to registry failed\n");
		RegCloseKey(hKey);
		return JHI_ERROR_REGISTRY;
	}

	RegCloseKey(hKey);
	return JHI_SUCCESS;
}