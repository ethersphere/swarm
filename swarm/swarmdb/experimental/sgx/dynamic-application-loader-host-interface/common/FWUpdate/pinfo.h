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

  pinfo.h

Abstract:

  Contains definition of the pinfo structure, pull from MkhiMsfs.h.

  
Author:

    Karl Cheng
  
--*/

#ifndef PINFO_H
#define PINFO_H

typedef struct _PINFO
{
   UINT32   PINFO_ID;      //value in EAX with EAX=1, version information
                           // (Type family, model and stepping ID)
   UINT32   PINFO_CP;      // read from MSR 0x36
   UINT32   PINFO_UG_0;    //read from brand string – 
                           // [7:0] = CPUID.0x8000003.EDX[31:24]  - e.g. ‘T’
                           // [15:8]= CPUID.0x8000004.EAX[7:0] – e.g. ‘7’
                           // [23:16]= CPUID.0x8000004.EAX[15:8] – e.g. ‘3’
                           // [31:24]= CPUID.0x8000004.EAX[23:16] – e.g. ‘0’
   UINT32   PINFO_UG_1;    // [7:0]= CPUID.0x8000004.EAX[31:24] – e.g. ‘0’
                           // [63:8]= CPUID.0x8000004.EBX[7:0] – e.g. ‘ ‘
}PINFO;

#endif
