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

import android.os.Handler;
import android.os.Message;
import android.os.RemoteException;
import android.util.Log;

/**
 * class EventsCallbackListener implements
 * Stub of session asynchronous callback interface:
 * IDALServiceCallbackListener.Stub
 * Provides callback EventData to application
 * using message Handler instance
 */
public final class EventsCallbackListener extends IDALServiceCallbackListener.Stub
{
	private final Handler responseHandler;
	private static final String TAG = "DAL" + EventsCallbackListener.class.getSimpleName();
	
    /**
     * Public constructor
     * @param handler - Handler instance
     */
	public EventsCallbackListener(Handler handler) 
	{		
		responseHandler = handler;
	}
	/**
	 * Implements Asynchronous IPC call
	 * called from Dal Java Service, provides parcelable data to
	 * application as Handler message, containing EventData container
	 * @param response - DALCallback class instance - parcelable data 
	 */
	public final void onResponse(DALCallback response) 
	{
		
		Log.d(TAG, "Got response ");
        EventData data = new EventData(response.getHandle(), response.getData(), response.getDataType()); 
		Message message = this.responseHandler.obtainMessage(1, data);
		if (this.responseHandler.sendMessage(message) == false) {
			Log.e(TAG, "Can't add msg to queue ");
			return;
		}
	}

}
