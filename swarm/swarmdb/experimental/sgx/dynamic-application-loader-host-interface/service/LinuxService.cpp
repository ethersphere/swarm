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
**	@file LinuxService.cpp
**
**	@brief  Contains linux service implementation for JHI
**
**	@author Alexander Usyskin
**
********************************************************************************
*/

#include <string.h>
#include <iostream>
#include <csignal>
#include "JHIMain.h"
#include "dbg.h"
#include "jhi.h"

// -----------------------------------------------------------------------
// JHI service command line error codes
// ------------------------------------------------------------------------
#define JHI_SERVICE_INIT_FAILED				-1

pthread_t g_main_thread_id;

void * termination_thread(void *)
{
	intel_dal::jhi_stop();

	return nullptr;
}

void termination_handler(int signum)
{
	TRACE1("Got the signal: %d. Terminating...", signum);

	pthread_t thread;
	pthread_create(&thread, NULL, termination_thread, NULL);
}

int main(int argc, char *argv[])
{
	g_main_thread_id = pthread_self();

    if(argc > 1)
    {
        if(!strcmp(argv[1], "-v"))
        {
            std::cout<<JHI_VERSION<<std::endl;
            return 0;
        }
        if(!strcmp(argv[1], "-d"))
            daemon(0,0);
    }

    signal(SIGPIPE, SIG_IGN);
	signal(SIGINT,  termination_handler);
	signal(SIGHUP,  termination_handler);
	signal(SIGTERM, termination_handler);

	if (!intel_dal::jhi_init())
	{
		LOG0("jhi service init failed");
		return JHI_SERVICE_INIT_FAILED;
	}

	intel_dal::jhi_main();

	LOG0("jhid main thread end");
	return 0;
}