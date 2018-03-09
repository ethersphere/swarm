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

#include "jhi_event.h"
#include "misc.h"

namespace intel_dal
{
	JhiEvent::JhiEvent():
		_name(NULL), _event(NULL), _created(false)
	{}

	JhiEvent::~JhiEvent()
	{
		clean();
	}

	bool JhiEvent::is_created()
	{
		return _created;
	}

	void JhiEvent::clean()
	{
		if (NULL != _event)
		{
			CloseHandle(_event);
			_event = NULL;
		}
		if (NULL != _name)
		{
			JHI_DEALLOC(_name);
			_name = NULL;
		}
		_created = false;
	}

	bool JhiEvent::__open_create(const char* name, bool open)
	{
		if (_created || NULL == name)
			return false;

		clean();
		uint32_t length = (uint32_t) strlen(name) + 1;
		_name = (char*)JHI_ALLOC(length);
		if (NULL == _name)
			return false;
		strcpy_s(_name, length, name);
		if (open)
			_event = OpenEventA(EVENT_MODIFY_STATE, FALSE, _name);
		else
			_event = CreateEventA(NULL, FALSE, FALSE, _name);
		if (NULL == _event)
			return false;
		_created = true;
		return true;
	}

	bool JhiEvent::create(const char* name)
	{
		return __open_create(name, false);
	}

	bool JhiEvent::open(const char* name)
	{
		return __open_create(name, true);
	}

	bool JhiEvent::close()
	{
		if (!_created)
			return false;
		clean();
		return true;
	}

	bool JhiEvent::wait()
	{
		if (!_created || NULL == _event)
			return false;

		DWORD ret = WaitForSingleObject(_event, INFINITE);
		return (ret == WAIT_OBJECT_0);
	}

	bool JhiEvent::set()
	{
		if (!_created || NULL == _event)
			return false;

		SetEvent(_event);
		return true;
	}
}
