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
*    @file        TEETransportTeeLibClientMetadata.c
*    @brief       Implementation of functions for TEE clients meta data management.
*    @author      Artum Tchachkin
*    @date        August 2014
*/

#include "teetransport_libtee_client_metadata.h"
#include "teetransport_internal.h"

TEE_COMM_STATUS SetupContext(IN TEE_CLIENT_META_DATA_CONTEXT* pContext)
{
	if(NULL == pContext)
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	if(TEE_COMM_SUCCESS != TEEMutexCreate(&pContext->client_mutex))
	{
		pContext->client_mutex = NULL;
		return TEE_COMM_INTERNAL_ERROR;
	}

	return TEE_COMM_SUCCESS;
}

static void ReleaseLinkedList(TEE_CLIENT_META_DATA* pHead)
{
	TEE_CLIENT_META_DATA* pCurrent = pHead;

	while(pCurrent != NULL)
	{
		TEE_CLIENT_META_DATA* ptr = pCurrent;
		pCurrent = pCurrent->pNext;

		if(NULL != ptr->buffer)
		{
			free(ptr->buffer);
			ptr->buffer = NULL;
		}
		free(ptr);
	}
}

TEE_COMM_STATUS TeardownContext(IN TEE_CLIENT_META_DATA_CONTEXT* pContext)
{
	if(NULL == pContext)
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	if(TEE_COMM_SUCCESS == TEEMutexLock(pContext->client_mutex))
	{
		ReleaseLinkedList(pContext->client_list);
		pContext->client_list = NULL;

		if(TEE_COMM_SUCCESS != TEEMutexUnlock(pContext->client_mutex))
		{
			// Log error
		}
	}
	else
	{
		// Log error
	}

	if(TEE_COMM_SUCCESS != TEEMutexDestroy(pContext->client_mutex))
	{
		// Log error
	}
	pContext->client_mutex = NULL;
	return TEE_COMM_SUCCESS;
}

TEE_COMM_STATUS RegisterClient(IN TEE_CLIENT_META_DATA_CONTEXT* pContext, IN TEE_CLIENT_META_DATA* pClient)
{
	if((NULL == pContext) || (NULL == pClient))
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	pClient->buffer = malloc(pClient->tee_context.maxMsgLen);
	if(NULL == pClient->buffer)
	{
		return TEE_COMM_OUT_OF_MEMORY;
	}

	memset(pClient->buffer, 0, pClient->tee_context.maxMsgLen);
	pClient->capacity = 0;
	pClient->curr_pos = 0;

	if(TEE_COMM_SUCCESS != TEEMutexLock(pContext->client_mutex))
	{
		free(pClient->buffer);
		pClient->buffer = NULL;

		return TEE_COMM_INTERNAL_ERROR;
	}

	// Count active clients
	pContext->client_count++;

	// advance this internal counter on each new client
	pContext->internal_counter++;
	if(pContext->internal_counter == (uintptr_t)TEE_TRANSPORT_INVALID_HANDLE_VALUE)
	{
		pContext->internal_counter++;
	}

	// use unique counter as HANDLE generator
	pClient->handle = pContext->internal_counter; 

	if(NULL == pContext->client_list)
	{
		pContext->client_list = pClient;
	}
	else
	{
		pClient->pNext = pContext->client_list;
		pContext->client_list = pClient;
	}

	if(TEE_COMM_SUCCESS != TEEMutexUnlock(pContext->client_mutex))
	{
		// Log error
	}

	return TEE_COMM_SUCCESS;
}

TEE_COMM_STATUS UnregisterClient(IN TEE_CLIENT_META_DATA_CONTEXT* pContext, IN uintptr_t handle, OUT TEE_CLIENT_META_DATA** ppClient)
{
	if((NULL == pContext) || (NULL == ppClient) || (handle == (uintptr_t)TEE_TRANSPORT_INVALID_HANDLE_VALUE))
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	// set default value in case of error
	*ppClient = NULL;

	if(TEE_COMM_SUCCESS != TEEMutexLock(pContext->client_mutex))
	{
		return TEE_COMM_INTERNAL_ERROR;
	}

	if(NULL != pContext->client_list)
	{
		TEE_CLIENT_META_DATA* pCurrent = NULL;
		TEE_CLIENT_META_DATA* pPrev = NULL;
		for(pCurrent = pContext->client_list; pCurrent != NULL; pCurrent = pCurrent->pNext)
		{
			if(pCurrent->handle == handle)
			{
				*ppClient = pCurrent;
				pContext->client_count--;
				if(NULL == pPrev)
				{
					pContext->client_list = pCurrent->pNext;
				}
				else
				{
					pPrev->pNext = pCurrent->pNext;
				}

				break;
			}
			pPrev = pCurrent;
		}
	}

	if(TEE_COMM_SUCCESS != TEEMutexUnlock(pContext->client_mutex))
	{
		// Log error
	}

	return TEE_COMM_SUCCESS;
}

TEE_CLIENT_META_DATA* GetClientByHandle(IN TEE_CLIENT_META_DATA_CONTEXT* pContext, uintptr_t handle)
{
	TEE_CLIENT_META_DATA* pCurrent = NULL;
	TEE_CLIENT_META_DATA* pResult = NULL;

	if((NULL == pContext) || (handle == (uintptr_t)TEE_TRANSPORT_INVALID_HANDLE_VALUE))
	{
		// Log error
		return NULL;
	}

	if(TEE_COMM_SUCCESS != TEEMutexLock(pContext->client_mutex))
	{
		// Log error
		return NULL;
	}

	for(pCurrent = pContext->client_list; pCurrent != NULL; pCurrent = pCurrent->pNext)
	{
		if(pCurrent->handle == handle)
		{
			pResult = pCurrent;
			break;
		}
	}

	if(TEE_COMM_SUCCESS != TEEMutexUnlock(pContext->client_mutex))
	{
		// Log error
	}

	return pResult;
}

TEE_CLIENT_META_DATA* NewClient()
{
	TEE_CLIENT_META_DATA* pNewClient = (TEE_CLIENT_META_DATA*)malloc(sizeof(TEE_CLIENT_META_DATA));
	if(NULL != pNewClient)
	{
		memset(pNewClient, 0, sizeof(TEE_CLIENT_META_DATA));
	}

	return pNewClient;
}

TEE_COMM_STATUS DeleteClient(IN TEE_CLIENT_META_DATA* pClient)
{
	if(NULL != pClient)
	{
		if(NULL != pClient->buffer)
		{
			free(pClient->buffer);
			pClient->buffer = NULL;
		}
		free(pClient);
	}

	return TEE_COMM_SUCCESS;
}

/*TEE_COMM_STATUS BeginUsingMetaData(IN TEE_CLIENT_META_DATA_CONTEXT* pContext, IN TEE_CLIENT_META_DATA* pClient)
{
	if((NULL == pContext) || (NULL == pClient))
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	if(TEE_COMM_SUCCESS != TEEMutexLock(pContext->client_mutex))
	{
		return TEE_COMM_INTERNAL_ERROR;
	}

	pClient->users++;

	if(TEE_COMM_SUCCESS != TEEMutexUnlock(pContext->client_mutex))
	{
		// Log error
	}

	return TEE_COMM_SUCCESS;
}*/

/*TEE_COMM_STATUS EndUsingMetaData(IN TEE_CLIENT_META_DATA_CONTEXT* pContext, IN TEE_CLIENT_META_DATA* pClient)
{
	if((NULL == pContext) || (NULL == pClient))
	{
		return TEE_COMM_INVALID_PARAMS;
	}

	if(TEE_COMM_SUCCESS != TEEMutexLock(pContext->client_mutex))
	{
		return TEE_COMM_INTERNAL_ERROR;
	}

	pClient->users--;

	if(TEE_COMM_SUCCESS != TEEMutexUnlock(pContext->client_mutex))
	{
		// Log error
	}

	return TEE_COMM_SUCCESS;
}*/
