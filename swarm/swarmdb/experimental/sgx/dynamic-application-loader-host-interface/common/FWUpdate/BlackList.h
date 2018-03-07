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

/*++

File Name:

   BlackList.h

Abstract:

   Contains data structures and constants used for FWU to black list FW version that is not
   downgradeable.
   
   
Authors:
   Key Phomsopha

**/
#ifndef _BLACK_LIST_H
#define _BLACK_LIST_H

#define BLACK_LIST_ENTRY_MAX     10


#pragma pack(1)
typedef struct    _BLACK_LIST_ENTRY
{
   UINT16   ExpressionType;
   UINT16   MinorVer;
   UINT16   HotfixVer1;
   UINT16   BuildVer1;
   UINT16   HotfixVer2;
   UINT16   BuildVer2;
}BLACK_LIST_ENTRY;



#pragma pack()

#endif //_BLACK_LIST_H



