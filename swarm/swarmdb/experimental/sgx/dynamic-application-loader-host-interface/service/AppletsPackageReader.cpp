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

#include "AppletsPackageReader.h"
#include "AppletsManager.h"
#include "dbg.h"
#include <sstream>
#include <algorithm>
#include "misc.h"
#include "string_s.h"

using std::copy;
using std::sort;
namespace intel_dal
{

	//init the private const - cannot be init'd in the .h file
	const string AppletsPackageReader::INVALID_PLATFORM_NAME = "INVALID_PLATFORM_NAME";

	AppletsPackageReader::AppletsPackageReader(const FILESTRING& packagePath)
	{
		_xmlReader = NULL;
		_packagePath = packagePath;

		_xmlReader = XmlReaderFactory::createInstance(JHI_DALP_VALIDATION_SCHEMA);
		if (NULL == _xmlReader)
		{
			TRACE0("Failed to receive IXmlReader instance\n");
			_packageValid = false;
		}
		else
		{
			_packageValid = (_xmlReader->LoadXml(_packagePath) && _xmlReader->Validate());
		}
	}

	AppletsPackageReader::~AppletsPackageReader()
	{
		if (_xmlReader != NULL)
			delete _xmlReader;
	}

	bool AppletsPackageReader::isPackageValid()
	{
		return _packageValid;
	}

	// The function returns true if all the versions format is valid (no matter if the currentVersion was updated or not)
	// and false otherwise
	//
	// In addition the function compare the current major version with the new major version
	// and update the current major version to be the highest version that is equal or lower than the FW major version.
	bool AppletsPackageReader::compareFWVersions(const string& fwVersion, int* currentMajorVersion, const string& newVersion)
	{
		int fwVer1, fwVer2, fwVer3;		// the FW version breaked to 3 parts
		int newVer1, newVer2, newVer3;     // the new version breaked to 3 parts 

		char c1, c2;

		if (newVersion.empty() || fwVersion.empty() || currentMajorVersion == NULL)
			return false;

		std::istringstream fwVersionStream(fwVersion);
		fwVersionStream >> fwVer1 >> c1 >> fwVer2 >> c2 >> fwVer3;
		if (fwVersionStream.fail())
			return false;

		std::istringstream newVersionStream(newVersion);
		newVersionStream >> newVer1 >> c1 >> newVer2 >> c2 >> newVer3;
		if (newVersionStream.fail())
			return false;

		// all versions are broken apart, start the comparison:

		// make sure the new version is below the fw version.
		if ((fwVer1 < newVer1) ||
			((fwVer1 == newVer1) && (fwVer2 < newVer2)) ||
			((fwVer1 == newVer1) && (fwVer2 == newVer2) && (fwVer3 < newVer3)))
			return true; // no update, just return that the version format is valid.

		// compare the new version with the current version
		if (newVer1 > *currentMajorVersion)
		{
			// update the current Major version to the new major version version
			*currentMajorVersion = newVer1;
		}

		return true;
	}

	bool compareAppletVersions(APPLET_DETAILS version_a, APPLET_DETAILS version_b)
	{
		// compare by fw version and then by applet version
		if ((version_a.fwVersion.Major > version_b.fwVersion.Major) ||
			((version_a.fwVersion.Major == version_b.fwVersion.Major) && (version_a.fwVersion.Minor > version_b.fwVersion.Minor)) ||
			((version_a.fwVersion.Major == version_b.fwVersion.Major) && (version_a.fwVersion.Minor == version_b.fwVersion.Minor) && (version_a.fwVersion.Hotfix > version_b.fwVersion.Hotfix)) ||
			((version_a.fwVersion.Major == version_b.fwVersion.Major) && (version_a.fwVersion.Minor == version_b.fwVersion.Minor) && (version_a.fwVersion.Hotfix == version_b.fwVersion.Hotfix) && (version_a.appVersion.majorVersion > version_b.appVersion.majorVersion)) ||
			((version_a.fwVersion.Major == version_b.fwVersion.Major) && (version_a.fwVersion.Minor == version_b.fwVersion.Minor) && (version_a.fwVersion.Hotfix == version_b.fwVersion.Hotfix) && (version_a.appVersion.majorVersion == version_b.appVersion.majorVersion) && (version_a.appVersion.minorVersion > version_b.appVersion.minorVersion)))
			return true;

		return false;
	}

	/*
	compare applet version ( only the applet version )
	will create the effect of the applets sorted in desecnding order (1.5, 1.4...)
	*/
	bool compareAppletVersionsSignOnce(const APPLET_DETAILS& versionA, const APPLET_DETAILS& versionB)
	{
		APPLET_VERSION A = versionA.appVersion;
		APPLET_VERSION B = versionB.appVersion;

		if ((A.majorVersion > B.majorVersion) ||
			((A.majorVersion == B.majorVersion) && (A.minorVersion >= B.minorVersion)))
			return true;

		return false;

	}

	/**
	* return all candidate blobs from the dalp file.
	* blobs will be returned in sorted manner such that the first applet will be installed first ( or at least a try to install it will be performed)
	*/
	bool AppletsPackageReader::getAppletBlobs(const string fwVersion, list<vector<uint8_t> >& blobsList)
	{
		bool ret = false;
		int numAppletRecords;
		int status;
		bool isSignOnceSupported;

		string platformName = getPlatformName();
		if (platformName == INVALID_PLATFORM_NAME)
		{
			return false;
		}

		if (!_packageValid)
			return false;

		string platformXPath = string("//applets/applet[normalize-space(platform) = \"") + platformName + string("\"]");
		numAppletRecords = _xmlReader->GetNodeCount(platformXPath);

		if (numAppletRecords < 1)
		{
			TRACE0("no applets records in DALP file match the current platform\n");
			return false;
		}

		//check if we're "SIGN_ONCE" or not
		status = isSignOnce(fwVersion, isSignOnceSupported);
		if (status != JHI_SUCCESS)
		{
			TRACE0("getAppletBlobs(): isSignOnce() failed\n");
			return false;
		}

		//sign once procedure.
		if (isSignOnceSupported)
		{
			ret = getSignOnceAppletBlobs(fwVersion, blobsList);
		}

		//non sign once procedure.
		else
		{
			ret = getNonSignOnceAppletBlobs(numAppletRecords, platformName, fwVersion, blobsList);
		}

		return ret;
	}

	bool AppletsPackageReader::getAppletAndFwVersionAsStruct(APPLET_DETAILS* version, string& appletVersionString, string& fwVersionString)
	{
		char c1, c2;

		std::istringstream appVersionStream(appletVersionString);
		appVersionStream >> version->appVersion.majorVersion >> c1 >> version->appVersion.minorVersion;
		if (appVersionStream.fail()) {
			TRACE0("invalid applet version in dalp file\n");
			return false;
		}

		std::istringstream fwVersionStream(fwVersionString);
		fwVersionStream >> version->fwVersion.Major >> c1 >> version->fwVersion.Minor >> c2 >> version->fwVersion.Hotfix;
		if (fwVersionStream.fail()) {
			TRACE0("invalid fw version in dalp file\n");
			return false;
		}

		return true;
	}

	bool AppletsPackageReader::getNonSignOnceAppletBlobs(int numAppletRecords, string platformName, const string fwVersion, list<vector<uint8_t> >& blobsList)
	{
		int i;
		bool valid = true, status, ret = false;
		int selectedMajorVersion = 0;
		std::stringstream majorVersionStream;

		do
		{
			// get the latest FW Major version that is compatible with the current FW version
			for (i = 1; i <= numAppletRecords; i++)
			{
				// convert i to string
				std::stringstream sstream;
				sstream << i;
				string snum = sstream.str();
				string platform_xpath = string("//applets/applet[normalize-space(platform) = \"") + platformName + string("\"]");

				// get the applet[i] fwVersion
				string applet_xpath = platform_xpath + string("[") + snum + string("]");
				string version_xpath = applet_xpath + "/fwVersion";
				string appFWVersion;
				status = _xmlReader->GetNodeText(version_xpath, appFWVersion);

				if (!status)
				{
					TRACE0("invalid applet record in DALP file\n");
					valid = false;
					break;
				}

				status = compareFWVersions(fwVersion, &selectedMajorVersion, appFWVersion);

				if (!status)
				{
					TRACE0("invalid applet fw version in DALP file\n");
					valid = false;
					break;
				}
			}

			if (!valid)
			{
				TRACE0("failed to find a compatible FW version in the DALP file\n");
				break;
			}

			if (selectedMajorVersion == 0)
			{
				// DALP is valid but no compatible versions where found.
				ret = true;
				break;
			}

			majorVersionStream << selectedMajorVersion;

			// get all the applet versions that match the selected FW Major vesion and sort them from highest to the lowest according to Fw version and applet version.
			list<APPLET_DETAILS> versionsList;
			valid = getMatchingAppletsToMajorFwVersion(selectedMajorVersion, versionsList);

			if (!valid)
			{
				TRACE0("failed getting all the applet versions that match the selected FW vesion\n");
				break;
			}

			// do the sort
			versionsList.sort(compareAppletVersions);

			// copy all the blobs according to the sorted version list 
			valid = copyBlobsFromList(selectedMajorVersion, versionsList, blobsList);
			if (!valid)
			{
				TRACE0("getNonSignOnceAppletBlobs(): copyBlobsFromList() failed.\n");
				break;
			}

			ret = true;

		} while (0);

		return ret;
	}

	bool AppletsPackageReader::getSignOnceAppletBlobs(const string& fwVersion, list<vector<uint8_t> >& blobsList)
	{

		bool status = false;

		do
		{
			// get all the applets with versions that match the sign once fw version ( 11.x.x )
			list<APPLET_DETAILS> versionsList;
			status = getMatchingAppletsToMajorFwVersion(SIGN_ONCE_FW__MAJOR_VERSION, versionsList);

			if (!status)
			{
				TRACE0("failed getting all the applet versions that match the SIGN_ONCE_FW_VERSION\n");
				break;
			}

			//remove all the apples with an API level that's higher than the one supported on the platform.
			status = removeHigherApiLevelApplets(versionsList);

			if (!status)
			{
				TRACE0("failed removing higher API level applets from list\n");
				break;
			}

			// sort ( only applet version )
			versionsList.sort(compareAppletVersionsSignOnce);

			//copy all candidate blobs.
			status = copyBlobsFromList(SIGN_ONCE_FW__MAJOR_VERSION, versionsList, blobsList);
			if (!status)
			{
				TRACE0("getSignOnceAppletBlobs(): copyBlobsFromList() failed.\n");
				break;
			}

		} while (0);

		return status;
	}

	int AppletsPackageReader::isSignOnce(const string& fwVersion, bool& result)
	{
		int major, minor, hotfix;		
		int status = JHI_SUCCESS;

		do
		{
			if (fwVersion.empty())
			{
				status = JHI_UNKNOWN_ERROR;
				break;
			}

			int count = sscanf_s(TrimString(fwVersion).c_str(), "%d.%d.%d", &major, &minor, &hotfix);
			if (count != 3)
			{
				status = JHI_UNKNOWN_ERROR;
				break;
			}

			//this is BH V1 or TL ==> no sign once
			if (major == VLV_FW_MAJOR_VERSION  || major == CHV_FW_MAJOR_VERSION   ||
				major == ME_7_FW_MAJOR_VERSION || major == ME_8_FW_MAJOR_VERSION  ||
				major == ME_9_FW_MAJOR_VERSION || major == ME_10_FW_MAJOR_VERSION)
			{
				result = false;
				break;
			}

			//we got to here ==> sign once.
			result = true;

		} while (0);

		return status;
	}

	//return a list of applets that match the given majorFwVersion
	bool AppletsPackageReader::getMatchingAppletsToMajorFwVersion(int majorFwVersion, list<APPLET_DETAILS>& versionsList)
	{
		bool status = false, valid = true;
		std::stringstream majorVersionStream;
		majorVersionStream << majorFwVersion;

		string platformName = getPlatformName();
		if (platformName == INVALID_PLATFORM_NAME)
		{
			return false;
		}

		// //applets/applet[normalize-space(platform) = "ME" and starts-with(normalize-space(fwVersion),'8.')]
		string appletVersionsXpath = string("//applets/applet[normalize-space(platform) = \"") + platformName + string("\" and starts-with(normalize-space(fwVersion),\"") + majorVersionStream.str() + string(".\")]");
		int numAppletRecords = _xmlReader->GetNodeCount(appletVersionsXpath);

		for (int i = 1; i <= numAppletRecords; i++)
		{
			// convert i to string
			std::stringstream sstream;
			sstream << i;
			string snum = sstream.str();

			string appletXpath = appletVersionsXpath + string("[") + snum + string("]");
			string appletVersionXpath = appletXpath + "/appletVersion";
			string fwVersionXpath = appletXpath + "/fwVersion";
			string fwVersion;
			string appletVersion;

			status = _xmlReader->GetNodeText(appletVersionXpath, appletVersion);

			if (!status)
			{
				TRACE0("invalid applet record in DALP file\n");
				valid = false;
				break;
			}

			status = _xmlReader->GetNodeText(fwVersionXpath, fwVersion);

			if (!status)
			{
				TRACE0("invalid applet record in DALP file\n");
				valid = false;
				break;
			}

			APPLET_DETAILS version = { { 0, 0 }, { 0, 0, 0, 0 }, 0 };
			if (!getAppletAndFwVersionAsStruct(&version, appletVersion, fwVersion))
			{
				TRACE0("invalid applet version in DALP file\n");
				valid = false;
				break;
			}

			version.indexInDalp = i;

			versionsList.push_back(version);
		}

		return valid;
	}

	bool AppletsPackageReader::removeHigherApiLevelApplets(list<APPLET_DETAILS>& appletsList)
	{
		int supportedApiLevel = getPlatformApiLevel();

		if (supportedApiLevel == INVALID_API_LEVEL)
		{
			return false;
		}

		//we have the supported API level, iterate and remove applets with higher API levels.
		auto it = appletsList.begin();

		while (it != appletsList.end())
		{
			//get applet's api level
			VERSION appletFwVersion = (*it).fwVersion;
			int appletApiLevel = appletFwVersion.Minor;

			//if it's higher than supported, remove
			if (appletApiLevel > supportedApiLevel)
			{
				it = appletsList.erase(it);
			}
			else
			{
				++it;
			}
		}

		return true;
	}

	int AppletsPackageReader::getPlatformApiLevel()
	{
		VM_Plugin_interface* plugin = NULL;
		UINT32 status = JHI_UNKNOWN_ERROR;
		int apiLevel = INVALID_API_LEVEL;

		//input to JHI_Plugin_QueryTeeMetadata
		unsigned char* metadata = NULL;
		unsigned int length = 0;

		if ((!GlobalsManager::Instance().getPluginTable(&plugin)) || (plugin == NULL))
		{
			TRACE0("getSupportedApiLevel(): getPluginTable() failed.");
			return INVALID_API_LEVEL;
		}

		status = plugin->JHI_Plugin_QueryTeeMetadata(&metadata, &length);
		if (status == TEE_STATUS_SUCCESS)
		{
			apiLevel = ((dal_tee_metadata*)metadata)->api_level;
			JHI_DEALLOC(metadata);
		}
		else
		{
			TRACE1("getSupportedApiLevel(): JHI_Plugin_QueryTeeMetadata() failed with status = %d", status);
		}

		return apiLevel;
	}

	//copy the blobs (in sorted order) to appletsBlobsList.
	bool AppletsPackageReader::copyBlobsFromList(int fwMajorVersion, list<APPLET_DETAILS>& sortedAppletsList, list<vector<uint8_t> >& appletsBlobsList)
	{
		long blobSize = 0;
		bool status = false;
		std::stringstream fwMajorVersionStream;


		string platformName = getPlatformName();
		if (platformName == INVALID_PLATFORM_NAME)
		{
			return false;
		}

		fwMajorVersionStream << fwMajorVersion;

		for (list<APPLET_DETAILS>::iterator it = sortedAppletsList.begin(); it != sortedAppletsList.end(); it++)
		{
			// convert indexInDalp to string
			std::stringstream sstream;
			sstream << (*it).indexInDalp;
			string snum = sstream.str();
			string appletVersionsXpath = string("//applets/applet[normalize-space(platform) = \"") + platformName + string("\" and starts-with(normalize-space(fwVersion),\"") + fwMajorVersionStream.str() + string(".\")]");

			string applet_xpath = appletVersionsXpath + string("[") + snum + string("]");
			string blobXpath = applet_xpath + "/appletBlob";

			uint8_t * appletblob = NULL;
			status = _xmlReader->GetNodeTextAsBase64(blobXpath, &appletblob, &blobSize);

			if (!status || appletblob == NULL)
			{
				TRACE0("failed reading applet blob from DALP file\n");
				break;
			}

			vector<uint8_t> blob;
			blob.resize(blobSize);
			copy(appletblob, appletblob + blobSize, blob.begin());

			JHI_DEALLOC(appletblob);
			appletblob = NULL;

			appletsBlobsList.push_back(blob);

		}

		return status;
	}

	string AppletsPackageReader::getPlatformName()
	{
		string platformName;
		_JHI_PLATFROM_ID platformID = GlobalsManager::Instance().getPlatformId();

		if (platformID == ME)
		{
			platformName = "ME";
		}
		else if (platformID == SEC)
		{
			platformName = "SEC";
		}
		else if (platformID == CSE)
		{
			platformName = "CSE";
		}
		else
		{
			TRACE1("Invalid platform ID - %d", platformID);
			platformName = INVALID_PLATFORM_NAME;
		}

		return platformName;
	}
}
