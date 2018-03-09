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

import android.util.Log;

/** 
 * abstract class TrustedApp
 * Instantiate  TrustedApp data 
 */
public abstract class TrustedApp
{
	private final String appId;
	private static final String TAG = "DAL" + TrustedApp.class.getSimpleName();
	
	/**
	 * Private constructor, validates TrustedAppId
	 * @param trustedAppId - String, containing AppId
	 * @throws IllegalArgumentException
	 */
	TrustedApp(String trustedAppId) throws IllegalArgumentException
	{
		if (DalUtils.isStringValidAppId(trustedAppId) == false)
		{	
			Log.e (TAG, "invalid DAL AppID");
			throw new IllegalArgumentException ("invalid DAL AppID");
		}	
		appId = trustedAppId;
	}
    
	/**
	 * Retrieves TrustedAppId
	 * @return String, containing AppId
	 */
	public final String getAppId()
	{
		return appId;
	}
	
	/**
	 * abstruct TrustedApp installation method
	 * @param adminInterface - IDALAdminManager.Stub.Proxy
	 * @return int - transaction return code
	 */
	abstract int installTrustedApp(IDALAdminManager adminInterface);
	/**
	 * abstruct TrustedApp close() method
	 */
	public abstract void close(); // close the memory file
	 
	/**
	 * Retrieves TrustedAppId as String
	 * @return String, containing AppId
	 */
	public String toString()
	{
		return appId.toString();
	}
	
}
