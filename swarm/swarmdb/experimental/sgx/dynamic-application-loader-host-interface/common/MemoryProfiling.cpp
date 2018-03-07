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


#include "MemoryProfiling.h"
#include "dbg.h"

using std::list;

//---------------------------------------------------------------------------------------------------------------------
//
// This class is used for memory profiling, meaning detect cases of allocated memory that wwasn't freed correctly.
// The function addAllocation() is called by any allocation macro, and the function removeAllocation() is called
// by any dealloc macro.
// This class is Singleton. In order to print the current allocations, add the following line in the required location:
// MemoryProfiling::Instance().printAllocations();
//
//---------------------------------------------------------------------------------------------------------------------

/*
	Add an allocation to the list (in case of a new allocation using one of the macros: JHI_ALLOC, JHI_ALLOC_T etc.)
*/
void MemoryProfiling::addAllocation(void* pMem, int size, const char* file, int line)
{
	JHI_ALLOC_NODE node;

	node.pMem = pMem;
	node.size = size;
	node.file = file;
	node.line = line;

	allocList.push_back(node);
}

/*
	Search for a mathing poiner in the list. If it foum, then it will be removed.
	This function is called by the deallocation macros (such as JHI_DEALLOC, JHI_DEALLOC_1 etc.).
*/
void MemoryProfiling::removeAllocation(void* pMem)
{
	std::list<JHI_ALLOC_NODE>::iterator iter = allocList.begin();
	std::list<JHI_ALLOC_NODE>::iterator end = allocList.end();

	while (iter != end)
	{
		JHI_ALLOC_NODE node = *iter;

		if (node.pMem == pMem)
		{
			iter = allocList.erase(iter);
		}
		else
		{
			++iter;
		}
	}
}

/*
	Print the allocations list
*/
void MemoryProfiling::printAllocations()
{
	TRACE0("----------------------------------------------------------------------------------------------------------");

	TRACE1("Allocations list size = %d", allocList.size());

	int totalSize = 0;
	int index = 1;

	for (list<JHI_ALLOC_NODE>::iterator it = allocList.begin(); it != allocList.end(); it++, index++)
	{
		TRACE4("(%d)  allocation size = %d, file name = %s, line number = %d\n", index, (*it).size, (*it).file, (*it).line);
		totalSize += (*it).size;
	}

	TRACE1("Total allocations = %d bytes", totalSize);

	TRACE0("----------------------------------------------------------------------------------------------------------");
}