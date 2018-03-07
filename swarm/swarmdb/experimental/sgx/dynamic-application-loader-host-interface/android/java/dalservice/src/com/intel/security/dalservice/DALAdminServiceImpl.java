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


import java.io.File;
import java.io.FileInputStream;
import java.io.FileNotFoundException;
import java.io.FileOutputStream;
import java.io.IOException;
import java.nio.channels.FileChannel;

import android.content.Context;
import android.content.pm.PackageManager;
import android.os.Binder;
import android.os.ParcelFileDescriptor;
import android.os.Process;
import android.os.RemoteException;
import android.util.Log;

import com.intel.security.dalinterface.DALVersionInfo;
import com.intel.security.dalinterface.IDALAdminManager;
import com.intel.security.dalinterface.IDALServiceCallbackListener;
import com.intel.security.dalinterface.DALCallback;
import com.intel.security.dalservice.JNIDALAdmin;
import com.intel.security.dalinterface.DalConstants;
import com.intel.security.dalservice.JNIDALAdmin;

public class DALAdminServiceImpl extends IDALAdminManager.Stub {
	private static final String TAG = DALAdminServiceImpl.class.getSimpleName();
	private JNIDALAdmin jhi;
	private final Context context;


	DALAdminServiceImpl(Context context) {

		Log.d(TAG, "DALAdminServiceImpl creating JNIDALTransport instance");
		jhi = new JNIDALAdmin();
		this.context = context;
	}


	@Override
	public int DAL_Install(String AppId, String AppPath) throws RemoteException {
		int ret = DalConstants.DAL_INTERNAL_ERROR;

		// TODO: also check if appId is a valid UUID
		if (AppId == null || AppPath == null|| AppPath.length() == 0) {
			Log.e(TAG,"DAL Install - invalid parameters");
			return DalConstants.DAL_INVALID_PARAMS;
		}

		try {
			ret = jhi.DAL_Install(AppId, AppPath);
		} catch (Exception e) {
			Log.e(TAG, "DALService Install App path native exeption");
			throw new RemoteException("DAL Install App path failed");
		}

		return ret;
	}

	@Override
	public int DAL_Install_FD(String AppId, ParcelFileDescriptor AppFd, int AppSize) throws RemoteException {
		int ret = DalConstants.DAL_INTERNAL_ERROR;

		// TODO: also check if appId is a valid UUID
		if (AppId == null || AppFd == null|| AppFd.getFileDescriptor() == null || AppSize <= 0) {
			Log.e(TAG,"DAL Install - invalid parameters");
			return DalConstants.DAL_INVALID_PARAMS;
		}

		try {
			ret = jhi.DAL_Install_FD(AppId, AppFd.getFd(), AppSize);
		} catch (Exception e) {
			Log.e(TAG, "DALService Install FD native exeption");
			throw new RemoteException("DAL Install FD failed");
		}

		return ret;
	}

	@Override
	public int DAL_Uninstall(String AppId) throws RemoteException {
		int ret = DalConstants.DAL_INTERNAL_ERROR;

		// TODO: also check if appId is a valid UUID
		if (AppId == null || AppId.length() == 0) {
			Log.e(TAG,"DAL Install - invalid parameters");
			return DalConstants.DAL_INVALID_PARAMS;
		}

		try {
			ret = jhi.DAL_Uninstall(AppId);
		} catch (Exception e) {
			Log.e(TAG, "DALService Uninstall App native exeption");
			throw new RemoteException ("DAL Uninstall App failed");
		}

		return ret;
	}

	@Override
	public final DALVersionInfo DAL_GetVersionInfo(int[] ret) throws RemoteException  {
		DALVersionInfo info = null;
		if (ret == null || ret.length == 0) {
			Log.e(TAG,"DAL Install - invalid parameters");
			return null;
		}

		try {
			info = jhi.DAL_GetVersionInfo(ret);
		} catch (Exception e) {
			Log.e(TAG, "DALService DAL_GetVersionInfo native exeption");
			throw new RemoteException ("DAL DAL_GetVersionInfo failed");
		}

		return info;
	}

}
