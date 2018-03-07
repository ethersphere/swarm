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
**    @file misc.h
**
**    @brief  Defines miscellaneous util functions for JHI.DLL and JHI_SERVICE
**
**    @author Niveditha Sundaram
**
********************************************************************************
*/   


#ifndef __MISC_H__
#define __MISC_H__

#ifdef _WIN32
#pragma warning (push)
#pragma warning( disable : 4995 )
#endif
#include <string>
#include <stdio.h>
#include "jhi_i.h"
#include "teemanagement.h"
#include "dbg.h"

using std::string;

// NOTE: Enable this defintion in order to activate JHI memory profiling
//#define JHI_MEMORY_PROFILING

#ifdef JHI_MEMORY_PROFILING
#include "MemoryProfiling.h"
#endif //JHI_MEMORY_PROFILING

#ifdef JHI_MEMORY_PROFILING
void * JHI_ALLOC1(uint32_t bytes_alloc, const char* file, int line);
void JHI_DEALLOC1(void* handle, const char* file, int line);
#define JHI_ALLOC(x) JHI_ALLOC1(x, __FILE__, __LINE__)
#define JHI_DEALLOC(x) JHI_DEALLOC1(x, __FILE__, __LINE__)
#else
void* JHI_ALLOC(uint32_t bytes_alloc);
void JHI_DEALLOC(void* handle);
#endif //JHI_MEMORY_PROFILING
//------------------------------------------------------------------------------
//
//------------------------------------------------------------------------------
template<class T>
#ifdef JHI_MEMORY_PROFILING
T* JHI_ALLOC_T(const char* file, int line)
#else
T* JHI_ALLOC_T()
#endif //JHI_MEMORY_PROFILING
{
	T* var = NULL;
	var = new (std::nothrow) T;
	if (NULL == var)
	{
		LOG1("JHI memory allocation of size %d failed .", sizeof(T));
	}

#ifdef JHI_MEMORY_PROFILING
	TRACE4("JHI_ALLOC_T: address = %#08x, allocated size = %d, file = %s, line = %d\n", var, sizeof(T), file, line);
	MemoryProfiling::Instance().addAllocation(var, sizeof(T), file, line);
#endif //JHI_MEMORY_PROFILING

	return var;
}

template<class T>
#ifdef JHI_MEMORY_PROFILING
void JHI_DEALLOC_T(T* handle, const char* file, int line)
#else
void JHI_DEALLOC_T(T* handle)
#endif //JHI_MEMORY_PROFILING
{
	if (handle != NULL)
		delete handle;

#ifdef JHI_MEMORY_PROFILING
	TRACE4("JHI_DEALLOC_T: address = %#08x, allocated size = %d, file = %s, line = %d\n", handle, sizeof(T), file, line);
	MemoryProfiling::Instance().removeAllocation((void*)handle);
#endif //JHI_MEMORY_PROFILING
}

template<class T>
T* JHI_ALLOC_T_ARRAY(size_t count)
{
	T* var = NULL;

	var = new (std::nothrow) T[count];
	if (NULL == var)
	{
		LOG1("JHI memory allocation of size %d failed .", sizeof(T) * count);
	}

	return var;
}

template<class T>
void JHI_DEALLOC_T_ARRAY(T* handle)
{
	if (handle != NULL)
		delete[] handle;
}

uint32_t JhiUtilCopyFile (const char *pDstFile,const char *pSrcFile);
uint32_t JhiUtilCreateFile_fromBuff (const char *pDstFile, const char * blobBuf, uint32_t len);

int
JhiUtilUUID_Validate(
	const char*   AppId, 
	UINT8*  ucAppId
);

string strToUppercase(const string& str);

bool validateUuidList(UUID_LIST* uuidList);
bool validateUuidChar(const char* index);
bool validateUuidString(const string& str);

#ifdef _WIN32
using std::wstring;
wstring ConvertStringToWString(const string& str);
string ConvertWStringToString(const wstring& wstr);
#else
string ConvertStringToWString(const string& str);
string ConvertWStringToString(const string& wstr);
#endif // _WIN32

string TrimString(const string& str);

#ifdef JHI_MEMORY_PROFILING
#define JHI_ALLOC_T(k) JHI_ALLOC_T<k>(__FILE__, __LINE__)
#define JHI_DEALLOC_T(x) JHI_DEALLOC_T(x, __FILE__, __LINE__)
#else
#define JHI_ALLOC_T(k) JHI_ALLOC_T<k>()
#endif //JHI_MEMORY_PROFILING

#ifdef _WIN32

bool isVistaOrLater();

#endif // _WIN32

TEE_STATUS jhiErrorToTeeError(JHI_RET jhiError);
bool isJhiError(uint32_t error);

#ifdef SCHANNEL_OVER_SOCKET //(emulation mode)
JHI_RET freeLoadedAppletsList(IN JHI_LOADED_APPLET_GUIDS* appGUIDs);
#endif //SCHANNEL_OVER_SOCKET

#ifdef __linux__
bool isProcessDead (uint32_t pid, FILETIME& filetime);
JHI_RET getProcStartTime(uint32_t pid, FILETIME& filetime);
#ifdef __ANDROID__
bool isServiceRunning ();
#endif //ANDROID
#endif //__linux__

#ifdef _WIN32
#pragma warning (pop)
#endif

#endif // __MISC_H__
