// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package http

import (
	"encoding/hex"
	"fmt"
	"html/template"
	"path"

	"github.com/ethersphere/swarm/api"
)

type htmlListData struct {
	URI  *api.URI
	List *api.ManifestList
}

var TemplatesMap = make(map[string]*template.Template)
var faviconBytes []byte

func init() {
	for _, v := range []struct {
		templateName string
		partial      string
		funcs        template.FuncMap
	}{
		{
			templateName: "error",
			partial:      errorResponse,
		},
		{
			templateName: "bzz-list",
			partial:      bzzList,
			funcs: template.FuncMap{
				"basename": path.Base,
				"leaflink": leafLink,
			},
		},
		{
			templateName: "landing-page",
			partial:      landing,
		},
	} {
		TemplatesMap[v.templateName] = template.Must(template.New(v.templateName).Funcs(v.funcs).Parse(baseTemplate + js + css + v.partial + logo))
	}

	bytes, err := hex.DecodeString(favicon)
	if err != nil {
		panic(err)
	}
	faviconBytes = bytes
}

func leafLink(URI api.URI, manifestEntry api.ManifestEntry) string {
	return fmt.Sprintf("/bzz:/%s/%s", URI.Addr, manifestEntry.Path)
}

const bzzList = `{{ define "content" }}
<h3 class="top-space">Swarm index of {{ .URI }}</h3>
<hr>
<table>
	<thead>
		<tr>
			<th>Path</th>
			<th>Type</th>
			<th>Size</th>
		</tr>
	</thead>

	<tbody>
		{{ range .List.CommonPrefixes }}
		<tr>
			<td>
				<a class="normal-link" href="{{ basename . }}/">{{ basename . }}/</a>
			</td>
			<td>DIR</td>
			<td>-</td>
		</tr>
		{{ end }}
		{{ range .List.Entries }}
		<tr>
			<td>
				<a class="normal-link" href="{{ leaflink $.URI . }}">{{ basename .Path }}</a>
			</td>
			<td>{{ .ContentType }}</td>
			<td>{{ .Size }}</td>
		</tr>
		{{ end }}
</table>
<hr>

 {{ end }}`

const errorResponse = `{{ define "content" }}
<div class="errorContainer">
	<div class="errorHeader"><h1>Error</h1></div>
	<div class="errorMessage"><h3>{{.Msg}}</h3></div>

	<div class="errorCode"><h5>Error code: {{.Code}}</h5></div>
</div>
{{ end }}`

const landing = `{{ define "content" }}
<div class="control">
			<div class="controlHeader">
				<div id="controlHeaderDownload" class="controlHeaderItem active">
					<a>
						<span>Browse</span>
						<span class="controlHeaderItemIcon">
							<svg version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"
								 viewBox="0 0 20 20" style="enable-background:new 0 0 16 20;" xml:space="preserve">
								<path id="XMLID_1617_" d="M10,0C4.5,0,0,4.5,0,10s4.5,10,10,10c5.5,0,10-4.5,10-10C20,4.5,15.5,0,10,0z M9.8,17.9
			c-4.4,0-7.9-3.6-7.9-7.9c0-4.4,3.6-7.9,7.9-7.9c0.4,0,0.9,0,1.3,0.1l0,0.8c0,0.3-0.3,0.6-0.6,0.6l-1,0C9.3,3.6,9.1,3.7,9,3.9
			L8.7,4.4C8.6,4.5,8.5,4.6,8.3,4.6L7.7,4.8C7.5,4.8,7.3,5.1,7.3,5.4l0,0.2c0,0.3,0.3,0.6,0.6,0.6l3.6,0c0.2,0,0.3,0.1,0.4,0.2
			l0.3,0.3c0.1,0.1,0.3,0.2,0.4,0.2l0.4,0c0.3,0,0.6,0.3,0.6,0.6c0,0.3-0.2,0.5-0.4,0.6l-1.9,0.6c-0.2,0.1-0.3,0-0.5,0l-0.6-0.3
			c-0.3-0.2-0.6-0.2-1-0.2h0C8.8,8.1,8.3,8.2,8,8.5L6.9,9.3C6.3,9.8,6,10.4,6,11.1l0,0.6c0,0.6,0.2,1.1,0.6,1.5
			c0.4,0.4,1,0.6,1.5,0.6l1,0c0.3,0,0.6,0.3,0.6,0.6l0,1.2c0,0.5,0.1,1,0.3,1.4c0.2,0.4,0.6,0.6,1,0.6c0.4,0,0.7-0.2,0.9-0.5l0.5-0.8
			c0.3-0.4,0.6-0.8,1-1.2c0.1-0.1,0.2-0.2,0.2-0.3L14,14c0-0.1,0.1-0.2,0.1-0.3l0.7-1c0.1-0.1,0.1-0.2,0.1-0.4l0-0.5
			c0-0.3-0.3-0.6-0.6-0.6l-0.3,0c-0.2,0-0.4-0.1-0.5-0.3L13,10.2c-0.2-0.3-0.1-0.8,0.3-0.9l0.1,0c0.2-0.1,0.4,0,0.5,0.1l0.7,0.5
			c0.2,0.1,0.4,0.1,0.6,0l0.6-0.3c0.2-0.1,0.3-0.3,0.3-0.6l0-0.3c0-0.3,0.3-0.6,0.6-0.6l0.7,0c0.2,0.6,0.2,1.3,0.2,1.9
			C17.8,14.4,14.2,17.9,9.8,17.9z"/>
							</svg>
						</span>
					</a>
				</div>
				<div id="controlHeaderUpload" class="controlHeaderItem">
					<a>
						<span>Upload</span>
						<span class="controlHeaderItemIcon">
							<svg version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"
								 viewBox="0 0 16 20" style="enable-background:new 0 0 16 20;" xml:space="preserve">
								<path id="XMLID_1412_" d="M8.8,5.3V0H0.9C0.4,0,0,0.4,0,0.9v18.1C0,19.6,0.4,20,0.9,20h13.1c0.5,0,0.9-0.4,0.9-0.9V6.2
									H9.7C9.2,6.2,8.8,5.8,8.8,5.3z M11.3,13.8H8.8v3.1c0,0.3-0.3,0.6-0.6,0.6H6.9c-0.3,0-0.6-0.3-0.6-0.6v-3.1H3.7
									c-0.6,0-0.8-0.7-0.4-1.1L7,8.9c0.3-0.3,0.7-0.3,0.9,0l3.8,3.7C12.1,13.1,11.9,13.8,11.3,13.8z M14.7,4.1l-3.8-3.8
									C10.7,0.1,10.5,0,10.2,0H10v5h5V4.8C15,4.5,14.9,4.3,14.7,4.1z"/>
							</svg>
						</span>
					</a>
				</div>
				<div id="controlHeaderInfo" class="controlHeaderItem hidden">
					<a href="#">
						<span>Info</span>
						<span class="controlHeaderItemIcon">
							<svg version="1.1" id="Layer_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"
								 viewBox="0 0 20 20" style="enable-background:new 0 0 20 20;" xml:space="preserve">
							<path d="M10,0C4.5,0,0,4.5,0,10c0,5.5,4.5,10,10,10s10-4.5,10-10C20,4.5,15.5,0,10,0z M10,4.4c0.9,0,1.7,0.8,1.7,1.7
								S10.9,7.8,10,7.8S8.3,7.1,8.3,6.1S9.1,4.4,10,4.4z M12.3,14.7c0,0.3-0.2,0.5-0.5,0.5H8.2c-0.3,0-0.5-0.2-0.5-0.5v-1
								c0-0.3,0.2-0.5,0.5-0.5h0.5v-2.6H8.2c-0.3,0-0.5-0.2-0.5-0.5v-1c0-0.3,0.2-0.5,0.5-0.5h2.6c0.3,0,0.5,0.2,0.5,0.5v4h0.5
								c0.3,0,0.5,0.2,0.5,0.5V14.7z"/>
							</svg>
						</span>
					</a>
				</div>
			</div>
			<div class="controlMain">
				<div class="controlComponent fades" id="downloadComponent">
					<div class="controlComponentMessage">
						Enter a hash or ENS name.
					</div>
					<form id="downloadForm" spellcheck="false">
						<input type="text" id="downloadHashField" name="downloadHash" placeholder="swarm.eth"/>
						<button type="submit" value="Go" class="orangeButton wideButton">
							Go
							<svg version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"
								 viewBox="0 0 20 20" style="enable-background:new 0 0 16 20;" xml:space="preserve">
								<path id="XMLID_1617_" d="M10,0C4.5,0,0,4.5,0,10s4.5,10,10,10c5.5,0,10-4.5,10-10C20,4.5,15.5,0,10,0z M9.8,17.9
			c-4.4,0-7.9-3.6-7.9-7.9c0-4.4,3.6-7.9,7.9-7.9c0.4,0,0.9,0,1.3,0.1l0,0.8c0,0.3-0.3,0.6-0.6,0.6l-1,0C9.3,3.6,9.1,3.7,9,3.9
			L8.7,4.4C8.6,4.5,8.5,4.6,8.3,4.6L7.7,4.8C7.5,4.8,7.3,5.1,7.3,5.4l0,0.2c0,0.3,0.3,0.6,0.6,0.6l3.6,0c0.2,0,0.3,0.1,0.4,0.2
			l0.3,0.3c0.1,0.1,0.3,0.2,0.4,0.2l0.4,0c0.3,0,0.6,0.3,0.6,0.6c0,0.3-0.2,0.5-0.4,0.6l-1.9,0.6c-0.2,0.1-0.3,0-0.5,0l-0.6-0.3
			c-0.3-0.2-0.6-0.2-1-0.2h0C8.8,8.1,8.3,8.2,8,8.5L6.9,9.3C6.3,9.8,6,10.4,6,11.1l0,0.6c0,0.6,0.2,1.1,0.6,1.5
			c0.4,0.4,1,0.6,1.5,0.6l1,0c0.3,0,0.6,0.3,0.6,0.6l0,1.2c0,0.5,0.1,1,0.3,1.4c0.2,0.4,0.6,0.6,1,0.6c0.4,0,0.7-0.2,0.9-0.5l0.5-0.8
			c0.3-0.4,0.6-0.8,1-1.2c0.1-0.1,0.2-0.2,0.2-0.3L14,14c0-0.1,0.1-0.2,0.1-0.3l0.7-1c0.1-0.1,0.1-0.2,0.1-0.4l0-0.5
			c0-0.3-0.3-0.6-0.6-0.6l-0.3,0c-0.2,0-0.4-0.1-0.5-0.3L13,10.2c-0.2-0.3-0.1-0.8,0.3-0.9l0.1,0c0.2-0.1,0.4,0,0.5,0.1l0.7,0.5
			c0.2,0.1,0.4,0.1,0.6,0l0.6-0.3c0.2-0.1,0.3-0.3,0.3-0.6l0-0.3c0-0.3,0.3-0.6,0.6-0.6l0.7,0c0.2,0.6,0.2,1.3,0.2,1.9
			C17.8,14.4,14.2,17.9,9.8,17.9z"/>
							</svg>
						</button>
					</form>				
				</div>
				<div class="controlComponent fades hidden" id="uploadComponent">
					<div class="controlComponentMessage">
						Select a file to upload it to the Swarm network.
					</div>				
					<form id="uploadForm" enctype="multipart/form-data" spellcheck="false">
						<input type="text" id="uploadSelectedFile" name="uploadSelected" placeholder="select file or directory..."/>
						<input type="file" id="uploadSelectFile" name="file"/>
						<!-- inline js used to deal with browser security restrictions -->
						<button title="Upload a file to Swarm." type="submit" value="Upload File" class="orangeButton wideButton" onclick="if(document.querySelector('#uploadSelectFile').value === '')document.querySelector('#uploadSelectFile').click();" >
							Upload
							<svg version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"
								 viewBox="0 0 16 20" style="enable-background:new 0 0 16 20;" xml:space="preserve">
								<path id="XMLID_1412_" d="M8.8,5.3V0H0.9C0.4,0,0,0.4,0,0.9v18.1C0,19.6,0.4,20,0.9,20h13.1c0.5,0,0.9-0.4,0.9-0.9V6.2
									H9.7C9.2,6.2,8.8,5.8,8.8,5.3z M11.3,13.8H8.8v3.1c0,0.3-0.3,0.6-0.6,0.6H6.9c-0.3,0-0.6-0.3-0.6-0.6v-3.1H3.7
									c-0.6,0-0.8-0.7-0.4-1.1L7,8.9c0.3-0.3,0.7-0.3,0.9,0l3.8,3.7C12.1,13.1,11.9,13.8,11.3,13.8z M14.7,4.1l-3.8-3.8
									C10.7,0.1,10.5,0,10.2,0H10v5h5V4.8C15,4.5,14.9,4.3,14.7,4.1z"/>
							</svg>
						</button>
					</form>
				</div>
				<div class="controlComponent fades hidden" id="uploadFeedbackComponent">
					<div id="uploadTextFeedback" class="controlComponentMessage">
						<div class="">
							<span id="uploadStatusMessage">Uploading</span> <span id="uploadFilename"></span>
						</div>
						<div class="uploadTextFeedbackSub">
							<span id="uploadSwarmhash"><i>Waiting for hash...</i></span>
						</div>
					</div>
					<div class="uploadActions">
						<div class="uploadAction">
							<button id="uploadCancelButton" class="uploadCancelButton" title="Cancel upload">
								<svg version="1.1" id="Layer_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"
									 viewBox="0 0 100 100" style="enable-background:new 0 0 100 100;" xml:space="preserve">
								<path id="XMLID_1460_" d="M50,3.7C24.4,3.7,3.7,24.4,3.7,50S24.4,96.3,50,96.3S96.3,75.6,96.3,50S75.6,3.7,50,3.7z
									 M72.7,62.1c0.9,0.9,0.9,2.3,0,3.2l-7.4,7.4c-0.9,0.9-2.3,0.9-3.2,0L50,60.5L37.9,72.7c-0.9,0.9-2.3,0.9-3.2,0l-7.4-7.4
									c-0.9-0.9-0.9-2.3,0-3.2L39.5,50L27.3,37.9c-0.9-0.9-0.9-2.3,0-3.2l7.4-7.4c0.9-0.9,2.3-0.9,3.2,0L50,39.5l12.1-12.2
									c0.9-0.9,2.3-0.9,3.2,0l7.4,7.4c0.9,0.9,0.9,2.3,0,3.2L60.5,50L72.7,62.1z"/>
								</svg>
							</button>
						</div>
						<div class="uploadAction">
							<button id="uploadButtonLink" class="orangeButton fadeOut" title="Copy link to clipboard.">
								<svg version="1.1" id="Layer_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"
									 viewBox="0 0 20 20" style="enable-background:new 0 0 20 20;" xml:space="preserve">
								<path id="XMLID_2806_" d="M12.4,7.8c2,2,2,5.3,0,7.2c0,0,0,0,0,0l-2.3,2.3c-2,2-5.3,2-7.3,0c-2-2-2-5.3,0-7.3l1.3-1.3
									C4.4,8.5,5,8.7,5,9.2c0,0.6,0.1,1.2,0.3,1.8c0.1,0.2,0,0.4-0.1,0.6L4.8,12c-0.9,0.9-1,2.5,0,3.4c0.9,1,2.5,1,3.5,0l2.3-2.3
									c1-1,0.9-2.5,0-3.4c-0.1-0.1-0.3-0.2-0.3-0.3C10,9.3,9.9,9.2,9.9,9c0-0.4,0.1-0.7,0.4-1L11,7.3c0.2-0.2,0.5-0.2,0.7-0.1
									C11.9,7.4,12.2,7.6,12.4,7.8L12.4,7.8z M17.1,3c-2-2-5.3-2-7.3,0L7.6,5.3c0,0,0,0,0,0c-2,2-2,5.2,0,7.2c0.2,0.2,0.4,0.4,0.7,0.6
									c0.2,0.2,0.5,0.1,0.7-0.1l0.7-0.7c0.3-0.3,0.4-0.6,0.4-1c0-0.2-0.1-0.3-0.2-0.4c-0.1-0.1-0.2-0.2-0.3-0.3c-0.9-0.9-1-2.5,0-3.4
									L11.8,5c1-1,2.5-0.9,3.5,0c0.9,1,0.9,2.5,0,3.4l-0.4,0.4c-0.1,0.1-0.2,0.4-0.1,0.6c0.2,0.6,0.3,1.2,0.3,1.8c0,0.5,0.6,0.7,0.9,0.4
									l1.3-1.3C19.1,8.3,19.1,5,17.1,3L17.1,3z"/>
								</svg>

							</button>
						</div>
						<div class="uploadAction">
							<button id="uploadButtonHash" class="orangeButton fadeOut" title="Copy hash to clipboard.">
								<svg version="1.1" id="Layer_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"
									 viewBox="0 0 20 20" style="enable-background:new 0 0 20 20;" xml:space="preserve">
								<path id="XMLID_2810_" d="M18.9,6.9l0.3-1.7c0.1-0.3-0.2-0.6-0.5-0.6h-3.1l0.6-3.4c0.1-0.3-0.2-0.6-0.5-0.6h-1.7
									c-0.2,0-0.4,0.2-0.5,0.4L13,4.7H8.9l0.6-3.4C9.6,1,9.3,0.8,9,0.8H7.3c-0.2,0-0.4,0.2-0.5,0.4L6.2,4.7H3c-0.2,0-0.4,0.2-0.5,0.4
									L2.2,6.8C2.1,7.1,2.4,7.4,2.7,7.4h3.1l-0.9,5.3H1.5c-0.2,0-0.4,0.2-0.5,0.4l-0.3,1.7c-0.1,0.3,0.2,0.6,0.5,0.6h3.1l-0.6,3.4
									c-0.1,0.3,0.2,0.6,0.5,0.6h1.7c0.2,0,0.4-0.2,0.5-0.4L7,15.3h4.1l-0.6,3.4c-0.1,0.3,0.2,0.6,0.5,0.6h1.7c0.2,0,0.4-0.2,0.5-0.4
									l0.6-3.6H17c0.2,0,0.4-0.2,0.5-0.4l0.3-1.7c0.1-0.3-0.2-0.6-0.5-0.6h-3.1l0.9-5.3h3.3C18.7,7.4,18.9,7.2,18.9,6.9z M11.6,12.6H7.5
									l0.9-5.3h4.1L11.6,12.6z"/>
								</svg>
							</button>
						</div>
					</div>
					<div id="uploadFeedbackBarsWrapper">
						<div id="uploadFeedbackBars">
							<div class="uploadFeedbackBar uploadFeedbackColor1" id="uploadReceivedBar"></div>
							<div class="uploadFeedbackBar uploadFeedbackColorx hidden" id="uploadSeenBar"></div>							
							<div class="uploadFeedbackBar uploadFeedbackColor2 hidden" id="uploadSplitBar"></div>
							<div class="uploadFeedbackBar uploadFeedbackColor3 hidden" id="uploadStoredBar"></div>
							<div class="uploadFeedbackBar uploadFeedbackColor4 hidden" id="uploadSentBar"></div>
							<div class="uploadFeedbackBar uploadFeedbackColor4" id="uploadSyncedBar"></div>
						</div>
						<div class="incrementLine incrementLine1"></div>
						<div class="incrementLine incrementLine2"></div>
						<div class="incrementLine incrementLine3"></div>
					</div>
					<div class="uploadFeedbackCounts">
						<div class="uploadFeedbackCount uploadFeedbackCountColor1">Uploaded <span class="uploadFeedbackCountNumbers" id="uploadReceivedCount"></span></div>
						<div class="uploadFeedbackCount uploadFeedbackCountColor2 hidden">Stored <span class="uploadFeedbackCountNumbers" id="uploadStoredCount"></span></div>
						<div class="uploadFeedbackCount uploadFeedbackCountColor3 hidden">Seen <span class="uploadFeedbackCountNumbers" id="uploadSeenCount"></span></div>						
						<div class="uploadFeedbackCount uploadFeedbackCountColor3 hidden">Split <span class="uploadFeedbackCountNumbers" id="uploadSplitCount"></span></div>
						<div class="uploadFeedbackCount uploadFeedbackCountColor5 hidden">Sent <span class="uploadFeedbackCountNumbers" id="uploadSentCount"></span></div>
						<div class="uploadFeedbackCount uploadFeedbackCountColor5" id="uploadSyncedCount"></div>
					</div>
					<div class="clearBoth"></div>
				</div>
			</div>
		</div>
		<input type="text" id="uploadLinkInput" class="invisible"/>
		<input type="text" id="uploadHashInput" class="invisible"/>
{{ end }}`

const baseTemplate = `
<!DOCTYPE html>
<html>
<head>
	<title>Swarm</title>
	<meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>
	<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0"/>
	<meta http-equiv="X-UA-Compatible" ww="chrome=1"/>
	<link rel="icon" type="image/x-icon" href="favicon.ico"/> 
	<style>
		{{ template "css" . }}
	</style>
	<script>
		{{ template "js" .}}
	</script>
</head>

<body>
<div class="wrapper">
	<div class="main">
		<div class="header">
			<div class="title">
				<a href="/">
					<svg version="1.1" id="Layer_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"
						 viewBox="240 -40 500 100" style="enable-background:new 240 -40 500 100;" xml:space="preserve">
					<style type="text/css">
						.st0{enable-background:new    ;}
						.st1{fill:#333333;}
						.st2{fill:#CDCCCC;}
						.st3{fill:#4F4E4E;}
						.st4{fill:#3C3A3A;}
						.st5{fill:#161616;}
						.st6{fill:#191919;}
						.st7{fill:#2E2E2E;}
						.st8{fill:#343130;}
						.st9{fill:#191918;}
					</style>
					<g id="XMLID_1_">
						<g id="XMLID_1611_">
							<g class="st0">
								<path class="st1" d="M370.1,19.4c0.3,4.8,3.7,5.8,5.4,5.8c2.9,0,5.3-2.3,5.3-5.1c0-3.6-3-4.4-7.2-5.9c-2.5-0.8-8.1-2.7-8.1-8.7
									c0-6,5.1-9.2,10.1-9.2c4.1,0,9.3,2.3,9.7,8.9h-5c-0.3-1.7-1.1-4.5-4.8-4.5c-2.6,0-4.9,1.8-4.9,4.5c0,3.1,2.4,3.9,7.6,5.8
									c4,1.7,7.7,3.7,7.7,8.9s-3.5,9.7-10.4,9.7c-6.4,0-10.5-3.9-10.6-10.1H370.1z"/>
								<path class="st1" d="M387.7,5.3h4.7l5.6,17.8l5-16.5h3.8l5.1,16.5l5.5-17.8h4.7l-7.9,23.6h-4.3l-5.1-16.5l-5,16.5h-4.3L387.7,5.3
									z"/>
								<path class="st1" d="M448.5,28.9H444v-4.1c-1.7,3.2-5,4.8-8.8,4.8c-7.6,0-12.1-5.9-12.1-12.5c0-7.2,5.3-12.5,12.1-12.5
									c4.7,0,7.7,2.6,8.8,4.9V5.3h4.5V28.9z M427.6,17.2c0,3.5,2.5,8.2,8.3,8.2c3.6,0,6.2-2,7.3-4.7c0.5-1.1,0.8-2.3,0.8-3.5
									c0-1.2-0.2-2.4-0.7-3.5c-1.1-2.7-3.8-4.9-7.6-4.9C430.8,8.8,427.6,12.7,427.6,17.2L427.6,17.2z"/>
								<path class="st1" d="M454.5,5.3h4.3v3.5c1.3-2.8,3.8-4.1,7-4.2v4.5h-0.3c-4.3,0-6.5,2.3-6.5,7v12.8h-4.5V5.3z"/>
								<path class="st1" d="M469.3,5.3h4.3v3.3c1-2,3.7-4,7.2-4c2.2,0,4.4,0.7,6.1,2.5c0.6,0.6,1.3,1.9,1.5,2.3c0.4-0.7,0.9-1.5,1.7-2.3
									c1.5-1.5,3.7-2.5,6.3-2.5c2.2,0,4.6,0.6,6.3,2.4c2.1,2.1,2.6,4.4,2.6,8.9v13h-4.5V16.1c0-2-0.3-3.9-1.3-5.3c-0.8-1.2-2.1-2-4.2-2
									c-2,0-3.7,0.8-4.6,2.3c-1,1.5-1.1,2.8-1.1,4.8v13H485v-13c0-2-0.2-3.4-1-4.7c-0.9-1.4-2.3-2.3-4.5-2.3c-2.1,0-3.7,1-4.5,2.3
									c-0.9,1.3-1.3,2.9-1.3,4.8v12.9h-4.5V5.3z"/>
								<path class="st1" d="M549.1,5.9c-0.8-1.2-1.8-2.2-3-3c-1.9-1.3-4.1-2.2-6.8-2.2c-6.1,0-11.8,4.6-11.8,12.1
									c0,7.7,5.7,12.2,11.8,12.2c3,0,5.5-0.9,7.4-2.4c2-1.5,3.3-3.5,3.8-5.7h-13.9v-4.2h19.8c0,2-0.3,4.8-1.4,7.2
									c-3,6.4-9.6,9.6-15.7,9.6c-9.5,0-16.9-7.3-16.9-16.8c0-9.6,7.6-16.6,16.9-16.6c6.9,0,13.1,4.2,15.4,9.7H549.1z"/>
								<path class="st1" d="M584.9,28.9h-4.5v-4.1c-1.7,3.2-5,4.8-8.8,4.8c-7.6,0-12.1-5.9-12.1-12.5c0-7.2,5.3-12.5,12.1-12.5
									c4.7,0,7.7,2.6,8.8,4.9V5.3h4.5V28.9z M564.1,17.2c0,3.5,2.5,8.2,8.3,8.2c3.6,0,6.2-2,7.3-4.7c0.5-1.1,0.8-2.3,0.8-3.5
									c0-1.2-0.2-2.4-0.7-3.5c-1.1-2.7-3.8-4.9-7.6-4.9C567.3,8.8,564.1,12.7,564.1,17.2L564.1,17.2z"/>
								<path class="st1" d="M592.6,8.8h-4.3V5.3h4.3v-8.5h4.5v8.5h4.6v3.5h-4.6v20.1h-4.5V8.8z"/>
								<path class="st1" d="M606.8,18.6c0.2,4.8,4.4,7.2,7.9,7.2c3.2,0,5.5-1.5,6.8-4h4.7c-1.1,2.6-2.8,4.6-4.8,5.9c-2,1.3-4.3,2-6.7,2
									c-7.7,0-12.5-6.2-12.5-12.5c0-6.8,5.3-12.5,12.5-12.5c3.4,0,6.5,1.3,8.7,3.5c2.8,2.8,4,6.5,3.6,10.5H606.8z M622.5,15
									c-0.2-3.1-3.3-6.6-7.8-6.6c-4.6,0-7.6,3.3-7.8,6.6H622.5z"/>
								<path class="st1" d="M628.1,5.3h4.7l5.6,17.8l5-16.5h3.8l5.1,16.5l5.5-17.8h4.7l-7.9,23.6h-4.3l-5.1-16.5l-5,16.5H636L628.1,5.3z
									"/>
								<path class="st1" d="M688.9,28.9h-4.5v-4.1c-1.7,3.2-5,4.8-8.8,4.8c-7.6,0-12.1-5.9-12.1-12.5c0-7.2,5.3-12.5,12.1-12.5
									c4.7,0,7.7,2.6,8.8,4.9V5.3h4.5V28.9z M668,17.2c0,3.5,2.5,8.2,8.3,8.2c3.6,0,6.2-2,7.3-4.7c0.5-1.1,0.8-2.3,0.8-3.5
									c0-1.2-0.2-2.4-0.7-3.5c-1.1-2.7-3.8-4.9-7.6-4.9C671.2,8.8,668,12.7,668,17.2L668,17.2z"/>
								<path class="st1" d="M700.7,28.3l-9-22.9h4.9l6.5,17.3l6.6-17.3h4.9l-12.9,32.2h-4.8L700.7,28.3z"/>
							</g>
						</g>
						<g class="st0">
							<path class="st2" d="M692.9-12.1h0.4l2.5,8.7l2.4-8.7h0.3l2.4,8.7l2.5-8.7h0.4l-2.6,9.2h-0.5l-2.3-8.2l-2.3,8.2h-0.5L692.9-12.1z"
								/>
							<path class="st2" d="M703.9-6.1c0,2.2,1.8,3.1,3.1,3.1c1.2,0,2.3-0.6,2.8-1.7h0.4c-0.3,0.7-0.8,1.2-1.3,1.6
								c-0.6,0.3-1.2,0.5-1.9,0.5c-2.5,0-3.5-2-3.5-3.4c0-2,1.5-3.5,3.5-3.5c0.9,0,1.7,0.3,2.4,0.9c0.7,0.6,1.1,1.6,1.1,2.6H703.9z
								 M710.1-6.5c-0.1-1.6-1.6-2.9-3.1-2.9c-1.6,0-3,1.2-3.1,2.9H710.1z"/>
							<path class="st2" d="M711.9-12.1h0.3v2.6v1.7c0.5-1,1.5-1.9,3.1-1.9c2.1,0,3.5,1.6,3.5,3.5c0,1.8-1.3,3.5-3.5,3.5
								c-1.4,0-2.5-0.7-3.1-1.9v1.7h-0.3V-12.1z M718.4-6.2c0-1.8-1.4-3.1-3.1-3.1c-1.8,0-3.1,1.3-3.1,3.1c0,1.7,1.3,3.1,3.1,3.1
								C717.4-3,718.4-4.7,718.4-6.2L718.4-6.2z"/>
							<path class="st2" d="M720.5-10.2c0.1-1.4,1.2-2.1,2.2-2.1c1.1,0,2.2,0.7,2.2,2.1c0,0.8-0.4,1.5-1.3,2c1.6,0.5,1.9,1.9,1.9,2.6
								c0,1.7-1.2,2.9-2.8,2.9c-0.8,0-1.7-0.4-2.2-1c-0.4-0.5-0.6-1.1-0.6-1.8h0.4c0.1,1.7,1.4,2.5,2.5,2.5c1.5,0,2.4-1.1,2.4-2.5
								c0-1-0.7-2.5-2.6-2.5h-0.5v-0.3h0.5c1.9,0,2-1.4,2-1.8c0-1-0.7-1.8-1.7-1.8c-1.1,0-1.8,0.8-1.8,1.8H720.5z"/>
						</g>
						<g id="XMLID_2_">
							<polygon id="XMLID_1692_" class="st3" points="272.1,56.1 251.3,44.1 272.1,32.1 		"/>
							<polygon id="XMLID_1691_" class="st4" points="271.5,56.1 292.3,44.1 271.5,32.1 		"/>
							<polygon id="XMLID_1690_" class="st5" points="272.1,32.9 251.3,20.9 272.1,8.9 		"/>
							<polygon id="XMLID_1684_" class="st6" points="271.5,32.9 292.3,20.9 271.5,8.9 		"/>
							<polygon id="XMLID_1681_" class="st7" points="292.3,44.5 271.5,32.5 292.3,20.5 		"/>
							<polygon id="XMLID_1680_" class="st7" points="251.3,44.5 272.1,32.5 251.3,20.5 		"/>
							<polygon id="XMLID_1679_" class="st3" points="317.5,56.1 296.7,44.1 317.5,32.1 		"/>
							<polygon id="XMLID_1678_" class="st4" points="316.9,56.1 337.6,44.1 316.9,32.1 		"/>
							<polygon id="XMLID_1677_" class="st5" points="317.5,32.9 296.7,20.9 317.5,8.9 		"/>
							<polygon id="XMLID_1676_" class="st6" points="316.9,32.9 337.6,20.9 316.9,8.9 		"/>
							<polygon id="XMLID_1675_" class="st7" points="337.6,44.5 316.9,32.5 337.6,20.5 		"/>
							<polygon id="XMLID_1674_" class="st7" points="296.7,44.5 317.5,32.5 296.7,20.5 		"/>
							<polygon id="XMLID_1673_" class="st3" points="294.8,16.4 274,4.4 294.8,-7.6 		"/>
							<polygon id="XMLID_1672_" class="st4" points="294.2,16.4 315,4.4 294.2,-7.6 		"/>
							<polygon id="XMLID_1671_" class="st5" points="294.8,-6.8 274,-18.8 294.8,-30.8 		"/>
							<polygon id="XMLID_1670_" class="st7" points="274,4.8 294.8,-7.2 274,-19.2 		"/>
							<polygon id="XMLID_1669_" class="st8" points="304.7,-24.8 304.7,-12.8 294.3,-18.8 		"/>
							<polygon id="XMLID_1668_" class="st8" points="318.6,-32.7 318.6,-20.8 308.3,-26.7 		"/>
							<polygon id="XMLID_1667_" class="st8" points="328.7,-26.9 328.7,-14.9 318.3,-20.9 		"/>
							<polygon id="XMLID_1666_" class="st5" points="304.7,-13.2 304.7,-1.2 294.3,-7.2 		"/>
							<polygon id="XMLID_1665_" class="st7" points="314.8,-7.4 314.8,4.6 304.4,-1.4 		"/>
							<polygon id="XMLID_1663_" class="st5" points="318.3,-20.8 318.3,-32.7 328.7,-26.7 		"/>
							<polygon id="XMLID_1662_" class="st3" points="318.3,-9.1 318.3,-21.1 328.7,-15.1 		"/>
							<polygon id="XMLID_1661_" class="st9" points="304.4,-1.2 304.4,-13.2 314.8,-7.2 		"/>
							<polygon id="XMLID_1659_" class="st8" points="294.3,-7 294.3,-19 304.7,-13 		"/>
							<polygon id="XMLID_1658_" class="st6" points="294.3,-18.7 294.3,-30.7 304.7,-24.7 		"/>
						</g>
					</g>
					</svg>
				</a>					
			</div>
			<div class="subTitle">
				Censorship resistant storage and communication infrastructure for a sovereign digital society.
			</div>
		</div>
		{{ template "content" . }}
		<div class="footer">
			<div class="footerItems">
				<div class="footerItem">
					<a href="/bzz:/swarm.eth" target="_blank" rel="noopener noreferrer">
						<div class="footerItemLogo">
							<svg version="1.1" id="Layer_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"
								 viewBox="0 0 100 100" style="enable-background:new 0 0 100 100;" xml:space="preserve">
							<style type="text/css">
								.stt0{fill:#4F4E4E;}
								.stt1{fill:#3C3A3A;}
								.stt2{fill:#161616;}
								.stt3{fill:#191919;}
								.stt4{fill:#2E2E2E;}
								.stt5{fill:#343130;}
								.stt6{fill:#191918;}
							</style>
							<g id="XMLID_1693_">
								<polygon id="XMLID_1692_" class="stt0" points="24.8,98.9 2.3,86 24.8,73 	"/>
								<polygon id="XMLID_1691_" class="stt1" points="24.8,98.9 47.2,86 24.8,73 	"/>
								<polygon id="XMLID_1690_" class="stt2" points="24.8,73 2.3,60.1 23.6,47.8 24.8,47.1 	"/>
								<polygon id="XMLID_1684_" class="stt3" points="24.8,73 47.2,60.1 24.8,47.1 	"/>
								<polygon id="XMLID_1681_" class="stt4" points="47.2,86 24.8,73 47.2,60.1 	"/>
								<polygon id="XMLID_1680_" class="stt4" points="2.3,86 24.8,73 2.3,60.1 	"/>
								<polygon id="XMLID_1679_" class="stt0" points="75.2,98.9 52.8,86 75.2,73 	"/>
								<polygon id="XMLID_1678_" class="stt1" points="75.2,98.9 97.7,86 75.2,73 	"/>
								<polygon id="XMLID_1677_" class="stt2" points="75.2,73 52.8,60.1 75.2,47.1 	"/>
								<polygon id="XMLID_1676_" class="stt3" points="75.2,73 97.7,60.1 75.2,47.1 	"/>
								<polygon id="XMLID_1675_" class="stt4" points="97.7,86 75.2,73 97.7,60.1 	"/>
								<polygon id="XMLID_1674_" class="stt4" points="52.8,86 75.2,73 52.8,60.1 	"/>
								<polygon id="XMLID_1673_" class="stt0" points="50,54.8 27.6,41.8 50,28.9 	"/>
								<polygon id="XMLID_1672_" class="stt1" points="50,54.8 72.4,41.8 50,28.9 	"/>
								<polygon id="XMLID_1671_" class="stt2" points="50,28.9 27.6,15.9 50,3 	"/>
								<polygon id="XMLID_1670_" class="stt4" points="27.6,41.8 50,28.9 27.6,15.9 	"/>
								<polygon id="XMLID_1669_" class="stt5" points="61.2,9.4 61.2,22.4 50,15.9 	"/>
								<polygon id="XMLID_1668_" class="stt5" points="76.7,0.6 76.7,13.6 65.5,7.1 	"/>
								<polygon id="XMLID_1667_" class="stt5" points="87.9,7.1 87.9,20.1 76.7,13.6 	"/>
								<polygon id="XMLID_1666_" class="stt2" points="61.2,22.4 61.2,35.3 50,28.9 	"/>
								<polygon id="XMLID_1665_" class="stt4" points="72.4,28.9 72.4,41.8 61.2,35.3 	"/>
								<polygon id="XMLID_1663_" class="stt2" points="76.7,13.6 76.7,0.6 87.9,7.1 	"/>
								<polygon id="XMLID_1662_" class="stt0" points="76.7,26.5 76.7,13.6 87.9,20.1 	"/>
								<polygon id="XMLID_1661_" class="stt6" points="61.2,35.3 61.2,22.4 72.4,28.9 	"/>
								<polygon id="XMLID_1659_" class="stt5" points="50,28.9 50,15.9 61.2,22.4 	"/>
								<polygon id="XMLID_1658_" class="stt3" points="50,15.9 50,3 61.2,9.4 	"/>
							</g>
							</svg>

						</div>
						<div class="footerItemLink">
							swarm.eth
						</div>
					</a>
				</div>
				<div class="footerItem">
					<a href="https://github.com/ethersphere" target="_blank" rel="noopener noreferrer">
						<div class="footerItemLogo">
							<svg version="1.1" id="Layer_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"
								 viewBox="0 0 100 100" style="enable-background:new 0 0 100 100;" xml:space="preserve">
							<style type="text/css">
								.sttt0{fill-rule:evenodd;clip-rule:evenodd;fill:#191717;}
								.sttt1{fill:#191717;}
							</style>
							<g id="XMLID_1580_">
								<path id="XMLID_1602_" class="sttt0" d="M50,5.3C24.7,5.3,4.2,25.8,4.2,51.1c0,20.2,13.1,37.4,31.3,43.5c2.3,0.4,3.1-1,3.1-2.2
									c0-1.1,0-4.7-0.1-8.5c-12.7,2.8-15.4-5.4-15.4-5.4c-2.1-5.3-5.1-6.7-5.1-6.7c-4.2-2.8,0.3-2.8,0.3-2.8c4.6,0.3,7,4.7,7,4.7
									c4.1,7,10.7,5,13.3,3.8c0.4-3,1.6-5,2.9-6.1c-10.2-1.2-20.9-5.1-20.9-22.6c0-5,1.8-9.1,4.7-12.3c-0.5-1.2-2-5.8,0.4-12.1
									c0,0,3.8-1.2,12.6,4.7c3.7-1,7.6-1.5,11.5-1.5c3.9,0,7.8,0.5,11.5,1.5c8.7-5.9,12.6-4.7,12.6-4.7c2.5,6.3,0.9,11,0.4,12.1
									c2.9,3.2,4.7,7.3,4.7,12.3c0,17.6-10.7,21.5-20.9,22.6c1.6,1.4,3.1,4.2,3.1,8.5c0,6.1-0.1,11.1-0.1,12.6c0,1.2,0.8,2.6,3.1,2.2
									c18.2-6.1,31.3-23.2,31.3-43.5C95.8,25.8,75.3,5.3,50,5.3z"/>
								<path id="XMLID_1599_" class="sttt1" d="M21.5,71.1c-0.1,0.2-0.5,0.3-0.8,0.1c-0.3-0.1-0.5-0.5-0.4-0.7c0.1-0.2,0.5-0.3,0.8-0.1
									C21.5,70.6,21.6,70.9,21.5,71.1L21.5,71.1z M21,70.7"/>
								<path id="XMLID_1596_" class="sttt1" d="M23.4,73.2c-0.2,0.2-0.6,0.1-0.9-0.2c-0.3-0.3-0.4-0.7-0.1-1c0.2-0.2,0.6-0.1,0.9,0.2
									C23.6,72.5,23.6,73,23.4,73.2L23.4,73.2z M22.9,72.7"/>
								<path id="XMLID_1593_" class="sttt1" d="M25.2,75.8c-0.3,0.2-0.7,0-1-0.4c-0.3-0.4-0.3-0.9,0-1.1c0.3-0.2,0.7,0,1,0.4
									C25.5,75.1,25.5,75.6,25.2,75.8L25.2,75.8z M25.2,75.8"/>
								<path id="XMLID_1590_" class="sttt1" d="M27.7,78.4c-0.3,0.3-0.8,0.2-1.2-0.2c-0.4-0.4-0.5-0.9-0.3-1.2c0.3-0.3,0.8-0.2,1.2,0.2
									C27.8,77.6,27.9,78.1,27.7,78.4L27.7,78.4z M27.7,78.4"/>
								<path id="XMLID_1587_" class="sttt1" d="M31.1,79.8c-0.1,0.4-0.6,0.5-1.1,0.4c-0.5-0.2-0.9-0.6-0.8-0.9c0.1-0.4,0.6-0.5,1.1-0.4
									C30.9,79.1,31.2,79.5,31.1,79.8L31.1,79.8z M31.1,79.8"/>
								<path id="XMLID_1584_" class="sttt1" d="M34.8,80.1c0,0.4-0.4,0.7-1,0.7c-0.5,0-1-0.3-1-0.7c0-0.4,0.4-0.7,1-0.7
									C34.4,79.4,34.8,79.7,34.8,80.1L34.8,80.1z M34.8,80.1"/>
								<path id="XMLID_1581_" class="sttt1" d="M38.3,79.5c0.1,0.4-0.3,0.7-0.9,0.8c-0.5,0.1-1-0.1-1.1-0.5c-0.1-0.4,0.3-0.8,0.9-0.9
									C37.8,78.9,38.3,79.1,38.3,79.5L38.3,79.5z M38.3,79.5"/>
							</g>
							</svg>
						</div>
						<div class="footerItemLink">
							github.com/ethersphere
						</div>
					</a>
				</div>
				<div class="footerItem">
					<a href="https://ethereum.org" target="_blank" rel="noopener noreferrer">
						<div class="footerItemLogo">
						<svg version="1.1" id="Layer_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"
							 viewBox="0 0 100 100" style="enable-background:new 0 0 100 100;" xml:space="preserve">
						<style type="text/css">
							.sttttt0{fill:#343434;}
							.sttttt1{fill:#8C8C8C;}
							.sttttt2{fill:#3D3D3C;}
							.sttttt3{fill:#161616;}
							.sttttt4{fill:#3A3A39;}
						</style>
						<g id="XMLID_1556_">
							<g id="XMLID_1557_">
								<polygon id="XMLID_1565_" class="sttttt0" points="50,5.7 49.4,7.7 49.4,66.3 50,66.9 77.2,50.8 		"/>
								<polygon id="XMLID_1563_" class="sttttt1" points="50,5.7 22.8,50.8 50,66.9 50,38.5 		"/>
								<polygon id="XMLID_1561_" class="sttttt2" points="50,72.1 49.7,72.5 49.7,93.4 50,94.3 77.2,56 		"/>
								<polygon id="XMLID_1560_" class="sttttt1" points="50,94.3 50,72.1 22.8,56 		"/>
								<polygon id="XMLID_1559_" class="sttttt3" points="50,66.9 77.2,50.8 50,38.5 		"/>
								<polygon id="XMLID_1558_" class="sttttt4" points="22.8,50.8 50,66.9 50,38.5 		"/>
							</g>
						</g>
						</svg>
						</div>
						<div class="footerItemLink">
							ethereum.org
						</div>
					</a>
				</div>
			</div>
			<div class="footerLicense">
				GNU Lesser General Public License v3.0
			</div>
		</div>
	</div>
</div>
</body>
</html>
`

const css = `{{ define "css" }}
.hidden{
	display: none;
}

.invisible{
	position: absolute;
	opacity: 0;
}

.fades{
	opacity: 1;	
	transition: opacity 0.5s;
}

.fadeOut{
	pointer-events: none;
	opacity: 0;
	transition: opacity 0.5s;
}


* {
	-webkit-font-smoothing: antialiased;
	-moz-osx-font-smoothing: grayscale;
	-webkit-text-size-adjust: 100%;
}

input:-internal-autofill-selected {
	background-color: none;
	background-image: none;
	color: rgb(0, 0, 0) !important;
}

:focus {
	outline: none;
}

.clearBoth{
	visibility: hidden;
	display: block;
	content: "";
	clear: both;
	height: 0;
}

body, input{
	font-family: "Helvetica Neue", HelveticaNeue, Helvetica, Arial, sans-serif;
	color: #343434;
}

a{
	text-decoration: none;
}

@font-face {
	font-family: 'urw_gothic_lbook';
	src: url(data:application/font-woff2;charset=utf-8;base64,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) format('woff2'),
		 url(data:application/font-woff;charset=utf-8;base64,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) format('woff');
	font-weight: normal;
	font-style: normal;
}

@font-face {
	font-family: 'urw_gothic_ldemi';
	src: url(data:application/font-woff2;charset=utf-8;base64,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) format('woff2'),
		 url(data:application/font-woff;charset=utf-8;base64,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) format('woff');
	font-weight: normal;
	font-style: normal;

}

.wrapper{
	min-width: 750px;
	min-height: 750px;
	position: relative;
}

.title{
	text-align: center;
	margin-top: 50px;
}

.title svg{
	width:660px;
}

.subTitle{
	width: 709px;
	margin: auto;
	margin-top: 36px;
	font-size: 17px;
}

.control{
	margin-top: 62px;
}

.controlHeader {
	width:606px;
	margin:auto;
	padding-left: 9px;
}

.controlHeaderItem a{
	cursor: pointer;
	font-family: urw_gothic_ldemi;
	font-size: 20px;

	color: #FFA500;

	opacity: 0.3;
	float: left;
	margin-right: 30px;
}

.controlHeaderItemIcon svg{
	height:20px;
	fill: #FFA500;
}

#controlHeaderInfo .controlHeaderItemIcon,
#controlHeaderDownload .controlHeaderItemIcon{
	top: 2px;
	position: relative;
}


.controlHeaderItem.active a{
	opacity: 1;
}


.controlComponent {
	border-radius: 6px;
	width: 606px;
	padding: 9px;
	border: 1px solid #F5F5F5;
	margin-bottom: 6px;
}

.controlMain {
	clear: both;
	padding: 3px;
	width: 606px;
	margin: auto;
}

.controlComponentMessage{
	margin-bottom: 6px;
}

.controlMain input[type=text] {
	font-size: 13px;
	width: 473px;
	margin-right: 8px;
	height: 29px;
	border: 1px solid #E0E0E0;
	padding-left: 11px;
}

.controlMain input[type=text]::placeholder {
	font-size: 13px;
	color: #E0E0E0;
	width: 430px; 
	font-style: italic;
}

.controlMain button {
	transition: opacity 1s;
	cursor: pointer;
	position: relative;
	font-size: 20px;
	width: 104px;
	border-radius: 6px;
	text-align: left;
	height: 32px;
	padding-right: 3px;
	top: 2px;
}

.controlMain button svg {
	height: 20px;
	float: right;
	fill: white;
	top: 0px;
	right: 3px;
	position: relative;
}

.controlMain button::after {
	content: '';
	display: block;
	position: absolute;
	top: -6px;
	width: 112px;
	left: -6px;
	border-radius: 9px;
	height: 40px;
	opacity:0;
	transition: opacity 0.3s;
}


.controlMain button.orangeButton {
	background: #FFA500;
	color: #ffffff;
	border: 1px solid #FFA500;

}

.controlMain button.orangeButton.fadeOut {
	opacity: 0.2;
	transition: opacity 1s;
}


.controlMain button.orangeButton svg {
	fill: #ffffff;
}

.controlMain button.orangeButton:hover:after {
	border: 1px solid #FFA500;
	opacity: 0.5;
	transition: opacity 1s;
}

.controlMain .uploadAction button{
	width: 34px;
}

.controlMain .uploadAction button::after{
	width: 43px;
}


.controlMain button.uploadCancelButton {
	background: #ffffff;
	color: #8c8c8c;
	border-radius: 6px;
	border: 1px solid #cdcccc;
}

.controlMain button.uploadCancelButton svg {
	fill: #8c8c8c;
}

.controlMain button.uploadCancelButton:hover:after {
	border: 1px solid #8c8c8c;
	opacity: 0.5;
	transition: opacity 1s;
}

#uploadForm{
	position: relative;
}

#uploadForm input[type="text"]{
	pointer-events: none;
}

#uploadForm input[type="file"]{
	position: absolute;
	left: 0;
	height: 38px;
	width: 484px;
	opacity: 0;
}


.uploadActions{
	float: right;
	padding-right: 4px;
}

.uploadAction{
	margin-left: 9px;
	float: right;
}

#uploadTextFeedback{
	margin-left: 2px;
	float:left;
}

.uploadTextFeedbackSub{
	margin-top: 4px;
}

#uploadFilename{
	font-weight: bold;
}

#uploadSwarmhash {
	font-weight: bold;
	font-size: 12px;
	cursor: pointer;
}

#uploadSwarmhash i{
	font-size: normal;
	font-style: italic;
}

.uploadFeedbackCounts{
	float: left;
	margin-left: 5px;
	width: 125px;
}

.uploadFeedbackCount{
	float: left;
	width: 120px;
	font-size: 11px;
	height: 11px;
	margin-bottom: 2px;
	font-weight: bold;
	line-height: 12px;	
}

.uploadFeedbackCountNumbers{
	float: right;
	font-family: "Courier New", Courier, monospace;
}

#uploadFeedbackBarsWrapper{
	position: relative;
	clear: left;
	width: 472px;
}

#uploadFeedbackBarsWrapper .incrementLine{
	position: absolute;
	content: "";
	width: 0px;
	height: 30px;
	top: -2px;
	border-width: 0px 1px 0px 0px ;
	border-color: lightgray;
	border-style: solid;
	padding: 1px;
	z-index: 0;
}

#uploadFeedbackBarsWrapper .incrementLine1{
	left: 25%;
}

#uploadFeedbackBarsWrapper .incrementLine2{
	left: 50%;
}

#uploadFeedbackBarsWrapper .incrementLine3{
	left: 75%;
}


#uploadFeedbackBars{
	position: relative;
	float: left;
	width: 100%;
	border-width: 0px 1px;
	border-color: lightgray;
	border-style: solid;
	padding: 1px;
	z-index: 2;
	background: #ECF6FF;
}


#uploadFeedbackBars .uploadFeedbackBar{
	height: 11px;
	margin-bottom: 2px;
	z-index: 1;
	transition: width 0.1s;
	width: 0;
}

.uploadFeedbackColor1 {background-color: #7979F2;}
.uploadFeedbackColor2 {background-color: #5E45FC;}
.uploadFeedbackColor3 {background-color: #3A1BFF;}
.uploadFeedbackColor4 {background-color: #1A00BF;}
.uploadFeedbackColor5 {background-color: #150194;}

.uploadFeedbackCountColor1 {color: #7979F2;}
.uploadFeedbackCountColor2 {color: #5E45FC;}
.uploadFeedbackCountColor3 {color: #3A1BFF;}
.uploadFeedbackCountColor4 {color: #1A00BF;}
.uploadFeedbackCountColor5 {color: #150194;}

.uploadFeedbackShare {
	clear:both;
}

.uploadFeedbackShareLink {

}

.uploadFeedbackShareButton {

}

.footer{
	position: absolute;
	width: 100%;
	bottom: 44px;
}

.footer a{
	color: #343434;
}

.footerItems{
	text-align: center;
	margin-bottom: 44px;
}

.footerItem{
	width: 200px;
	display: inline-block;
}

.footerItem a{
	font-size: 13px;	
	text-decoration: none;
}

.footerItemLogo svg{
	height: 70px;
	margin-bottom: 8px;
}

.footerLicense{
	text-align: center;
	clear: both;
}

.errorHeader, .errorMessage, .errorCode{
	text-align: center;
	width: 660px;
	margin: auto;
}

.errorHeader h1{
	font-family: urw_gothic_ldemi;
	font-size: 20px;
	color: #FFA500;
}
{{ end }}`

const js = `{{ define "js" }}
let gatewayHost = window.location.protocol+"//"+window.location.hostname+(window.location.port ? ":"+window.location.port : "");

class SwarmProgressBar {
	constructor(gateway){
		this.gateway = gateway;
		this.uploadProgressPercent = 0;
		this.tagId = false;
		this.pollEvery = 1 * 1000;
		this.checkInterval = false;
		this.onProgressCallback = false;
		this.onErrorCallback = false;
		this.onStartCallback = false;

		this.status = {
			Total: false,
			Received: false,
			Seen: false,
			Sent: false,
			Split: false,
			Stored: false,
			Synced: false,
			Complete: true,
			swarmHash: false,
			gatewayLink: false
		};

		this.isComplete = false;
	}

	setStatus(newStatus){
		for(var key in newStatus) { 
			this.status[key] = newStatus[key];
		}
	}

	upload(formData) {
		this.startCheckProgress();

		let url = this.gateway + '/bzz:/';

		let uploadURL = url + '?defaultpath=' + formData.get('file').name;

		return this.sendUploadRequest(uploadURL, 'POST', 'text', formData, formData.get('file').size).then((response) => {
			let swarmHash = response.responseText;
			this.setStatus({swarmHash: swarmHash});  
			this.setStatus({gatewayLink: url + swarmHash + "/" + formData.get('file').name});
			this.tagId = response.getResponseHeader('x-swarm-tag');
			this.onUploadedCallback(response);
		}).catch((error) => {
			throw new Error(error);
		});
	}

	startCheckProgress(){
		this.checkProgressInterval = setInterval(()=>{
			this.checkProgress();
		}, this.pollEvery);
		this.checkProgress();
	}

	checkProgress(){
		let responseData;
		if(this.tagId !== false){
			let url = this.gateway + '/bzz-tag:/?Id=' + this.tagId;
			return this.sendRequest(url, 'GET', 'json').then((response) => {
				if(response.responseText){
					responseData = JSON.parse(response.responseText);
					this.setStatus({
						Total: responseData.Total,
						Seen: responseData.Seen,
						Sent: responseData.Sent,
						Split: responseData.Split,
						Stored: responseData.Stored,
						Synced: responseData.Synced
					});
				}
				if(this.onProgressCallback){
					this.onProgressCallback(this.status);
				}
				if(responseData.Total === (responseData.Synced - responseData.Seen)){
					this.isCompleted = true;
					clearInterval(this.checkProgressInterval);
				}
			}).catch((error) => {
				this.isErrored = true;
				clearInterval(this.checkProgressInterval);
				throw new Error(error);
			});
		}else{
			if(this.onProgressCallback){
				this.isErrored = true;
				this.onProgressCallback(this.status);
			}
		}
	}

	sendUploadRequest(url, requestType, responseType = 'text', data, dataLength) {
		return new Promise((resolve,reject) => {
			let xhr = new XMLHttpRequest();

			xhr.onloadstart = () => {
				if(this.onStartCallback){
					this.onStartCallback(event);
				}
			};

			xhr.onreadystatechange = function(){   
				if(xhr.readyState === 4 && xhr.status === 200){
					resolve(xhr);  
				}
			}

			xhr.upload.onprogress = (event) => {
				let received;
				if(event.loaded > dataLength){
					received = 100;
				}else{
					received = Math.floor((event.loaded/dataLength) * 100, 2);
				}
				this.setStatus({Received: received});
			};

			xhr.onerror = (error) => {
				reject(error);
			};

			xhr.open(requestType, url, true);

			xhr.setRequestHeader('Accept', responseType);

			xhr.send(data);
		});

	}

	sendRequest(url, requestType, responseType = 'text', data, dataLength) {
		return new Promise((resolve,reject) => {
			let xhr = new XMLHttpRequest();

			xhr.onreadystatechange = function(){ 
				if(xhr.readyState === 4 && xhr.status === 200){
					resolve(xhr);  
				}
			}

			xhr.onerror = (error) => {
				reject(error);
			};

			xhr.open(requestType, url, true);

			xhr.setRequestHeader('Accept', responseType);

			xhr.send(data);
		});

	}

	onProgress(fn){
		this.onProgressCallback = fn;
	}

	onStart(fn){
		this.onStartCallback = fn;
	}

	onError(fn){
		this.onErrorCallback = fn;
	}   

	onUploaded(fn){
		this.onUploadedCallback = fn;
	}

	cancel(){
		clearInterval(this.checkProgressInterval);
	}

}

let humanFileSize = (size) => {
	var i = Math.floor( Math.log(size)/Math.log(1024) );
	return ( size/Math.pow(1024, i) ).toFixed(0) * 1 + ' ' + ['bytes', 'kb', 'mb', 'gb', 'tb'][i];
};


let fadeAndReplace = (selector, content, time=600) => {
	let element = document.querySelector(selector);
	element.classList.add("fades");
	element.classList.add("fadeOut");
	setTimeout(()=>{
		element.innerHTML = content;
		element.classList.remove("fadeOut");
	}, time);
};

let padNumber = (n, width, z) => {
  z = z || '0';
  n = n + '';
  return n.length >= width ? n : new Array(width - n.length + 1).join(z) + n;
}


let truncateEnd = function (string, length, separator = '...') {
	if (string.length <= length) return string;
	return string.substr(0, length) + separator;
};

let components = [
	['#controlHeaderDownload','#downloadComponent'],
	['#controlHeaderUpload','#uploadComponent'],
	['#controlHeaderInfo','#uploadFeedbackComponent']
];

let fadeInComponent = (headerSelectorIn, selectorIn, time=600) => {
	let elementIn = document.querySelector(selectorIn);
	let headerIn = document.querySelector(headerSelectorIn);

	if(headerSelectorIn){
		headerIn.classList.add("active");
	}

	for (var i = components.length - 1; i >= 0; i--) {
		if(components[i][1] !== selectorIn){
			if(headerSelectorIn){
				document.querySelector(components[i][0]).classList.remove("active");
			}
			document.querySelector(components[i][1]).classList.add("fadeOut");
		}
	}

	setTimeout(()=>{
		for (var i = components.length - 1; i >= 0; i--) {
			if(components[i][1] !== selectorIn){  
				document.querySelector(components[i][1]).classList.add("hidden");
			}
		}


		elementIn.classList.add("fadeOut"); 
		elementIn.classList.remove("hidden");
		setTimeout(()=>{   
			elementIn.classList.remove("fadeOut");
		},200);
	}, time);
};


let goToPage = () => {
  var page = document.getElementById('downloadHashField').value;
  if (page == "") {
	return false;
  }
  var address = "/bzz:/" + page;
  location.href = address;
}

let copyHashAction = (e) => {
	e.preventDefault();

	let copyText = document.querySelector('#uploadHashInput');
	copyText.select();
	copyText.setSelectionRange(0, 99999); /*For mobile devices*/
	document.execCommand("copy");
	alert("Copied Swarm hash to clipboard!"); 
};


let copyLinkAction = (e) => {
	e.preventDefault();

	let copyText = document.querySelector('#uploadLinkInput');
	copyText.select();
	copyText.setSelectionRange(0, 99999); /*For mobile devices*/
	document.execCommand("copy");
	alert("Copied link to clipboard!"); 
};

let isUploading = false;
let currentProgressBar = null;

document.addEventListener('DOMContentLoaded', function(){ 
	let form = document.querySelector('#uploadForm');
	let uploadComponent = document.querySelector('#uploadComponent');
	let uploadFeedbackComponent = document.querySelector('#uploadFeedbackComponent');

	let resetUpload = () => {
		document.querySelector('#uploadSelectFile').value = "";
		document.querySelector('#uploadSelectedFile').value = "";
		document.querySelector('#uploadSwarmhash').innerHTML = "Waiting for hash...";
		document.querySelector('#uploadHashInput').classList.remove('hidden');
		document.querySelector('#uploadButtonHash').classList.add('fadeOut');
		document.querySelector('#uploadButtonLink').classList.add('fadeOut');

		document.querySelector('#uploadComponent .controlComponentMessage').innerHTML = "Select your file to upload it to the Swarm network.";
		isUploading = false;
		if(currentProgressBar !== null){
			currentProgressBar.cancel();			
		}
	}

	form.addEventListener("submit", (e)=>{
		e.preventDefault();

		if(currentProgressBar){
			currentProgressBar.cancel();
		}

		if(document.querySelector('#uploadSelectFile').value === ""){
			return false;
		}

		if(isUploading === true){
			return false;
		}

		isUploading = true;

		let formData = new FormData(form);

		document.querySelector('#uploadFilename').innerHTML = truncateEnd(formData.get('file').name, 45);

		if(formData.get('file')){
			swb = new SwarmProgressBar(gatewayHost);
			currentProgressBar = swb;
			swb.onProgress((status)=>{
				let totalLength = status.Total.toString().length;
				let syncedString = "";
				let syncedPercent = 0;

				if(
					status.Synced !== false &&
					status.Total !== false && 
					status.Seen !== false
				){

					if(status.Total - status.Seen > 0){
						syncedPercent = Math.ceil((status.Synced/(status.Total - status.Seen)) * 100, 2);				
					}else{
						syncedPercent = 100;
					}

					if(
						status.Total - ( status.Synced + status.Seen ) > 0
					){
						syncedString = 'Syncing <span class="uploadFeedbackCountNumbers">'+syncedPercent+'%</span>';
					}else{
						syncedString = 'Synced <span class="uploadFeedbackCountNumbers">'+syncedPercent+'%</span>';
					}
				}

				document.querySelector('#uploadReceivedCount').innerHTML = status.Received !== false ? padNumber(status.Received, 3) + "%" : "";
				document.querySelector('#uploadSyncedCount').innerHTML = syncedString;

				document.querySelector('#uploadReceivedBar').setAttribute('style', status.Received !== false ? "width: "+ status.Received + "%" : "");
				document.querySelector('#uploadSyncedBar').setAttribute('style', status.Synced !== false ? "width: "+ syncedPercent + "%" : "");
			});
			swb.onStart((event)=>{
				fadeInComponent(false, '#uploadFeedbackComponent')
			})
			swb.onError((event)=>{
				console.log('error', event);
			})
			swb.onUploaded((response)=>{
				document.querySelector('#uploadStatusMessage').innerHTML = "Uploaded";
				fadeAndReplace(
					'#uploadSwarmhash', 
					swb.status.swarmHash !== false ? swb.status.swarmHash : ""
				);
				document.querySelector('#uploadButtonLink').classList.remove("fadeOut");
				document.querySelector('#uploadLinkInput').value = swb.status.gatewayLink;
				document.querySelector('#uploadButtonHash').classList.remove("fadeOut");
				document.querySelector('#uploadHashInput').value = swb.status.swarmHash;
			})
			swb.upload(formData);
		}
	}, false);

	let uploadSelectFile = document.querySelector('#uploadSelectFile');
	let uploadSelectedFile = document.querySelector('#uploadSelectedFile');
	form.addEventListener("change", (e)=>{
		e.preventDefault();
		if(e.target.files.length > 0){
			fadeAndReplace(
				'#uploadComponent .controlComponentMessage', 
				"Upload '" + truncateEnd(e.target.files[0].name,50) + "' (" + humanFileSize(e.target.files[0].size) +") ?"
				);
			uploadSelectedFile.value = truncateEnd(e.target.files[0].name, 96);
		}else{
			uploadSelectedFile.value = "";
		}
	}, false);

	document.querySelector('#uploadButtonLink').addEventListener('click', copyLinkAction)   ; 

	document.querySelector('#uploadButtonHash').addEventListener('click', copyHashAction);

	document.querySelector('#controlHeaderDownload').addEventListener('click', (e) => {
		fadeInComponent('#controlHeaderDownload', '#downloadComponent')
	});


	document.querySelector('#controlHeaderUpload').addEventListener('click', (e) => {
		resetUpload();
		fadeInComponent('#controlHeaderUpload', '#uploadComponent'); 
	});

	document.querySelector('#uploadCancelButton').addEventListener('click', (e) => {
		resetUpload();
		fadeInComponent('#controlHeaderUpload', '#uploadComponent'); 
	});

	document.querySelector('#downloadForm button').addEventListener('click', (e) => {
		e.preventDefault();
		goToPage();
	});

}, false);
{{ end }}`

const logo = `{{ define "logo" }}
<a href="/bzz:/swarm.eth"><img src="data:image/png;base64,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"/></a>
{{ end }}`

const favicon = `000001000400101000000000200068040000460000002020000000002000a8100000ae0400003030000000002000a825000056150000404000000000200028420000fe3a000028000000100000002000000001002000000000004004000000000000000000000000000000000000ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017e7e7e0362626263545454c548484849ffffff01ffffff01ffffff01ffffff01646464375b5b5bbf4545457758585809ffffff01ffffff01ffffff0164646443626262cf626262ff535353ff454545ff454545b74949492b6868681d626262a5626262fd5c5c5cff464646ff454545dd47474755ffffff01ffffff013f3f3feb565656ff636363ff535353ff464646ff3f3f3fff373737ab393939894d4d4dff626262ff5c5c5cff464646ff424242ff3a3a3af7ffffff01ffffff01383838e9353535ff424242ff474747ff383838ff353535ff363636ab35353587363636ff3a3a3aff4a4a4aff3b3b3bff353535ff363636f5ffffff01ffffff01383838e9303030ff181818ff131313ff232323ff343434ff363636ab35353587343434ff202020ff101010ff1d1d1dff303030ff373737f5ffffff01ffffff01232323c50c0c0cff0d0d0dff131313ff171717ff171717ff2929298b2727276b0f0f0ffd0d0d0dff101010ff171717ff161616ff232323d9ffffff01ffffff014d4d4d030f0f0f650c0c0ce7131313ff161616d51d1d1d4b63636363464646691717173b0d0d0dc50f0f0fff161616ef171717752e2e2e07ffffff01ffffff01ffffff01ffffff011d1d1d0f1515155360606045626262cf636363ff464646ff454545d3484848491414144d24242417ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013c3c3c374f4f4fff636363ff636363ff464646ff464646ff3f3f3fff3c3c3c41ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013636363d353535ff3c3c3cff575757ff363636ff181818ff282828ff37373747ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013636363d363636ff303030ff181818ff292929ff131313ef17171771696969136565653bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01323232371e1e1eff0d0d0dff0c0c0cff363636ff363636a3ffffff0185858515606060ff4747476bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01111111450d0d0dd10c0c0cff1b1b1bff2a2a2a993e3e3e0b30303085292929ff37373787ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01363636030e0e0e671616166b45454505323232432e2e2ed9151515c31d1d1d2dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff014e4e4e05ffffff01ffffff01ffffff01ffffff010000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff28000000200000004000000001002000000000008010000000000000000000000000000000000000ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017272721b646464a54646466f72727205ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0168686845575757b74f4f4f39ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017e7e7e0b6262627d616161f3636363ff424242ff444444d74f4f4f49ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff016c6c6c27636363b5616161ff555555ff434343ff464646a35858581dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff016666665d616161e3626262ff636363ff636363ff444444ff464646ff434343ff454545b95252522bffffff01ffffff01ffffff01ffffff016c6c6c1363636393616161fb636363ff636363ff555555ff464646ff464646ff444444f5464646836666660bffffff01ffffff01ffffff01ffffff01ffffff016a6a6a3f626262c9616161ff636363ff636363ff636363ff636363ff444444ff464646ff464646ff464646ff434343fb48484897545454135b5b5b036868686f616161ef626262ff636363ff636363ff636363ff555555ff464646ff464646ff464646ff454545ff444444e54a4a4a5fffffff01ffffff01ffffff01ffffff013b3b3bd7505050ff646464ff636363ff636363ff636363ff636363ff444444ff464646ff464646ff464646ff454545ff3a3a3aff33333357313131113c3c3cff5a5a5aff646464ff636363ff636363ff636363ff555555ff464646ff464646ff464646ff464646ff424242ff383838f1ffffff01ffffff01ffffff01ffffff013a3a3ad5353535ff3a3a3aff575757ff646464ff626262ff636363ff444444ff464646ff464646ff3d3d3dff353535ff363636ff3636365535353511363636ff343434ff434343ff606060ff636363ff636363ff555555ff464646ff464646ff444444ff393939ff353535ff373737edffffff01ffffff01ffffff01ffffff013a3a3ad5363636ff363636ff343434ff3f3f3fff5d5d5dff646464ff444444ff404040ff363636ff353535ff363636ff363636ff3636365535353511363636ff363636ff363636ff343434ff4a4a4aff636363ff555555ff454545ff3c3c3cff353535ff363636ff363636ff373737edffffff01ffffff01ffffff01ffffff013a3a3ad5363636ff363636ff363636ff363636ff353535ff3f3f3fff363636ff353535ff363636ff363636ff363636ff363636ff3636365535353511363636ff363636ff363636ff363636ff353535ff383838ff3a3a3aff373737ff353535ff363636ff363636ff363636ff373737edffffff01ffffff01ffffff01ffffff013a3a3ad5363636ff363636ff363636ff323232ff181818ff0e0e0eff171717ff282828ff373737ff363636ff363636ff363636ff3636365535353511363636ff363636ff353535ff373737ff292929ff0f0f0fff111111ff1b1b1bff2f2f2fff373737ff363636ff363636ff373737edffffff01ffffff01ffffff01ffffff013a3a3ad5363636ff363636ff1e1e1eff0b0b0bff0d0d0dff0f0f0fff171717ff161616ff191919ff2c2c2cff373737ff363636ff3636365535353511363636ff373737ff2f2f2fff141414ff0b0b0bff0d0d0dff131313ff171717ff151515ff1f1f1fff333333ff363636ff373737edffffff01ffffff01ffffff01ffffff013b3b3bd5252525ff0d0d0dff0c0c0cff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff151515ff1c1c1cff313131ff3535355734343411333333ff1a1a1aff0b0b0bff0d0d0dff0d0d0dff0d0d0dff131313ff171717ff171717ff171717ff161616ff242424ff373737efffffff01ffffff01ffffff01ffffff012020205d0b0b0be50b0b0bff0d0d0dff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff171717ff131313ff161616b73333331f3b3b3b05111111970a0a0afb0d0d0dff0d0d0dff0d0d0dff0d0d0dff131313ff171717ff171717ff171717ff161616ff141414f51c1c1c7fffffff01ffffff01ffffff01ffffff01ffffff014d4d4d0b1212127f0a0a0af50d0d0dff0d0d0dff0f0f0fff171717ff171717ff151515ff151515d522222249ffffff017373731b51515121ffffff011d1d1d2b101010b50a0a0aff0d0d0dff0d0d0dff131313ff171717ff171717ff131313ff181818a12e2e2e1dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff012c2c2c1b0f0f0fa10a0a0afd0f0f0fff161616ff141414e91b1b1b69656565057878780b6363637b626262f3464646f7454545896969690fffffff011c1c1c470c0c0cd30b0b0bff131313ff141414ff151515c32a2a2a37ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff011d1d1d35111111bd1a1a1a8d2f2f2f11ffffff0166666659616161e1626262ff646464ff474747ff454545ff444444e9494949677b7b7b054040400517171769131313cd24242455ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0169696939626262c7616161ff636363ff636363ff646464ff474747ff464646ff464646ff444444ff454545d14e4e4e45ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01424242615e5e5eff636363ff636363ff636363ff636363ff646464ff474747ff464646ff464646ff464646ff464646ff434343ff3f3f3f77ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0136363679343434ff494949ff636363ff636363ff636363ff646464ff474747ff464646ff464646ff474747ff3d3d3dff353535ff3a3a3a8dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0136363679363636ff353535ff363636ff505050ff646464ff636363ff474747ff484848ff2f2f2fff1c1c1cff323232ff363636ff3a3a3a8dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0136363679363636ff363636ff363636ff353535ff3a3a3aff5a5a5aff393939ff0f0f0fff040404ff111111ff151515ff232323ff3535358fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0136363679363636ff363636ff363636ff363636ff323232ff171717ff2a2a2aff0c0c0cff030303ff111111ff141414fb171717992e2e2e17a3a3a305ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0136363679363636ff363636ff363636ff1f1f1fff0b0b0bff0d0d0dff363636ff383838ff242424ff121212bf2a2a2a2dffffff01ffffff018484842b636363bf6d6d6d2fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0136363679373737ff252525ff0d0d0dff0c0c0cff0d0d0dff0d0d0dff373737ff363636ff353535ff39393949ffffff01ffffff01ffffff0186868629646464ff656565fb6464649b55555505ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff012e2e2e650e0e0eff0c0c0cff0d0d0dff0d0d0dff0d0d0dff0c0c0cff353535ff363636ff353535ff37373749ffffff01ffffff01ffffff0185858529656565ff525252ff353535ff4b4b4b0fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff011c1c1c430d0d0dcf0b0b0bff0d0d0dff0d0d0dff0d0d0dff171717ff282828ff363636ff37373749ffffff01ffffff01ffffff0144444459363636ff353535ff353535ff4e4e4e0fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0162626203161616630b0b0be70c0c0cff0d0d0dff171717ff161616ff171717ed3737372fffffff013e3e3e2b303030b72a2a2aff151515ff262626ff363636ff4b4b4b0fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013636360d101010850a0a0af7141414f91717178f45454511ffffff014c4c4c252c2c2cdb303030ff2d2d2dff151515ff131313ff1b1b1bad5a5a5a07ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff012b2b2b2121212127ffffff01ffffff01ffffff01ffffff0161616109313131752b2b2bf1131313cd26262641ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff014e4e4e1359595903ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000028000000300000006000000001002000000000008025000000000000000000000000000000000000ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0173737357545454997c7c7c11ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0176767663515151916c6c6c0dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017676762d636363bb636363ff4d4d4dff434343eb4f4f4f6d7f7f7f05ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0176767635616161c3626262ff494949ff424242e94f4f4f6392929203ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017e7e7e19626262955f5f5ffd626262ff666666ff4f4f4fff464646ff424242ff434343d75a5a5a49ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017777771d6464649f5f5f5fff636363ff656565ff4b4b4bff464646ff424242ff444444d158585841ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff018585850966666677606060ef626262ff636363ff636363ff666666ff4f4f4fff464646ff464646ff464646ff414141ff464646b75d5d5d2dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff018989890d6868687f5f5f5ff5626262ff636363ff636363ff656565ff4b4b4bff464646ff464646ff464646ff404040ff484848b160606027ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff016a6a6a55626262df606060ff636363ff636363ff636363ff636363ff666666ff4f4f4fff464646ff464646ff464646ff464646ff454545ff424242fd484848956a6a6a17ffffff01ffffff01ffffff01ffffff01ffffff016969695f606060e3606060ff636363ff636363ff636363ff636363ff656565ff4b4b4bff464646ff464646ff464646ff464646ff454545ff414141f94a4a4a8d65656513ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff016e6e6e3b656565c15f5f5fff636363ff636363ff636363ff636363ff636363ff636363ff666666ff4f4f4fff464646ff464646ff464646ff464646ff464646ff464646ff444444ff424242ed52525277ffffff01ffffff016c6c6c37676767c95f5f5fff636363ff636363ff636363ff636363ff636363ff636363ff656565ff4b4b4bff464646ff464646ff464646ff464646ff464646ff464646ff434343ff444444e94d4d4d6dffffff01ffffff01ffffff01ffffff01ffffff01ffffff013c3c3cc5454545ff646464ff646464ff636363ff636363ff636363ff636363ff636363ff666666ff4f4f4fff464646ff464646ff464646ff464646ff464646ff464646ff474747ff424242ff333333fb34343409ffffff0131313199494949ff656565ff646464ff636363ff636363ff636363ff636363ff636363ff656565ff4b4b4bff464646ff464646ff464646ff464646ff464646ff464646ff474747ff414141ff373737ebffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040bf333333ff343434ff4f4f4fff666666ff636363ff636363ff636363ff636363ff666666ff4f4f4fff464646ff464646ff464646ff464646ff474747ff444444ff383838ff343434ff363636f737373707ffffff0135353597343434ff343434ff525252ff666666ff636363ff636363ff636363ff636363ff656565ff4b4b4bff464646ff464646ff464646ff464646ff474747ff444444ff383838ff343434ff393939e3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040bf343434ff363636ff333333ff383838ff585858ff676767ff636363ff636363ff666666ff4f4f4fff464646ff464646ff474747ff464646ff3b3b3bff343434ff363636ff363636ff363636f737373707ffffff0135353597363636ff363636ff333333ff383838ff5a5a5aff666666ff636363ff636363ff656565ff4b4b4bff464646ff464646ff474747ff454545ff3a3a3aff343434ff363636ff363636ff393939e3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040bf343434ff363636ff363636ff363636ff323232ff3d3d3dff5d5d5dff666666ff666666ff4f4f4fff464646ff474747ff3e3e3eff353535ff353535ff363636ff363636ff363636ff363636f737373707ffffff0135353597363636ff363636ff363636ff363636ff313131ff3f3f3fff5f5f5fff666666ff656565ff4b4b4bff464646ff474747ff3d3d3dff353535ff353535ff363636ff363636ff363636ff393939e3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040bf343434ff363636ff363636ff363636ff363636ff353535ff323232ff444444ff676767ff525252ff404040ff363636ff353535ff363636ff363636ff363636ff363636ff363636ff363636f737373707ffffff0135353597363636ff363636ff363636ff363636ff363636ff353535ff323232ff464646ff676767ff4e4e4eff404040ff363636ff353535ff363636ff363636ff363636ff363636ff363636ff393939e3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040bf343434ff363636ff363636ff363636ff363636ff363636ff353535ff383838ff2d2d2dff2b2b2bff373737ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636f737373707ffffff0135353597363636ff363636ff363636ff363636ff363636ff363636ff363636ff383838ff2c2c2cff2a2a2aff373737ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff393939e3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040bf343434ff363636ff363636ff363636ff353535ff383838ff343434ff171717ff090909ff151515ff171717ff2d2d2dff383838ff363636ff363636ff363636ff363636ff363636ff363636f737373707ffffff0135353597363636ff363636ff363636ff363636ff353535ff383838ff333333ff151515ff090909ff151515ff181818ff2f2f2fff383838ff363636ff363636ff363636ff363636ff363636ff393939e3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040bf343434ff363636ff363636ff373737ff373737ff1f1f1fff090909ff0c0c0cff0c0c0cff171717ff171717ff141414ff1b1b1bff323232ff383838ff363636ff363636ff363636ff363636f737373707ffffff0135353597363636ff363636ff363636ff373737ff373737ff1d1d1dff0a0a0aff0c0c0cff0c0c0cff171717ff171717ff141414ff1c1c1cff333333ff383838ff353535ff363636ff363636ff393939e3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040bf343434ff363636ff393939ff272727ff0c0c0cff0b0b0bff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff161616ff141414ff202020ff353535ff373737ff363636ff363636f737373707ffffff0135353597363636ff363636ff383838ff252525ff0b0b0bff0b0b0bff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff161616ff141414ff222222ff363636ff373737ff363636ff393939e3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040bf383838ff2d2d2dff101010ff0a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff171717ff171717ff161616ff141414ff262626ff373737ff373737f737373707ffffff0136363697393939ff2b2b2bff0f0f0fff0a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff171717ff171717ff161616ff151515ff272727ff383838ff393939e3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013a3a3abd131313ff090909ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff171717ff171717ff171717ff171717ff151515ff171717ff262626fb38383807ffffff012a2a2a97121212ff090909ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff171717ff171717ff171717ff171717ff151515ff161616ff2a2a2ae7ffffff01ffffff01ffffff01ffffff01ffffff01ffffff015f5f5f0b1616167b090909ef0a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff171717ff171717ff171717ff171717ff0f0f0fff181818b74040402dffffff01ffffff014646461118181883080808f30b0b0bff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff171717ff171717ff171717ff161616ff101010ff181818b141414127ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff014d4d4d171212129b090909fd0c0c0cff0d0d0dff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff171717ff171717ff111111ff141414d335353547ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013838381d131313a5060606ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff171717ff171717ff111111ff181818cd2e2e2e3dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01333333310f0f0fbb070707ff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff141414ff121212e72424246d86868603ffffff01ffffff017373732b656565b9464646c95e5e5e3bffffff01ffffff01ffffff01323232370e0e0ec3080808ff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff121212ff161616e525252563ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff012525254d0e0e0ed9090909ff0c0c0cff171717ff151515ff121212f91d1d1d894d4d4d13ffffff01ffffff0178787815656565935f5f5ffb646464ff484848ff404040ff454545a96a6a6a1fffffff01ffffff01ffffff011b1b1b570e0e0edf080808ff0d0d0dff171717ff151515ff0f0f0ff3212121815656560dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01636363071a1a1a710a0a0aed0f0f0fff1b1b1bad2f2f2f23ffffff01ffffff018d8d8d0566666675616161eb616161ff636363ff646464ff484848ff464646ff454545ff424242f54c4c4c856262620fffffff01ffffff014040400b21212179080808f10f0f0fff1b1b1ba15757571dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff014141411740404037ffffff01ffffff01ffffff016a6a6a4d616161db606060ff636363ff636363ff636363ff646464ff484848ff464646ff464646ff464646ff434343ff434343e751515167ffffff01ffffff01ffffff014646461d30303033ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0176767631616161c35f5f5fff636363ff636363ff636363ff636363ff636363ff646464ff484848ff464646ff464646ff464646ff464646ff464646ff424242ff454545d158585841ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff015252527f636363ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff646464ff484848ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff434343ff454545a1ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01313131b53b3b3bff5b5b5bff676767ff636363ff636363ff636363ff636363ff636363ff646464ff484848ff464646ff464646ff464646ff464646ff464646ff474747ff444444ff393939ff383838d3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01363636b3363636ff323232ff404040ff616161ff656565ff626262ff636363ff636363ff646464ff484848ff464646ff464646ff454545ff494949ff474747ff3b3b3bff343434ff353535ff3a3a3ad3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01363636b3363636ff363636ff353535ff323232ff484848ff656565ff646464ff636363ff646464ff484848ff464646ff474747ff494949ff242424ff282828ff383838ff363636ff363636ff3a3a3ad3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01363636b3363636ff363636ff363636ff363636ff343434ff343434ff515151ff666666ff656565ff484848ff4b4b4bff323232ff070707ff040404ff151515ff181818ff2f2f2fff383838ff3a3a3ad3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01363636b3363636ff363636ff363636ff363636ff363636ff363636ff333333ff383838ff5f5f5fff3c3c3cff0f0f0fff020202ff050505ff050505ff171717ff171717ff141414ff1c1c1cff323232d7ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01363636b3363636ff363636ff363636ff363636ff363636ff353535ff383838ff343434ff161616ff2a2a2aff0c0c0cff020202ff050505ff050505ff171717ff171717ff101010ff161616bf2e2e2e35ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01363636b3363636ff363636ff363636ff363636ff373737ff383838ff1f1f1fff0a0a0aff0c0c0cff373737ff3a3a3aff262626ff060606ff040404ff121212ff151515dd30303051ffffff01ffffff01ffffff018787872d6b6b6b47ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01363636b3363636ff363636ff363636ff393939ff272727ff0d0d0dff0b0b0bff0d0d0dff0d0d0dff373737ff363636ff373737ff383838ff1c1c1cf92020207568686807ffffff01ffffff01ffffff01ffffff018686863d5f5f5fff676767af77777721ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01363636b3363636ff393939ff2e2e2eff101010ff0a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff373737ff363636ff363636ff353535ff373737ebffffff01ffffff01ffffff01ffffff01ffffff01ffffff018686863d626262ff666666ff646464f76969698d9494940fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01383838b5333333ff161616ff090909ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff373737ff363636ff363636ff363636ff353535ebffffff01ffffff01ffffff01ffffff01ffffff01ffffff018686863d626262ff676767ff6b6b6bff555555ff3a3a3a93ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0125252589030303ff0c0c0cff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff333333ff383838ff353535ff363636ff353535ebffffff01ffffff01ffffff01ffffff01ffffff01ffffff018585853d666666ff5f5f5fff3c3c3cff313131ff3a3a3a93ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff012d2d2d3f0e0e0ecb080808ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff141414ff222222ff363636ff373737ff353535ebffffff01ffffff01ffffff01ffffff01ffffff01ffffff0177777741414141ff313131ff363636ff353535ff3a3a3a93ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff011e1e1e5f0a0a0ae50a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff171717ff161616ff151515ff282828ff353535f3ffffff01ffffff01ffffff01ffffff016e6e6e0b37373781242424f1191919ff333333ff383838ff343434ff3a3a3a93ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff015a5a5a0d1919197f0a0a0af30b0b0bff0d0d0dff0d0d0dff171717ff171717ff161616ff0f0f0ffb24242489ffffff01ffffff01ffffff013e3e3e5d2d2d2de52e2e2eff2b2b2bff151515ff141414ff212121ff363636ff3b3b3b95ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013636361b111111a3080808ff0c0c0cff181818ff0f0f0fff171717b545454525ffffff01ffffff017f7f7f05363636c7282828ff313131ff313131ff2b2b2bff151515ff171717ff161616ff0c0c0cfb3434346bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01303030350f0f0fc7121212d337373741ffffff01ffffff01ffffff01ffffff01ffffff016b6b6b0b3a3a3a7d2c2c2cf12f2f2fff2b2b2bff151515ff101010ff171717bb4646462dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01515151193535359b242424ff131313d72828284bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff014e4e4e2b59595905ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff28000000400000008000000001002000000000000042000000000000000000000000000000000000ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0176767635666666914e4e4e457c7c7c09ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff018080801569696989545454696c6c6c0bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff018484840d70707061616161d5606060fb3d3d3ddf4e4e4e9172727213ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017070704d626262b35f5f5ffb464646f1454545a16a6a6a33ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017676760f67676753646464cf5e5e5eff656565ff626262ff414141ff404040ff444444e54b4b4b7b69696919ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01979797036c6c6c45676767a95d5d5dff616161ff626262ff484848ff424242ff3e3e3efd4e4e4e8958585831ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017e7e7e2b616161a75f5f5fef616161ff636363ff656565ff626262ff424242ff464646ff444444ff414141fd434343b961616153ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017777771969696981606060e7606060ff636363ff636363ff626262ff484848ff464646ff454545ff424242fd414141d95656566569696911ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01858585056e6e6e29656565995f5f5ff1616161ff636363ff636363ff636363ff656565ff626262ff424242ff464646ff464646ff464646ff444444ff3f3f3fff484848af5353534b86868607ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01797979216a6a6a6f616161ed5e5e5eff636363ff636363ff636363ff636363ff626262ff484848ff464646ff464646ff464646ff464646ff3e3e3eff474747d75151515762626213ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01838383036f6f6f755f5f5fd3606060ff626262ff636363ff636363ff636363ff636363ff656565ff626262ff424242ff464646ff464646ff464646ff464646ff454545ff434343ff404040e94e4e4e8d5f5f5f1bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff018f8f8f056b6b6b45616161c95f5f5ff7616161ff636363ff636363ff636363ff636363ff636363ff626262ff484848ff464646ff464646ff464646ff464646ff464646ff444444ff424242f1434343b16666662dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017070700f6969695f626262d35e5e5eff626262ff636363ff636363ff636363ff636363ff636363ff636363ff656565ff626262ff424242ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff404040ff444444f14d4d4d776a6a6a23ffffff01ffffff01ffffff01ffffff017b7b7b096c6c6c39636363c15f5f5ffb626262ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff626262ff484848ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff434343ff414141f54a4a4aa35b5b5b2d70707007ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0171717143676767a7616161f3616161ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff656565ff626262ff424242ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff444444ff414141f7474747cd54545447ffffff01ffffff015b5b5b096b6b6b99646464e1606060ff626262ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff626262ff484848ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff444444ff424242ff414141d552525277ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040b33b3b3bff5c5c5cff656565ff646464ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff656565ff626262ff424242ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff474747ff454545ff3a3a3aff313131ad34343407ffffff012e2e2e25383838ff535353ff656565ff656565ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff626262ff484848ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff474747ff464646ff3b3b3bff3a3a3ae9ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9313131ff363636ff484848ff636363ff676767ff636363ff636363ff636363ff636363ff636363ff636363ff656565ff626262ff424242ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff404040ff363636ff343434ff353535a537373705ffffff0135353521333333ff333333ff434343ff5c5c5cff686868ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff626262ff484848ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff484848ff414141ff393939ff313131ff3c3c3cdbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff363636ff323232ff353535ff4b4b4bff636363ff656565ff636363ff626262ff636363ff636363ff656565ff626262ff424242ff464646ff464646ff464646ff464646ff474747ff464646ff414141ff363636ff343434ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff333333ff313131ff484848ff5e5e5eff666666ff646464ff626262ff636363ff636363ff636363ff626262ff484848ff464646ff464646ff464646ff464646ff464646ff474747ff424242ff3a3a3aff343434ff353535ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff363636ff363636ff343434ff333333ff3d3d3dff555555ff686868ff656565ff626262ff636363ff656565ff626262ff424242ff464646ff464646ff464646ff484848ff444444ff393939ff353535ff353535ff363636ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff363636ff353535ff323232ff363636ff515151ff646464ff656565ff636363ff636363ff636363ff626262ff484848ff464646ff464646ff464646ff484848ff454545ff3d3d3dff353535ff343434ff363636ff363636ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff363636ff363636ff363636ff363636ff343434ff303030ff3f3f3fff575757ff666666ff656565ff646464ff626262ff424242ff464646ff474747ff454545ff3a3a3aff343434ff353535ff363636ff363636ff363636ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff363636ff363636ff363636ff363636ff303030ff373737ff535353ff636363ff656565ff636363ff626262ff484848ff464646ff474747ff454545ff3e3e3eff353535ff343434ff363636ff363636ff363636ff363636ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff363636ff363636ff363636ff363636ff363636ff363636ff333333ff333333ff484848ff606060ff696969ff626262ff434343ff474747ff3e3e3eff363636ff353535ff353535ff363636ff363636ff363636ff363636ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff363636ff363636ff363636ff363636ff353535ff343434ff333333ff3e3e3eff5d5d5dff686868ff626262ff484848ff474747ff424242ff373737ff353535ff353535ff363636ff363636ff363636ff363636ff363636ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff323232ff323232ff505050ff616161ff3d3d3dff373737ff343434ff353535ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff343434ff313131ff434343ff606060ff464646ff383838ff343434ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff3a3a3aff2b2b2bff1e1e1eff2d2d2dff383838ff373737ff353535ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff393939ff323232ff1c1c1cff262626ff373737ff383838ff353535ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff363636ff363636ff363636ff363636ff363636ff353535ff373737ff383838ff303030ff191919ff080808ff101010ff141414ff1a1a1aff303030ff383838ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff363636ff363636ff363636ff363636ff353535ff363636ff383838ff363636ff1d1d1dff0b0b0bff0c0c0cff141414ff181818ff292929ff373737ff373737ff363636ff363636ff363636ff363636ff363636ff363636ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff363636ff363636ff363636ff363636ff353535ff393939ff363636ff222222ff0c0c0cff0a0a0aff0c0c0cff121212ff171717ff151515ff161616ff212121ff353535ff393939ff363636ff363636ff363636ff363636ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff363636ff363636ff363636ff353535ff383838ff3a3a3aff262626ff121212ff0a0a0aff0c0c0cff0f0f0fff171717ff151515ff151515ff1e1e1eff2f2f2fff3a3a3aff363636ff363636ff363636ff363636ff363636ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff363636ff363636ff363636ff383838ff363636ff262626ff0d0d0dff090909ff0d0d0dff0d0d0dff0c0c0cff121212ff171717ff171717ff171717ff141414ff151515ff232323ff353535ff383838ff363636ff353535ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff363636ff353535ff383838ff383838ff292929ff131313ff080808ff0c0c0cff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff151515ff131313ff202020ff313131ff383838ff363636ff363636ff363636ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff353535ff363636ff3a3a3aff2e2e2eff131313ff0a0a0aff0b0b0bff0d0d0dff0d0d0dff0d0d0dff0c0c0cff121212ff171717ff171717ff171717ff171717ff161616ff141414ff1a1a1aff2a2a2aff393939ff373737ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff363636ff3a3a3aff313131ff1c1c1cff0a0a0aff0a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff171717ff161616ff151515ff161616ff282828ff363636ff383838ff363636ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9353535ff383838ff313131ff151515ff080808ff0b0b0bff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff121212ff171717ff171717ff171717ff171717ff171717ff171717ff161616ff131313ff1b1b1bff2d2d2dff373737ff373737ff363636a537373705ffffff0134343421363636ff383838ff333333ff1e1e1eff090909ff0a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff171717ff171717ff171717ff171717ff131313ff171717ff2a2a2aff363636ff353535ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444af353535ff1e1e1eff0d0d0dff0a0a0aff0c0c0cff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff121212ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff151515ff151515ff222222ff333333ff353535ad30303007ffffff0134343423373737ff282828ff0d0d0dff0a0a0aff0c0c0cff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff151515ff141414ff1b1b1bff2e2e2eff3e3e3ee1ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013e3e3e6f0f0f0fd5040404ff0b0b0bff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff121212ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff101010ff0e0e0ee72f2f2f7347474703ffffff013b3b3b13141414cd050505f70a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff121212ff0c0c0cf12a2a2aa5ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff015f5f5f052020202b1a1a1aa1080808f1070707ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff121212ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff141414ff0c0c0cff212121af2a2a2a496d6d6d07ffffff01ffffff01ffffff01333333231d1d1d730b0b0beb060606ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff151515ff0e0e0eff181818d72626265546464615ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff014d4d4d29121212af080808ef0a0a0aff0c0c0cff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff121212ff171717ff171717ff171717ff171717ff171717ff171717ff141414ff121212f9141414b93b3b3b4fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0138383819151515890a0a0ae5080808ff0c0c0cff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff171717ff171717ff171717ff161616ff101010fb151515d72c2c2c614444440dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0133333311262626510f0f0fd7050505ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff121212ff171717ff171717ff171717ff171717ff171717ff101010ff141414e7242424733a3a3a19ffffff01ffffff01ffffff01878787097272725f4d4d4d736a6a6a11ffffff01ffffff01ffffff016060600524242445191919ad040404ff0a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff171717ff171717ff111111ff0e0e0efd242424873232322dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff015c5c5c0d2525255f090909d7080808fb0b0b0bff0d0d0dff0c0c0cff121212ff171717ff171717ff161616ff121212ff121212df2121218965656511ffffff01ffffff01ffffff018080800d6767674b646464d1606060ff454545ff464646df4f4f4f6165656517ffffff01ffffff01ffffff01ffffff012d2d2d4b101010b5060606fb0a0a0aff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff161616ff131313ff101010ef2020209d4242422dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff012c2c2c2d1f1f1f83080808fb080808ff0d0d0dff121212ff171717ff141414ff0f0f0ff91e1e1eb12c2c2c354d4d4d09ffffff01ffffff01ffffff0178787825646464a75f5f5feb616161ff656565ff4a4a4aff414141ff424242f3414141bd69696937ffffff01ffffff01ffffff01ffffff0142424219171717710d0d0de3060606ff0c0c0cff0f0f0fff171717ff151515ff0d0d0dff171717c3292929575656560dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013737372d1212129d080808ef0d0d0dff121212f5191919bf2e2e2e3d70707003ffffff01ffffff018c8c8c037676762564646497606060ed606060ff636363ff636363ff656565ff4a4a4aff444444ff464646ff444444ff404040f74a4a4aad5555553162626207ffffff01ffffff01ffffff014040401125252589090909dd0a0a0aff121212ff141414c738383869ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff015b5b5b0b1f1f1f591d1d1daf292929673f3f3f19ffffff01ffffff01ffffff01ffffff016d6d6d715f5f5fcd606060ff626262ff636363ff636363ff636363ff656565ff4a4a4aff444444ff464646ff464646ff454545ff434343ff414141db4f4f4f857b7b7b11ffffff01ffffff01ffffff0153535307222222331d1d1da91b1b1b8d4141412365656503ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017c7c7c0f6868685d636363cb5e5e5eff626262ff636363ff636363ff636363ff636363ff636363ff656565ff4a4a4aff444444ff464646ff464646ff464646ff464646ff464646ff404040ff454545e14c4c4c6b69696917ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0177777733626262a3606060f3616161ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff656565ff4a4a4aff444444ff464646ff464646ff464646ff464646ff464646ff464646ff444444ff424242f9454545b55d5d5d49ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff014b4b4b0f5e5e5e85626262ff626262ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff656565ff4a4a4aff444444ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff444444ff414141ff454545a16464641dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0132323225333333cf4e4e4eff646464ff666666ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff656565ff4a4a4aff444444ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff474747ff404040ff303030e35757573bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd313131ff363636ff515151ff636363ff656565ff636363ff636363ff636363ff636363ff636363ff636363ff656565ff4a4a4aff444444ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff414141ff373737ff343434ff323232e159595939ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff343434ff333333ff3c3c3cff5b5b5bff686868ff636363ff626262ff636363ff636363ff636363ff656565ff4a4a4aff444444ff464646ff464646ff454545ff464646ff4c4c4cff454545ff393939ff353535ff353535ff353535ff323232e159595939ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff363636ff363636ff353535ff313131ff3f3f3fff5d5d5dff666666ff646464ff626262ff636363ff656565ff4a4a4aff444444ff454545ff474747ff4a4a4aff404040ff212121ff2f2f2fff373737ff373737ff353535ff363636ff323232e159595939ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff363636ff363636ff363636ff353535ff333333ff363636ff484848ff646464ff676767ff626262ff656565ff4a4a4aff444444ff4b4b4bff4a4a4aff262626ff0b0b0bff090909ff171717ff252525ff353535ff393939ff363636ff323232e159595939ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff363636ff363636ff363636ff363636ff363636ff363636ff323232ff363636ff4c4c4cff646464ff676767ff4d4d4dff484848ff2c2c2cff0b0b0bff020202ff040404ff0b0b0bff171717ff141414ff161616ff282828ff353535ff343434e359595939ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff343434ff323232ff3f3f3fff5f5f5fff3a3a3aff161616ff030303ff030303ff050505ff040404ff0b0b0bff171717ff171717ff161616ff151515ff1a1a1aff242424e55555553bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff363636ff363636ff363636ff363636ff363636ff363636ff353535ff363636ff383838ff2e2e2eff191919ff262626ff111111ff030303ff030303ff050505ff040404ff0b0b0bff171717ff171717ff151515ff111111f9121212cd272727557d7d7d09ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff363636ff363636ff363636ff363636ff363636ff363636ff383838ff373737ff242424ff0b0b0bff0a0a0aff393939ff393939ff222222ff080808ff020202ff030303ff0b0b0bff181818ff0f0f0fff151515f32424247935353525ffffff01ffffff01ffffff01a3a3a30fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff363636ff363636ff363636ff363636ff383838ff373737ff272727ff0c0c0cff090909ff0c0c0cff0e0e0eff373737ff363636ff3a3a3aff393939ff1e1e1eff080808ff080808ff0f0f0feb232323914040401dffffff01ffffff01ffffff01ffffff01ffffff018282825d626262c36d6d6d4d8d8d8d09ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff363636ff353535ff363636ff3a3a3aff2f2f2fff131313ff0b0b0bff0b0b0bff0d0d0dff0c0c0cff0e0e0eff373737ff363636ff353535ff363636ff393939ff303030ff1c1c1cc92626264d68686807ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01868686515e5e5eff646464e9696969957878781fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff373737ff383838ff313131ff161616ff090909ff0b0b0bff0d0d0dff0d0d0dff0d0d0dff0c0c0cff0e0e0eff373737ff363636ff363636ff363636ff353535ff353535ff3c3c3c8fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0186868651616161ff676767ff646464ff656565f16a6a6a7d7f7f7f25ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723353535cd393939ff373737ff1f1f1fff0d0d0dff0a0a0aff0c0c0cff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff0e0e0eff373737ff363636ff363636ff363636ff363636ff353535ff37373791ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0186868651616161ff676767ff666666ff676767ff686868f9555555cd55555511ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0134343425323232cf212121ff0e0e0eff090909ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff0e0e0eff383838ff363636ff363636ff363636ff363636ff353535ff37373791ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0186868651616161ff686868ff696969ff5f5f5fff3d3d3dff303030ff4848481dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01474747132323238f020202ff080808ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff0c0c0cff2e2e2eff393939ff363636ff353535ff363636ff353535ff37373791ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0185858551666666ff676767ff494949ff353535ff323232ff353535ff4e4e4e1bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0130303045101010af080808f70a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff0d0d0dff131313ff1c1c1cff303030ff373737ff363636ff353535ff37373791ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0181818151494949ff363636ff313131ff363636ff353535ff363636ff4e4e4e1bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0141414113191919690f0f0fdb060606ff0c0c0cff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff0d0d0dff171717ff151515ff161616ff222222ff363636ff383838ff37373791ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff014d4d4d53272727c1242424ff373737ff373737ff353535ff353535ff363636ff4e4e4e1bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01626262091c1c1c830b0b0bd7090909ff0c0c0cff0d0d0dff0d0d0dff0c0c0cff0d0d0dff171717ff171717ff171717ff141414ff151515ff202020ff35353595ffffff01ffffff01ffffff01ffffff017474740540404049343434af2a2a2aff262626ff101010ff191919ff2e2e2eff373737ff363636ff363636ff4e4e4e1bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff015a5a5a073636362d141414a7080808f5080808ff0d0d0dff0c0c0cff0d0d0dff171717ff171717ff171717ff151515ff0e0e0efb1b1b1bbb3d3d3d29ffffff01ffffff01ffffff0151515119393939892a2a2ae92d2d2dff323232ff282828ff141414ff151515ff151515ff1f1f1fff343434ff393939ff4949491dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013636362f111111b5070707f30a0a0aff0d0d0dff171717ff141414ff111111f5111111c74343433d70707005ffffff01ffffff017c7c7c034e4e4e632a2a2af7292929ff323232ff313131ff323232ff282828ff141414ff171717ff171717ff151515ff0e0e0efd222222e153535315ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff012d2d2d151f1f1f590e0e0edb040404ff0f0f0fff171717e7262626673f3f3f1dffffff01ffffff01ffffff01ffffff01ffffff01444444293535358b2d2d2deb2b2b2bff313131ff323232ff282828ff141414ff171717ff121212ff0d0d0dff2222229d2626263dbebebe03ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01505050112626266f1d1d1d7f36363617ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01616161213333339d2c2c2ce92f2f2fff282828ff111111ff111111f7191919ab3c3c3c41ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff015151510b3b3b3b43383838c51f1f1fff141414d71e1e1e654f4f4f13ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff015858580b4d4d4d4159595909ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000`
