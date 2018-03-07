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

import java.io.BufferedInputStream;
import java.io.ByteArrayOutputStream;
import java.io.FileDescriptor;
import java.io.IOException;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;

import android.content.Context;
import android.os.MemoryFile;
import android.os.ParcelFileDescriptor;
import android.os.RemoteException;
import android.util.Log;

/** 
 * TrustedAppAsset extends TrustedApp abstract class
 * for Applets, provided by Application as /assets/ resource
 */
final class TrustedAppAsset extends TrustedApp 
{
	private static final String TAG = "DAL" + TrustedAppAsset.class.getSimpleName();
	private final ParcelFileDescriptor appFd;
	private final int appSize;
	private MemoryFile applet;
	
	/**
	 * Private constructor
	 * allocates ashmem containing Applet File content - MemoryFile instance.
	 * @param context - application Context
	 * @param trustedAppId - String, containing App Id
	 * @param trustedAppPath - String, containing App File related path from assets/
	 * @throws IllegalArgumentException
	 * @throws IOException
	 */
	TrustedAppAsset(Context context, String trustedAppId, String trustedAppPath) throws IllegalArgumentException, IOException	 
	{
		super( trustedAppId );
		if (context == null) {
			Log.e (TAG,"Null Context parameter in Constructor ");
			throw new RuntimeException ("Null Context parameter in Constructor");
		}
		
		if (trustedAppPath == null) {
			Log.e (TAG, "invalid DAL AppID");
			throw new IllegalArgumentException ("invalid DAL App Path");
		}
		
		BufferedInputStream assetsStream = null;
		ByteArrayOutputStream tmpStream = null;
		byte[] appletData;
		try {
			assetsStream = new BufferedInputStream(context.getAssets().open(trustedAppPath));
		} catch (IOException e) {
				Log.e (TAG, "Invalid Trusted App Assets File - stream");
				throw new IllegalArgumentException ("Inavalid Trusted App Assets File - stream");
		}
		Log.d (TAG, "Trusted App Assets File - stream created");
		try {
			byte [] tmp = new byte [1024];
			tmpStream = new ByteArrayOutputStream(4*1024);
			int bytesCount = 0;
           	while ((bytesCount = assetsStream.read(tmp, 0, tmp.length)) != -1) {
				tmpStream.write(tmp, 0, bytesCount);
				Log.d (TAG, String.format("read assets bytes %d", bytesCount));
			}
			appletData = tmpStream.toByteArray();
		}
		catch (Throwable ex) 
		{
			assetsStream.close();
			if (tmpStream != null) {
				tmpStream.close();
			}
			Log.d (TAG, "Can't copy App to asmem");
			throw new RuntimeException("Can't copy App to asmem", ex);
		}

		try {
			applet = new MemoryFile("applet", appletData.length);
		} catch (IOException e) {
			Log.e (TAG,"Can't create MemoryFile");
			throw new RuntimeException("Can't create MemoryFile", e);
		}

		FileDescriptor mfd = null;
		Method getFD = null;

		try {
			try {
				getFD = MemoryFile.class.getDeclaredMethod("getFileDescriptor");
			} catch (Exception e){
				Log.e (TAG, "Can't get getFileDescriptor method of the MemoryFilep");
				throw new RuntimeException("Trusted App failed - getFD", e);
			}
			if (getFD == null) {
				Log.e (TAG, "Can't get getFileDescriptor method of the MemoryFile - null");
				throw new RuntimeException("Trusted App failed - getFD");
			}
			try {
				mfd = (FileDescriptor)getFD.invoke(applet);
			} catch (IllegalAccessException e1) {
				Log.e (TAG, "Can't invoke getFileDescriptor method of the MemoryFile");
				throw new RuntimeException("Trusted App failed - getFD,invoke", e1);
			} catch (InvocationTargetException e1) {
				Log.e (TAG, "Can't invoke getFileDescriptor method of the MemoryFile1");
				throw new RuntimeException("Trusted App failed - getFD.invoke", e1);
			}

			try {
				applet.writeBytes(appletData, 0, 0, appletData.length);
			} catch (IOException e)	 {
				Log.e (TAG, "Can't write to the MemoryFile");
				throw new RuntimeException("Trusted App failed - writeBuffer ", e);
			}

			try {
				appFd = ParcelFileDescriptor.dup(mfd);
			} catch (IOException e1) {
				Log.e (TAG, "Can't ParcelFileDescriptor.dup()");
				throw new RuntimeException("Trusted App failed - ParcelFileDescriptor.dup()", e1);
			}

			if (appFd == null) {
				Log.e (TAG, "Can't get ParcelFD  of the MemoryFile - null");
				throw new RuntimeException("Trusted App failed - Can't get ParcelFD  of the MemoryFile - null");
			
			}
			
			if ((appSize = applet.length()) == 0){
				Log.e (TAG, "Can't get Applet size");
				throw new RuntimeException("Trusted App failed - size is 0");
			
			}
		}	
		catch (RuntimeException ex1)
		{
			applet.close();
			throw ex1;
		}
	}
	/**
	 * TrustedApp::installTrustedApp() implementation
	 * Applet, provided as Application /assets/ file,
	 * is installed using AppId and ashmem allocation
	 */
	int installTrustedApp(IDALAdminManager adminInterface) 
	{
		int ret = DalConstants.DAL_INTERNAL_ERROR;
		try {
			ret = adminInterface.DAL_Install_FD( getAppId(), appFd, appSize );
		}	
		catch (Throwable ex )
		{
				ret = DalUtils.encodeException( ex );
		}
		return ret;
	}
	
	/**
	 * TrustedApp::close() implementation
	 * close ashmem
	 */
	public void close() 
	{
		if (applet != null)
		{
			applet.close();
		}
	}
	
}
