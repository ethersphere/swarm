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

package com.intel.security.dalinterface;

import java.io.FileDescriptor;
import java.io.IOException;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;

import android.os.MemoryFile;
import android.os.ParcelFileDescriptor;
import android.util.Log;
/**
 * class SessionApp implements Dal session,
 * created for Trusted Applet and Session API-s
 */

public final class SessionApp
{
	public static final int SESSION_SHARED_FLAG = 1;
	public static final int SESSION_NO_FLAG     = 0;
	private static final int SESSION_HANDLE_INVALID = -1;
	private static final String TAG = "DAL" + SessionApp.class.getSimpleName();
	
	private final SessionManager sessionManager;
	private int sessionFlags;
	private final String appId;
	private long sessionHandle;
	private boolean isSharedSession;
		
	/**
	 * PrivateConstructor
	 * @param sm - SessionManager reference 
	 * @param trustedAppId -String, containing AppId
	 * @throws IllegalArgumentException
	 */
	SessionApp(SessionManager sm, String trustedAppId) throws IllegalArgumentException		
	{
		if (sm == null || DalUtils.isStringValidAppId(trustedAppId) == false) {
		   Log.e (TAG, "Invalid Constructor param(s)");
		   throw new IllegalArgumentException ("Invalid DALSessionManager Constructor patam(s)");
		}   
		sessionManager = sm;		
		appId = trustedAppId;
		clearSession();
	}
	
	/**
	 * Clear Session data
	 */
	private void clearSession ()
	{
		sessionFlags = 0;
		sessionHandle = SESSION_HANDLE_INVALID;
		isSharedSession = false;
	}
	
	/**
	 * Opens Session for Trusted Application enabling communication {@link SessionApp#sendAndReceive()}
	 * @param flags - Session parameters, currently supported SESSION_SHARED_FLAG, SESSION_NO_FLAG
	 * @param init - initial Session data, provided to Trusted Application (optional, can be null)
	 * @return transaction result or encoded exception, defined in {@link DalConstants} Class
	 */
	public final synchronized int openSession(byte[] init, int flags) 
	{
		long[] sessionHandles = new long[1];
		int ret = DalConstants.DAL_INTERNAL_ERROR;
		if (sessionHandle != SESSION_HANDLE_INVALID) {
			Log.e(TAG, "Create session: not shared session already created");
			return DalConstants.DAL_SESSION_EXISTS;
		}
    	if (init == null) {
    		Log.w(TAG, "Create session: null initBuffer object");
		//	return DalConstants.DAL_INVALID_PARAMS;
		}
		
		try
		{
			ret = sessionManager.getInterface().DAL_CreateSession( appId, flags, init, sessionHandles );
		}
		catch (Throwable ex )
		{
			ret = DalUtils.encodeException( ex );
			Log.e(TAG, String.format("Create session exception %s", DalUtils.exToString (ret)));
			
		}
		
		if ( ret == DalConstants.DAL_SUCCESS )
		{
			sessionHandle = sessionHandles[0];
			sessionFlags = flags;
			isSharedSession = ((sessionFlags & 0x1) == 0x1);
		}
		
		return ret;
	}
	
	/**
	 * close existing Session
	 * @return transaction result or encoded exception, defined in {@link DalConstants} Class
	 */
	public final synchronized int closeSession() 
	{		
		int ret = DalConstants.DAL_INTERNAL_ERROR;
		if (sessionHandle == SESSION_HANDLE_INVALID) {
			Log.e(TAG, "Close session: not created session");
			return DalConstants.DAL_INVALID_SESSION_HANDLE;
		}
		try
		{
			ret = sessionManager.getInterface().DAL_CloseSession( sessionHandle );
		}
		catch (Throwable ex )
		{
			ret = DalUtils.encodeException( ex );
			Log.e(TAG, String.format("Close session exception %s", DalUtils.exToString (ret)));
			
		}
		
		clearSession();
		
		return ret;
	}
	
	/**
	 * Register callback for asynchronous event from Trusted Application to existing not shared session
	 * applicable for not shared session only  
	 * @param listener - {@link EventsCallbackListener} Class instance
	 * @return transaction result or encoded exception, defined in {@link DalConstants} Class
	 */
	public final synchronized int registerEvents(EventsCallbackListener listener)
	{
		int ret = DalConstants.DAL_INTERNAL_ERROR;
		
		if (sessionHandle == SESSION_HANDLE_INVALID) {
			Log.e(TAG, "Register Events: not created session");
			return DalConstants.DAL_INVALID_SESSION_HANDLE;
		} 
		
		if (isSharedSession) {
			Log.e(TAG, "Register Events: shared session events not supported");
			return DalConstants.DAL_EVENTS_NOT_SUPPORTED;
		} 
		
		if (listener == null) {
    		Log.e(TAG, "RegisterEvents: null event listener object");
    		return DalConstants.DAL_INVALID_PARAMS;
		}
		try {
			ret =  sessionManager.getInterface().DAL_RegisterEvents( sessionHandle, listener );
		} catch (Throwable ex )	{
			ret = DalUtils.encodeException( ex );
			Log.e(TAG, String.format("Register events exception %s", DalUtils.exToString (ret)));
			clearSession();
			
		}		
		if (ret ==  DalConstants.DAL_INVALID_SESSION_HANDLE) {
			Log.e(TAG, "Register events failure, clear session");
			clearSession();
		}
		return ret;
	}
	
	/**
	 * Unregister callback for asynchronous event from Trusted Application to existing session
	 * @return transaction result or encoded exception, defined in {@link DalConstants} Class
	 */
	public final synchronized int unregisterEvents()
	{
		int ret = DalConstants.DAL_INTERNAL_ERROR;
		if (sessionHandle == SESSION_HANDLE_INVALID) {
			Log.e(TAG, "UnRegister Events: not created session");
			return DalConstants.DAL_INVALID_SESSION_HANDLE;
		} 
		
		if (isSharedSession) {
			Log.e(TAG, "UnRegister Events: shared session events not supported");
			return DalConstants.DAL_EVENTS_NOT_SUPPORTED;
		} 
		
		try {
			ret = sessionManager.getInterface().DAL_UnregisterEvents( sessionHandle );
		} catch (Throwable ex ) {
			ret = DalUtils.encodeException( ex );
			Log.e(TAG, String.format("Unregister events %s", DalUtils.exToString (ret)));
			clearSession();
			
		}	
		if (ret ==  DalConstants.DAL_INVALID_SESSION_HANDLE) {
			Log.e(TAG, "UnRegister events failure, clear session");
			clearSession();
		}
		return ret;
	}
	
	/**
	 * Synchronous API to provide command and data to Trusted Application and receive response
	 * @param command - Trusted Application specific command
	 * @param data - TransactionData container
	 *        on invocation: 
	 * 		     data.request should be set to the request buffer which is sent to the applet
	 * 		     data.maxResponseLength should be set to the max expected response length, 
	 *           if (-1) is specified then data.maxResponseLength = request.length
	 *        on return:
	 * 		     data.response is set to the exact response buffer received from the applet
	 * 	         data.appletResponseCode is set to the response code returned from the applet    
	 * @return transaction result or encoded exception, defined in {@link DalConstants} Class
	 */
	public synchronized int sendAndReceive(int command, TransactionData data)
	{
		int ret = DalConstants.DAL_INTERNAL_ERROR;
		
		if (sessionHandle == SESSION_HANDLE_INVALID) {
			Log.e(TAG, "SendAndReceive: not created session");
			return DalConstants.DAL_INVALID_SESSION_HANDLE;
		}
		if (data == null) {
    		Log.e(TAG, "SendAndReceive: null data container object");
    		return DalConstants.DAL_INVALID_PARAMS;
		}
		
		if (data.request != null && data.request.length > DalConstants.DAL_BUFFER_MAX) {
			Log.e(TAG, "SendAndReceive: request data too long");
			return DalConstants.DAL_INVALID_BUFFER_SIZE;
		}
		
		if (data.maxResponseLength == -1) {
			Log.e(TAG, "SendAndReceive: response data isn't assigned");
			return DalConstants.DAL_INVALID_PARAMS;
		}
		if (data.maxResponseLength > DalConstants.DAL_BUFFER_MAX) {
			Log.w(TAG, "SendAndReceive: expected response data too long");
			data.maxResponseLength = DalConstants.DAL_BUFFER_MAX;
		}
		
		int[] responseCode = new int[1];
		MemoryFile txrx = null;
		int[] rxCount = new int[1];
		rxCount[0] = data.maxResponseLength;
		FileDescriptor mfd = null;
		int txLength = (data.request == null) ? 0 : data.request.length;
		//Method getFD = null;
		ParcelFileDescriptor pfd = null;
			try {
				txrx = new MemoryFile ("txrx", ((txLength + data.maxResponseLength) > 0 ? (txLength + data.maxResponseLength) : 16));
				//allocate min ashmem for binder transaction if data size is 0
				Method getFD = MemoryFile.class.getDeclaredMethod("getFileDescriptor");
				mfd = (FileDescriptor)getFD.invoke(txrx);
				if (data.request != null) {
					txrx.writeBytes (data.request, 0, 0, txLength);
				}
				pfd = ParcelFileDescriptor.dup (mfd);
			
			} catch (Throwable ex) {
				ret = DalUtils.encodeException( ex );
				Log.e(TAG, String.format("SendAndReceive create ashmem %s", DalUtils.exToString (ret)));
				if (txrx != null)
					txrx.close();
				clearSession();
				return ret;
			}
		
        try {
        	ret = sessionManager.getInterface().DAL_SHMemTxRxTrans(
		    		sessionHandle,
	       			command,
		   			pfd,
		   			txLength,
		   			rxCount,
		   			responseCode);
        } catch (Throwable ex) {
			ret = DalUtils.encodeException( ex );
			Log.e(TAG, String.format("SendAndReceive transaction %s", DalUtils.exToString (ret)));
			clearSession();
		}	
		
		if (ret ==  DalConstants.DAL_INVALID_SESSION_HANDLE || ret == DalConstants.JHI_APPLET_FATAL) {
			Log.e(TAG, "SendAndReceive failure, clear session");
			clearSession();
		}
		
		if (ret == DalConstants.DAL_SUCCESS) {
			Log.e(TAG, "DAL SendAndReceive success");
			data.appletResponseCode = responseCode[0];
			if (rxCount[0] != 0 && txrx != null) {
			
				data.response = new byte [rxCount[0]];
				try {
					txrx.readBytes (data.response, txLength, 0, rxCount[0]);
				} catch (Throwable ex) {
					ret = DalUtils.encodeException( ex );
					Log.e(TAG, String.format("SendAndReceive read from ashmem %s", DalUtils.exToString (ret)));
					clearSession();
				}
			}
		}
		if (txrx != null)
		txrx.close();
		return ret;
	}
	
	/**
	 * Retrieves Session Trusted Application Id
	 * @return Trusted Application Id
	 */
	public final String getTrustedAppId() 
	{
    	return appId;
    }
    
	/**
	 * Retrieves Session status
	 * @return true  - opened Session exists;
	 * 		   false - Session is not created / closed
	 */
    public synchronized final boolean isOpened()
    {
    	return (sessionHandle != SESSION_HANDLE_INVALID);
    }
    
    /**
     * Retrieves isSharedSession property 
     * @return  true  - shared session;
     * 			false - not shared session; 
     */
    public final boolean isShared() 
    {
    	return isSharedSession;
    }
    /**
     * Retrieves Session identification handle 
     * @return Session handle or -1 if session isn't opened;
     */
    public final long getSessionHandle() 
    {
    	return sessionHandle;
    }

}
