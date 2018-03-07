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
 *
 * @file  bh_shared_conf.h
 * @brief This file declares the shared configuration parameters for entire Beihai
 *        system, including host and firmware part.
 * @author
 * @version
 *
 */
#ifndef __BH_SHARED_CONF_H
#define __BH_SHARED_CONF_H

#ifdef __cplusplus
extern "C" {
#endif

//Support SVM for SubSD feature or not
#define BEIHAI_ENABLE_SVM 0

//Support Native TA feature or not
#define BEIHAI_ENABLE_NATIVETA 0

//Support DAL OEM Siging for IoTG feature or not, which is exclusive to SVM
#define BEIHAI_ENABLE_OEM_SIGNING_IOTG 1
#ifdef __cplusplus
}
#endif

#endif
