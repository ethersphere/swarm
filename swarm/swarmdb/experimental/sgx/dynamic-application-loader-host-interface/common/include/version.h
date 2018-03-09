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

@file: version.h

--*/

#ifndef __VERSION_H__
#define __VERSION_H__

#include "Build.h"

#define VER_MAJOR		11
#define VER_MINOR		0
#define VER_HOTFIX		0

#define _MAKE_VER_STRING(maj, min, submin, bld)    #maj "." #min "." #submin "." #bld
#define MAKE_VER_STRING(maj, min, submin, bld)    _MAKE_VER_STRING(maj, min, submin, bld)

#define VER_PRODUCTVERSION          VER_MAJOR,VER_MINOR,VER_HOTFIX,VER_BUILD
#define VER_PRODUCTVERSION_STR      MAKE_VER_STRING(VER_MAJOR, VER_MINOR, VER_HOTFIX, VER_BUILD)

/**  Adding the current year date string for use in updating the year component of the 
***  displayed copyright message.
***  the string needs to be the full 4 char value without spaces or special characters.
***  For example:
***        "2012"
*/
#define CURRENT_YEAR_STRING   "2017"

#endif
