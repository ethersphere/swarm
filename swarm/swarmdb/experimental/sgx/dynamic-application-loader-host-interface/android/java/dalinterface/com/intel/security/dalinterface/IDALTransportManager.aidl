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
 * System-private Client API for communicating with DAL.
 *
 * A service can only have a single interface (AIDL) so this file will
 * contain all the API for all libraries that access DAL and are
 * linked to the DAl Transport service.
 * {@hide}
 */

package com.intel.security.dalinterface;
import com.intel.security.dalinterface.IDALServiceCallbackListener;
import android.os.ParcelFileDescriptor;

interface IDALTransportManager {

    int DAL_CreateSession (
	in String AppId,
	in int flags,
	in byte[] initBuffer,
	out long[] SessionHandle
    );

    int DAL_CloseSession(
	in long SessionHandle
    );

    int DAL_SendAndRecv(
	in long SessionHandle,
	in int nCommandId,
	in byte[] CommTx,
	out byte[] CommRx,
	out int[] rxBuffLength,
	out int[] responseCode
    );

    int DAL_SHMemTxRxTrans(
	in long SessionHandle,
	in int nCommandId,
	in ParcelFileDescriptor mfd,
	in int txLength,
	inout int[] rxLength,
	out int[] responseCode);

    int DAL_RegisterEvents(
	in long SessionHandle,
	in IDALServiceCallbackListener listener);

    int DAL_UnregisterEvents(
	in long SessionHandle
    );
    int DAL_ClearSessions();


}
