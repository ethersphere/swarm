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

import android.util.Log;

/** 
 * TrustedAppAsset extends TrustedApp abstract class
 * for Applets, accessible in device storage File System
 */
final class TrustedAppFile extends TrustedApp
{
	private final String appPath;
	private static final String TAG = "DAL" + TrustedAppFile.class.getSimpleName();
	
	/**
	 * Private constructor
	 * provides validated App Id and App Path to native service
	 * @param trustedAppId - String, containing App Id
	 * @param trustedAppPath - String, containing App File absolute path in device FileSystem
	 * @throws IllegalArgumentException
	 */
	TrustedAppFile(String trustedAppId, String trustedAppPath) throws IllegalArgumentException 
	{
		super( trustedAppId );
		if (trustedAppPath == null) {
			Log.e (TAG, "invalid DAL App Path");
			throw new IllegalArgumentException ("invalid DAL App Path");
		}
		appPath = trustedAppPath;		
	}
	
	/**
	 * TrustedApp::installTrustedApp() implementation
	 * Applet, accessible on device storage FS,
	 * is installed using AppId and full path
	 */
	int installTrustedApp(IDALAdminManager adminInterface) 
	{	
		int ret = DalConstants.DAL_INTERNAL_ERROR;
		try {
			ret = adminInterface.DAL_Install( super.getAppId(), appPath );
		}	
		catch (Throwable ex )
		{
				ret = DalUtils.encodeException( ex );
		}
		return ret;
	}
	
	/**
	 * TrustedApp::close() isn't implemented
	 */
	public void close() {}
	
	
}
