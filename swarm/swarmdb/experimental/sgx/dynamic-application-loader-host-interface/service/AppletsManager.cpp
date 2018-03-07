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

#include "AppletsManager.h"
#include "misc.h"
#include <stdio.h>
#include "AppletsPackageReader.h"
#include "reg.h"
#ifdef _WIN32
#include <io.h>
#else
#include <string.h>
#include <sys/stat.h>
#include "string_s.h"
#include <ctype.h>
#include <sys/types.h>
#include <dirent.h>
#include <errno.h>
#endif //_WIN32


#include <fstream>
#include <iterator>
#include <algorithm>


namespace intel_dal
{

	AppletsManager::AppletsManager() : _appletTable() {}

	AppletsManager::~AppletsManager(void) {}

	JHI_RET AppletsManager::prepareInstallFromFile(const FILESTRING& file, list<vector<uint8_t> >& appletBlobs,const string& appletId, bool isAcp)
	{
		JHI_RET ulRetCode = JHI_SUCCESS;
		int copy_status;
		FILESTRING DstFile;

		do
		{
			// copy the file to the repository and set applet source
			DstFile = getPendingFileName(appletId, isAcp);

#ifdef _WIN32

			copy_status = CopyFile(file.c_str(),(LPCWSTR) DstFile.c_str(), FALSE);

			if (!copy_status) // zero if copy file fails
			{
				TRACE0 ("Copy file to repository failed!!\n");
				ulRetCode = JHI_FILE_ERROR_COPY;
				break;
			} 

			// remove all attributes (readonly, hidden etc.) from the copied file
			if (SetFileAttributes((LPCWSTR) DstFile.c_str(), FILE_ATTRIBUTE_NORMAL) == 0)
			{
				TRACE0 ("failed removing all attributes from file\n");
				ulRetCode = JHI_FILE_ERROR_COPY;
				break;
			}


#else //!_win32
			copy_status = JhiUtilCopyFile(DstFile.c_str(), file.c_str());
			if (copy_status)
			{
				TRACE0 ("Copy file to repository failed!!\n");
				ulRetCode = JHI_FILE_ERROR_COPY;
				break;
			}
			if (chmod(DstFile.c_str(), S_IRWXO | S_IRWXG | S_IRWXU) != 0)
			{
				TRACE0 ("failed removing all attributes from file\n");
				ulRetCode = JHI_FILE_ERROR_COPY;
				break;
			}
#endif //win32


			// 3. get the applet blob from the dalp file
			ulRetCode = getAppletBlobs(DstFile,appletBlobs, isAcp);
			if (ulRetCode != JHI_SUCCESS)
			{
				TRACE0("failed getting applet blobs from dalp file\n");
				break;
			}

			//4. if the applet is not installed (we dont have a record in the app table) 
			//   create entry for the applet under its ID and set its state to PENDING.
			//   otherwise do nothing (the applet is installed but we install it again in case there is a version update)
			if (getAppletState(appletId) == NOT_INSTALLED)
			{
				AppletRecord record;
				record.status = PENDING_INSTALL;
				record.sharedSessionSupport = false;
				record.sharedSessionSupportRetrievedFromFW = false;
				addAppRecordEntry(appletId,record);
			}
		}
		while(0);

		// cleanup
		if (ulRetCode != JHI_SUCCESS)
		{
			// delete the file copied to the repository
			if ((!DstFile.empty()) && (_waccess_s(DstFile.c_str(),0) == 0))
			{
				_wremove(DstFile.c_str());
			}
		}

		return ulRetCode;
	}

	JHI_RET AppletsManager::prepareInstallFromBuffer(vector<uint8_t>& appletBlob, const string& appletId)
	{
		JHI_RET ulRetCode = JHI_SUCCESS;
		FILESTRING DstFile;

		do
		{
			// copy the file to the repository and set applet source
			DstFile = getPendingFileName(appletId, true);

#ifdef _WIN32
			std::fstream fWriter(DstFile.c_str(), std::ios::out | std::ios::binary);
			fWriter.write((const char*)&appletBlob[0], appletBlob.size());
			fWriter.close();

			// verify the applet file
			if (_waccess_s(DstFile.c_str(), 0) != 0)
			{
				TRACE0("prepere install failed - applet file not written properly");
				ulRetCode = JHI_FILE_ERROR_COPY;
				break;
			}

			// remove all attributes (readonly, hidden etc.) from the copied file
			if (SetFileAttributes((LPCWSTR) DstFile.c_str(), FILE_ATTRIBUTE_NORMAL) == 0)
			{
				TRACE0 ("failed removing all attributes from file\n");
				ulRetCode = JHI_FILE_ERROR_COPY;
				break;
			}

#else //!_WIN32
			if (JhiUtilCreateFile_fromBuff (DstFile.c_str(),reinterpret_cast<const char*>(&appletBlob[0]), appletBlob.size()))
			{
				TRACE0("prepere install failed - applet file is not created");
				ulRetCode = JHI_FILE_ERROR_COPY;
				break;
			}
			if (_waccess_s(DstFile.c_str(), 0) != 0)
			{
				TRACE0("prepere install failed - applet file not written properly");
				ulRetCode = JHI_FILE_ERROR_COPY;
				break;
			}

			if (chmod(DstFile.c_str(), S_IRWXO | S_IRWXG | S_IRWXU) != 0)
			{
				TRACE0 ("failed removing all attributes from file\n");
				ulRetCode = JHI_FILE_ERROR_COPY;
				break;
			}

#endif //_WIN32

			//4. if the applet is not installed (we dont have a record in the app table) 
			//   create entry for the applet under its ID and set its state to PENDING.
			//   otherwise do nothing (the applet is installed but we install it again in case there is a version update)
			if (getAppletState(appletId) == NOT_INSTALLED)
			{
				AppletRecord record;
				record.status = PENDING_INSTALL;
				record.sharedSessionSupport = false;
				record.sharedSessionSupportRetrievedFromFW = false;
				addAppRecordEntry(appletId,record);
			}
		}
		while(0);

		// cleanup
		if (ulRetCode != JHI_SUCCESS)
		{
			// delete the file copied to the repository
			if ((!DstFile.empty()) && (_waccess_s(DstFile.c_str(),0) == 0))
			{
				_wremove(DstFile.c_str());
			}
		}

		return ulRetCode;
	}

	bool AppletsManager::compareFileExtention(const FILESTRING& file, const string& extention)
	{
		FILESTRING ext; 
		size_t index = file.rfind('.');

		if (index==string::npos)
			return false; // no extention found.

		ext = file.substr(index); // get the extention

		if ( ext.size() != extention.size() ) // compare sizes
			return false;

		for ( string::size_type i = 0; i < extention.size(); ++i ) //compare chars
			if (toupper(extention[i]) != toupper(ext[i]))
				return false;

		return true;
	}

	void AppletsManager::addAppRecordEntry(const string& AppId, const AppletRecord& record)
	{
		_locker.Lock();
		_appletTable.insert(pair<string, AppletRecord>(AppId, record));
		_locker.UnLock();
	}

	bool AppletsManager::completeInstall(const string& appletId, bool isAcp)
	{
		int result;

		// rename the applet file in the repository from PENNDING_<UUID>.dalp to <UUID>.dalp
		FILESTRING pendingFileName = getPendingFileName(appletId, isAcp);
		FILESTRING newfilename = getFileName(appletId, isAcp);
		FILESTRING otherExistingFilename = getFileName(appletId, !isAcp); // needed to remove old file in case it was with a different extension.

		// delete an exsiting file with newfilename
		_wremove(newfilename.c_str()); // no need to check since rename will fail
		_wremove(otherExistingFilename.c_str()); // no need to check since rename will fail

		// rename the temp file to the newfilename
		result = _wrename( pendingFileName.c_str() , newfilename.c_str() );

		if ( result != 0 )
		{
			TRACE0("rename file failed\n");
			return false;
		}

		// change the status in the applet table to INSTALLED
		_locker.Lock();
		_appletTable[appletId].status = INSTALLED;
		_locker.UnLock();

		return true;
	}

	bool AppletsManager::appletExistInRepository(IN const string& appletId, OUT FILESTRING* outFileName, OUT bool& isAcp)
	{
		bool exists = false;

		FILESTRING dalpfilename = getFileName(appletId, false);
		FILESTRING acpfilename = getFileName(appletId, true);

		if (_waccess_s(dalpfilename.c_str(), 0) == 0)
		{
			exists = true;
			isAcp = false;

			if (outFileName != NULL)
			{
				*outFileName = dalpfilename;
			}
		}
		else
		{
			if (_waccess_s(acpfilename.c_str(), 0) == 0)
			{
				exists = true;
				isAcp = true;

				if (outFileName != NULL)
				{
					*outFileName = acpfilename;
				}
			}
		}
		return exists;
	}

	bool AppletsManager::remove(const string& appletId)
	{
		_locker.Lock();
		size_t ret = _appletTable.erase(appletId);
		_locker.UnLock();

		return (ret != 0);
	}


	bool AppletsManager::get(const string& appletId, AppletRecord& appRecord)
	{
		bool status = true;

		_locker.Lock();

		do
		{
			if (!isAppletRecordPresent(appletId))
			{
				status = false;
				break;
			}
			appRecord = _appletTable[appletId];
		}
		while(0);

		_locker.UnLock();

		return status;
	}

	bool AppletsManager::isAppletRecordPresent(const string& appletId)
	{
		map<string, AppletRecord>::iterator it;

		it = _appletTable.find(appletId);

		return (it != _appletTable.end());
	}

	JHI_RET AppletsManager::readFileAsBlob(const FILESTRING& filepath, list< vector<uint8_t> >& appletBlobs)
	{
		JHI_RET ret = JHI_INVALID_PARAMS;
		std::ifstream is(filepath.c_str(), std::ios::binary);

		if (!is)
		{
			return JHI_INTERNAL_ERROR;
		}

		try
		{
			is >> std::noskipws;
			is.seekg (0, is.end);
			std::streamoff len = is.tellg();
			is.seekg (0, is.beg);

			if (len >= MAX_APPLET_BLOB_SIZE)
			{
				ret = JHI_INVALID_PACKAGE_FORMAT;
			}
			std::istream_iterator<uint8_t> start(is), end;
			vector<uint8_t> blob(start, end);
			appletBlobs.push_back(blob);
			is.close();
			ret = JHI_SUCCESS;
		}
		catch(...)
		{
			if (is.is_open())
			{
				is.close();				
			}
			ret =  JHI_INVALID_PARAMS;
		}

		return ret;
	}

	JHI_RET AppletsManager::getAppletBlobs(const FILESTRING& filepath, list< vector<uint8_t> >& appletBlobs, bool isAcp)
	{
		VERSION fwVersion = GlobalsManager::Instance().getFwVersion();

		if (isAcp)
		{
			return readFileAsBlob(filepath, appletBlobs);
		}

		char FWVersionStr[FW_VERSION_STRING_MAX_LENGTH];

		if (compareFileExtention(filepath,dalpFileExt))
		{
			AppletsPackageReader reader(filepath);

			if (!reader.isPackageValid())
			{
				TRACE0 ("Invalid package file received\n");
				return JHI_INVALID_PACKAGE_FORMAT;
			}

			// create a fw version string to compare against the versions in the dalp file.
			sprintf_s(FWVersionStr, FW_VERSION_STRING_MAX_LENGTH, "%d.%d.%d",fwVersion.Major,fwVersion.Minor,fwVersion.Hotfix);

			if (!reader.getAppletBlobs(FWVersionStr,appletBlobs))
			{
				TRACE0 ("get applet blob from dalp file failed!!\n");
				return JHI_READ_FROM_FILE_FAILED;
			}

			if (appletBlobs.empty())
			{
				TRACE0 ("No compatible applets where found in the dalp file\n");
				return JHI_INSTALL_FAILED;
			}
		}
		else
		{
			return JHI_INVALID_FILE_EXTENSION;
		}

		return JHI_SUCCESS;
	}


	JHI_APPLET_STATUS AppletsManager::getAppletState(const string& appletId)
	{
		JHI_APPLET_STATUS status;

		_locker.Lock();

		do
		{
			if (!isAppletRecordPresent(appletId))
			{
				status = NOT_INSTALLED;
				break;
			}

			status = _appletTable[appletId].status;
		}
		while(0);

		_locker.UnLock();

		return status;
	}

	bool AppletsManager::isSharedSessionSupported(const string& appletId)
	{
		bool shareSupported = false;

		_locker.Lock();

		if (isAppletRecordPresent(appletId))
		{
			if (!_appletTable[appletId].sharedSessionSupportRetrievedFromFW)
			{
				updateSharedSessionSupport(appletId);
			}
			shareSupported = _appletTable[appletId].sharedSessionSupport;
		}

		_locker.UnLock();

		return shareSupported;
	}

	ostream& operator <<(ostream& os, const AppletsManager& am)
	{	
		map<string, AppletRecord>::const_iterator it;

		for ( it = am._appletTable.begin() ; it != am._appletTable.end(); it++ )
		{
			os << "Applet ID: "<<(*it).first << "\n";
			os << "Session State: "<<(*it).second.status << "\n";
			os << "\n";
		}

		return os;
	}

	void AppletsManager::resetAppletTable()
	{
		_locker.Lock();
		_appletTable.clear();
		_locker.UnLock();
	}

	bool AppletsManager::UnloadUnusedApplets()
	{
		list<string> appidList;
		map<string, AppletRecord>::const_iterator it;
		bool unloaded = false;
		char Appid[LEN_APP_ID+1];

		_locker.Lock();

		for ( it = _appletTable.begin() ; it != _appletTable.end(); it++ )
		{
			if (it->second.status == INSTALLED)
				appidList.push_back(it->first);
		}

		for (list<string>::iterator list_it = appidList.begin(); list_it != appidList.end(); list_it++)
		{
			strcpy_s(Appid,LEN_APP_ID + 1, list_it->c_str());
			if (jhis_unload(Appid) == JHI_SUCCESS)
			{
				TRACE1("unloaded applet with appid: %s\n",Appid);
				unloaded = true;
			}
		}

		_locker.UnLock();

		return unloaded;
	}

	void AppletsManager::updateSharedSessionSupport(const string& appletId)
	{
		char appId[LEN_APP_ID+1];
		JVM_COMM_BUFFER ioBuffer;
		const char * appProperty = "applet.shared.session.support";
		const int responseLen = 6;
		char responseStr[responseLen];
		bool shareEnabled = false;

		JHI_RET status;

		strcpy_s(appId,LEN_APP_ID+1,appletId.c_str());

		ioBuffer.TxBuf->buffer = (void*)appProperty;
		ioBuffer.TxBuf->length = (uint32_t)strlen(appProperty)+1;

		ioBuffer.RxBuf->buffer = responseStr;
		ioBuffer.RxBuf->length = responseLen;

		status = jhis_get_applet_property(appId,&ioBuffer);

		if (status == JHI_SUCCESS)
		{
			_appletTable[appletId].sharedSessionSupportRetrievedFromFW = true;
			if (strcmp(responseStr, "true") == 0)
				shareEnabled = true;
		}

		_locker.Lock();
		_appletTable[appletId].sharedSessionSupport = shareEnabled;
		_locker.UnLock();
	}

	void AppletsManager::updateAppletsList()
	{
		vector<string> uuidsInFw, uuidsInRepo;
		FILESTRING repositoryDir;
#ifdef _WIN32
		// code based on http://msdn.microsoft.com/en-us/library/windows/desktop/aa365200(v=vs.85).aspx
		WIN32_FIND_DATA ffd;
		FILESTRING searchStr;
		HANDLE hFind = INVALID_HANDLE_VALUE;

		GlobalsManager::Instance().getAppletsFolder(repositoryDir);
		repositoryDir.append(FILE_SEPERATOR);

		do	// search dalp files
		{
			TRACE0("Searching dalp TAs in the repository...");
			searchStr = FILESTRING(repositoryDir + FILEPREFIX("*") + ConvertStringToWString(dalpFileExt));

			// Gets the first file in the folder
			hFind = FindFirstFile(searchStr.c_str(), &ffd);

			if (INVALID_HANDLE_VALUE == hFind) 
			{
				TRACE0("FindFirstFile failed.");
				break;
			}

			do
			{
				ffd.cFileName[LEN_APP_ID] = FILEPREFIX('\0');
				string fileName(ConvertWStringToString(ffd.cFileName));
				if (validateUuidString(fileName))
				{
					TRACE1("The TA %s was found in the repository.", fileName.c_str());
					uuidsInRepo.push_back(fileName);
				}
			}
			// Continue on the next dalp file in the folder.
			while (FindNextFile(hFind, &ffd) != 0);
		} while(0);

		do	// search acp files
		{
			TRACE0("Searching acp TAs in the repository...");
			searchStr = FILESTRING(repositoryDir + FILEPREFIX("*") + ConvertStringToWString(acpFileExt));

			// Gets the first file in the folder
			hFind = FindFirstFile(searchStr.c_str(), &ffd);

			if (INVALID_HANDLE_VALUE == hFind) 
			{
				TRACE0("FindFirstFile failed.");
				break;
			}

			do
			{
				ffd.cFileName[LEN_APP_ID] = FILEPREFIX('\0');
				string fileName(ConvertWStringToString(ffd.cFileName));
				if (validateUuidString(fileName))
				{
					TRACE1("The TA %s was found in the repository.", fileName.c_str());
					uuidsInRepo.push_back(fileName);
				}
			}
			// Continue on the next acp file in the folder.
			while (FindNextFile(hFind, &ffd) != 0);
		} while (0);

		// Register the found applets.
		for (auto uuid = uuidsInRepo.begin(); uuid != uuidsInRepo.end(); ++uuid)
		{ 
			if (_stricmp(uuid->c_str(), SPOOLER_APPLET_UUID) == 0)
			{
				continue;
			}
			AppletRecord record;
			record.status = INSTALLED;
			record.sharedSessionSupport = false;
			record.sharedSessionSupportRetrievedFromFW = false;
			addAppRecordEntry(*uuid, record);
		}

#else //ANDROID
		DIR *dir;
		struct dirent *entry;
       		struct stat info;

		GlobalsManager::Instance().getAppletsFolder(repositoryDir);
		repositoryDir.append(FILE_SEPERATOR);

		if ((dir = opendir(const_cast < const char*>(repositoryDir.c_str()))) == NULL)
			TRACE2("Cannot open applets repository dir %s, %s\n",
			       repositoryDir.c_str(), strerror(errno));
		else {
			while ((entry = readdir(dir)) != NULL) {
				std::string filename (entry->d_name);
				std::string appName = repositoryDir + filename;
				if (stat(appName.c_str(), &info) != 0) {
					TRACE2 ("Can't stat %, %s\n", appName.c_str(),strerror(errno));
					continue;
				}

				if ((filename.find (dalpFileExt)) == LEN_APP_ID + 1) {
					uuidsInRepo.push_back(filename.substr (0, LEN_APP_ID));
					//std::string uuid = filename.substr (0, LEN_APP_ID);
					//jhis_install (const_cast <const char*>(uuid.c_str()), const_cast <const FILECHAR*>(appName.c_str()), true, false);
				} else if ((filename.find (acpFileExt)) == LEN_APP_ID + 1) {
					uuidsInRepo.push_back(filename.substr (0, LEN_APP_ID));
					//std::string uuid = filename.substr (0, LEN_APP_ID);
					//jhis_install (const_cast <const char*>(uuid.c_str()), const_cast <const FILECHAR*>(appName.c_str()), true, true);
				} else {
					 continue;
				}
			}
			closedir(dir);
		}
		for (size_t ii = 0; ii < uuidsInRepo.size(); ii++)
		{
			if (strcmp(uuidsInRepo[ii].c_str(), SPOOLER_APPLET_UUID) == 0)
			{
				continue;
			}
			AppletRecord record;
			record.status = INSTALLED;
			record.sharedSessionSupport = false;
			record.sharedSessionSupportRetrievedFromFW = false;
			addAppRecordEntry(uuidsInRepo[ii], record);
		}

#endif //_WIN32

		return;
	}



	void AppletsManager::getLoadedAppletsList(list<string>& appletsList)
	{
		map<string, AppletRecord>::const_iterator it;
		list<string> applets;

		_locker.Lock();

		for ( it = _appletTable.begin() ; it != _appletTable.end(); it++ )
		{
			if (it->second.status == INSTALLED)
				appletsList.push_back(it->first);
		}

		_locker.UnLock();
	}

	FILESTRING AppletsManager::getPendingFileName(const string& appletId, bool isAcp)
	{
		FILESTRING repositoryDir;
		GlobalsManager::Instance().getAppletsFolder(repositoryDir);
		string fileExt;
		if (isAcp)
		{
			fileExt = acpFileExt;
		}
		else
		{
			fileExt = dalpFileExt;
		}

		return repositoryDir + ConvertStringToWString(pendingHeader + appletId + fileExt);
	}

	FILESTRING AppletsManager::getFileName(const string& appletId, bool isAcp)
	{
		FILESTRING repositoryDir;
		GlobalsManager::Instance().getAppletsFolder(repositoryDir);
		string fileExt;
		if (isAcp)
		{
			fileExt = acpFileExt;
		}
		else
		{
			fileExt = dalpFileExt;
		}
		return repositoryDir + ConvertStringToWString("/" + appletId + fileExt);
	}
}