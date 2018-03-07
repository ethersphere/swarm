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

#include "FWInfoWin32Sockets.h"
#include "dbg.h"

namespace intel_dal
{

	bool FWInfoWin32Sockets::Connect()
	{
		MB_RETURN_STATUS status;
		// init the mailbox.
		MBMailBoxInit(&jhi_mailbox);

		// open the mailbox 
		status = MBOpenMailBox(&jhi_mailbox, JHI_MAILBOX_NAME, MB_MODE_READ);
		if (status != MB_STATUS_OK) {
			TRACE1("ERR: error opening mailbox status = %d\n", status);
			return false;
		}

		return true;
	}
	bool FWInfoWin32Sockets::Disconnect()
	{
		MB_RETURN_STATUS status;

		status = MBCloseMailBox(&jhi_mailbox);
		if (status != MB_STATUS_OK) {
			TRACE1("ERR: error closing mailbox status = %d\n", status);
			return false;
		}

		return true;
	}

	bool FWInfoWin32Sockets::GetFwVersion(VERSION* fw_version)
	{
		DWORD msgCount;
		MBMessage request_msg;
		MBMessage response_msg;
		MB_RETURN_STATUS status;

		// build the message.
		MBMessageBuild(&request_msg, "QRYREP", JHI_MAILBOX_NAME, DEVPLATFORM_MAILBOX_NAME, JHI_FW_VERSION_REQUEST);

		// send using single message
		status = MBSendSingleMessage(DEVPLATFORM_MAILBOX_NAME, request_msg);
		if (status != MB_STATUS_OK) {
			TRACE1("ERR: error sending message status = %d\n", status);
			return false;
		}

		TRACE0("Sleeping for 3 seconds before reading form mailbox");
		Sleep(3000);

		status = MBCheckMail(&jhi_mailbox, &msgCount);
		if (status != MB_STATUS_OK) {
			TRACE1("ERR: error creating mailbox status = %d\n", status);
			return false;
		}

		if (msgCount > 0)
		{
			status = MBReadNextMessage(&jhi_mailbox, &response_msg);
			if (status != MB_STATUS_OK) {
				TRACE1("ERR: error reading message, status = %d\n", status);
				return false;
			}

			if (sscanf_s(response_msg.data,"My FW Version is %hd.%hd.%hd (%hd)",&fw_version->Major,&fw_version->Minor,&fw_version->Hotfix,&fw_version->Build) != 4)
			{
				TRACE0("recieved invalid fw version format from devplatform\n");
				return false;
			}

		}
		else
		{
			printf("ERR: no response recieved from devplatform.\n");
			return false;
		}

		return true;
	}
}