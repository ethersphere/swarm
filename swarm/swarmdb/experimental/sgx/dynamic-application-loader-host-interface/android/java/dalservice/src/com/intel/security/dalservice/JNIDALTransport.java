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

package com.intel.security.dalservice;
/**
 * @author novsyani
 *
 */

import android.os.ParcelFileDescriptor;

import  com.intel.security.dalinterface.IDALServiceCallbackListener;

public class JNIDALTransport extends java.lang.Object{
	static {
		System.load("libclientjni_jhi.so");
	}

	public native int DAL_CreateSession(
			String AppId,
			int AppPID,
			int flags,
			byte[] initBuffer,
			long[] SessionHandle
			);

	public native int DAL_CloseSession(
			long SessionHandle
			);

	public native int DAL_SendAndRecv(
			long SessionHandle,
			int nCommandId,
			byte[] CommTx,
			byte[] CommRx,
			int [] rxCount,
			int[] responseCode
			);

	public native int DAL_RegisterEvents(
			long SessionHandle
			);

	public native int DAL_UnregisterEvents(
			long SessionHandle
			);

	public native int DAL_SHMemTxRxTrans(
			long SessionHandle,
			int nCommandId,
			int rfd,
			int txLength,
			int[] rxLength,
			int[] responseCode
			);
	public native int DAL_ClearSessions (
			int AppPID
			);
}
