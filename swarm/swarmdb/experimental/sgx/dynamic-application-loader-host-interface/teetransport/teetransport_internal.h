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

#ifndef _TEE_TRANSPORT_INTERNAL_H_
#define _TEE_TRANSPORT_INTERNAL_H_

#include <stdint.h>
#ifdef _WIN32
#include <Windows.h>
#else
#include "typedefs_i.h"
#undef uuid_le
#include <pthread.h>
#endif // _WIN32
#include "teetransport.h"

#ifdef __cplusplus
extern "C" {
#endif

	/**
	* Number of valid ports in the TEE_TRANSPORT_ENTITY enum
	*/
#define TEE_TRANSPORT_ENTITY_COUNT    (4)
#if defined(_WIN32)
	typedef void* TEE_MUTEX_HANDLE;
#else
	typedef pthread_mutex_t* TEE_MUTEX_HANDLE;
#endif

	TEE_COMM_STATUS TEEMutexCreate(OUT TEE_MUTEX_HANDLE* mutex);
	TEE_COMM_STATUS TEEMutexLock(IN TEE_MUTEX_HANDLE mutex);
	TEE_COMM_STATUS TEEMutexUnlock(IN TEE_MUTEX_HANDLE mutex);
	TEE_COMM_STATUS TEEMutexDestroy(IN TEE_MUTEX_HANDLE mutex);

	int isEntityValid(TEE_TRANSPORT_ENTITY entity);
	const GUID* ParseGUID(IN const char* params, OUT GUID* guid);


#ifdef __cplusplus
};
#endif


#endif //_TEE_TRANSPORT_INTERNAL_H_