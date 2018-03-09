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

/**
 * DAL Version Info container
 */
public final class DalVersion 
{
	public static final int COMM_TYPE_SOCKETS = 0x0;
	public static final int COMM_TYPE_HECI = 0x1;
	
	public static final int PLATFORM_ID_ME = 0x0;
	public static final int PLATFORM_ID_SEC = 0x1;
	public static final int PLATFORM_ID_CSE = 0x2;
	
	private final String jhiVersion;
	private final String fwVersion;
	
	private final int commType;
	private final int platformId;
	
	/**
	 * Private constructor
	 * @param jhiVer: native SW JHI version 
	 * @param fwVer: FW version
	 * @param comm:  communication type     0 - COMM_TYPE_SOCKETS (VP)
	 * 										1 - COMM_TYPE_HECI 
	 * @param platform: platform type   0 - PLATFORM_ID_ME
	 *                                  1 - PLATFORM_ID_SEC
	 *                                  2 - PLATFORM_ID_CSE 
	 */
	DalVersion(String jhiVer, String fwVer, int comm, int platform)
	{
		jhiVersion 	= jhiVer;
		fwVersion 	= fwVer;
		commType 	= comm;
		platformId 	= platform; 
	}
	
	/**
	 * Retrieves native SW JHI version 
	 * @return String, containing SW version
	 */
	public final String getJhiVersion() 
	{
		return jhiVersion;
	}

	/**
	 * Retrieves FW version
	 * @return String, containing FW version
	 */
	public final String getFwVersion() 
	{
		return fwVersion;
	}
    
	/**
	 * Retrieves  communication type:  0 - COMM_TYPE_SOCKETS (VP)
	 * 								   1 - COMM_TYPE_HECI
	 * @return communication type
	 */
	public final int getCommType() 
	{
		return commType;
	}

	/**
	 * Retrieves  platform type:   0 - PLATFORM_ID_ME
	 *                             1 - PLATFORM_ID_SEC
	 *                             2 - PLATFORM_ID_CSE
	 * @return platform type
	 */
	public final int getPlatformId() 
	{
		return platformId;
	}
	
}
