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


import android.os.Parcel;
import android.os.Parcelable;


/**
 * Parcelable implementation of Callback Event data container 
 * {@hide}
 */
public final class DALCallback implements Parcelable {
	private  long handle;
	private  byte[] data;
	private  byte dataType;

	
	public DALCallback(long handle, byte[] data, byte dataType) {
		this.handle = handle;
		this.dataType = dataType;
		this.data = data;
	}

	public DALCallback(Parcel parcel) {
		handle = parcel.readLong();
		data = parcel.createByteArray();
		dataType = parcel.readByte();
	}

	public final long getHandle () {
		return handle;
	}

	public final byte getDataType () {
		return dataType;
	}

	public final byte[] getData () {
		return data;
	}

	public final void setHandle (long handle) {
		handle = handle;
	}

	public final void setDataType (byte dataType) {
		dataType = dataType;
	}

	public final void setData (byte[] data) {
		data = data;
	}

	public int describeContents() {
		return 0;
	}

	public void writeToParcel(Parcel parcel, int flags) {
		parcel.writeLong(handle);
		parcel.writeByteArray(data);
		parcel.writeByte(dataType);


	}

	public static final Parcelable.Creator<DALCallback> CREATOR = new Parcelable.Creator<DALCallback>() {
		public DALCallback createFromParcel(Parcel in) {
			return new DALCallback(in);
		}

		public DALCallback[] newArray(int size) {
			return new DALCallback[size];
		}
	};
}
