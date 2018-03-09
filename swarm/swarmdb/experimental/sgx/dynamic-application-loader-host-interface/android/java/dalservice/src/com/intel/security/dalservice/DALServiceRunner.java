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

import com.intel.security.dalinterface.IDALTransportManager;
import com.intel.security.dalinterface.IDALAdminManager;
import android.app.Application;
import android.util.Log;
import android.content.Intent;
import android.os.ServiceManager;

public final class DALServiceRunner extends Application {
	private DALTransportServiceImpl transportService ;
	private DALAdminServiceImpl adminService;
	static	final String TAG = DALServiceRunner.class.getSimpleName();
	private static final String ADMIN_SERVICE_NAME = IDALAdminManager.class.getName();
	private static final String TRANSPORT_SERVICE_NAME = IDALTransportManager.class.getName();
	static  {
		  System.load("libjhi.so");
	}
	public void onCreate () {
		super.onCreate();
		Log.i (TAG, "onCreate");
		transportService = new DALTransportServiceImpl(this);
		ServiceManager.addService(TRANSPORT_SERVICE_NAME, this.transportService);

		adminService = new DALAdminServiceImpl(this);
		ServiceManager.addService(ADMIN_SERVICE_NAME, this.adminService);
		Intent intent = new Intent();
		intent.setAction("com.intel.security.dalservice.DAL_ACCESS");
		intent.addFlags(Intent.FLAG_INCLUDE_STOPPED_PACKAGES);
		sendBroadcast(intent);
		Log.d(TAG,"DAL Service Broadcast sent");
	}

	public void onTerminate() {
		super.onTerminate();
		Log.d(TAG, "Terminated");
     }

}
