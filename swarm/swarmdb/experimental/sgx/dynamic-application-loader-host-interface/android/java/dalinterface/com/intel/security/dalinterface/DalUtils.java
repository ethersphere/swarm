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

import java.io.IOException;
import java.lang.reflect.InvocationTargetException;

import android.os.RemoteException;

/**
 * Class DalItils is utilized for Dal Data validation 
 * and return codes unification
 */
final class DalUtils 
{
    /**
     * Encoding caught exception 
     * @param exception - Throwable exception instance
     * @return int - Exception encoding from DalConstants
     */
	final static int encodeException(Throwable exception)
	{
		int errorCode = DalConstants.DAL_INTERNAL_ERROR;

		// sample exception should be the real exceptions as RemoteException
		if ( exception instanceof NullPointerException ) {
				errorCode = DalConstants.DAL_INVALID_PARAMS;
		} else if ( exception instanceof RemoteException ) {
	            errorCode = DalConstants.DAL_REMOTE_EXCEPTION;
		} else if (exception instanceof IOException) {	
				errorCode = DalConstants.DAL_IO_EXCEPTION;
		} else if (exception instanceof IllegalAccessException) {	
				errorCode = DalConstants.DAL_ILLEGAL_ACCESS_EXCEPTION;
		} else if (exception instanceof InvocationTargetException) {	
				errorCode = DalConstants.DAL_INVOCATION_TARGET_EXCEPTION;
	    } else {
	    	    errorCode = DalConstants.DAL_THROWABLE_EXCEPTION;
	    }

		return errorCode;
	}
	/**
	 * Evaluate exception encoding value to text string 
	 * @param encoding: int, Dal Exception Encoding (DalConstants)
	 * @return String, containing Dal Exception name
	 */
	final static String exToString (int encoding) {
		String exception = null;
		if (encoding == DalConstants.DAL_INVALID_PARAMS) {
			exception = "DAL_INVALID_PARAMS";
		} else if (encoding == DalConstants.DAL_REMOTE_EXCEPTION) {
			exception = "DAL_REMOTE_EXCEPTION";
		} else if (encoding == DalConstants.DAL_IO_EXCEPTION) {
			exception = "DAL_IO_EXCEPTION";
		} else if (encoding == DalConstants.DAL_ILLEGAL_ACCESS_EXCEPTION) {
			exception = "DAL_ILLEGAL_ACCESS_EXCEPTION";
		} else if (encoding == DalConstants.DAL_INVOCATION_TARGET_EXCEPTION) {
			exception = "DAL_INVOCATION_TARGET_EXCEPTION";	
		} else {
			exception = "DAL_THROWABLE_EXCEPTION";
		}
		return exception;
	}

	/**
	 *  validate that a given appId is a valid sequence of 32 hex digits
	 * @param appId - String, containing App Id
	 * @return boolean: true  - valid AppId
	 * 					false - invalid AppId	 
	 */
	final static boolean isStringValidAppId(String appId) 
	{
		final int uuid_hex_str_len = DalConstants.APP_ID_LEN; 
		 

		if (appId == null) 
		{
			return false;
		}

		if (appId.length() == uuid_hex_str_len) 
		{
			for (int i = 0; i < uuid_hex_str_len; i++) 
			{
				char c = appId.charAt(i);
				if (!is_hex_digit(c)) 
				{
					return false;
				}
			}

			return true;
		}

		return false;
	}
	/**
	 * validate character value as hex digit
	 * @param c - char
	 * @return boolean: true  - valid character
	 * 					false - invalid character
	 */
	private final static boolean is_hex_digit(char c) 
	{
        if ('0' <= c && c <= '9') return true;
        if ('a' <= c && c <= 'f') return true;
        if ('A' <= c && c <= 'F') return true;

        return false;
    }

}
