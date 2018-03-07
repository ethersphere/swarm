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

#include "bh_shared_types.h"

static int char2hex(char c)
{
    if (c>='0' && c<='9') return (c - '0');
    else if (c>='a' && c<='f') return (c - 'a' + 0xA);
    else return (c - 'A' + 0xA);
}

static int string_check1_uuid(const char* str)
{
    int i;

    for(i=0; i<BH_GUID_LENGTH*2; i++, str++) {
        if(! ((*str >= '0' && *str <= '9') || 
              (*str >= 'a' && *str <= 'f') || 
              (*str >= 'A' && *str <= 'F')))
            return 0;
    }
    if (*str != 0) return 0; //incorrect string length

    return 1;
}

static int string_check2_uuid(const char* str)
{
    int i;

    for(i=0; i<BH_GUID_LENGTH*2; i++, str++) {
        if (*str == '-' && (i==8 || i==12 || i==16 || i== 20))
            str++;
        if (! ((*str >= '0' && *str <= '9') || 
               (*str >= 'a' && *str <= 'f') || 
               (*str >= 'A' && *str <= 'F')))
            return 0;
    }
    if (*str != 0) return 0; //incorrect string length

    return 1;
}

BH_I32 hexstring_to_binary(const BH_I8* str, BH_U32 str_len, BH_I8* out)
{
    unsigned int i, hex_len = str_len / 2;
    /*str_len must be even, so one byte could be got from two hex char*/
    if (str_len & 1)
        return 0;

    for(i=0; i<hex_len; i++, out++)
    {
        if(! ((*str >= '0' && *str <= '9') || 
              (*str >= 'a' && *str <= 'f') || 
              (*str >= 'A' && *str <= 'F')))
            return 0;
        *out = char2hex(*str++);
        *out <<= 4;

        if(! ((*str >= '0' && *str <= '9') || 
              (*str >= 'a' && *str <= 'f') || 
              (*str >= 'A' && *str <= 'F')))
            return 0;
        *out += char2hex(*str++);
    }

    return 1;
}

BH_I32 string_to_uuid(const BH_I8* str, BH_I8* uuid)
{
    int i;
    if (!string_check1_uuid(str) && !string_check2_uuid(str)) return 0;

    for(i=0; i<BH_GUID_LENGTH; i++, uuid++) {
        if(*str == '-')	str++;

        *uuid = char2hex(*str++);
        *uuid <<= 4;
        *uuid += char2hex(*str++);
    }
    return 1;
}

static int hex2asc (char c)
{
    if (c < 10) return '0' + c;
    else return 'a' + c - 10;
}

void uuid_to_string(const BH_I8* uuid, BH_I8* str)
{
    int i;

    str[BH_GUID_LENGTH * 2] = 0;
    for (i=0; i<BH_GUID_LENGTH; i++, uuid++) {
        *str++ = hex2asc((*uuid & 0xf0) >> 4); 
        *str++ = hex2asc(*uuid & 0xf);
    }
}
