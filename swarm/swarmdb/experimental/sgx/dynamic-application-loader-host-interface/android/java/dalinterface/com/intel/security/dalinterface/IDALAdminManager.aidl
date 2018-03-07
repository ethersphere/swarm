/*
 * Copyright 2010-2016 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/**
 * System-private Admin API for communicating with DAL.
 *
 * A service can only have a single interface (AIDL) so this file will
 * contain all the API for all libraries that access JHI and are
 * linked to the DAL Admin Service.
 *
 * {@hide}
 */

package com.intel.security.dalinterface;
import android.os.ParcelFileDescriptor;
import com.intel.security.dalinterface.DALVersionInfo;

interface IDALAdminManager {

    int DAL_Install (
        in String   AppId,
        in String   AppPath
    );

    int DAL_Install_FD (
        in String   AppId,
        in ParcelFileDescriptor AppFd,
        in int AppSize
    );

    int DAL_Uninstall (
        in String   AppId
    );

    DALVersionInfo DAL_GetVersionInfo (
        out int[] ret
    );

}
