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
**    @file misc.c
**
**    @brief  Miscellaneous util functions for JHI.DLL and JHI_SERVICE
**
**    @author Niveditha Sundaram
**	  @author Venky Gokulrangan	
**
********************************************************************************
*/

#include "misc.h"
#include <algorithm>
#include <iterator>

#ifdef _WIN32
#include <Windows.h>
#else
#include <string.h>
#include "string_s.h"
#include <sstream>
#include <ctype.h>
#include <sys/stat.h>
#endif // _WIN32

//------------------------------------------------------------------------------
//
//------------------------------------------------------------------------------
#ifdef __ANDROID__

#ifdef JHI_MEMORY_PROFILING
void * JHI_ALLOC1(uint32_t bytes_alloc, const char* file, int line)
#else
void * JHI_ALLOC(uint32_t bytes_alloc)
#endif //JHI_MEMORY_PROFILING
{
	void* var = NULL;
	var = (void*) new uint8_t[bytes_alloc];
	if (NULL == var)
	{
		TRACE1("JHI memory allocation of size %d failed .", bytes_alloc);
	}
#ifdef JHI_MEMORY_PROFILING
	TRACE4("JHI_ALLOC1: address = %#08x, bytes allocated = %d, file = %s, line = %d\n",
	       var, bytes_alloc, file, line);

	MemoryProfiling::Instance().addAllocation(var, bytes_alloc, file, line);
#endif
	return var;
}

#ifdef JHI_MEMORY_PROFILING
void JHI_DEALLOC1(void* handle, const char* file, int line)
#else
void JHI_DEALLOC(void* handle)
#endif
{
	if (handle != NULL)
		delete [] (uint8_t*)handle;
#ifdef JHI_MEMORY_PROFILING
	TRACE3("JHI_DEALLOC: address = %#08x, file = %s, line = %d\n",
	       handle, file, line);

	MemoryProfiling::Instance().removeAllocation((void*)handle);
#endif
}

#else //NOT ANDROID

#ifdef JHI_MEMORY_PROFILING
void * JHI_ALLOC1(uint32_t bytes_alloc, const char* file, int line)
#else
void * JHI_ALLOC(uint32_t bytes_alloc)
#endif //JHI_MEMORY_PROFILING
{
	void* var = NULL;
	try
	{
		var = (void*) new uint8_t[bytes_alloc];
	}
	catch (...) 
	{
		TRACE1("JHI memory allocation of size %d failed .",bytes_alloc);
	}

#ifdef JHI_MEMORY_PROFILING
	MemoryProfiling::Instance().addAllocation(var, bytes_alloc, file, line);

	TRACE4("JHI_ALLOC1: address = %#08x, bytes allocated = %d, file = %s, line = %d\n",
	       var, bytes_alloc, file, line);
#endif
	return var;
}

//------------------------------------------------------------------------------
//
//------------------------------------------------------------------------------
#ifdef JHI_MEMORY_PROFILING
void JHI_DEALLOC1(void* handle, const char* file, int line)
#else
void JHI_DEALLOC(void* handle)
#endif //JHI_MEMORY_PROFILING
{
	try 
	{
		if (handle != NULL)
		{
			delete [] (uint8_t*)handle;
		}
	}
	catch (...) 
	{
		TRACE0("JHI free memory failed.");
	}
#ifdef JHI_MEMORY_PROFILING
	MemoryProfiling::Instance().removeAllocation((void*)handle);
	TRACE3("JHI_DEALLOC1: address = %#08x, file = %s, line = %d\n",
	       handle, file, line);
#endif
	handle = NULL;
}

#endif //ANDROID

//------------------------------------------------------------------------------
// OS-Neutral Copy
//------------------------------------------------------------------------------
#ifndef _WIN32
#define LEN_MAX_BUFFER 4096
uint32_t
JhiUtilCopyFile (const char *pDstFile,const  char *pSrcFile)
{
	FILE *fpDst = NULL, *fpSrc = NULL;
	uint32_t ulRetCode = JHI_SUCCESS, len;
	char buf1[LEN_MAX_BUFFER];

	if ( !(pDstFile && pSrcFile) )
		return JHI_INVALID_PARAMS ;

	TRACE2 ("Copy file params: src: %s dest: %s\n", pSrcFile, pDstFile);

	// Open the destination file
	fpDst = fopen(pDstFile, "wb");
	if (NULL == fpDst) 
	{
		TRACE0("dest file fopen failed");
		ulRetCode = JHI_FILE_ERROR_OPEN ;
	}

	// Open the source file
	if( JHI_SUCCESS == ulRetCode ) 
	{
		fpSrc = fopen(pSrcFile, "rb");
		if (NULL == fpSrc)
		{
			TRACE0("src file fopen failed");
			fclose(fpDst);  // Close the previous one as well
			ulRetCode = JHI_FILE_ERROR_OPEN ;
		}
	}

	// Check for some maximum size after which we fail the copy
	while ( (ulRetCode == JHI_SUCCESS) &&
			(len = fread(buf1, 1, LEN_MAX_BUFFER, fpSrc)) > 0 )
	{
		if( fwrite(buf1, 1, len, fpDst) != len )
		{
			TRACE0("fwrite failed") ;
			ulRetCode = JHI_FILE_ERROR_COPY ;
			break;
		}
	}

	if(fpSrc)
		fclose(fpSrc);

	if(fpDst)
		fclose(fpDst);

	return ulRetCode ; // How prodigous
}

uint32_t JhiUtilCreateFile_fromBuff (const char *pDstFile, const char * blobBuf, uint32_t len)
{
	FILE* fpDst = NULL;
	uint32_t ulRetCode = JHI_SUCCESS;

	if ((pDstFile == NULL || blobBuf == NULL) )
		return JHI_INVALID_PARAMS ;
	fpDst = fopen(pDstFile, "wb");
	if (NULL == fpDst)
	{
		return JHI_FILE_ERROR_OPEN ;
	}
	if( fwrite(blobBuf, 1, len, fpDst) != len )
	{
		TRACE0( "WRITE FILE FROM BLOB FAILURE\n");
		ulRetCode = JHI_FILE_ERROR_COPY ;
	}
	if(fpDst)
		fclose(fpDst);

	return ulRetCode ;
}
#endif // !WIN32


//------------------------------------------------------------------------------
// Function: JhiUtilUUID_Validate
//------------------------------------------------------------------------------

int
JhiUtilUUID_Validate(
	const char*   AppId, 
	UINT8*  ucAppId
)
{
	int ulRetCode = JHI_SUCCESS ;
	size_t len, i;

	if( !(AppId && ucAppId) ) 
		return JHI_INVALID_APPLET_GUID;

	if (AppId[LEN_APP_ID] != '\0')
		return JHI_INVALID_APPLET_GUID;

	len = strlen(AppId) ;
	if( LEN_APP_ID != len )
		return JHI_INVALID_APPLET_GUID;

	// Go thru each of the byte and convert to upper case if need be.
	for( i=0; i<len; i++, AppId++ )
	{
		UINT8 c = (*AppId & 0xff ) ;

		if( isdigit(*AppId) ||
			( isalpha(*AppId) && ((toupper(c) >= 'A') && (toupper(c) <= 'F'))))
		{
			ucAppId[i] = toupper(c) ; // even for digits, this is fine
		}
		else 
		{
			return JHI_INVALID_APPLET_GUID;
		}
	}
	ucAppId[LEN_APP_ID] = 0 ;
	return ulRetCode ;
}

string strToUppercase(const string& str)
{
	string uppercaseStr;
	std::transform(str.begin(), str.end(), std::back_inserter(uppercaseStr), ::toupper);
	return uppercaseStr;
}

bool validateUuidList(UUID_LIST* uuidList)
{
	if ( (uuidList == NULL) 
		|| ( (uuidList->uuidCount != 0) && (uuidList->uuids == NULL) ) )
	{
		return false;
	}
	
	char* index = (char*)uuidList->uuids;
	
	for (uint32_t i = 0; i < uuidList->uuidCount; ++i)
	{
		if (!validateUuidChar(index))
		{
			return false;
		}
		index += UUID_LEN;
	}
	return true;
}

bool validateUuidChar(const char* index)
{
	if (index[LEN_APP_ID] != '\0')
	{
		return false;
	}

	for(int i = 0; i < LEN_APP_ID; ++i, ++index) 
	{
		if(! ((*index >= '0' && *index <= '9') || 
			(*index >= 'a' && *index <= 'f') || 
			(*index >= 'A' && *index <= 'F')))
			return false;
	}
	return true;
}

bool validateUuidString(const string& str)
{
	if (str.length() != LEN_APP_ID)
	{
		return false; //incorrect string length
	}

	char* index = (char*) str.c_str();

	return validateUuidChar(index);
}

#ifdef _WIN32
// this function converts a string into a wstring
wstring ConvertStringToWString(const string& str)
{
	wstring wstr(str.length(), L' ');
	copy(str.begin(), str.end(), wstr.begin());
	return wstr;
}

// this function converts a wstring into a string
string ConvertWStringToString(const wstring& wstr)
{
	string str(wstr.begin(), wstr.end());
	return str;
}
#else
//Stubs for non-wchar Linux	
string ConvertStringToWString(const string& str){return str;}	
string ConvertWStringToString(const string& wstr){return wstr;}	
#endif // _WIN32

// this function removes all leading and trailing spaces,tabs,newlines etc from a string
string TrimString(const string& str)
{

	size_t startStr = str.find_first_not_of(" \t\r\n");
	size_t endStr    = str.find_last_not_of(" \t\r\n");

	if (startStr == string::npos)
		return "";

	return str.substr(startStr,endStr - startStr + 1);
}

#ifdef _WIN32

// this function verify that we are running in windows vista and above
bool isVistaOrLater()
{
	OSVERSIONINFOEX versionInfo = {0};
	DWORDLONG conditionMask = 0;
	BYTE comparisonMethod=VER_GREATER_EQUAL;
		
	BOOL isVistaOrLater = FALSE;

	versionInfo.dwOSVersionInfoSize = sizeof(OSVERSIONINFOEX);
	versionInfo.dwMajorVersion = 6;	// windows vista is defined as version 6.0 
	versionInfo.dwMinorVersion = 0; // OS version list can be found at: http://msdn.microsoft.com/en-us/library/ms724833%28v=vs.85%29.aspx

	VER_SET_CONDITION( conditionMask, VER_MAJORVERSION, comparisonMethod );
	VER_SET_CONDITION( conditionMask, VER_MINORVERSION, comparisonMethod );

	isVistaOrLater = VerifyVersionInfo(&versionInfo,VER_MAJORVERSION | VER_MINORVERSION,conditionMask);

	TRACE1("OS is vista or later flag: %d\n", isVistaOrLater);

	return (isVistaOrLater == TRUE);
}

#endif // _WIN32

bool isJhiError(uint32_t error)
{
	if (error < 0x2000)
	{
		return true;
	}
	return false;
}

TEE_STATUS jhiErrorToTeeError(JHI_RET jhiError)
{
	if (!isJhiError(jhiError))
	{
		return (TEE_STATUS) jhiError;
	}

	TEE_STATUS teeError = TEE_STATUS_INTERNAL_ERROR;

	switch (jhiError)
	{
		// General errors
	case JHI_SUCCESS:
		teeError = TEE_STATUS_SUCCESS;
		break;

	case JHI_INTERNAL_ERROR:
		teeError = TEE_STATUS_INTERNAL_ERROR;
		break;

	case JHI_INVALID_PARAMS:
		teeError = TEE_STATUS_INVALID_PARAMS;
		break;

	case JHI_INVALID_APPLET_GUID:
		teeError = TEE_STATUS_INVALID_UUID;
		break;

	case JHI_NO_CONNECTION_TO_FIRMWARE:
		teeError = TEE_STATUS_NO_FW_CONNECTION;
		break;

		// Service errors
	case JHI_SERVICE_UNAVAILABLE:
		teeError = TEE_STATUS_SERVICE_UNAVAILABLE;
		break;

	case JHI_ERROR_REGISTRY:
		teeError = TEE_STATUS_REGISTRY_ERROR;
		break;

	case JHI_ERROR_REPOSITORY_NOT_FOUND:
	case JHI_DELETE_FROM_REPOSITORY_FAILURE:
	case JHI_FILE_ERROR_COPY:
		teeError = TEE_STATUS_REPOSITORY_ERROR;
		break;

	case JHI_SPOOLER_NOT_FOUND:
		teeError = TEE_STATUS_SPOOLER_MISSING;
		break;

	case JHI_INVALID_SPOOLER:
		teeError = TEE_STATUS_SPOOLER_INVALID;
		break;

	case JHI_VM_DLL_FILE_NOT_FOUND:
		teeError = TEE_STATUS_PLUGIN_MISSING;
		break;

	case JHI_VM_DLL_VERIFY_FAILED:
		teeError = TEE_STATUS_PLUGIN_VERIFY_FAILED;
		break;

		// Package errors
	case JHI_INVALID_PACKAGE_FORMAT:
		teeError = TEE_STATUS_INVALID_PACKAGE;
		break;

	case JHI_FILE_ERROR_AUTH:
		teeError = TEE_STATUS_INVALID_SIGNATURE;
		break;

	case JHI_MISSING_ACCESS_CONTROL:
		teeError = TEE_STATUS_MISSING_ACCESS_CONTROL;
		break;

		// Install / uninstall TA errors:
	case JHI_MAX_INSTALLED_APPLETS_REACHED:
		teeError = TEE_STATUS_MAX_TAS_REACHED;
		break;

	case JHI_INSTALL_FAILURE_SESSIONS_EXISTS:
	case JHI_UNINSTALL_FAILURE_SESSIONS_EXISTS:
		teeError = TEE_STATUS_CMD_FAILURE_SESSIONS_EXISTS;
		break;

	case JHI_SVL_CHECK_FAIL:
		teeError = TEE_STATUS_INVALID_TA_SVN;
		break;

	case JHI_APPLET_NOT_INSTALLED:
		teeError = TEE_STATUS_TA_DOES_NOT_EXIST;
		break;

	case JHI_FILE_IDENTICAL:
		teeError = TEE_STATUS_IDENTICAL_PACKAGE;
		break;

	case JHI_INSTALL_FAILED:
		teeError = TEE_STATUS_CMD_FAILURE;
		break;

	case JHI_ILLEGAL_PLATFORM_ID:
		teeError = TEE_STATUS_ILLEGAL_PLATFORM_ID;
		break;

	default:
		break;
	}
	return teeError;
}

#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
JHI_RET freeLoadedAppletsList(IN JHI_LOADED_APPLET_GUIDS* appGUIDs)
{
	UINT32 rc = JHI_INTERNAL_ERROR;
	try
	{
		if (appGUIDs && appGUIDs->appsGUIDs)
		{
			for ( UINT32 i = 0; i < appGUIDs->loadedAppletsCount; ++i)
			{
				if (appGUIDs->appsGUIDs[i])
				{
					JHI_DEALLOC(appGUIDs->appsGUIDs[i]);
					appGUIDs->appsGUIDs[i] = NULL;
				}
			}
			JHI_DEALLOC_T_ARRAY(appGUIDs->appsGUIDs);
			appGUIDs->appsGUIDs = NULL;
		}
		rc = JHI_SUCCESS;
	}
	catch(...)
	{
		rc = JHI_INVALID_PARAMS;
	}
	return rc;
}
#endif

#ifdef __linux__
JHI_RET getProcStartTime(uint32_t pid, FILETIME& filetime)
{
	std::stringstream fname;
	long long unsigned int data;
	FILE *fd;

	fname << "/proc/" << pid << "/stat";
	fd = fopen(fname.str().c_str(), "r");

	if (NULL == fd)
	{
		TRACE1("Can't open stat for process %d\n", pid);
		return JHI_INTERNAL_ERROR;
	}

	if ( 1 != fscanf(fd, "%*d %*s %*c %*d %*d %*d %*d %*d %*u %*u %*u %*u %*u %*u %*u %*d %*d %*d %*d %*d %*d %llu", &data))
	{
		TRACE1("Can't sscanf stat for process %d\n", pid);
        fclose(fd);
		return JHI_INTERNAL_ERROR;
	}
	fclose(fd);
	memcpy(&filetime, &data, sizeof(data));// not the real FILETIME, but unique enough for our purpose
	return JHI_SUCCESS;
}

bool isProcessDead (uint32_t pid, FILETIME& savedTime)
{
	char pAddr[17];
	memset(pAddr, 0, 17);
	FILETIME creationTime;
	strcpy_s(pAddr, strlen("/proc/"), "/proc/");
    sprintf_s(pAddr + strlen("/proc/"), 11, "%u", pid); // itoa
	struct stat status;
	if (stat(pAddr, &status) == -1 && errno == ENOENT)
	{
		TRACE0("OpenProcess returned NULL\n");
		return true; // there is no such process with the given id
	}
	if (JHI_SUCCESS != getProcStartTime(pid, creationTime))
	{
		TRACE0("failed to get process creation time\n");
		return false; // internal error
	}
	if ((savedTime.dwHighDateTime || savedTime.dwLowDateTime) && /* if savedTime is 0 - do nothing */
			(creationTime.dwHighDateTime != savedTime.dwHighDateTime
			|| creationTime.dwLowDateTime != savedTime.dwLowDateTime))
	{
		return true;
	}
	return false;
}
#ifdef __ANDROID__
bool isServiceRunning ()
{
	const char* service_name = "jhi_service";
	std::stringstream cmd;
	cmd << "ps | grep " << service_name;
	FILE* app = popen(cmd.str().c_str(), "r");
	char instances = '0';
	if (app)
	{
		fread(&instances, sizeof(instances), 1, app);
		pclose(app);
	}
	return (instances != '0');
}
#endif //ANDROID
#endif //__linux__
