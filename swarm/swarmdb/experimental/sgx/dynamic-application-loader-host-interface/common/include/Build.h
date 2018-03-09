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
**    @file build.h
**
**    @brief Contains the global version strings
**
**    @author Christopher Spiegel
**
********************************************************************************
*/   

/* Sentry Header
 *******************/
#ifndef _BUILD_H_
#define _BUILD_H_

#ifdef __cplusplus
extern "C" {
#endif

/* Global Declarations
 **************************/

/** Build Version Number.  Updated by Build Server */
#define VER_BUILD		  3000
/** Build Version String.  Updated by Build Server */
#define VER_BUILD_STR     "3000"

#ifdef __cplusplus
}
#endif

#define POST_PV   0  // Set this flag to zero when are at PRE-PV and 1 when we are at POST_PV
                     // This flag will take care of including or excluding WW kill-pill check
                     // and turn on the update logic to prevent downgrade from POST_PV to PRE_PV

/*** Global Functions
 **************************/

#endif // _BUILD_H_

