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

#include "FWInfoWin32.h"
#include "dbg.h"

#define _CRT_RAND_S
#include <stdlib.h>
#include "misc.h"
#include <cfgmgr32.h>
#include <Objbase.h>
#include <Devpkey.h>
#include <Strsafe.h>

namespace intel_dal
{

	FWInfoWin32::FWInfoWin32()
	{
		isconnected = false;
		ConnectionAttemptNum = 0;
		MAX_BUFFER_SIZE = 0;
		hDevice = INVALID_HANDLE_VALUE;
	}

	FWInfoWin32::~FWInfoWin32()
	{
		if (isconnected)
			Disconnect();
	}

	bool FWInfoWin32::Connect()
	{
		bool status = false;
		ConnectionAttemptNum++;
		HECI_CLIENT_PROPERTIES pProperties;
		WCHAR DevicePath[256] = { 0 };
		unsigned int randomValue = 0;

		do
		{
			if (ConnectionAttemptNum>1)
			{
				// after first try we must wait a random time
				// Sleep randomly between 100ms and 300ms 
				rand_s(&randomValue);
				Sleep((randomValue % 201) + 100);
			}

			// connect to heci
			if (!GetHeciDeviceDetail(&DevicePath[0]))
				break;

			hDevice = GetHandle(DevicePath); //get handle from HECI driver

			if (hDevice == INVALID_HANDLE_VALUE)
				break;

			// connect to HCI client
			if(!HeciConnectHCI(&hDevice, &pProperties))
				break;

			isconnected = true;
			status = true;

		}
		while (0);

		return status;
	}

	bool FWInfoWin32::GetFwVersion(VERSION* fw_version)
	{
		if (!isconnected)
			return false;

		if (!SendGetFwVersionRequest())  
			return false;

		if(!ReceiveGetFwVersionResponse(fw_version)) 
			return false;

		return true;
	}

	bool FWInfoWin32::Disconnect()
	{
		if (CloseHandle(hDevice)== 0 )
			return false;

		isconnected = false;
		return true;
	}

	bool FWInfoWin32::SendGetFwVersionRequest()
	{
		DWORD LastError;

		GEN_GET_FW_VERSION request;

		request.Header.Fields.Command = GEN_GET_FW_VERSION_CMD;
		request.Header.Fields.GroupId = MKHI_GEN_GROUP_ID;
		request.Header.Fields.IsResponse = 0;

		if(HeciWrite(hDevice, (BYTE*)&request , sizeof(request), INFINITE))		
		{
			TRACE0("Sent FWU_GET_VERSION to HECI.\n");
			return true;
		}
		else
		{		
			TRACE0("Error: sending FWU_GET_VERSION request to HECI failed.\n");
			LastError = GetLastError();
			TRACE1("error: %d",LastError);	
			return false;
		}
	}

	bool FWInfoWin32::ReceiveGetFwVersionResponse(VERSION* fw_version)
	{
		bool status = false;
		BYTE *HECIReply = NULL;
		DWORD BytesRead = 0; // get number bytes that were actually read	
		GEN_GET_FW_VERSION_ACK *ResponseMessage; // struct to analyzed the response

		do
		{
			if (fw_version == NULL)
				break;

			if (MAX_BUFFER_SIZE == 0)
				break;

			HECIReply = (BYTE*) JHI_ALLOC(MAX_BUFFER_SIZE); // buffer to receive the response
			if (HECIReply == NULL)
			{
				break;
			}

			memset(HECIReply,0,MAX_BUFFER_SIZE); 

			// if HeciRead succeed
			if(HeciRead(hDevice, HECIReply, MAX_BUFFER_SIZE, &BytesRead, INFINITE))		
			{
				TRACE1("Number bytes read from HECI: %d\n",BytesRead);

				// analyze response messgae
				ResponseMessage = (GEN_GET_FW_VERSION_ACK*) HECIReply;

				if ( ResponseMessage->Header.Fields.Result != ME_SUCCESS )
				{
					TRACE0("Got error status from HCI_GET_FW_VERSION.\n");
					break;
				}

				fw_version->Major = ResponseMessage->Data.FWVersion.CodeMajor;
				fw_version->Minor = ResponseMessage->Data.FWVersion.CodeMinor;
				fw_version->Hotfix = ResponseMessage->Data.FWVersion.CodeHotFix;
				fw_version->Build = ResponseMessage->Data.FWVersion.CodeBuildNo;

				status = true;
			}
			else
			{
				TRACE1("HeciRead Error. LastError = %d\n",GetLastError());
				break;
			}

		}
		while(0);

		// cleanup
		if (HECIReply != NULL)
		{
			JHI_DEALLOC(HECIReply);
			HECIReply = NULL;
		}

		return status;
	}

	// ******************************  HECI HELPER FUNCTIONS ******************************************//

	////////////////////////////////////////////////////////////////////////////////////
	// GetHeciDeviceDetail:                                                           //
	//  Get HECI device details (by Windows API)                                      //
	//                                                                                //
	// Input:                                                                         //
	//  DeviceDetail structure.                                                       //
	//                                                                                //
	// Output:                                                                        //
	//  Device details in DeviceDetail structure.                                     //
	////////////////////////////////////////////////////////////////////////////////////
	bool FWInfoWin32::GetHeciDeviceDetail(WCHAR *DevicePath)
	{
		DWORD Status = ERROR_SUCCESS;
		CONFIGRET cr = CR_SUCCESS;
		PWSTR deviceInterfaceList = NULL;
		ULONG deviceInterfaceListLength = 0;
		PWSTR nextInterface;
		HRESULT hr = E_FAIL;
		size_t BufLen = 256;
		HANDLE   DeviceHandle = INVALID_HANDLE_VALUE;

		cr = CM_Get_Device_Interface_List_Size(&deviceInterfaceListLength, (LPGUID)&GUID_DEVINTERFACE_HECI, NULL, CM_GET_DEVICE_INTERFACE_LIST_PRESENT);
		if (cr != CR_SUCCESS)
		{
			TRACE1("Error 0x%x retrieving device interface list size.", cr);
			return false;
		}

		if (deviceInterfaceListLength <= 1)
		{
			TRACE0("Error: No active device interfaces found.");
			return false;
		}

		deviceInterfaceList = (PWSTR)malloc(deviceInterfaceListLength * sizeof(WCHAR));
		if (deviceInterfaceList == NULL)
		{
			TRACE0("Error allocating memory for device interface list.");
			return false;
		}

		ZeroMemory(deviceInterfaceList, deviceInterfaceListLength * sizeof(WCHAR));

		cr = CM_Get_Device_Interface_List((LPGUID)&GUID_DEVINTERFACE_HECI, NULL, deviceInterfaceList, deviceInterfaceListLength, CM_GET_DEVICE_INTERFACE_LIST_PRESENT);
		if (cr != CR_SUCCESS)
		{
			TRACE1("Error 0x%x retrieving device interface list.\n", cr);
			return false;
		}

		nextInterface = deviceInterfaceList + wcslen(deviceInterfaceList) + 1;
		if (*nextInterface != UNICODE_NULL)
		{
			TRACE0("Warning: More than one device interface instance found. Selecting first matching device.");
		}

		hr = StringCchCopy(DevicePath, BufLen, deviceInterfaceList);

		if (FAILED(hr))
		{
			TRACE1("Error: StringCchCopy failed with HRESULT 0x%x", hr);
			return false;
		}


		return true;
	}

	////////////////////////////////////////////////////////////////////////////////////
	// GetHandle:                                                                     //
	//  Get driver handle using fixed parameters.                                     //
	//                                                                                //
	// Input:                                                                         //
	//  DeviceDetail - structure that hold device details.                            //
	//                                                                                //
	// Output:                                                                        //
	//  return driver handle (valid/invalid handle).                                  //
	////////////////////////////////////////////////////////////////////////////////////
	HANDLE FWInfoWin32::GetHandle(WCHAR *DevicePath) {
		DWORD LastError;
		HANDLE hDevice;
		//print device path
		//wcout << L"DeviceDetail->DevicePath " << DeviceDetail->DevicePath;
		//cout << endl;

		// CreateFile function returns a handle that can be used to access an object.
		// CreateFile(FileName,DesiredAccess,ShareMode,SecurityAttributes,CreationDisposition,FlagsAndAttributes,TemplateFile}
		// GENERIC_READ | GENERIC_WRITE : ask device for READ & WRITE
		hDevice = CreateFile(DevicePath, GENERIC_READ | GENERIC_WRITE,
			FILE_SHARE_READ | FILE_SHARE_WRITE, NULL, OPEN_EXISTING, 0, NULL );

		//free( DeviceDetail );

		// if getting handle failed
		if( hDevice == INVALID_HANDLE_VALUE )
		{
			LastError = GetLastError();
			TRACE1("error: %d",LastError);
			return INVALID_HANDLE_VALUE;
		}

		// if getting handle succeed return the handle
		return  hDevice; 
	}

	// Connect to HCI client
	bool FWInfoWin32::HeciConnectHCI( HANDLE * pHandle,HECI_CLIENT_PROPERTIES * pProperties)
	{
		DWORD PropertiesSize;
		HECI_CLIENT DrvClientProp;
		BYTE ProtocolId[16];
		DWORD LastError;

		// Connect to HCI client in the FW
		memcpy_s( ProtocolId,sizeof(ProtocolId), (PVOID)&HCI_HECI_DYNAMIC_CLIENT_GUID, sizeof(ProtocolId) );
		if( DeviceIoControl( *pHandle, IOCTL_HECI_CONNECT_CLIENT,ProtocolId, 16, 
			&DrvClientProp, sizeof(DrvClientProp), &PropertiesSize, NULL ) == 0 )
		{
			TRACE0("Error in HeciConnectHCI.DeviceIoControl:\n");
			LastError = GetLastError();
			TRACE1("error: %d",LastError);
			return false;
		}

		// if actually return Properties Size (PropertiesSize) differ from expected size (DrvClientProp)
		if( PropertiesSize != sizeof(DrvClientProp) ) {
			TRACE0("In HeciConnectHCI: return PropertiesSize != expected size (DrvClientProp)\n");
			return false;
		}

		// if HECI successfully connected to HCI client		
		TRACE1("DrvClientProp.MaxMessageLength = %d\n",DrvClientProp.MaxMessageLength);
		pProperties->MaxMessageSize  = DrvClientProp.MaxMessageLength;
		MAX_BUFFER_SIZE=DrvClientProp.MaxMessageLength;
		pProperties->ProtocolVersion = DrvClientProp.ProtocolVersion;
		return true;
	}

	//////////////////////////////////////////////////////////////////////////////
	// Write to HECI driver														//
	// Parameters:																//
	//  Handle - HECI driver handle												//
	//  pData - data to be write												//
	//  DataSize - size of the data												//
	//  msTimeous - timeout														//
	//////////////////////////////////////////////////////////////////////////////
	bool FWInfoWin32::HeciWrite(HANDLE Handle, void * pData, DWORD DataSize, DWORD msTimeous)
	{
		DWORD BytesWritten;

		if( msTimeous != INFINITE )
		{
			return false;
		}

		if( WriteFile( Handle, pData, DataSize, &BytesWritten, NULL ) == 0 )
			return false;

		return true;
	}

	//////////////////////////////////////////////////////////////////////////////
	// HeciRead:                                                                //
	//  Read from HECI driver.                                                  //
	//                                                                          //
	// Input:                                                                   //
	//  Handle - HECI driver handle.                                            //
	//  pBuffer - buffer to received the data responsed.                        //
	//  BufferSize - size of the buffer.                                        //
	//  pBytesRead - number bytes that actualy read.                            //
	//  msTimeous - timeout.                                                    //
	//                                                                          //
	// Output:                                                                  //
	//  boolean value that represent reading success.                           //
	//////////////////////////////////////////////////////////////////////////////
	bool FWInfoWin32::HeciRead(HANDLE Handle, void * pBuffer, DWORD BufferSize, DWORD * pBytesRead, DWORD msTimeous)
	{
		if( msTimeous != INFINITE )
		{
			return false;
		}

		if( ReadFile( Handle, pBuffer, BufferSize, pBytesRead, NULL ) == 0 )
			return false;

		return true;
	}
}