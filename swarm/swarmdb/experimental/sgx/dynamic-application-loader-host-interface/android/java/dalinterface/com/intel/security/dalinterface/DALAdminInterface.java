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


import android.os.IBinder;
import android.os.RemoteException;
import android.os.ServiceManager;
import android.util.Log;
import java.util.concurrent.Semaphore;
import java.lang.InterruptedException;

/**
 * DALAdminInterface is Singletone class utilized to retrieve
 * synchronously registered IDALAdminInterface instance from
 * SessionManager.
 * {@hide}
 */
final class DALAdminInterface {
    private static final String TAG = DALAdminInterface.class.getSimpleName();
    private static final String REMOTE_SERVICE_NAME = IDALAdminManager.class.getName();
    private final static Semaphore locker = new Semaphore(1);
    private static DALAdminInterface instance = null;

    /**
     * Retrieves static instance of DALAdminInterface class
     * @return DALAdminInterface class
     */
    protected static DALAdminInterface getInstance() {
		Log.d(TAG, "DALAdminInterface.getInstance()");
		if (instance == null)
	  		instance = new DALAdminInterface();
		return instance;
    }
    /**
     * Retrieve IDALAdminManager
     * interface instance
     * @return IDALAdminManager interface or null
     */

    protected IDALAdminManager getInterface () {
		Log.d(TAG, "getting to IDALAdminManager by name [" + REMOTE_SERVICE_NAME + "]");
		return getAdminInterface();
    }

    /**
     * Private method to retrieve IDALAdminManager
     * interface instance from SessionManager
     * @return IDALAdmintManager interface or null
     */
    private IDALAdminManager getAdminInterface () {
		IBinder iAdmin = null;
		Log.d(TAG, "Check  IBinder by name [" + REMOTE_SERVICE_NAME + "]");
		iAdmin = ServiceManager.checkService(REMOTE_SERVICE_NAME);
		if (iAdmin != null) {
			return IDALAdminManager.Stub.asInterface (iAdmin);
		}
		Log.d(TAG, "Get locked IBinder by name [" + REMOTE_SERVICE_NAME + "]");
		try {
			locker.acquire();
			iAdmin = ServiceManager.getService(REMOTE_SERVICE_NAME);
		} catch (InterruptedException e) {
			Log.e(TAG, "Semaphore.acquire() InterruptedException");
			return null;
		}
		finally {
			locker.release();
		}

		if (iAdmin != null) {
			return IDALAdminManager.Stub.asInterface (iAdmin);
		} else {
			Log.e(TAG, "Error: Can't get IBinder by name [" + REMOTE_SERVICE_NAME + "]");
			return null;
		}
   }
}
