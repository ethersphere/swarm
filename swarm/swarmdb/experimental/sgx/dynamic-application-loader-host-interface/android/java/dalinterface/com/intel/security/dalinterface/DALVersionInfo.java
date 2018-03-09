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
 * Parcelable DalVersion implementation for IPC transaction 
 * {@hide}
 */
public final class DALVersionInfo implements Parcelable {
	private  final String jhiVersion;
	private  final String fwVersion;
	private  final int   commType;
	private  final int   platformId;
	
	protected DALVersionInfo() {
		this("", "", 2, 2);		
	}
	public DALVersionInfo(String jhiVersion, String fwVersion,
			   int commType, int platformId) {
		this.jhiVersion = jhiVersion;
		this.fwVersion = fwVersion;
		this.commType = commType;
		this.platformId = platformId;
	}

	public DALVersionInfo(Parcel parcel) {
    	jhiVersion = parcel.readString();
		fwVersion = parcel.readString();
		commType = parcel.readInt();
		platformId = parcel.readInt();
	}

	public final String getJhiVersion () {
	 	return jhiVersion;
   	}

   	public final String getFwVersion () {
	 	return fwVersion;
   	}

	public final int getCommType () {
	 	return commType;
   	}

	public final int getPlatformId () {
	 	return platformId;
   	}

	public int describeContents() {
		return 0;
	}

	public final void writeToParcel(Parcel parcel, int flags) {
		parcel.writeString(jhiVersion);
		parcel.writeString(fwVersion);
		parcel.writeInt(commType);
		parcel.writeInt(platformId);
	}

	public static final Parcelable.Creator<DALVersionInfo> CREATOR = new Parcelable.Creator<DALVersionInfo>() {
		public DALVersionInfo createFromParcel(Parcel in) {
			return new DALVersionInfo(in);
		}

		public DALVersionInfo[] newArray(int size) {
			return new DALVersionInfo[size];
		}
	};
}
