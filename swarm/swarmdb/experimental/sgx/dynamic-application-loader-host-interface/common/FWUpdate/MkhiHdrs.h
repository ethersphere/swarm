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

/*
File Name:
   MkhiHdr.h
Abstract:
Authors:
   Tam Nguyen
*/

#ifndef _MKHI_HDRS_H
#define _MKHI_HDRS_H
#include "typedefs.h"

#ifdef _WIN32
#pragma warning (disable: 4214 4200)
#endif //_WIN32

#pragma pack(1)


//MKHI host message header. This header is part of HECI message sent from MEBx via
//Host Configuration Interface (HCI). ME Configuration Manager or Power Configuration
//Manager also include this header with appropriate fields set as part of the 
//response message to the HCI.
typedef union _MKHI_MESSAGE_HEADER
{
   UINT32     Data;
   struct
   {
      UINT32  GroupId     :8;
      UINT32  Command     :7;
      UINT32  IsResponse  :1;
      UINT32  Reserved    :8;
      UINT32  Result      :8;
   }Fields;
}MKHI_MESSAGE_HEADER;
C_ASSERT(sizeof(MKHI_MESSAGE_HEADER) == 4);

#pragma pack()

#endif //_MKHI_HDRS_H

