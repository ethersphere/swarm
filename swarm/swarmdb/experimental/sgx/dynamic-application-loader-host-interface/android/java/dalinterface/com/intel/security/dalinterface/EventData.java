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
 * Class EventData implements data container,
 * received in asynchronous event
 */

public final class EventData 
{
	public static final int DATA_FROM_APPLET = 0x0;
	public static final int DATA_FROM_SERVICE = 0x1;
	
	private final long handle;
	private final byte[] data;
	private final byte dataType;

	/**
	 * Public constructor
	 * @param handle - Session handle;
	 * @param data - Event data buffer;
	 * @param dataType - Event data type:
	 * 					 DATA_FROM_APPLET  = 0x0
	 * 					 DATA_FROM_SERVICE = 0x1 
	 */
	public EventData(long handle, byte[] data, byte dataType) 
	{
		this.handle = handle;
		this.dataType = dataType;
		this.data = data;
	}
    /**
     * Retrieves Session handle from EventData container 
     * @return Session Handle 
     */
	public final long getHandle() 
	{
		return this.handle;
	}
	
	/**
     * Retrieves Event data type from EventData container 
     * @return data type : 0 - data from FW, 1 - data from Dal/JHI service  
     */
	public final byte getDataType() 
	{
		return this.dataType;
	}
    
	/**
     * Retrieves Event data  binary data from EventData container 
     * @return binary data 
     */
	public final byte[] getData() 
	{
		return this.data;
	}

}
