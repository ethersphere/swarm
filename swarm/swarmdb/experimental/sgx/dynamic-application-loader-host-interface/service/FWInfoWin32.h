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
**    @file FWInfoWin32.h
**
**    @brief  Contains implementation for IFirmwareInfo using FU client.
**
**    @author Elad Dabool
**
********************************************************************************
*/
#ifndef _FW_INFO_WIN32_H_
#define _FW_INFO_WIN32_H_

#include "IFirmwareInfo.h"

#include <Windows.h>
#include <setupapi.h>
#include <initguid.h>
#include <WinIoCtl.h>
#include "MkhiMsgs.h"

namespace intel_dal
{
	typedef enum
	{
	   MKHI_CBM_GROUP_ID = 0,
	   MKHI_PM_GROUP_ID,
	   MKHI_PWD_GROUP_ID,
	   MKHI_FWCAPS_GROUP_ID,
	   MKHI_APP_GROUP_ID,      // Reserved (no longer used).
	   MKHI_FWUPDATE_GROUP_ID, // This is for manufacturing downgrade
	   MKHI_FIRMWARE_UPDATE_GROUP_ID,
	   MKHI_BIST_GROUP_ID,
	   MKHI_MDES_GROUP_ID,
	   MKHI_ME_DBG_GROUP_ID,
	   MKHI_MAX_GROUP_ID,
	   MKHI_GEN_GROUP_ID = 0xFF
	}MKHI_GROUP_ID;

	#define FWCAPS_GET_RULE_CMD            0x02
	#define FWCAPS_GET_RULE_CMD_ACK        0x82

	#define ME_RULE_FEATURE_ID                           0
	#define MEFWCAPS_PCV_OEM_PLAT_TYPE_CFG_RULE          29

	typedef union _RULE_ID
	{
	   UINT32      Data;
	   struct
	   {
		  UINT32   RuleTypeId     :16;
		  UINT32   FeatureId      :8;
		  UINT32   Reserved       :8;
	   }Fields;
	}RULE_ID;

	typedef struct _GET_RULE_DATA
	{
	   RULE_ID  RuleId;
	}GET_RULE_DATA;

	typedef struct _GET_RULE_ACK_DATA
	{
	   RULE_ID  RuleId;
	   UINT8    RuleDataLen;
	   UINT8    RuleData[0];
	}GET_RULE_ACK_DATA;

	typedef struct _FWCAPS_GET_RULE
	{
	   MKHI_MESSAGE_HEADER     Header;
	   GET_RULE_DATA           Data;
	}FWCAPS_GET_RULE;

	typedef struct _FWCAPS_GET_RULE_ACK
	{
	   MKHI_MESSAGE_HEADER     Header;
	   GET_RULE_ACK_DATA       Data;
	}FWCAPS_GET_RULE_ACK;

	// HECI GUID(Global Unique ID): {E2D1FF34-3458-49A9-88DA-8E6915CE9BE5}
	DEFINE_GUID(GUID_DEVINTERFACE_HECI, 0xE2D1FF34, 0x3458, 0x49A9,
	  0x88, 0xDA, 0x8E, 0x69, 0x15, 0xCE, 0x9B, 0xE5);

	// HCI HECI DYNAMIC CLIENT GUID
	DEFINE_GUID(HCI_HECI_DYNAMIC_CLIENT_GUID,0x8e6a6715, 0x9abc, 0x4043, 0x88, 0xef, 0x9e, 0x39, 0xc6, 0xf6, 0x3e, 0xf);

	#define FILE_DEVICE_HECI  0x8000

	// for connecting to HECI client in the FW(e.g LME)
	#define IOCTL_HECI_CONNECT_CLIENT \
		CTL_CODE(FILE_DEVICE_HECI, 0x801, METHOD_BUFFERED, FILE_READ_ACCESS|FILE_WRITE_ACCESS)

	typedef struct _HECI_CLIENT_PROPERTIES
	{  
	   BYTE  ProtocolVersion;
	   DWORD  MaxMessageSize;
	} HECI_CLIENT_PROPERTIES;

	#pragma pack(1)
	typedef struct _HECI_CLIENT
	{
		UINT32                  MaxMessageLength;
		UINT8                   ProtocolVersion;
	} HECI_CLIENT;
	#pragma pack()

	class FWInfoWin32 : public IFirmwareInfo
	{
	public:
	   FWInfoWin32();
	   ~FWInfoWin32();
	   bool GetFwVersion(VERSION* fw_version);
	   bool Connect();
	   bool Disconnect();
	   
	   bool GetHeciDeviceDetail(WCHAR *DevicePath);
	   HANDLE GetHandle(WCHAR *DevicePath);

	private:
		bool isconnected;
		int ConnectionAttemptNum;

		HANDLE hDevice;
		int MAX_BUFFER_SIZE;

		bool HeciConnectHCI( HANDLE * pHandle,HECI_CLIENT_PROPERTIES * pProperties);
		bool HeciWrite(HANDLE Handle, void * pData, DWORD DataSize, DWORD msTimeous);
		bool HeciRead(HANDLE Handle, void * pBuffer, DWORD BufferSize, DWORD * pBytesRead, DWORD msTimeous);

		bool SendGetFwVersionRequest();
		bool ReceiveGetFwVersionResponse(VERSION* fw_version);
	};

}

#endif //_FW_INFO_WIN32_H_