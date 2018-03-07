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

#ifndef __DALP_SCHEMA_H
#define __DALP_SCHEMA_H


#define JHI_DALP_VALIDATION_SCHEMA "\
\
<?xml version=\"1.0\" encoding=\"utf-8\"?>\
<xsd:schema xmlns:xsd=\"http://www.w3.org/2001/XMLSchema\" targetNamespace=\"urn:dalp\">\
  <xsd:element name=\"package\">\
    <xsd:complexType>\
      <xsd:sequence>\
\
        <xsd:element name=\"packageInfo\">\
          <xsd:complexType>\
            <xsd:sequence>\
              <xsd:element name=\"name\" type=\"xsd:string\"/>\
              <xsd:element name=\"description\" type=\"xsd:string\"/>\
              <xsd:element name=\"vendor\" type=\"xsd:string\"/>\
              <xsd:element name=\"appletId\" type=\"xsd:string\"/>\
            </xsd:sequence>\
          </xsd:complexType>\
        </xsd:element>\
\
        <xsd:element name=\"applets\">\
          <xsd:complexType>\
            <xsd:sequence>\
\
              <xsd:element name=\"applet\" minOccurs=\"1\" maxOccurs=\"50\">\
                <xsd:complexType>\
                  <xsd:sequence>\
                    <xsd:element name=\"platform\" type=\"xsd:string\"/>\
                    <xsd:element name=\"appletVersion\" type=\"xsd:string\"/>\
                    <xsd:element name=\"fwVersion\" type=\"xsd:string\"/>\
                    <xsd:element name=\"appletBlob\" type=\"xsd:base64Binary\"/>\
                  </xsd:sequence>\
                </xsd:complexType>\
              </xsd:element>\
\
            </xsd:sequence>\
          </xsd:complexType>\
        </xsd:element>\
\
      </xsd:sequence>\
\
      <xsd:attribute name=\"dalpVersion\" type=\"xsd:string\" use=\"required\"/>\
    </xsd:complexType>\
  </xsd:element>\
\
</xsd:schema>\
\
"

#endif