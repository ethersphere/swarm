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

#ifdef __cplusplus
extern "C" {
#endif

typedef enum _bh_hash_alg_t
{
   bh_hash_alg_unknown  = 0,
   bh_hash_alg_sha1     = 1,
   bh_hash_alg_sha256   = 2,
   bh_hash_alg_max      = 3
} bh_hash_alg_t;

unsigned int DAL_get_hash_len(int hash_alg) {
    if (bh_hash_alg_sha1 == hash_alg)
        return 20;
    else if (bh_hash_alg_sha256 == hash_alg)
        return 32;
    else
        return 0;
}

#ifdef __cplusplus
}
#endif
