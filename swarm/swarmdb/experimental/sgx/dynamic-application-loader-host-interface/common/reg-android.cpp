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
**    @file reg-android.cpp
**
**    @brief  Defines registry related functions
**
**    @author Alexander Usyskin
**
********************************************************************************
*/

#include <stdio.h>
#include <sys/socket.h>
#include <stdlib.h>
#include <sys/system_properties.h>
#include "reg.h"
#include "misc.h"

JHI_RET_I
JhiQueryAppFileLocationFromRegistry (char* outBuffer, uint32_t outBufferSize)
{
	char data[PROP_VALUE_MAX];
	int ret = __system_property_get("persist.jhi.APPLETS_LOCALE", data);
	if (0 == ret)
		strncpy(outBuffer, "/data/intel/dal/applet_repository", outBufferSize);
	else
		strncpy(outBuffer, data, outBufferSize);
	return JHI_SUCCESS;
}

JHI_RET_I
JhiQuerySpoolerLocationFromRegistry (char* outBuffer, uint32_t outBufferSize)
{
	char data[PROP_VALUE_MAX];
	int ret = __system_property_get("persist.jhi.SPOOLER_LOCALE", data);
	if (0 == ret)
		strncpy(outBuffer, "/system/vendor/intel/dal", outBufferSize);
	else
		strncpy(outBuffer, data, outBufferSize);

	return JHI_SUCCESS;
}

JHI_RET_I
JhiQueryTransportTypeFromRegistry(uint32_t* transportType)
{
	char data[PROP_VALUE_MAX];
	int ret = __system_property_get("persist.jhi.TRANSPORT_TYPE", data);
	if (0 == ret)
		*transportType = 2; //TEE_TRANSPORT_TYPE_TEE_LIB
	else
		*transportType = atoi(data);

	return JHI_SUCCESS;
}

JHI_RET_I
JhiQueryServiceFileLocationFromRegistry (char* outBuffer, uint32_t outBufferSize)
{
	char data[PROP_VALUE_MAX];
	int ret = __system_property_get("persist.jhi.FILE_LOCALE", data);
	if (0 == ret)
		strncpy(outBuffer, "/system/bin", outBufferSize);
	else
		strncpy(outBuffer, data, outBufferSize);

	return JHI_SUCCESS;
}

JHI_RET_I
JhiQueryPluginLocationFromRegistry(FILECHAR* outBuffer, uint32_t outBufferSize) {
	char data[PROP_VALUE_MAX];
	int ret = __system_property_get("persist.jhi.PLUGIN_LOCALE", data);
	if (0 == ret)
		strncpy(outBuffer, "/system/vendor/intel/dal/lib", outBufferSize);
	else
		strncpy(outBuffer, data, outBufferSize);
	return JHI_SUCCESS;
}
JHI_RET_I
JhiQueryEventSocketsLocationFromRegistry(FILECHAR* outBuffer, uint32_t outBufferSize) {
	char data[PROP_VALUE_MAX];
	int ret = __system_property_get("persist.jhi.EVENT_LOCALE", data);
	if (0 == ret)
		strncpy(outBuffer, "/data/intel/dal/dynamic_sockets", outBufferSize);
	else
		strncpy(outBuffer, data, outBufferSize);
	return JHI_SUCCESS;
}
JHI_RET_I
JhiQueryServicePortFromRegistry(uint32_t* portNumber)
{
	char data[PROP_VALUE_MAX];
	int ret = __system_property_get("persist.jhi.SERVICE_PORT", data);
	if (0 == ret)
		*portNumber = 49176;
	else
		*portNumber = atoi(data);

	return JHI_SUCCESS;
}


JHI_RET_I
JhiQueryAddressTypeFromRegistry(uint32_t* addressType)
{
	char data[PROP_VALUE_MAX];
	int ret = __system_property_get("persist.jhi.ADDRESS_TYPE", data);
	if (0 == ret)
		*addressType = AF_INET;
	else
		*addressType = atoi(data);

	return JHI_SUCCESS;
}

JHI_RET_I
JhiQueryLogLevelFromRegistry(JHI_LOG_LEVEL *loglevel)
{
        *loglevel = JHI_LOG_LEVEL_DEBUG;
        return JHI_SUCCESS;
}

JHI_RET_I
JhiWritePortNumberToRegistry(uint32_t portNumber)
{
	return JHI_SUCCESS;
}

JHI_RET_I
JhiWriteAddressTypeToRegistry(uint32_t addressType)
{
	return JHI_SUCCESS;
}

JHI_RET_I
RestartJhiService ()
{
	int ret = __system_property_set ("persist.service.jhi.enable", "0");
	if (ret == JHI_SUCCESS)
		ret = __system_property_set ("persist.service.jhi.enable", "1");
	return ret;
}
