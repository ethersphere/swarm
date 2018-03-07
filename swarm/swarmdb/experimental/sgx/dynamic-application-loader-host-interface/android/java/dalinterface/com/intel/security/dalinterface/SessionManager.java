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

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.ContextWrapper;
import android.content.Intent;
import android.content.IntentFilter;
import android.util.Log;

/**
 * SessionManager is utilized to create/instantiate {@link SessionApp} Class
 * per Trusted Applets.
 */
public final class SessionManager
{
	public final static int SESSION_NO_FLAGS = 0x0;
	public final static int SESSION_SHARED = 0x1;
	private static final String TAG = "DAL" + SessionManager.class.getSimpleName();
	private static ContextWrapper context;

	private IDALTransportManager sessionInterface;
	private DALTransportInterface transportBinder;
	private static SessionManager instance = null;
	private static boolean interfaceUpdateRequest = true;

	/**
	 * Private constructor
	 * creates class static instance;
	 * retrieves  IDALTransportManager interface
	 * @param ctx - Activity Context
	 * @throws RuntimeException
	 */
	private SessionManager(Context ctx)
	{	
		if (ctx == null) {
			Log.e(TAG, "Invalid Context parameter");
			throw new RuntimeException("Invalid Context parameter");
		}

		try {
			transportBinder = DALTransportInterface.getInstance();
		} catch (Throwable e) {
			Log.e( TAG, "can't get  DALTransportInterface");
			throw new RuntimeException ("can't get  DALTransportInterface", e);
		}
    	try {
    			sessionInterface = getInterface();
    	} catch (RuntimeException e) {
			Log.e(TAG, "Failed to get IDALTransportManager");
			throw e;
    	}

		context = new ContextWrapper(ctx.getApplicationContext());
		if (context == null) {
			Log.e(TAG, "Failed to get App context");
			throw new RuntimeException("Failed to get App context");
		}

		context.getApplicationContext().registerReceiver(new ServerNotification(),
		         new IntentFilter("com.intel.security.dalservice.DAL_ACCESS"));

	}

	/**
     * Retrieves static SessionManager instance 
     * @param ctx - Activity Context
     * @return SessionManager instance or null
     */
	public final static synchronized SessionManager getInstance(Context ctx)
	{
		if ( instance == null )
		{
			try {
				instance = new SessionManager(ctx);
			} catch (RuntimeException ex) {
				Log.e (TAG, "Can't get SessionManager instance");
				return null;
			}
		}

		return instance;
	}
	/**
	 * Broadcast receiver extension
	 * receives broadcast notifications from DAL Transport Service
	 */
	final class ServerNotification extends BroadcastReceiver {
		@Override
		public void onReceive(Context context, Intent intent)
		{
			Log.d(TAG, "Broadcast received");
			sessionInterface = transportBinder.getInterface();
		}
	}


	/**
	 * create  SessionApp instance for TrustedApp 
	 * @param app - TrustedApp implementation instance
	 * @return code, defined in {@link DalConstants} Class
	 * @throws IllegalArgumentException
	 */
	public final synchronized SessionApp createSession(TrustedApp app) throws IllegalArgumentException
	{
		if (app == null) {
			Log.e (TAG, "Illegal TrustedApp");
			throw new IllegalArgumentException ("Illegal TrustedApp");
		}
		return new SessionApp( this, app.getAppId());
	}

	/**
	 * create  SessionApp instance for trustedAppId
	 * @param trustedAppId - String, containing AppId
	 * @return code, defined in {@link DalConstants} Class
	 * @throws IllegalArgumentException
	 */
	public final synchronized SessionApp createSession(String trustedAppId)
	{
		
		return new SessionApp( this, trustedAppId);
	}
	

	/**
	 * Retrieves IDALTransportManager.Stub.Proxy interface
	 * @return IDALTransportManager reference
	 */
	final synchronized IDALTransportManager getInterface()
	{   if (interfaceUpdateRequest == true) {
			sessionInterface = transportBinder.getInterface();
			if (sessionInterface == null){
				throw new RuntimeException ("Failed to get IDALTransportManager");
			}
			interfaceUpdateRequest = false;
		}

		return sessionInterface;
	}
	/**
	 * Clear all opened session in Firmware
	 * Should be invoke from Activity#onDestroy() 
	 */
	public final void onDestroy () {
		  int ret = DalConstants.JHI_INTERNAL_ERROR;
		  Log.e (TAG, "Destroy App Sessions...");
		  try {
			  ret = instance.getInterface().DAL_ClearSessions();
		  } catch (Throwable ex) {
			    ret = DalUtils.encodeException( ex );
			    Log.e(TAG, String.format("Destroy sessions %s", DalUtils.exToString (ret)));
		  }



	}

}
