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
**    @file jhi_plugin_types.h
**
**    @brief  Conatins plugin defintions taken from jhi.h and jhi_service.h
**
**    @author Elad Dabool
**
********************************************************************************
*/
#ifndef __JHI_PLUGIN_TYPES_H__
#define __JHI_PLUGIN_TYPES_H__

#ifdef _WIN32
#include <windows.h>
#include <Shlwapi.h>
#else
#include <typedefs_i.h>
#endif

#include "jhi.h"

#define JHI_APPLET_AUTHENTICATION_FAILURE		JHI_FILE_ERROR_AUTH     // FW rejected the applet while trying to install it
#define JHI_BAD_APPLET_FORMAT					0x2001	

#define JHI_EVENT_DATA_BUFFER_SIZE 1024
#define SPOOLER_COMMAND_GET_EVENT 1

typedef UUID JHI_SESSION_ID;

#endif // __JHI_PLUGIN_TYPES_H__