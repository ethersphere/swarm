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

import java.io.IOException;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.ContextWrapper;
import android.content.Intent;
import android.content.IntentFilter;
import android.util.Log;

/**
 * AdmimManager 
 * implements DAL Administration API-s
 */

public final class AdminManager
{
	private static AdminManager instance = null;
	private static ContextWrapper context;
	private static IDALAdminManager adminInterface;
	private static DALAdminInterface adminBinder;
	private static boolean interfaceUpdateRequest = true;
	private static final String TAG = "DAL" + AdminManager.class.getSimpleName();
	
	/**
	 * Private constructor
	 * creates static AdminManager instance
	 * and retrieves IDALAdminManager interface instance 
	 * @param ctx - Activity Context
	 * @throws RuntimeException if IDALAdminManager can't be retrieved
	 */
    private AdminManager(Context ctx) throws RuntimeException
	{		
    	if (ctx == null) {
			Log.e(TAG, "Invalid Context parameter");
			throw new RuntimeException("Invalid Context parameter");
		}
    	try {
			adminBinder = DALAdminInterface.getInstance();
		} catch (Throwable e) {
			Log.e( TAG, "can't get  DALAdminInterface");
			throw new RuntimeException ("can't get  DALAdminInterface", e);
		}
    	try {
    		adminInterface = retrieveInterface();
    	} catch (RuntimeException e) {
			Log.e(TAG, "Failed to get IDALAdminManager");
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
    
    final synchronized IDALAdminManager retrieveInterface () {
    	if (interfaceUpdateRequest == true) {
    		adminInterface = adminBinder.getInterface();
    		if (adminInterface == null){
    			throw new RuntimeException ("Failed to get IDALAdminManager");
    		}
    		interfaceUpdateRequest = false;
    	}
    	return adminInterface;
    }
    /**
     * Retrieves static AdminMctxanager instance 
     * @param ctx - Activity Context
     * @return AdminManager instance
     * @throws RuntimeException from class constructor
     */
	public final static synchronized AdminManager getInstance(Context ctx) 
	{
		if ( instance == null )
		{
			try {
				instance = new AdminManager(ctx);
			} catch (RuntimeException ex) {
				Log.e(TAG, "Cant' get AdminManager instance");
				instance = null;
			}
		}
		return instance;
	}
	/**
	 * Broadcast receiver extension
	 * receives broadcast notifications from DAL Admin Service
	 */
	final class ServerNotification extends BroadcastReceiver {
		@Override
		public void onReceive(Context context, Intent intent)
		{
			Log.d(TAG, "Broadcast received");
			interfaceUpdateRequest = true;
		}
	}
    /**
     * Creates Trusted Application instance for application, stored in File System 
     * @param trustedAppId - Trusted Application ID
     * @param trustedAppPath - Trusted Application File System path
     * @return {@link TrustedApp} instance 
     * @throws IllegalArgumentException
     */
	public final synchronized TrustedApp createTrustedAppFromFile(String trustedAppId, String trustedAppPath) throws IllegalArgumentException
	{
		return new TrustedAppFile( trustedAppId, trustedAppPath );
	}
	
	/**
     * Creates Trusted Application instance for application, stored as assets resource 
     * @param trustedAppId - Trusted Application ID
     * @param trustedAppPath - application path from assets/
     * @return {@link TrustedApp} instance 
     * @throws IllegalArgumentException
	 * @throws IOException 
     * @throws RuntimeException 
     */
	public final synchronized TrustedApp createTrustedAppFromAsset(Context context, String trustedAppId, String trustedAppPath) throws IllegalArgumentException, IOException 
	{
		return new TrustedAppAsset(context, trustedAppId, trustedAppPath );
	}
	/** 
	 * Install Trusted Application API
	 * @param app - {@link TrustedApp} Class instance
	 * @return transaction result or encoded exception, defined in {@link DalConstants} Class
	 */
	public final synchronized int installTrustedApp(TrustedApp app)
	{
		int ret = DalConstants.DAL_INTERNAL_ERROR;
		if (app == null) {
			Log.e(TAG, "TrustedApp is null");
			return DalConstants.DAL_INVALID_PARAMS;
		}	
		try
		{
			ret = app.installTrustedApp( retrieveInterface());
		}		
		catch (Throwable ex )
		{
			ret = DalUtils.encodeException( ex );
			Log.e(TAG, String.format("Install TrustedApp exception %s", DalUtils.exToString (ret)));
		}

		return ret;
	}
	/** 
	 * Uninstall Trusted Application API
	 * @param app - {@link TrustedApp} Class instance
	 * @return transaction result or encoded exception, defined in {@link DalConstants} Class
	 */
	public final synchronized int uninstallTrustedApp(TrustedApp app)
	{
		if (app == null)
			return DalConstants.DAL_INVALID_PARAMS;
		int ret = DalConstants.DAL_INTERNAL_ERROR;
		try {
			return uninstallTrustedApp( app.getAppId() );
		} catch (Throwable ex ) {
			ret = DalUtils.encodeException( ex );
			Log.e(TAG, String.format("Uninstall TrustedApp exception %s", DalUtils.exToString (ret)));
		} 
		return ret;
	}
	/** Uninstall Trusted Application API
	 * @param trustedAppId - Trusted Application ID
	 * @return transaction result or encoded exception, defined in {@link DalConstants} Class
	 */
	public final synchronized int uninstallTrustedApp(String trustedAppId)
	{	
		int ret = DalConstants.DAL_INTERNAL_ERROR;
		if (DalUtils.isStringValidAppId(trustedAppId) == false)
			return DalConstants.DAL_INVALID_PARAMS;
		try
		{
			ret = retrieveInterface().DAL_Uninstall( trustedAppId );
		}	
		catch (Throwable ex )
		{
				ret = DalUtils.encodeException( ex );
				Log.e(TAG, String.format("Uninstall TrustedApp exception %s", DalUtils.exToString (ret)));
		}
		return ret;
	}
	/**
	 * Get DAL FW & SW version API
	 * @return {@link DalVersion} Class instance or null on transaction error or exception
	 */
	public final synchronized DalVersion getVersionInfo()
	{
		DalVersion retval = null;
		DALVersionInfo info = null;
		int[] ret= new int[1];

		try
		{
				info = retrieveInterface().DAL_GetVersionInfo( ret );
		}
		catch (Throwable ex )
		{
			Log.e(TAG, String.format("GetVersionInfo exception %s", DalUtils.exToString (DalUtils.encodeException( ex ))));
			return null;
		}
		if ( (ret[0] == DalConstants.DAL_SUCCESS) && (info != null) )
		{
			retval = new DalVersion( info.getJhiVersion(), info.getFwVersion(), info.getCommType(), info.getPlatformId() );
		}

		return retval;
	}

}
