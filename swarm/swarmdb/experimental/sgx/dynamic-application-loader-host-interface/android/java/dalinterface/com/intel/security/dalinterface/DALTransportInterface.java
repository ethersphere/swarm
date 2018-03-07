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
 * DALTransportInterface is Singletone class utilized to retrieve
 * synchronously registered IDALTransportInterface instance from
 * SessionManager.
 */
final class DALTransportInterface {
    private static final String TAG = DALTransportInterface.class.getSimpleName();
    private static final String REMOTE_SERVICE_NAME = IDALTransportManager.class.getName();
    private final static Semaphore locker = new Semaphore(1);
    private static DALTransportInterface instance = null;
    /**
     * Retrieves static instance of DALTransportInterface class
     * @return DALTransportInterface class
     */
    public static DALTransportInterface getInstance() {
	Log.d(TAG, "DALTransportInterface.getInstance()");
	if (instance == null)
	  	instance = new DALTransportInterface();
		return instance;
    }

    /**
     * Retrieve IDALTransportManager
     * interface instance
     * @return IDALTransportManager interface or null
     */
    protected IDALTransportManager getInterface () {
		Log.d(TAG, "getting to IDALTransportManager by name [" + REMOTE_SERVICE_NAME + "]");
		return getTransportInterface();
	}
    /**
     * Private method to retrieve IDALTransportManager
     * interface instance from SessionManager
     * @return IDALTransportManager interface or null
     */
    private IDALTransportManager getTransportInterface () {
		IBinder iTransport = null;
		Log.d(TAG, "Check  IBinder by name [" + REMOTE_SERVICE_NAME + "]");
		iTransport = ServiceManager.checkService(REMOTE_SERVICE_NAME);
		if (iTransport != null) {
			return IDALTransportManager.Stub.asInterface (iTransport);
		}
		Log.d(TAG, "Get locked IBinder by name [" + REMOTE_SERVICE_NAME + "]");
		try {
			locker.acquire();
		} catch (InterruptedException e) {
			Log.e(TAG, "Semaphore.acquire() InterruptedException");
			return null;
		}
		iTransport = ServiceManager.getService(REMOTE_SERVICE_NAME);

		locker.release();

		if (iTransport != null) {
			return IDALTransportManager.Stub.asInterface (iTransport);
		} else {
			Log.e(TAG, "Error: Can't get IBinder by name [" + REMOTE_SERVICE_NAME + "]");
			return null;
		}
   }

}
