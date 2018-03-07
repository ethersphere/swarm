This folder contains mixins and sepolicy settings for enabling DAL on Android.
To enable DAL you should put the folder content in the following location:

- mixins/dal is the mixins for dal.
	put the inner folder "dal" under device/intel/mixins/groups/
- sepolicy/dal is the sepolicy settings for dal.
	put the inner folder "dal" under: device/intel/sepolicy/
	
You should also have to modify the mixin.spec file of your target product to include dal.
Find the mixins.spec file under device/intel/<YOUR_TARGET>
(e.g. for cht rvp it is under device/intel/cherrytrail/cht_rvp).
add line "dal: true" at the end of the spec file.
You can see the add_dal_to_mixins.patch