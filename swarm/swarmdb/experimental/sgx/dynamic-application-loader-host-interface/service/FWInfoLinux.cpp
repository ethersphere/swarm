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
**    @file FWInfoLinux.cpp
**
**    @brief  Contains implementation for IFirmwareInfo for linux.
**
**    @author Alexander Usyskin
**
********************************************************************************
*/

#include <errno.h>
#include <fcntl.h>
#include <sys/ioctl.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include "typedefs.h"
#include "typedefs_i.h"
#include "FWInfoLinux.h"
#include "dbg.h"
#include "misc.h"
#include "MkhiMsgs.h"

#pragma pack(1)
typedef struct _MEI_CLIENT {
  unsigned int  MaxMessageLength;
  unsigned char ProtocolVersion;
  unsigned char reserved[3];
} MEI_CLIENT;

typedef struct heci_ioctl_data
{
  union
  {
    uuid_le in_client_uuid;
    MEI_CLIENT out_client_properties;
  };
} heci_ioctl_data_t;

const uuid_le MEI_MKHIF = UUID_LE(0x8e6a6715, 0x9abc,0x4043, \
                                  0x88, 0xef, 0x9e, 0x39, 0xc6, 0xf6, 0x3e, 0xf);
#pragma pack()

/* IOCTL commands */
#undef IOCTL_HECI_CONNECT_CLIENT
#define HECI_IOCTL_TYPE 0x48
#define IOCTL_HECI_CONNECT_CLIENT \
  _IOWR(HECI_IOCTL_TYPE, 0x01, heci_ioctl_data_t)

#define MAX_BUFFER_SIZE 16384

#ifndef ARRAY_SIZE
#define ARRAY_SIZE(a) (sizeof (a) / sizeof ((a)[0]))
#endif
static inline const char *mei_default_device()
{
	static const char *devnode[] = {"/dev/mei0", "/dev/mei"};
	unsigned int i;

	for (i = 0; i < ARRAY_SIZE(devnode); i++) {
		if (access(devnode[i], F_OK) == 0)
			return devnode[i];
	}
	return NULL;
}

namespace intel_dal
{
  FWInfoLinux::FWInfoLinux(): _heciFd(-1), _isConnected(false), _connectionAttemptNum(0)
  {
  }

  FWInfoLinux::~FWInfoLinux()
  {
    if (_isConnected)
    {
      Disconnect();
    }
  }

  bool FWInfoLinux::GetFwVersion(VERSION* fw_version)
  {
    if (fw_version == NULL)
      return false;
    if (!_isConnected)
      return false;

    GEN_GET_FW_VERSION request;

    request.Header.Fields.Command = GEN_GET_FW_VERSION_CMD;
    request.Header.Fields.GroupId = MKHI_GEN_GROUP_ID;
    request.Header.Fields.IsResponse = 0;

    if(HeciWrite((uint8_t*)&request , sizeof(request), FWINFO_FW_COMMS_TIMEOUT))
    {
      return false;
    }

    uint8_t HECIReply[MAX_BUFFER_SIZE];
    int bytesRead = 0; // get number bytes that were actually read
    GEN_GET_FW_VERSION_ACK *ResponseMessage; // struct to analyzed the response

    memset(HECIReply, 0, MAX_BUFFER_SIZE);

                // if HeciRead succeed
    if(HeciRead(HECIReply, MAX_BUFFER_SIZE, &bytesRead))
    {
      return false;
    }

                // analyze response messgae
    ResponseMessage = (GEN_GET_FW_VERSION_ACK*) HECIReply;

    if ( ResponseMessage->Header.Fields.Result != ME_SUCCESS )
    {
      TRACE0("Got error status from HCI_GET_FW_VERSION.\n");
       return false;
    }

    fw_version->Major = ResponseMessage->Data.FWVersion.CodeMajor;
    fw_version->Minor = ResponseMessage->Data.FWVersion.CodeMinor;
    fw_version->Hotfix = ResponseMessage->Data.FWVersion.CodeHotFix;
    fw_version->Build = ResponseMessage->Data.FWVersion.CodeBuildNo;

    return true;
  }

  bool FWInfoLinux::Connect()
  {
    if (_isConnected)
      return true;
    _connectionAttemptNum++;

    if (_connectionAttemptNum > 1)
    {
      // after first try we must wait a random time
      // Sleep randomly between 100ms and 300ms
      usleep((rand() % 201) + 100);
    }

    _heciFd = open(mei_default_device(), O_RDWR);
    if (_heciFd == -1)
    {
      TRACE1("Failed to open device 0x%x", errno);
      return false;
    }

    uuid_le guid;
    memcpy(&guid, &MEI_MKHIF, sizeof(MEI_MKHIF));
    heci_ioctl_data_t client_connect;
    client_connect.in_client_uuid = guid;

    int result = ioctl(_heciFd, IOCTL_HECI_CONNECT_CLIENT, &client_connect);
    if (0 == result)
    {
      _isConnected = true;
    }
    else
    {
      TRACE1("Failed to connect to device 0x%x", result);
      close(_heciFd);
    }
    return (0 == result);
  }

  bool FWInfoLinux::Disconnect()
  {
    if (!_isConnected)
      return true;
    if (_heciFd != -1)
      return (close(_heciFd) == 0);
    else
      return true;
  }

  int FWInfoLinux::HeciRead(uint8_t *buffer, size_t len, int *bytesRead)
  {
    int rv = 0;
    *bytesRead = 0;
    rv = read(_heciFd, (void*)buffer, len);
    if (rv < 0)
    {
      TRACE2("Failed to read 0x%x 0x%x", rv, errno);
      return -1;
    }
    *bytesRead = rv;
    return 0;
  }

  int FWInfoLinux::HeciWrite(const uint8_t *buffer, size_t len, unsigned long timeout)
  {
    ssize_t rv = 0;
    fd_set set;
    struct timeval tv;

    tv.tv_sec =  timeout / 1000;
    tv.tv_usec =(timeout % 1000) * 1000000;

    rv = write(_heciFd, (void *)buffer, len);
    if (rv < 0)
    {
      TRACE2("Failed to write 0x%x 0x%x", rv, errno);
      return -1;
    }

    FD_ZERO(&set);
    FD_SET(_heciFd, &set);
    rv = select(_heciFd+1, &set, NULL, NULL, &tv);
    if (rv > 0 && FD_ISSET(_heciFd, &set))
    {
      return 0;
    }
    else if (rv == 0)
    {
      TRACE0("Failed to write (timeout)");
    }
    else //rv<0
    {
      TRACE1("Failed to write on select 0x%x", rv);
    }
    return -1;
  }

}//namespace intel_dal

