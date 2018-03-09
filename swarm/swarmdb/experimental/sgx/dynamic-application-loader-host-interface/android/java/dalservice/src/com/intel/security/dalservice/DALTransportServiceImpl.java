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


import java.util.Hashtable;
import java.util.Map;
import java.lang.reflect.InvocationHandler;
import java.lang.reflect.Method;
import android.content.Context;
import android.content.pm.PackageManager;
import android.os.Binder;
import android.os.Process;
import android.os.RemoteException;
import android.os.ParcelFileDescriptor;
import android.util.Log;
import com.intel.security.dalinterface.IDALTransportManager;
import com.intel.security.dalinterface.IDALServiceCallbackListener;
import com.intel.security.dalinterface.DALCallback;
import com.intel.security.dalservice.JNIDALTransport;
import com.intel.security.dalinterface.DalConstants;
import com.intel.security.dalservice.JNIDALTransport;



final class DALTransportServiceImpl extends IDALTransportManager.Stub {
	private static final String TAG = DALTransportServiceImpl.class.getSimpleName();;
	private final Context context;
	private JNIDALTransport jhi;

	private static Hashtable hListeners = new Hashtable();

    //Public constructor is required to call Static callback method from native JNI
	public DALTransportServiceImpl ()
	{
		Log.d(TAG, "Public constructor");
		jhi = null;
	    this.context = null;
	}


	DALTransportServiceImpl (Context context)
	{
		Log.d(TAG, "Creating JNIDALTransport instance");
		jhi = new JNIDALTransport();
		this.context = context;
	}

	public final static void DALcallbackHandler(DALCallback response)
	{
		Log.d(TAG, "DALService Callback \n");

		if (response == null) {
			Log.e(TAG, "DALcallbackHandler - invalid callback");
			return;
		}


		Object listener = hListeners.get(Long.toString((response.getHandle())));
		if (listener == null) {
			Log.e(TAG, String.format("DALService Callback for session %d/%s",
	    					  response.getHandle(), Long.toString(response.getHandle())));
			return;
		}
		Class [] callbackResp = new Class [1];
		callbackResp[0] = DALCallback.class;
		Method onResponse = null;

		try {
			onResponse = listener.getClass().getDeclaredMethod ("onResponse", callbackResp[0]);
		} catch (Exception e) {
			Log.e(TAG, "DALService Callback Reflection error");
			return;
		}

		try {
			Object jVoid = onResponse.invoke(listener, response);
		} catch (Throwable e) {
			Log.e(TAG, "DALService Callback Execution error");
			return;
		}
		Log.d(TAG, "DALService Callback success\n");
	}

	@Override
	public int DAL_CreateSession(String AppId, int flags, byte[] initBuffer, long[] SessionHandle) throws RemoteException {
		int ret = DalConstants.DAL_INTERNAL_ERROR;

		Log.d(TAG, String.format("DALService CreateSession, App Id %s, APP PID %d, service PID %s TID %S",
					 AppId, Binder.getCallingPid(), Process.myPid(), Process.myTid()));
        // we validate APP_ID in SessionManager amd also in native,
		try {
      		ret = jhi.DAL_CreateSession(AppId, Binder.getCallingPid(), flags, initBuffer, SessionHandle);
		} catch (Exception e) {
			Log.e(TAG, "DALService CreateSession native exception");
			throw new RemoteException("DAL CreateSession failed");
		}

		return ret;
	}

	@Override
	public int DAL_CloseSession(long SessionHandle) throws RemoteException
	{
		int ret = DalConstants.DAL_INTERNAL_ERROR;

		Log.d(TAG, "DALService CloseSession");
		try {
			ret = jhi.DAL_CloseSession (SessionHandle);

		} catch (Exception e) {
			Log.e(TAG, "DALService CloseSession native exception");
			throw new RemoteException("DAL CloseSession failed");
		}
		finally
		{
			this.hListeners.remove(Long.toString(SessionHandle));
		}

		return ret;
	}

	@Override
	public int DAL_SendAndRecv(long SessionHandle,
			int nCommandId,
			byte[] CommTx,
			byte[] CommRx,
			int[] rxCont,
			int[] responseCode) throws RemoteException
			{
		int ret = DalConstants.DAL_INTERNAL_ERROR;

		Log.d(TAG, "DALService SendAndReceive");
        //input validation is executed in native jni
		try {
			ret = jhi.DAL_SendAndRecv(SessionHandle,
							nCommandId,
							CommTx, CommRx,
							rxCont, responseCode);
		} catch (Exception e) {
			Log.e(TAG, "DALService SendAndReceive native exception");
			throw new RemoteException ("DAL SendAndReceive failed");
		}
		return ret;
	}

	@Override
	public int DAL_RegisterEvents(long SessionHandle, IDALServiceCallbackListener listener) throws RemoteException
	{
		int ret = DalConstants.DAL_INTERNAL_ERROR;

		Log.d(TAG, "DALService RegisterEvents");

		if (listener == null) {
			Log.e(TAG, "DAL_RegisterEvents - invalid listener");
			return ret;
		}

		if (hListeners.containsKey(Long.toString(SessionHandle))) {
			Log.e(TAG, String.format("DALService Register for session %d/%s - already registered\n",
					SessionHandle, Long.toString(SessionHandle)));
		} else {
		//early callback registration to avoid first event receive race
			this.hListeners.put (Long.toString(SessionHandle), listener);
		}

		Log.d(TAG, String.format("DALService RegisterEvents for session %d/%s",
					SessionHandle, Long.toString(SessionHandle)));
		try {
			ret = jhi.DAL_RegisterEvents( SessionHandle);
		} catch (Exception e) {
			this.hListeners.remove(Long.toString(SessionHandle));
			Log.e(TAG, "DALService RegisterEvents native exception");
			throw new RemoteException ("DAL  RegisterEvents failed");
		}
		if (ret != DalConstants.DAL_SUCCESS && ret != DalConstants.JHI_SESSION_ALREADY_REGSITERED) {
			this.hListeners.remove(Long.toString(SessionHandle));
			Log.e(TAG, "DALService RegisterEvents native error");
		} else {//DAL_SUCCESS
			if (hListeners.containsValue(listener)) {
				Log.d(TAG, "DALService Register: callback instance already registered\n");
			}
			this.hListeners.put (Long.toString(SessionHandle), listener);
			Log.d(TAG, String.format("DALService Register for session %d/%s - already registered\n, replace existing registration\n",
			      SessionHandle, Long.toString(SessionHandle)));
		}
		return ret;
	}
	@Override
	public int DAL_UnregisterEvents(long SessionHandle) throws RemoteException {
		int ret = DalConstants.DAL_INTERNAL_ERROR;

		Log.d(TAG, "DALService UnRegisterEvents");

		try {
			ret = jhi.DAL_UnregisterEvents( SessionHandle);
		} catch (Exception e) {
			Log.e(TAG, "DALService UnregisterEvents native exeption");
			throw new RemoteException ("DAL Unregister event failed");
		}
		finally
		{
			this.hListeners.remove(Long.toString(SessionHandle));
		}

		if (ret != DalConstants.DAL_SUCCESS) {
			Log.e(TAG, "DALService UnRegisterEvents native error");
		}

		return ret;
	}

	@Override
	public int DAL_SHMemTxRxTrans (long SessionHandle,
			int nCommandId,
			ParcelFileDescriptor mfd,
			int txLength,
			int[] rxLength,
			int[] responseCode) throws RemoteException {
		int ret = DalConstants.DAL_INTERNAL_ERROR;

		// input validation is executed in native JNI

		try {
			ret = jhi.DAL_SHMemTxRxTrans(
							SessionHandle,
							nCommandId,
							mfd.getFd(),
							txLength,
							rxLength,
							responseCode);
		} catch (Exception e) {
			Log.e(TAG, "DALService SHMemTxRxTrans native exeption");
			throw new RemoteException ("DAL SHMemTxRxTrans event failed");
		}
		if (ret == DalConstants.DAL_SUCCESS) {
			Log.d(TAG, "DALService SHMemTxRxTrans native success");
		} else {
			Log.e(TAG, "DALService SHMemTxRxTrans native error");
		}
		return ret;
	}

	@Override
	public int DAL_ClearSessions () {
	        int ret = DalConstants.DAL_INTERNAL_ERROR;
		try {
			ret = jhi.DAL_ClearSessions (Binder.getCallingPid());
		} catch (Throwable e) {
			ret = DalConstants.DAL_THROWABLE_EXCEPTION;
			Log.e(TAG, String.format("Clear sessions exception %d", ret));
		}
		return ret;
	}
}
