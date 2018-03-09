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

#include "teetransport_internal.h"

#ifndef _WIN32
#include <stdlib.h>
#endif

int isEntityValid(TEE_TRANSPORT_ENTITY entity)
{
	// should be black list and NOT range check since the TEE_TRANSPORT_PORT values are not consecutive range
	if( (entity == TEE_TRANSPORT_ENTITY_IVM) ||
		(entity == TEE_TRANSPORT_ENTITY_RTM) ||
		(entity == TEE_TRANSPORT_ENTITY_SDM) ||
		(entity == TEE_TRANSPORT_ENTITY_SVM) ||
		(entity == TEE_TRANSPORT_ENTITY_CUSTOM) )
	{
		return 1;
	}
	return 0;
}

int string_to_uuid2(IN const char* str, OUT GUID* guid)
{
    if (str == NULL || guid == NULL) return 0;
	//implement when enbling native TAs
    return 1;
}

const GUID* ParseGUID(IN const char* params, OUT GUID* guid)
{	
	if (string_to_uuid2(params, guid))
    {
        return guid;
    }

	return NULL;
}

TEE_COMM_STATUS TEEMutexCreate(OUT TEE_MUTEX_HANDLE* mutex)
{
	if(NULL == mutex)
	{
		return TEE_COMM_INVALID_PARAMS;
	}

#if defined(_WIN32)
	{
		CRITICAL_SECTION* pCriticalSection = (CRITICAL_SECTION*)malloc(sizeof(CRITICAL_SECTION));
		InitializeCriticalSection(pCriticalSection);
		*mutex = (TEE_MUTEX_HANDLE*)pCriticalSection;
	}
#else
    pthread_mutex_t* pm = (pthread_mutex_t*)malloc(sizeof(pthread_mutex_t));
    if(pm == NULL) return TEE_COMM_OUT_OF_MEMORY;

    if(pthread_mutex_init(pm, NULL) != 0) // If failed
    {
        free(pm);
        pm = NULL;
        return TEE_COMM_INTERNAL_ERROR;
    }

    *mutex = pm;
#endif

	return TEE_COMM_SUCCESS;
}
TEE_COMM_STATUS TEEMutexLock(IN TEE_MUTEX_HANDLE mutex)
{
	if(NULL == mutex)
	{
		return TEE_COMM_INVALID_PARAMS;
	}

#if defined(_WIN32)
	{
		CRITICAL_SECTION* pCriticalSection = (CRITICAL_SECTION*)mutex;
		EnterCriticalSection(pCriticalSection);
	}
#else
        {
		pthread_mutex_unlock (mutex);
	}
#endif

	return TEE_COMM_SUCCESS;
}
TEE_COMM_STATUS TEEMutexUnlock(IN TEE_MUTEX_HANDLE mutex)
{
	if(NULL == mutex)
	{
		return TEE_COMM_INVALID_PARAMS;
	}

#if defined(_WIN32)
	{
		CRITICAL_SECTION* pCriticalSection = (CRITICAL_SECTION*)mutex;
		LeaveCriticalSection(pCriticalSection);
	}
#else
	pthread_mutex_unlock (mutex);
#endif

	return TEE_COMM_SUCCESS;
}
TEE_COMM_STATUS TEEMutexDestroy(IN TEE_MUTEX_HANDLE mutex)
{
	if(NULL == mutex)
	{
		return TEE_COMM_INVALID_PARAMS;
	}

#ifdef _WIN32
	{
		CRITICAL_SECTION* pCriticalSection = (CRITICAL_SECTION*)mutex;
		DeleteCriticalSection(pCriticalSection);
		free(pCriticalSection);
	}
#else
	free (mutex);
#endif

	return TEE_COMM_SUCCESS;
}
