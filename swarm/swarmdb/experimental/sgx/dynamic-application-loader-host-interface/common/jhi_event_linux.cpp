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

#ifdef __ANDROID__
#include <android/log.h>
#endif // __ANDROID__

#include <sys/types.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <sys/stat.h>
#include "dbg.h"
#include "jhi_event.h"
#include "misc.h"
#include "errno.h"

namespace intel_dal
{
	JhiEvent::JhiEvent():
	_name(NULL), _isClient(false), _clFd(-1), _event(-1), _created(false), rx_cnt(0), tx_cnt(0)
	{}

	JhiEvent::~JhiEvent()
	{
		clean();
	}

	bool JhiEvent::is_created()
	{
		return _created;
	}

	bool JhiEvent::is_client()
	{
		return _isClient;
	}

	void JhiEvent::clean()
	{
		_created = false;

		if (_isClient && -1 != _event)
		{
			::close(_event);
			_event = -1;
			if (NULL != _name)
			{
				unlink(_name);
			}

		}
		else
		{
			if (-1 != _clFd)
			{
				::close(_clFd);
				_clFd = -1;
			}
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
		if (_created || NULL == name) {
		  TRACE1("socket__open_create error, l%d \n", __LINE__);
			return false;
		}
		_isClient = open;
		clean();
		_name = (char*)JHI_ALLOC(strlen(name) + 1);
		if (NULL == _name)
			return false;
		strcpy(_name, name);

		_event = socket(AF_UNIX, SOCK_STREAM, 0);
		if (-1 == _event) {
		  TRACE1("socket__open_create error, l%d \n", __LINE__);
			return false;
		}

		sockaddr_un my_addr;
		memset(&my_addr, 0, sizeof(struct sockaddr_un));
		my_addr.sun_family = AF_UNIX;
		strncpy(my_addr.sun_path, _name, sizeof(my_addr.sun_path) - 1);

		if (open) {
			if (connect(_event, (struct sockaddr *) &my_addr,
					sizeof(struct sockaddr_un)) == -1) {
				TRACE2("socket cl connect, l%d, %s \n",
					__LINE__, strerror(errno));
				return false;
			}
		} else {
			unlink(_name);
			if (bind(_event, (struct sockaddr *) &my_addr,
					sizeof(struct sockaddr_un)) == -1) {
				TRACE2("socket srv bind, l%d, %s \n",
					__LINE__, strerror(errno));
				return false;
			}
			/* NOTE: let everyone permissions, so jhid will be able to r/w the socket 
				the socket is created from the user context, using libjhi */
			if (chmod(_name, S_IRWXU | S_IRWXG | S_IRWXO)) {
				TRACE2("failed to give jhi socket permissions, l%d, %s\n",
					__LINE__, strerror(errno));
				return false;
			}

			TRACE0 ("Socket listen(ing) ...");
			if (listen(_event, 1) == -1) {
				TRACE2("socket srv lstn, l%d, %s \n",
				__LINE__, strerror(errno));
				return false;
			}
		}
		_created = true;
		return true;
	}

	bool JhiEvent::listenCl()
	{
		if (!_created || _isClient || -1 != _clFd) {
			TRACE2("socket srv listen, l%d, %s \n",
				__LINE__, strerror(errno));
			return false;
		}

		struct sockaddr_un remote;
		socklen_t len = sizeof(struct sockaddr_un);
		TRACE0 ("Socket accept(ing) ...");
		_clFd = accept(_event, (struct sockaddr *)&remote, &len);
		if (-1 == _clFd) {
			TRACE2("socket srv accept, l%d, %s \n",
			       __LINE__, strerror(errno));
			return false;
		}
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
		if (!_created || _isClient || -1 == _clFd) {
			TRACE2("socket srv wait, l%d, %s \n",
				__LINE__, strerror(errno));
			return false;
		}
		char buf[1];
		int ret = recv(_clFd, buf, 1, 0);
		rx_cnt++;
		if (ret != 1 || !_created) {
			TRACE2("socket srv recv error, l%d, %s \n",
				__LINE__, strerror(errno));
			return false;
		}
		return true;
	}

	bool JhiEvent::set()
	{
		if (!_created || -1 == _event || !_isClient){
			TRACE2("socket cl set, l%d, %s \n",
				__LINE__, strerror(errno));
			return false;
		}

		char buf[1] = {0x1};
		tx_cnt++;
		int ret = send(_event, buf, 1, 0);
		if (ret != 1) {
			TRACE2("socket cl send, l%d, %s \n",
				__LINE__, strerror(errno));
			return false;
		}
		return true;
	}
}
