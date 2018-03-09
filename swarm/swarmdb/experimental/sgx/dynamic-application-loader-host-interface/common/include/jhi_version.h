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
**    @file jhi_version.h
**
**    @brief Contains definitions of the MAJOR, MINOR, & HOTFIX versions used
**           throughout the code.
**
**    @author Christopher Spiegel and John Traver
**
********************************************************************************
*/   

/* Sentry Header
 *****************/
#ifndef _JHI_VERSION_H_
#define _JHI_VERSION_H_

/* Include Files
 *****************/
#include "version.h"

#ifdef __cplusplus
extern "C" {
#endif

/* Global Declarations
 **************************/

#define VER_SEPARATOR_STR "."   /**< Separator String used between major, minor, hotfix and build strings */

/* Flags set based on build type */
#if DBG
/** Define "DBG" string that is attached to version string when DBG build */
#define VER_DEBUG_TAG   " (DBG)"
#else
/** Define string that is attached to version string when release build */
#define VER_DEBUG_TAG   
#endif


/** Combined file version information  */
#define VERSION_STR VER_MAJOR_STR VER_SEPARATOR_STR VER_MINOR_STR VER_SEPARATOR_STR VER_HOTFIX_STR VER_SEPARATOR_STR VER_BUILD_STR VER_DEBUG_TAG

/** @brief Defines a generic version structure used in the software build process. This structure will be used to
*   represent versions of ROM, FW and Recovery modules.
*/
typedef struct _VERSION
{
   UINT16      Major;
   UINT16      Minor;
   UINT16      Hotfix;
   UINT16      Build;
}VERSION;

#ifdef __cplusplus
}
#endif

/* Global Functions
 **************************/

#endif // _VER_H_

