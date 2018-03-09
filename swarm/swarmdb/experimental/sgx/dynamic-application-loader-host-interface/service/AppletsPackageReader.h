/*
   Copyright 2010-2016 Intel Corporation

   This software is licensed to you in accordance
   with the agreement between you and Intel Corporation.

   Alternatively, you can use this file in compliance
   with the Apache license, Version 2.


   Apache License, Version 2.0

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

#ifndef __APPLETSPACKAGEREADER_H
#define __APPLETSPACKAGEREADER_H

#include <string>
#include <vector>
#include <list>
#include "typedefs.h"
#include "jhi_i.h"
#include "IXmlReader.h"
#include "dalpSchema.h"
#include "jhi_version.h"
#include "XmlReaderFactory.h"



namespace intel_dal
{
	using std::string;
	using std::vector;
	using std::list;

	typedef struct
	{
		int majorVersion;
		int minorVersion;
	} APPLET_VERSION;  // For internal usage

	typedef struct
	{
		APPLET_VERSION appVersion;
		VERSION  fwVersion;
		int indexInDalp;
	} APPLET_DETAILS; // For intrenal usage

	class AppletsPackageReader
	{

	private:

		//constant representing FW which supports sign once.
		static const int SIGN_ONCE_FW__MAJOR_VERSION = 11;
		static const int INVALID_API_LEVEL = -1;

		//constant representing an invalid platform name ( initialized in the .cpp file ).
		static const string INVALID_PLATFORM_NAME;

		//constants representing fw major versions.
		static const int VLV_FW_MAJOR_VERSION   = 1;
		static const int CHV_FW_MAJOR_VERSION   = 2;
		static const int ME_7_FW_MAJOR_VERSION  = 7;
		static const int ME_8_FW_MAJOR_VERSION  = 8;
		static const int ME_9_FW_MAJOR_VERSION  = 9;
		static const int ME_10_FW_MAJOR_VERSION = 10;


		FILESTRING _packagePath;
		bool _packageValid;
		IXmlReader* _xmlReader;

		bool compareFWVersions(const string& fwVersion, int* currentMajorVersion, const string& newVersion);

		/*
		Trasform the applet and fw version strings to a APPLET_DETAILS struct.
		Return true on success, false otherwise.
		*/
		bool getAppletAndFwVersionAsStruct(APPLET_DETAILS* version, string& appletVersionString, string& fwVersionString);

		/*
		get all applet blobs that don't support sign once.
		*/
		bool getNonSignOnceAppletBlobs(int numAppletRecords, string platformName, const string fwVersion, list<vector<uint8_t> >& blobsList);

		/*
		get all applet blobs that support sign once.
		*/
		bool getSignOnceAppletBlobs(const string& fwVersion, list<vector<uint8_t> >& blobsList);

		/*
		return a list of applets that match the given majorFwVersion
		*/
		bool getMatchingAppletsToMajorFwVersion(int majorFwVersion, list<APPLET_DETAILS>& versionsList);

		/*
		remove from given appletsList all applets with higher API level than supported API level on platform.
		*/
		bool removeHigherApiLevelApplets(list<APPLET_DETAILS>& appletsList);

		/*
		return a string representing the platform's name, i.e ME, CSE..
		*/
		string getPlatformName();

		/*
		return supported API level on platform ( using query tee meta data ).
		*/
		int getPlatformApiLevel();

		/*
		copy the blobs in sorted order to the output list.
		*/
		bool copyBlobsFromList(int fwMajorVersion, list<APPLET_DETAILS>& sortedAppletsList, list<vector<uint8_t> >& appletsBlobsList);
		/*
		return true if sign once is supported, false otherwise
		*/
		int isSignOnce(const string& fwVersion, bool& result);



		// disabling copy constructor and assignment operator by declaring them as private
		AppletsPackageReader&  operator = (const AppletsPackageReader& other) { return *this; }
		AppletsPackageReader(const AppletsPackageReader& other) { }

	public:
		AppletsPackageReader(const FILESTRING& packagePath);
		~AppletsPackageReader();

		/*
		Validate the packge file against an xsl template

		Return:
		true - package file is valid
		false - otherwise
		*/
		bool isPackageValid();

		/*
		reads an applet blob form a dalp file.
		the blob is selected accrding to the fw_version
		Paramters:
		fw_version	[In]		the FW version.
		blobsList		[Out]		the applet blob.

		Return:
		true - blob read successfuly
		false - read failed.

		*/
		bool getAppletBlobs(const string fwVersion, list<vector<uint8_t> >& blobsList);

	};
}

#endif 

