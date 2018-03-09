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
**
**    @file jhi_plugin.h
**
**    @brief  Defines API to work with JHI VM plugins 
**
**    @author Elad Dabool
**
********************************************************************************
*/

#ifndef __JHI_PLUGIN_H__
#define __JHI_PLUGIN_H__

#include "plugin_interface.h"

#ifdef _WIN32
#define TEE_FILENAME L"teePlugin.dll"
#define BH_FILENAME L"bhPlugin.dll"
#define BH_V2_FILENAME L"bhPluginV2.dll"
#elif defined(__ANDROID__)
#define TEE_FILENAME "teePlugin.so"
#define BH_FILENAME "libbhplugin1.so"
#define BH_V2_FILENAME "libbhplugin2.so"
#elif defined(__linux__)
#define TEE_FILENAME "teePlugin.so"
#define BH_FILENAME "libbhplugin1.so"
#define BH_V2_FILENAME "libbhplugin2.so"
#else
Unknown OS
#endif//WIN32

#define TEE_VENDORNAME FILEPREFIX("Intel(R) Embedded Subsystems and IP Blocks Group")
#define BH_VENDORNAME FILEPREFIX("Intel(R) Embedded Subsystems and IP Blocks Group")

#define JHI_PLUGIN_REGISTER_FUNCTION "pluginRegister"

JHI_RET JhiPlugin_Register (VM_Plugin_interface** plugin);
JHI_RET	JhiPlugin_Unregister(VM_Plugin_interface** plugin);

///end
#endif
