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

#ifndef __APPLETSMANAGER_H
#define __APPLETSMANAGER_H

// The H-Files
#include <cstdint>
#include <ostream>
#include <string>
#include <vector>
#include <list>
#include <map>
#include "Locker.h"
#include "Singleton.h"
#include "jhi_i.h"
#include "jhi_version.h"
#include "GlobalsManager.h"

namespace intel_dal
{
	using std::string;
	using std::map;
	using std::pair;
	using std::list;
	using std::vector;
	using std::ostream;

	enum JHI_APPLET_STATUS 
	{
		NOT_INSTALLED,		// Not Intalled
		PENDING_INSTALL,	// Pendning Intalltion
		INSTALLED,			// Installed - No Active Sessions
		MAX_APP_STATES
	};

	struct AppletRecord
	{
		JHI_APPLET_STATUS	status;
		bool				sharedSessionSupport;	
		bool				sharedSessionSupportRetrievedFromFW;	
	};

	// File Extentions of applet file
	const string dalpFileExt = ".dalp";
	const string acpFileExt = ".acp";

	const string pendingHeader = "/PENDING-";


	class AppletsManager : public Singleton<AppletsManager>
	{
		friend class Singleton<AppletsManager>;
	private:

		/*
			cheack if a file name has a given extention (case insensitive.
			Paramters:
				file		[In]			file name.
				extention	[In]			the extention to compare to.

			Return:
				true - the extention match
				false - the extention doesnt match
				
		*/


		bool isAppletRecordPresent(const string& appletId);


		AppletsManager();
		~AppletsManager(void);

		// Key == AppletId
		map<string, AppletRecord>	_appletTable;
		Locker						_locker;

	public:

		friend ostream& operator <<(ostream& os, const AppletsManager& sm);

		/* 
			Prepering the Applet for installation - extracting the blob from the file
			and copying it under PENDING to the repository.

		Paramters:
			file		[In]			DALP file to open.
			appletBlobs	[Out]			a list of applet blobs to try to install in FW.
			appletId	[Out, Optional]	Applet Id. 
			isAcp		[In]			Determines whether the file is ACP or DALP.

		Return:
			JHI_SUCCESS on success, error code otherwise.
		*/
		JHI_RET prepareInstallFromFile(const FILESTRING& file, list< vector<uint8_t> >& appletBlobs,const string& appletId, bool isAcp);
		
		/* 
			Prepering the Applet for installation - extracting the blob from the buffer
			and copying it under PENDING to the repository.

		Paramters:
			appletBlob	[In]			the buffer contains the ACP.
			appletId	[In]	Applet Id.

		Return:
			JHI_SUCCESS on success, error code otherwise.
		*/
		JHI_RET prepareInstallFromBuffer(vector<uint8_t>& appletBlob, const string& appletId);

		/* 
			Marking that Download Applet Blob to FW was ok, Moving Applet Status to Installed

		Paramters:
			appletId	[In]	Applet Id.
			isAcp		[In]	optional, indicates wether the file is acp or dalp.

		Return:
			true	- able to complate Installtion
			false	- unable to complate Installtion

		*/
		bool completeInstall(const string& appletId, bool isAcp = false);

		void updateAppletsList();

		void updateSharedSessionSupport(const string& appletId);

		/* 
			Remove Applet from repstory

		Paramters:
			appletId	[In]	Applet Id.

		Return:
			true	- able to remove applet
			false	- unable to remove applet

		*/		
		bool remove(const string& appletId);

		/* 
			Index Operator (operator[]) Returen the Applet Record according to a given applet ID

		Paramters:
			appletId		[In]	Applet Id.
			AppletRecord	[Out]	Applet Record of the given ID

		Return:
			true	- able to get applet record
			false	- unable to get applet record

		*/		
		bool get(const string& appletId, AppletRecord& appRecord);

		/* 
			return a state of a given applet id

		Paramters:
			appletId		[In]	Applet Id.

		Return:
			the applet state.

		*/			
		JHI_APPLET_STATUS getAppletState(const string& appletId);

		/* 
			return a list of applet id's of all loaded applets
			in case there is none, an empty list is returned.

			NOTE: It is the caller responstablity to free all strings in the list!

		*/	
		void getLoadedAppletsList(list<string>& appletsList);

		/*
			read applet blob from a pack file into a vector;
		
		Paramters:
			filepath		[In]	the full path of the file
			appletSource    [In]    the file type
			appletBlob		[Out]	the applet blob
		Return:
			JHI_SUCCESS on success, error code otherwise.
		*/
		JHI_RET readFileAsBlob(const FILESTRING& filepath, list< vector<uint8_t> >& appletBlobs);

		/*
			read applet blob from a pack file or a dalp file into a vector;
		
		Paramters:
			filepath		[In]	the full path of the file
			appletSource    [In]    the file type
			appletBlob		[Out]	the applet blob
			isAcp			[In]	Indicates wether the file is acp or dalp.
		Return:
			JHI_SUCCESS on success, error code otherwise.
		*/
		JHI_RET getAppletBlobs(const FILESTRING& filepath, list< vector<uint8_t> >& appletBlobs, bool isAcp);

		void addAppRecordEntry(const string& AppId, const AppletRecord& record);

		bool compareFileExtention(const FILESTRING& file,const string& extention);

		/*
			returns whether an applet file exists in the repository or not.
		
		Paramters:
			appletId		[In]	the applet ID
			outFileName		[Out]	optional, the file name in the repository.
			isAcp			[Out]	Indicates wether the file found is acp or dalp.
		Return:
			true	- applet file exists in the repository
			false	- otherwise
		*/
		bool appletExistInRepository(IN const string& appletId,OUT FILESTRING* outFileName, OUT bool& isAcp);

		void resetAppletTable();

		/** 
			try to unload applets that has no sessions from the VM, the applet file will remain
			in the applets repository in order to load (install) it again when needed.

			Return:
				true if at least one applet unloaded, false otherwise
		
		**/	
		bool UnloadUnusedApplets();

		/**
			this function returns true if the given applet support shared session, false otherwise.

			Paramters:
				appletId		[In]	the applet ID
		**/
		bool isSharedSessionSupported(const string& appletId);
		
		/*
		Returns full path name for pending applet in repository.

		Parameters:
		appletId 	[In] the applet ID
		isAcp		[In]	Indicates wether the file is acp or dalp.
		Return:
		path name - pending applet full path name
		*/
		FILESTRING getPendingFileName(const string& appletId, bool isAcp = false);

		/*
		Returns full path name for applet in repository.

		Parameters:
		appletId 	[In] the applet ID
		isAcp		[In]	Indicates wether the file is acp or dalp.
		Return:
		path name - applet full path name
		*/
		FILESTRING getFileName(const string& appletId, bool isAcp = false); 

	};
}

#endif 

