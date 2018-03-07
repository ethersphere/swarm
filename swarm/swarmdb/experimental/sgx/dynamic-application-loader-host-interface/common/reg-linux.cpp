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
**    @file reg-linux.cpp
**
**    @brief  Defines "registry" related functions
**
********************************************************************************
*/

#include <string>
#include <sstream>
#include <map>
#include <fstream>
#include <stdio.h>
#include <sys/socket.h>
#include <stdlib.h>
#include <Beihai/bhp/impl/bhp_platform.h>
#include "reg.h"
#include "misc.h"
#include "teetransport.h"
#include "Singleton.h"
using namespace std;

#define CONFIG_FILE_PATH "/etc/jhi/jhi.conf"

class ConfigFile : public intel_dal::Singleton<ConfigFile>
{
public:
    static TEE_TRANSPORT_TYPE getTransportType()
    {
        ConfigFile &config = ConfigFile::Instance();
        string s_transport = "MEI";
        TEE_TRANSPORT_TYPE t_transport = TEE_TRANSPORT_TYPE_TEE_LIB;

        map<string, string>::iterator it = config.settings.find("transport");
        if(it != config.settings.end())
            s_transport = it->second;

        if      (s_transport == "SOCKET")
            t_transport = TEE_TRANSPORT_TYPE_SOCKET;
        else if (s_transport == "MEI")
            t_transport = TEE_TRANSPORT_TYPE_TEE_LIB;
        else if (s_transport == "KERNEL")
            t_transport = TEE_TRANSPORT_TYPE_DAL_DEVICE;

        TRACE1("Using transport type: %d", t_transport);

        return t_transport;
    }

    static string getIpAddress()
    {
        ConfigFile &config = ConfigFile::Instance();

        string ip = "127.0.0.1";
        map<string, string>::iterator it = config.settings.find("socket_ip_address");
        if(it != config.settings.end())
            ip = it->second;

        return ip;
    }

	static JHI_LOG_LEVEL getLogLevel()
	{
		ConfigFile &config = ConfigFile::Instance();
		string s_loglevel = "RELEASE";
		JHI_LOG_LEVEL e_loglevel = JHI_LOG_LEVEL::JHI_LOG_LEVEL_RELEASE;

		map<string, string>::iterator it = config.settings.find("log_level");
		if(it != config.settings.end())
			s_loglevel = it->second;

		if     (s_loglevel == "OFF")
			e_loglevel = JHI_LOG_LEVEL_OFF;
		else if(s_loglevel == "RELEASE")
			e_loglevel = JHI_LOG_LEVEL_RELEASE;
		else if(s_loglevel == "DEBUG")
			e_loglevel = JHI_LOG_LEVEL_DEBUG;

		return e_loglevel;
	}

	static string getDaemonSocketPath()
	{
		ConfigFile &config = ConfigFile::Instance();
		string path = "/tmp/jhi_socket";

		map<string, string>::iterator it = config.settings.find("socket_path");
		if(it != config.settings.end())
			path = it->second;

		//LOG1("Daemon socket path: %s", path.c_str());

		return path;
	}

    map<string, string> settings;
private:
    // This allows only the Singleton template to instantiate ConfigFile
    friend class Singleton<ConfigFile>;

    ConfigFile()
    {
        ifstream config_file(CONFIG_FILE_PATH);
        
        if(!config_file.is_open())
            TRACE1("Config file not found. Using defaults. Path tried: %s", CONFIG_FILE_PATH);
		else
		{
			string line;

			while (getline(config_file, line))
			{
				if (line[0] != '#')
				{
					string key;
					string value;
					stringstream ss(line);
					ss >> key >> value;

					settings[key] = value;
				}
			}

			config_file.close();
		}
    }
};

JHI_RET_I
JhiQueryAppFileLocationFromRegistry (char* outBuffer, uint32_t outBufferSize)
{
    strncpy(outBuffer, "/var/lib/intel/dal/applet_repository", outBufferSize);
    return JHI_SUCCESS;
}

JHI_RET_I
JhiQuerySpoolerLocationFromRegistry (char* outBuffer, uint32_t outBufferSize)
{
    strncpy(outBuffer, "/var/lib/intel/dal/applets", outBufferSize);
    return JHI_SUCCESS;
}

JHI_RET_I
JhiQueryTransportTypeFromRegistry(uint32_t* transportType)
{
    *transportType = ConfigFile::getTransportType();
    return JHI_SUCCESS;
}

JHI_RET_I
JhiQuerySocketIpAddressFromRegistry(char *ip)
{
    string s_ip = ConfigFile::getIpAddress();
    strcpy(ip, s_ip.c_str());
    return JHI_SUCCESS;
}

JHI_RET_I
JhiQueryLogLevelFromRegistry(JHI_LOG_LEVEL *loglevel)
{
	*loglevel = ConfigFile::getLogLevel();
	return JHI_SUCCESS;
}

JHI_RET_I
JhiQueryDaemonSocketPathFromRegistry(char * path)
{
	string s_path = ConfigFile::getDaemonSocketPath();
	strcpy (path, s_path.c_str());
	return JHI_SUCCESS;
}

// Not being used?
JHI_RET_I
JhiQueryServiceFileLocationFromRegistry (char* outBuffer, uint32_t outBufferSize)
{
    strncpy(outBuffer, "/usr/sbin", outBufferSize);
    return JHI_SUCCESS;
}

// Not being used?
JHI_RET_I
JhiQueryPluginLocationFromRegistry(FILECHAR* outBuffer, uint32_t outBufferSize)
{
    strncpy(outBuffer, "/usr/lib64", outBufferSize);
    return JHI_SUCCESS;
}

// Not being used?
JHI_RET_I
JhiQueryEventSocketsLocationFromRegistry(FILECHAR* outBuffer, uint32_t outBufferSize)
{
    strncpy(outBuffer, "/data/intel/dal/dynamic_sockets", outBufferSize);
    return JHI_SUCCESS;
}

// Not being used?
JHI_RET_I
JhiQueryServicePortFromRegistry(uint32_t* portNumber)
{
    *portNumber = 49176;
    return JHI_SUCCESS;
}

// Not being used?
JHI_RET_I
JhiQueryAddressTypeFromRegistry(uint32_t* addressType)
{
    *addressType = AF_INET;
    return JHI_SUCCESS;
}

// Not being used?
JHI_RET_I
JhiWritePortNumberToRegistry(uint32_t portNumber)
{
    //TODO: Make this dynamic
    //      This doesn't work because setenv can't change global env
    if(setenv("JHI_SERVICE_PORT", std::to_string(portNumber).c_str(), true) != 0)
        TRACE1("Error: setenv(\"JHI_SERVICE_PORT\", ...) failed. Error code: %d", errno);
    return JHI_SUCCESS;
}

// Not being used?
JHI_RET_I
JhiWriteAddressTypeToRegistry(uint32_t addressType)
{
    //TODO: Make this dynamic
    //      This doesn't work because setenv can't change global env
    if(setenv("JHI_ADDRESS_TYPE", std::to_string(addressType).c_str(), true) != 0)
        TRACE1("Error: setenv(\"JHI_ADDRESS_TYPE\", ...) failed. Error code: %d", errno);
    return JHI_SUCCESS;
}

