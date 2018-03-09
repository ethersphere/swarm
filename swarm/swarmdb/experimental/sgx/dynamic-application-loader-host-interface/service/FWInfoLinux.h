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
**    @file FWInfoLinux.h
**
**    @brief  Contains implementation for IFirmwareInfo for linux.
**
**    @author Alexander Usyskin
**
********************************************************************************
*/
#ifndef _FW_INFO_LINUX_H_
#define _FW_INFO_LINUX_H_

#include "IFirmwareInfo.h"
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

#define FWINFO_FW_COMMS_TIMEOUT 100000

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

	class FWInfoLinux : public IFirmwareInfo
	{
	public:
		FWInfoLinux();
		virtual ~FWInfoLinux();
		virtual bool GetFwVersion(VERSION* fw_version);
		virtual bool Connect();
		virtual bool Disconnect();

	private:
		int _heciFd;
		bool _isConnected;
		int _connectionAttemptNum;

	private:
		int HeciRead(uint8_t *buffer, size_t len, int *bytesRead);
		int HeciWrite(const uint8_t *buffer, size_t len, unsigned long timeout);
	};

}

#endif
