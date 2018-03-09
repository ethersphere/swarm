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

#ifndef JHI_EVENT_H
#define JHI_EVENT_H

#ifdef _WIN32
# include "windows.h"
#else
# include "string.h"
#endif//WIN32
namespace intel_dal
{
	class JhiEvent
	{
	private:
		//Event name
		char* _name;
		//Event handle
#ifdef _WIN32
		HANDLE _event;
#else
		bool _isClient;
		int _clFd;
		int _event;
#endif//WIN32
		volatile bool _created;

		void clean();
		bool __open_create(const char* name, bool open);
	public:
		JhiEvent();
		~JhiEvent();

		bool create(const char* name);
		bool open(const char* name);
		bool close();

		bool wait();
		bool set();

		bool is_created();
		
#ifndef _WIN32
		int rx_cnt;
		int tx_cnt;
		bool is_client ();
		bool listenCl();
#endif //WIN32
	private:
		// We do not allow copying of this class
		JhiEvent(const JhiEvent&);
		JhiEvent& operator=(const JhiEvent&);
	};
}
#endif// JHI_EVENT_H
