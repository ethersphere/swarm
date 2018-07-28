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
	"html/template"
	"path"

	"github.com/ethereum/go-ethereum/swarm/api"
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
			funcs:        template.FuncMap{"basename": path.Base},
		},
		{
			templateName: "landing-page",
			partial:      landing,
		},
		{
			templateName: "multiple-choice",
			partial:      multipleChoice,
		},
	} {
		TemplatesMap[v.templateName] = template.Must(template.New(v.templateName).Funcs(v.funcs).Parse(baseTemplate + css + v.partial))
	}

	bytes, err := hex.DecodeString(favicon)
	if err != nil {
		panic(err)
	}
	faviconBytes = bytes
}

const bzzList = `{{ define "content" }}
<h1>Swarm index of {{ .URI }}</h1>
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
        <a href="{{ basename . }}/">{{ basename . }}/</a>
      </td>
      <td>DIR</td>
      <td>-</td>
    </tr>
    {{ end }} {{ range .List.Entries }}
    <tr>
      <td>
        <a href="{{ basename .Path }}">{{ basename .Path }}</a>
      </td>
      <td>{{ .ContentType }}</td>
      <td>{{ .Size }}</td>
    </tr>
    {{ end }}
</table>
<hr>

 {{ end }}`

const errorResponse = `{{ define "content" }}
<div class="container">
    <div class="logo">
      <a href="/bzz:/theswarm.eth">
        <svg width="180" version="1.1" viewBox="0 0 83.477 107.43" xmlns="http://www.w3.org/2000/svg" xmlns:osb="http://www.openswatchbook.org/uri/2009/osb">
          <g display="none">
            <text transform="scale(1.0173 .98298)" x="9.6367435" y="109.16406" fill="#000300" fill-opacity=".7098" font-family="Poppins"
              font-size="18.628px" font-weight="300" letter-spacing="0px" stroke-width=".4657" word-spacing="0px" style="font-feature-settings:normal;font-variant-caps:normal;font-variant-ligatures:normal;font-variant-numeric:normal;line-height:1.25"
              xml:space="preserve">
              <tspan x="9.6367435" y="109.16406" fill="#000300" fill-opacity=".7098" font-family="Poppins" font-size="18.628px" font-weight="300"
                stroke-width=".4657" style="font-feature-settings:normal;font-variant-caps:normal;font-variant-ligatures:normal;font-variant-numeric:normal">swarm</tspan>
            </text>
            <g transform="translate(-60.059 -85.696)">
              <path d="m79.635 149.73 19.576-11.306v22.612z" fill="#000301" fill-opacity=".76863" />
              <path d="m60.059 161.03 19.576 11.306-1.11e-4 -22.611-19.576 11.306" fill="#000302" fill-opacity=".60784" />
              <path d="m79.635 127.12-1.11e-4 22.612-19.576-11.306z" fill="#000201" fill-opacity=".90588" />
              <path d="m79.635 127.12 19.576 11.306-19.576 11.306 1.11e-4 -22.612" fill="#000201" fill-opacity=".87451" />
              <path d="m60.059 138.42v22.612l19.576-11.306z" fill="#000301" fill-opacity=".76863" />
              <path d="m79.635 172.34 19.576-11.306-19.576-11.306 1.11e-4 22.611" fill="#000301" fill-opacity=".7098" />
            </g>
            <path d="m22.162 36.961 19.576 11.306-1.08e-4 -22.611-19.576 11.306" fill="#000302" fill-opacity=".60784" />
            <path d="m41.739 3.0432-1.08e-4 22.612-19.576-11.306z" fill="#000201" fill-opacity=".90588" />
            <g fill="#000301">
              <path d="m22.162 14.349v22.612l19.576-11.306z" fill-opacity=".76863" />
              <path d="m41.739 48.267 19.576-11.306-19.576-11.306 1.08e-4 22.611" fill-opacity=".7098" />
              <path d="m51.527 8.7322-9.788 5.6529v11.306l9.788-5.6529z" fill-opacity=".76471" />
            </g>
            <path d="m66.559 11.323 9.788 5.6529-9.788 5.6529z" fill="#000302" fill-opacity=".59216" />
            <path d="m41.739 3.0793 9.788 5.6529-9.788 5.6529z" fill="#000201" fill-opacity=".87451" />
            <path d="m76.347 16.976-9.788-5.6529 9.788-5.6529z" fill="#000301" fill-opacity=".76863" />
            <path d="m51.527 20.038 9.788 5.6529-9.788 5.6529z" fill="#000201" fill-opacity=".87451" />
            <path d="m61.315 36.997-9.788-5.6529 9.788-5.6529z" fill="#000301" fill-opacity=".76863" />
            <path d="m51.527 31.344-9.788-5.6529 9.788-5.6529z" fill="#000103" fill-opacity=".93725" />
            <path d="m66.559 0.0177 9.788 5.6529-9.788 5.6529z" fill="#000201" fill-opacity=".87451" />
            <path d="m66.559 11.323-9.788-5.6529 9.788-5.6529z" fill="#000104" fill-opacity=".78431" />
            <g transform="translate(-15.734 -85.696)">
              <path d="m79.635 149.73 19.576-11.306v22.612z" fill="#000301" fill-opacity=".76863" />
              <path d="m60.059 161.03 19.576 11.306-1.11e-4 -22.611-19.576 11.306" fill="#000302" fill-opacity=".60784" />
              <path d="m79.635 127.12-1.11e-4 22.612-19.576-11.306z" fill="#000201" fill-opacity=".90588" />
              <path d="m79.635 127.12 19.576 11.306-19.576 11.306 1.11e-4 -22.612" fill="#000201" fill-opacity=".87451" />
              <path d="m60.059 138.42v22.612l19.576-11.306z" fill="#000301" fill-opacity=".76863" />
              <path d="m79.635 172.34 19.576-11.306-19.576-11.306 1.11e-4 22.611" fill="#000301" fill-opacity=".7098" />
            </g>
          </g>
          <g transform="translate(-60.994 -58.006)">
            <g transform="translate(44.325 -.00024)">
              <g fill="#000302">
                <path d="m60.994 110.73 19.576-11.306 19.576 11.306v22.612l-19.576 11.306-19.576-11.306 9e-5 -22.612" fill-opacity=".60784"
                />
                <path d="m80.57 99.426-19.576 11.306v22.612l19.576-11.306v22.612l19.576-11.306v-22.612z" fill-opacity=".25882" />
                <path d="m80.57 99.427-19.576 11.306v22.612l19.576-11.306 19.576 11.306v-22.612z" fill-opacity=".19608" />
                <path d="m80.57 99.426-19.576 11.306 39.152 22.612v-5.2e-4l-19.576-11.306 19.576-11.306z" fill-opacity=".46275" />
              </g>
              <path d="m80.57 99.426-1.1e-4 22.612-19.576-11.306z" fill="#000301" fill-opacity=".26275" />
            </g>
            <g transform="translate(93.919 -3.5547)" fill="#000202">
              <path d="m33.635 84.19 9.788-5.6529v-11.306l-9.788-5.6529-9.788 5.6529 9.788 5.6529z" fill-opacity=".59216" />
              <path d="m23.847 67.232 19.576 11.306v-11.306l-9.788-5.6529z" fill-opacity=".45098" />
              <path d="m33.635 61.579 9.788 5.6529-9.788 5.6529z" fill-opacity=".44314" />
            </g>
            <g transform="translate(71.486 -49.6)">
              <path d="m31.247 110.69-1.1e-4 22.612-19.576-11.306z" fill="#000201" fill-opacity=".19608" />
              <g fill="#000302">
                <path d="m31.247 155.87-19.576-11.306 9e-5 -22.612 19.576-11.306 9.7881 5.6891v11.306l9.788 5.6529v11.306z" fill-opacity=".60784"
                />
                <path d="m31.247 155.87 2e-5 -22.576-19.576 11.27 9e-5 -22.612 19.576-11.27 9.788 5.6529v11.306l9.788 5.6529 1e-5 11.27z"
                  fill-opacity=".25882" />
                <path d="m11.671 144.57 19.576-11.27 19.576 11.27-1e-5 -11.27-9.788-5.6529v-11.306l-9.788-5.6529-19.576 11.27z" fill-opacity=".19608"
                />
                <path d="m41.035 116.38-9.788 5.6529-1.3e-4 11.27-19.576-11.306 19.576-11.306z" fill-opacity=".47059" />
                <path d="m31.247 133.3 9.788-5.6529 9.788 5.6529-9.788 5.6529z" fill-opacity=".47059" />
              </g>
              <path d="m41.035 138.95-9.788-5.6529 9.788-5.6529z" fill="#000103" fill-opacity=".45098" />
            </g>
            <g fill="#000302">
              <path d="m60.994 110.73 19.576-11.306 19.576 11.306v22.612l-19.576 11.306-19.576-11.306 9e-5 -22.612" fill-opacity=".60784"
              />
              <path d="m80.57 99.426-19.576 11.306v22.612l19.576-11.306v22.612l19.576-11.306v-22.612z" fill-opacity=".25882" />
              <path d="m80.57 99.427-19.576 11.306v22.612l19.576-11.306 19.576 11.306v-22.612z" fill-opacity=".19608" />
              <path d="m80.57 99.426-19.576 11.306 39.152 22.612v-5.2e-4l-19.576-11.306 19.576-11.306z" fill-opacity=".46275" />
            </g>
            <path d="m80.57 99.426-1.1e-4 22.612-19.576-11.306z" fill="#000301" fill-opacity=".26275" />
            <text transform="scale(1.0173 .98298)" x="69.593071" y="168.1747" fill="#000300" fill-opacity=".7098" font-family="Poppins"
              font-size="18.628px" font-weight="300" letter-spacing="0px" stroke-width=".4657" word-spacing="0px" style="font-feature-settings:normal;font-variant-caps:normal;font-variant-ligatures:normal;font-variant-numeric:normal;line-height:1.25"
              xml:space="preserve">
              <tspan x="69.593071" y="168.1747" fill="#000300" fill-opacity=".7098" font-family="Poppins" font-size="18.628px" font-weight="300"
                stroke-width=".4657" style="font-feature-settings:normal;font-variant-caps:normal;font-variant-ligatures:normal;font-variant-numeric:normal">swarm</tspan>
            </text>
          </g>
        </svg>
      </a>
    </div>

    <div class="separate-block">
      <h2>{{.Msg}}</h2>
    </div>

    <div>
      <h2>Error code:</h2>
      <p class="orange">{{.Code}}</p>
    </div>

    <div class="footer">
      <p>Wed, 20 Jun 2018 15:39:37 UTC</p>
      <p>Swarm: Serverless Hosting Incentivised Peer-To-Peer Storage And Content Distribution</p>
    </div>
  </div>
{{ end }}`

const landing = `{{ define "content" }}

<script type="text/javascript">
function goToPage() {
  var page = document.getElementById('page').value;
  if (page == "") {
    var page = "theswarm.eth"
  }
  var address = "/bzz:/" + page;
  location.href = address;
  console.log(address)
}
</script>

<div class="container">
<div class="logo">
  <a href="/bzz:/theswarm.eth">
    <svg width="180" version="1.1" viewBox="0 0 83.477 107.43" xmlns="http://www.w3.org/2000/svg" xmlns:osb="http://www.openswatchbook.org/uri/2009/osb">
      <g display="none">
        <text transform="scale(1.0173 .98298)" x="9.6367435" y="109.16406" fill="#000300" fill-opacity=".7098" font-family="Poppins"
          font-size="18.628px" font-weight="300" letter-spacing="0px" stroke-width=".4657" word-spacing="0px" style="font-feature-settings:normal;font-variant-caps:normal;font-variant-ligatures:normal;font-variant-numeric:normal;line-height:1.25"
          xml:space="preserve">
          <tspan x="9.6367435" y="109.16406" fill="#000300" fill-opacity=".7098" font-family="Poppins" font-size="18.628px" font-weight="300"
            stroke-width=".4657" style="font-feature-settings:normal;font-variant-caps:normal;font-variant-ligatures:normal;font-variant-numeric:normal">swarm</tspan>
        </text>
        <g transform="translate(-60.059 -85.696)">
          <path d="m79.635 149.73 19.576-11.306v22.612z" fill="#000301" fill-opacity=".76863" />
          <path d="m60.059 161.03 19.576 11.306-1.11e-4 -22.611-19.576 11.306" fill="#000302" fill-opacity=".60784" />
          <path d="m79.635 127.12-1.11e-4 22.612-19.576-11.306z" fill="#000201" fill-opacity=".90588" />
          <path d="m79.635 127.12 19.576 11.306-19.576 11.306 1.11e-4 -22.612" fill="#000201" fill-opacity=".87451" />
          <path d="m60.059 138.42v22.612l19.576-11.306z" fill="#000301" fill-opacity=".76863" />
          <path d="m79.635 172.34 19.576-11.306-19.576-11.306 1.11e-4 22.611" fill="#000301" fill-opacity=".7098" />
        </g>
        <path d="m22.162 36.961 19.576 11.306-1.08e-4 -22.611-19.576 11.306" fill="#000302" fill-opacity=".60784" />
        <path d="m41.739 3.0432-1.08e-4 22.612-19.576-11.306z" fill="#000201" fill-opacity=".90588" />
        <g fill="#000301">
          <path d="m22.162 14.349v22.612l19.576-11.306z" fill-opacity=".76863" />
          <path d="m41.739 48.267 19.576-11.306-19.576-11.306 1.08e-4 22.611" fill-opacity=".7098" />
          <path d="m51.527 8.7322-9.788 5.6529v11.306l9.788-5.6529z" fill-opacity=".76471" />
        </g>
        <path d="m66.559 11.323 9.788 5.6529-9.788 5.6529z" fill="#000302" fill-opacity=".59216" />
        <path d="m41.739 3.0793 9.788 5.6529-9.788 5.6529z" fill="#000201" fill-opacity=".87451" />
        <path d="m76.347 16.976-9.788-5.6529 9.788-5.6529z" fill="#000301" fill-opacity=".76863" />
        <path d="m51.527 20.038 9.788 5.6529-9.788 5.6529z" fill="#000201" fill-opacity=".87451" />
        <path d="m61.315 36.997-9.788-5.6529 9.788-5.6529z" fill="#000301" fill-opacity=".76863" />
        <path d="m51.527 31.344-9.788-5.6529 9.788-5.6529z" fill="#000103" fill-opacity=".93725" />
        <path d="m66.559 0.0177 9.788 5.6529-9.788 5.6529z" fill="#000201" fill-opacity=".87451" />
        <path d="m66.559 11.323-9.788-5.6529 9.788-5.6529z" fill="#000104" fill-opacity=".78431" />
        <g transform="translate(-15.734 -85.696)">
          <path d="m79.635 149.73 19.576-11.306v22.612z" fill="#000301" fill-opacity=".76863" />
          <path d="m60.059 161.03 19.576 11.306-1.11e-4 -22.611-19.576 11.306" fill="#000302" fill-opacity=".60784" />
          <path d="m79.635 127.12-1.11e-4 22.612-19.576-11.306z" fill="#000201" fill-opacity=".90588" />
          <path d="m79.635 127.12 19.576 11.306-19.576 11.306 1.11e-4 -22.612" fill="#000201" fill-opacity=".87451" />
          <path d="m60.059 138.42v22.612l19.576-11.306z" fill="#000301" fill-opacity=".76863" />
          <path d="m79.635 172.34 19.576-11.306-19.576-11.306 1.11e-4 22.611" fill="#000301" fill-opacity=".7098" />
        </g>
      </g>
      <g transform="translate(-60.994 -58.006)">
        <g transform="translate(44.325 -.00024)">
          <g fill="#000302">
            <path d="m60.994 110.73 19.576-11.306 19.576 11.306v22.612l-19.576 11.306-19.576-11.306 9e-5 -22.612" fill-opacity=".60784"
            />
            <path d="m80.57 99.426-19.576 11.306v22.612l19.576-11.306v22.612l19.576-11.306v-22.612z" fill-opacity=".25882" />
            <path d="m80.57 99.427-19.576 11.306v22.612l19.576-11.306 19.576 11.306v-22.612z" fill-opacity=".19608" />
            <path d="m80.57 99.426-19.576 11.306 39.152 22.612v-5.2e-4l-19.576-11.306 19.576-11.306z" fill-opacity=".46275" />
          </g>
          <path d="m80.57 99.426-1.1e-4 22.612-19.576-11.306z" fill="#000301" fill-opacity=".26275" />
        </g>
        <g transform="translate(93.919 -3.5547)" fill="#000202">
          <path d="m33.635 84.19 9.788-5.6529v-11.306l-9.788-5.6529-9.788 5.6529 9.788 5.6529z" fill-opacity=".59216" />
          <path d="m23.847 67.232 19.576 11.306v-11.306l-9.788-5.6529z" fill-opacity=".45098" />
          <path d="m33.635 61.579 9.788 5.6529-9.788 5.6529z" fill-opacity=".44314" />
        </g>
        <g transform="translate(71.486 -49.6)">
          <path d="m31.247 110.69-1.1e-4 22.612-19.576-11.306z" fill="#000201" fill-opacity=".19608" />
          <g fill="#000302">
            <path d="m31.247 155.87-19.576-11.306 9e-5 -22.612 19.576-11.306 9.7881 5.6891v11.306l9.788 5.6529v11.306z" fill-opacity=".60784"
            />
            <path d="m31.247 155.87 2e-5 -22.576-19.576 11.27 9e-5 -22.612 19.576-11.27 9.788 5.6529v11.306l9.788 5.6529 1e-5 11.27z"
              fill-opacity=".25882" />
            <path d="m11.671 144.57 19.576-11.27 19.576 11.27-1e-5 -11.27-9.788-5.6529v-11.306l-9.788-5.6529-19.576 11.27z" fill-opacity=".19608"
            />
            <path d="m41.035 116.38-9.788 5.6529-1.3e-4 11.27-19.576-11.306 19.576-11.306z" fill-opacity=".47059" />
            <path d="m31.247 133.3 9.788-5.6529 9.788 5.6529-9.788 5.6529z" fill-opacity=".47059" />
          </g>
          <path d="m41.035 138.95-9.788-5.6529 9.788-5.6529z" fill="#000103" fill-opacity=".45098" />
        </g>
        <g fill="#000302">
          <path d="m60.994 110.73 19.576-11.306 19.576 11.306v22.612l-19.576 11.306-19.576-11.306 9e-5 -22.612" fill-opacity=".60784"
          />
          <path d="m80.57 99.426-19.576 11.306v22.612l19.576-11.306v22.612l19.576-11.306v-22.612z" fill-opacity=".25882" />
          <path d="m80.57 99.427-19.576 11.306v22.612l19.576-11.306 19.576 11.306v-22.612z" fill-opacity=".19608" />
          <path d="m80.57 99.426-19.576 11.306 39.152 22.612v-5.2e-4l-19.576-11.306 19.576-11.306z" fill-opacity=".46275" />
        </g>
        <path d="m80.57 99.426-1.1e-4 22.612-19.576-11.306z" fill="#000301" fill-opacity=".26275" />
        <text transform="scale(1.0173 .98298)" x="69.593071" y="168.1747" fill="#000300" fill-opacity=".7098" font-family="Poppins"
          font-size="18.628px" font-weight="300" letter-spacing="0px" stroke-width=".4657" word-spacing="0px" style="font-feature-settings:normal;font-variant-caps:normal;font-variant-ligatures:normal;font-variant-numeric:normal;line-height:1.25"
          xml:space="preserve">
          <tspan x="69.593071" y="168.1747" fill="#000300" fill-opacity=".7098" font-family="Poppins" font-size="18.628px" font-weight="300"
            stroke-width=".4657" style="font-feature-settings:normal;font-variant-caps:normal;font-variant-ligatures:normal;font-variant-numeric:normal">swarm</tspan>
        </text>
      </g>
    </svg>
  </a>
</div>

<div class="searchbar">
  <form class="separate-block" action="javascript:goToPage();">
    <input type="text" id="page" autofocus name="search" placeholder="Please enter an ENS name or swarm hash to retrieve ..">
    <button class="button" type="submit" value="submit" onclick="goToPage();">Go!</button>
  </form>
</div>
<div class="footer">
  <p>Swarm: Serverless Hosting Incentivised Peer-To-Peer Storage And Content Distribution</p>
</div>
</div>
  
{{ end }}`

const multipleChoice = `{{ define "content" }}
<content>

      <header>
        <div class="header-left">
          <img style="height:18vh;margin-left:40px" src="data:image/png;base64,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"/>
        </div>
        <div class="page-title">
          <h1>Swarm: disambiguation</h1>
        </div>
        <div class="header-right">
          <div id="timestamp">{{.Timestamp}}</div>
        </div>
      </header>

      <content-body>
        <section>
          <table>
            <thead>
              <td style="height: 150px; font-size: 1.3em; color: black; font-weight: bold">
                Your request may refer to {{ .Details}}.
              </td>
            </thead>
            <tbody>
              <tr>
                <td class="key">
                  Error code:
                </td>
              </tr>
              <tr>
                <td class="value">
                  {{.Code}}
                </td>
              </tr>
            </tbody>
          </table>
        </section>
      </content-body>
    </content>
{{ end }}`

const baseTemplate = `<html>
<head>
  <meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0"/>
  <meta http-equiv="X-UA-Compatible" ww="chrome=1"/>
  <link rel="icon" type="image/x-icon" href="favicon.ico"/>
  <style>
  {{ template "css" . }}
  </style>
</head>
<body>
  {{ template "content" . }}
</body>
</html>
`

const css = `{{ define "css" }} 
html {
  font-size: 18px;
  font-size: 1.13rem;
  -webkit-text-size-adjust: 100%;
  -ms-text-size-adjust: 100%;
  font-family: Helvetica, Arial, sans-serif;
  margin: 0;
  padding: 0;
}

body {
  background: #f6f6f6;
  margin: auto;
  color: #333;
}

a, a:visited, a:active {
  color: #fff;
  text-decoration: none;
}

h1 {
  margin: 0;
}

h2 {
  font-size: 20px;
  font-size: 1.25rem;
  font-weight: 400;
}

.container {
  max-width: 600px;
  margin: 40px auto 40px;
  text-align: center;
}

.separate-block {
  margin: 40px 0;
  word-wrap: break-word;
}

.footer {
  font-size: 12px;
  font-size: 0.75rem;
  text-align: center;
}

.orange {
  color: #ffa500;
}

/* SVG Logos, editable */

.searchbar {
  padding: 20px 20px 0;
}

.logo {
  margin: 100px 80px 0;
}

/* Tablet < 600p*/

@media only screen and (max-width: 600px) {}

/* Mobile phone < 360p*/

@media only screen and (max-width: 360px) {
  h1 {
      font-size: 20px;
      font-size: 1.5rem;
  }
  h2 {
      font-size: 0.88rem;
      margin: 0;
  }
  .logo {
      margin: 50px 40px 0;
  }
  .footer {
      font-size: 0.63rem;
      text-align: center;
  }
}

input[type=text] {
  width: 100%;
  box-sizing: border-box;
  border: 2px solid #777;
  border-radius: 2px;
  font-size: 16px;
  padding: 12px 20px 12px 20px;
  transition: border 250ms ease-in-out;
}

input[type=text]:focus {
  border: 2px solid #ffce73;
}

.button {
  background-color: #ffa500;
  margin: 20px 0;
  border: none;
  border-radius: 2px;
  color: #222;
  padding: 15px 32px;
  text-align: center;
  text-decoration: none;
  display: inline-block;
  font-size: 16px;
}
{{ end }}`

const faviconBase64 = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAQAAAC1+jfqAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAAAAmJLR0QA/4ePzL8AAAAHdElNRQfiBwUEAiAc99adAAABd0lEQVQoz32RzUsbYRDGf++7m13U3WZjTCJ+UWgEDaKCYMAWC4KhfvVUQRB6aFTQkxfxJl7Ei4iICP4D/g3iRSwVeioWVFRs60FQQVARikbN9NDdWkF85jLMzPPMMwyinor3oSDT/Ie0Dg8lRjJmU9vhVvxzVR0AD6xIu/vdkXJJLbUWNU7VSnr6kUJTs7tKPZTjFK7/7p1L4x0B/BswrVJe4XDFj+7S0aN86M3KIoD54EBweMsvzj1v9tvO11WfGLStXJQSwMPjhvtcUPdXVL3Yq4hkZBNAtgo6d53qeCCseozwoHMclfh2ZVc2mxlOvHY3whI/T4zV26KUqOSX0xaFhQkUz+QPT+bz2sBGcPcPakyIJV1OuMUhgn4pd64+4wKLKHaFbzJEihCX3Pu2ykhxSQ75e4WtE7hcINxwjaFLiJHjDptrDRqq+5xNAQRrpXiyYSG2TB7A+Ol+9H8xYXzKfth41xF8ZSDdv9Y+3lcgSpQSxbP4AzZRgGtq5JjRAAAAJXRFWHRkYXRlOmNyZWF0ZQAyMDE4LTA3LTA1VDA0OjAyOjMyLTA3OjAwZY9YVwAAACV0RVh0ZGF0ZTptb2RpZnkAMjAxOC0wNy0wNVQwNDowMjozMi0wNzowMBTS4OsAAAAASUVORK5CYII="

const favicon = `000001000400101000000000200068040000460000002020000000002000a8100000ae0400003030000000002000a825000056150000404000000000200028420000fe3a000028000000100000002000000001002000000000004004000000000000000000000000000000000000ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017e7e7e0362626263545454c548484849ffffff01ffffff01ffffff01ffffff01646464375b5b5bbf4545457758585809ffffff01ffffff01ffffff0164646443626262cf626262ff535353ff454545ff454545b74949492b6868681d626262a5626262fd5c5c5cff464646ff454545dd47474755ffffff01ffffff013f3f3feb565656ff636363ff535353ff464646ff3f3f3fff373737ab393939894d4d4dff626262ff5c5c5cff464646ff424242ff3a3a3af7ffffff01ffffff01383838e9353535ff424242ff474747ff383838ff353535ff363636ab35353587363636ff3a3a3aff4a4a4aff3b3b3bff353535ff363636f5ffffff01ffffff01383838e9303030ff181818ff131313ff232323ff343434ff363636ab35353587343434ff202020ff101010ff1d1d1dff303030ff373737f5ffffff01ffffff01232323c50c0c0cff0d0d0dff131313ff171717ff171717ff2929298b2727276b0f0f0ffd0d0d0dff101010ff171717ff161616ff232323d9ffffff01ffffff014d4d4d030f0f0f650c0c0ce7131313ff161616d51d1d1d4b63636363464646691717173b0d0d0dc50f0f0fff161616ef171717752e2e2e07ffffff01ffffff01ffffff01ffffff011d1d1d0f1515155360606045626262cf636363ff464646ff454545d3484848491414144d24242417ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013c3c3c374f4f4fff636363ff636363ff464646ff464646ff3f3f3fff3c3c3c41ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013636363d353535ff3c3c3cff575757ff363636ff181818ff282828ff37373747ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013636363d363636ff303030ff181818ff292929ff131313ef17171771696969136565653bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01323232371e1e1eff0d0d0dff0c0c0cff363636ff363636a3ffffff0185858515606060ff4747476bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01111111450d0d0dd10c0c0cff1b1b1bff2a2a2a993e3e3e0b30303085292929ff37373787ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01363636030e0e0e671616166b45454505323232432e2e2ed9151515c31d1d1d2dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff014e4e4e05ffffff01ffffff01ffffff01ffffff010000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff0000ffff28000000200000004000000001002000000000008010000000000000000000000000000000000000ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017272721b646464a54646466f72727205ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0168686845575757b74f4f4f39ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017e7e7e0b6262627d616161f3636363ff424242ff444444d74f4f4f49ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff016c6c6c27636363b5616161ff555555ff434343ff464646a35858581dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff016666665d616161e3626262ff636363ff636363ff444444ff464646ff434343ff454545b95252522bffffff01ffffff01ffffff01ffffff016c6c6c1363636393616161fb636363ff636363ff555555ff464646ff464646ff444444f5464646836666660bffffff01ffffff01ffffff01ffffff01ffffff016a6a6a3f626262c9616161ff636363ff636363ff636363ff636363ff444444ff464646ff464646ff464646ff434343fb48484897545454135b5b5b036868686f616161ef626262ff636363ff636363ff636363ff555555ff464646ff464646ff464646ff454545ff444444e54a4a4a5fffffff01ffffff01ffffff01ffffff013b3b3bd7505050ff646464ff636363ff636363ff636363ff636363ff444444ff464646ff464646ff464646ff454545ff3a3a3aff33333357313131113c3c3cff5a5a5aff646464ff636363ff636363ff636363ff555555ff464646ff464646ff464646ff464646ff424242ff383838f1ffffff01ffffff01ffffff01ffffff013a3a3ad5353535ff3a3a3aff575757ff646464ff626262ff636363ff444444ff464646ff464646ff3d3d3dff353535ff363636ff3636365535353511363636ff343434ff434343ff606060ff636363ff636363ff555555ff464646ff464646ff444444ff393939ff353535ff373737edffffff01ffffff01ffffff01ffffff013a3a3ad5363636ff363636ff343434ff3f3f3fff5d5d5dff646464ff444444ff404040ff363636ff353535ff363636ff363636ff3636365535353511363636ff363636ff363636ff343434ff4a4a4aff636363ff555555ff454545ff3c3c3cff353535ff363636ff363636ff373737edffffff01ffffff01ffffff01ffffff013a3a3ad5363636ff363636ff363636ff363636ff353535ff3f3f3fff363636ff353535ff363636ff363636ff363636ff363636ff3636365535353511363636ff363636ff363636ff363636ff353535ff383838ff3a3a3aff373737ff353535ff363636ff363636ff363636ff373737edffffff01ffffff01ffffff01ffffff013a3a3ad5363636ff363636ff363636ff323232ff181818ff0e0e0eff171717ff282828ff373737ff363636ff363636ff363636ff3636365535353511363636ff363636ff353535ff373737ff292929ff0f0f0fff111111ff1b1b1bff2f2f2fff373737ff363636ff363636ff373737edffffff01ffffff01ffffff01ffffff013a3a3ad5363636ff363636ff1e1e1eff0b0b0bff0d0d0dff0f0f0fff171717ff161616ff191919ff2c2c2cff373737ff363636ff3636365535353511363636ff373737ff2f2f2fff141414ff0b0b0bff0d0d0dff131313ff171717ff151515ff1f1f1fff333333ff363636ff373737edffffff01ffffff01ffffff01ffffff013b3b3bd5252525ff0d0d0dff0c0c0cff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff151515ff1c1c1cff313131ff3535355734343411333333ff1a1a1aff0b0b0bff0d0d0dff0d0d0dff0d0d0dff131313ff171717ff171717ff171717ff161616ff242424ff373737efffffff01ffffff01ffffff01ffffff012020205d0b0b0be50b0b0bff0d0d0dff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff171717ff131313ff161616b73333331f3b3b3b05111111970a0a0afb0d0d0dff0d0d0dff0d0d0dff0d0d0dff131313ff171717ff171717ff171717ff161616ff141414f51c1c1c7fffffff01ffffff01ffffff01ffffff01ffffff014d4d4d0b1212127f0a0a0af50d0d0dff0d0d0dff0f0f0fff171717ff171717ff151515ff151515d522222249ffffff017373731b51515121ffffff011d1d1d2b101010b50a0a0aff0d0d0dff0d0d0dff131313ff171717ff171717ff131313ff181818a12e2e2e1dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff012c2c2c1b0f0f0fa10a0a0afd0f0f0fff161616ff141414e91b1b1b69656565057878780b6363637b626262f3464646f7454545896969690fffffff011c1c1c470c0c0cd30b0b0bff131313ff141414ff151515c32a2a2a37ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff011d1d1d35111111bd1a1a1a8d2f2f2f11ffffff0166666659616161e1626262ff646464ff474747ff454545ff444444e9494949677b7b7b054040400517171769131313cd24242455ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0169696939626262c7616161ff636363ff636363ff646464ff474747ff464646ff464646ff444444ff454545d14e4e4e45ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01424242615e5e5eff636363ff636363ff636363ff636363ff646464ff474747ff464646ff464646ff464646ff464646ff434343ff3f3f3f77ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0136363679343434ff494949ff636363ff636363ff636363ff646464ff474747ff464646ff464646ff474747ff3d3d3dff353535ff3a3a3a8dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0136363679363636ff353535ff363636ff505050ff646464ff636363ff474747ff484848ff2f2f2fff1c1c1cff323232ff363636ff3a3a3a8dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0136363679363636ff363636ff363636ff353535ff3a3a3aff5a5a5aff393939ff0f0f0fff040404ff111111ff151515ff232323ff3535358fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0136363679363636ff363636ff363636ff363636ff323232ff171717ff2a2a2aff0c0c0cff030303ff111111ff141414fb171717992e2e2e17a3a3a305ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0136363679363636ff363636ff363636ff1f1f1fff0b0b0bff0d0d0dff363636ff383838ff242424ff121212bf2a2a2a2dffffff01ffffff018484842b636363bf6d6d6d2fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0136363679373737ff252525ff0d0d0dff0c0c0cff0d0d0dff0d0d0dff373737ff363636ff353535ff39393949ffffff01ffffff01ffffff0186868629646464ff656565fb6464649b55555505ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff012e2e2e650e0e0eff0c0c0cff0d0d0dff0d0d0dff0d0d0dff0c0c0cff353535ff363636ff353535ff37373749ffffff01ffffff01ffffff0185858529656565ff525252ff353535ff4b4b4b0fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff011c1c1c430d0d0dcf0b0b0bff0d0d0dff0d0d0dff0d0d0dff171717ff282828ff363636ff37373749ffffff01ffffff01ffffff0144444459363636ff353535ff353535ff4e4e4e0fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0162626203161616630b0b0be70c0c0cff0d0d0dff171717ff161616ff171717ed3737372fffffff013e3e3e2b303030b72a2a2aff151515ff262626ff363636ff4b4b4b0fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013636360d101010850a0a0af7141414f91717178f45454511ffffff014c4c4c252c2c2cdb303030ff2d2d2dff151515ff131313ff1b1b1bad5a5a5a07ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff012b2b2b2121212127ffffff01ffffff01ffffff01ffffff0161616109313131752b2b2bf1131313cd26262641ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff014e4e4e1359595903ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000028000000300000006000000001002000000000008025000000000000000000000000000000000000ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0173737357545454997c7c7c11ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0176767663515151916c6c6c0dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017676762d636363bb636363ff4d4d4dff434343eb4f4f4f6d7f7f7f05ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0176767635616161c3626262ff494949ff424242e94f4f4f6392929203ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017e7e7e19626262955f5f5ffd626262ff666666ff4f4f4fff464646ff424242ff434343d75a5a5a49ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017777771d6464649f5f5f5fff636363ff656565ff4b4b4bff464646ff424242ff444444d158585841ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff018585850966666677606060ef626262ff636363ff636363ff666666ff4f4f4fff464646ff464646ff464646ff414141ff464646b75d5d5d2dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff018989890d6868687f5f5f5ff5626262ff636363ff636363ff656565ff4b4b4bff464646ff464646ff464646ff404040ff484848b160606027ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff016a6a6a55626262df606060ff636363ff636363ff636363ff636363ff666666ff4f4f4fff464646ff464646ff464646ff464646ff454545ff424242fd484848956a6a6a17ffffff01ffffff01ffffff01ffffff01ffffff016969695f606060e3606060ff636363ff636363ff636363ff636363ff656565ff4b4b4bff464646ff464646ff464646ff464646ff454545ff414141f94a4a4a8d65656513ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff016e6e6e3b656565c15f5f5fff636363ff636363ff636363ff636363ff636363ff636363ff666666ff4f4f4fff464646ff464646ff464646ff464646ff464646ff464646ff444444ff424242ed52525277ffffff01ffffff016c6c6c37676767c95f5f5fff636363ff636363ff636363ff636363ff636363ff636363ff656565ff4b4b4bff464646ff464646ff464646ff464646ff464646ff464646ff434343ff444444e94d4d4d6dffffff01ffffff01ffffff01ffffff01ffffff01ffffff013c3c3cc5454545ff646464ff646464ff636363ff636363ff636363ff636363ff636363ff666666ff4f4f4fff464646ff464646ff464646ff464646ff464646ff464646ff474747ff424242ff333333fb34343409ffffff0131313199494949ff656565ff646464ff636363ff636363ff636363ff636363ff636363ff656565ff4b4b4bff464646ff464646ff464646ff464646ff464646ff464646ff474747ff414141ff373737ebffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040bf333333ff343434ff4f4f4fff666666ff636363ff636363ff636363ff636363ff666666ff4f4f4fff464646ff464646ff464646ff464646ff474747ff444444ff383838ff343434ff363636f737373707ffffff0135353597343434ff343434ff525252ff666666ff636363ff636363ff636363ff636363ff656565ff4b4b4bff464646ff464646ff464646ff464646ff474747ff444444ff383838ff343434ff393939e3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040bf343434ff363636ff333333ff383838ff585858ff676767ff636363ff636363ff666666ff4f4f4fff464646ff464646ff474747ff464646ff3b3b3bff343434ff363636ff363636ff363636f737373707ffffff0135353597363636ff363636ff333333ff383838ff5a5a5aff666666ff636363ff636363ff656565ff4b4b4bff464646ff464646ff474747ff454545ff3a3a3aff343434ff363636ff363636ff393939e3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040bf343434ff363636ff363636ff363636ff323232ff3d3d3dff5d5d5dff666666ff666666ff4f4f4fff464646ff474747ff3e3e3eff353535ff353535ff363636ff363636ff363636ff363636f737373707ffffff0135353597363636ff363636ff363636ff363636ff313131ff3f3f3fff5f5f5fff666666ff656565ff4b4b4bff464646ff474747ff3d3d3dff353535ff353535ff363636ff363636ff363636ff393939e3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040bf343434ff363636ff363636ff363636ff363636ff353535ff323232ff444444ff676767ff525252ff404040ff363636ff353535ff363636ff363636ff363636ff363636ff363636ff363636f737373707ffffff0135353597363636ff363636ff363636ff363636ff363636ff353535ff323232ff464646ff676767ff4e4e4eff404040ff363636ff353535ff363636ff363636ff363636ff363636ff363636ff393939e3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040bf343434ff363636ff363636ff363636ff363636ff363636ff353535ff383838ff2d2d2dff2b2b2bff373737ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636f737373707ffffff0135353597363636ff363636ff363636ff363636ff363636ff363636ff363636ff383838ff2c2c2cff2a2a2aff373737ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff393939e3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040bf343434ff363636ff363636ff363636ff353535ff383838ff343434ff171717ff090909ff151515ff171717ff2d2d2dff383838ff363636ff363636ff363636ff363636ff363636ff363636f737373707ffffff0135353597363636ff363636ff363636ff363636ff353535ff383838ff333333ff151515ff090909ff151515ff181818ff2f2f2fff383838ff363636ff363636ff363636ff363636ff363636ff393939e3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040bf343434ff363636ff363636ff373737ff373737ff1f1f1fff090909ff0c0c0cff0c0c0cff171717ff171717ff141414ff1b1b1bff323232ff383838ff363636ff363636ff363636ff363636f737373707ffffff0135353597363636ff363636ff363636ff373737ff373737ff1d1d1dff0a0a0aff0c0c0cff0c0c0cff171717ff171717ff141414ff1c1c1cff333333ff383838ff353535ff363636ff363636ff393939e3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040bf343434ff363636ff393939ff272727ff0c0c0cff0b0b0bff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff161616ff141414ff202020ff353535ff373737ff363636ff363636f737373707ffffff0135353597363636ff363636ff383838ff252525ff0b0b0bff0b0b0bff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff161616ff141414ff222222ff363636ff373737ff363636ff393939e3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040bf383838ff2d2d2dff101010ff0a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff171717ff171717ff161616ff141414ff262626ff373737ff373737f737373707ffffff0136363697393939ff2b2b2bff0f0f0fff0a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff171717ff171717ff161616ff151515ff272727ff383838ff393939e3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013a3a3abd131313ff090909ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff171717ff171717ff171717ff171717ff151515ff171717ff262626fb38383807ffffff012a2a2a97121212ff090909ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff171717ff171717ff171717ff171717ff151515ff161616ff2a2a2ae7ffffff01ffffff01ffffff01ffffff01ffffff01ffffff015f5f5f0b1616167b090909ef0a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff171717ff171717ff171717ff171717ff0f0f0fff181818b74040402dffffff01ffffff014646461118181883080808f30b0b0bff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff171717ff171717ff171717ff161616ff101010ff181818b141414127ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff014d4d4d171212129b090909fd0c0c0cff0d0d0dff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff171717ff171717ff111111ff141414d335353547ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013838381d131313a5060606ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff171717ff171717ff111111ff181818cd2e2e2e3dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01333333310f0f0fbb070707ff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff141414ff121212e72424246d86868603ffffff01ffffff017373732b656565b9464646c95e5e5e3bffffff01ffffff01ffffff01323232370e0e0ec3080808ff0d0d0dff0d0d0dff0c0c0cff171717ff171717ff171717ff121212ff161616e525252563ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff012525254d0e0e0ed9090909ff0c0c0cff171717ff151515ff121212f91d1d1d894d4d4d13ffffff01ffffff0178787815656565935f5f5ffb646464ff484848ff404040ff454545a96a6a6a1fffffff01ffffff01ffffff011b1b1b570e0e0edf080808ff0d0d0dff171717ff151515ff0f0f0ff3212121815656560dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01636363071a1a1a710a0a0aed0f0f0fff1b1b1bad2f2f2f23ffffff01ffffff018d8d8d0566666675616161eb616161ff636363ff646464ff484848ff464646ff454545ff424242f54c4c4c856262620fffffff01ffffff014040400b21212179080808f10f0f0fff1b1b1ba15757571dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff014141411740404037ffffff01ffffff01ffffff016a6a6a4d616161db606060ff636363ff636363ff636363ff646464ff484848ff464646ff464646ff464646ff434343ff434343e751515167ffffff01ffffff01ffffff014646461d30303033ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0176767631616161c35f5f5fff636363ff636363ff636363ff636363ff636363ff646464ff484848ff464646ff464646ff464646ff464646ff464646ff424242ff454545d158585841ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff015252527f636363ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff646464ff484848ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff434343ff454545a1ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01313131b53b3b3bff5b5b5bff676767ff636363ff636363ff636363ff636363ff636363ff646464ff484848ff464646ff464646ff464646ff464646ff464646ff474747ff444444ff393939ff383838d3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01363636b3363636ff323232ff404040ff616161ff656565ff626262ff636363ff636363ff646464ff484848ff464646ff464646ff454545ff494949ff474747ff3b3b3bff343434ff353535ff3a3a3ad3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01363636b3363636ff363636ff353535ff323232ff484848ff656565ff646464ff636363ff646464ff484848ff464646ff474747ff494949ff242424ff282828ff383838ff363636ff363636ff3a3a3ad3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01363636b3363636ff363636ff363636ff363636ff343434ff343434ff515151ff666666ff656565ff484848ff4b4b4bff323232ff070707ff040404ff151515ff181818ff2f2f2fff383838ff3a3a3ad3ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01363636b3363636ff363636ff363636ff363636ff363636ff363636ff333333ff383838ff5f5f5fff3c3c3cff0f0f0fff020202ff050505ff050505ff171717ff171717ff141414ff1c1c1cff323232d7ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01363636b3363636ff363636ff363636ff363636ff363636ff353535ff383838ff343434ff161616ff2a2a2aff0c0c0cff020202ff050505ff050505ff171717ff171717ff101010ff161616bf2e2e2e35ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01363636b3363636ff363636ff363636ff363636ff373737ff383838ff1f1f1fff0a0a0aff0c0c0cff373737ff3a3a3aff262626ff060606ff040404ff121212ff151515dd30303051ffffff01ffffff01ffffff018787872d6b6b6b47ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01363636b3363636ff363636ff363636ff393939ff272727ff0d0d0dff0b0b0bff0d0d0dff0d0d0dff373737ff363636ff373737ff383838ff1c1c1cf92020207568686807ffffff01ffffff01ffffff01ffffff018686863d5f5f5fff676767af77777721ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01363636b3363636ff393939ff2e2e2eff101010ff0a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff373737ff363636ff363636ff353535ff373737ebffffff01ffffff01ffffff01ffffff01ffffff01ffffff018686863d626262ff666666ff646464f76969698d9494940fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01383838b5333333ff161616ff090909ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff373737ff363636ff363636ff363636ff353535ebffffff01ffffff01ffffff01ffffff01ffffff01ffffff018686863d626262ff676767ff6b6b6bff555555ff3a3a3a93ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0125252589030303ff0c0c0cff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff333333ff383838ff353535ff363636ff353535ebffffff01ffffff01ffffff01ffffff01ffffff01ffffff018585853d666666ff5f5f5fff3c3c3cff313131ff3a3a3a93ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff012d2d2d3f0e0e0ecb080808ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff141414ff222222ff363636ff373737ff353535ebffffff01ffffff01ffffff01ffffff01ffffff01ffffff0177777741414141ff313131ff363636ff353535ff3a3a3a93ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff011e1e1e5f0a0a0ae50a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff171717ff161616ff151515ff282828ff353535f3ffffff01ffffff01ffffff01ffffff016e6e6e0b37373781242424f1191919ff333333ff383838ff343434ff3a3a3a93ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff015a5a5a0d1919197f0a0a0af30b0b0bff0d0d0dff0d0d0dff171717ff171717ff161616ff0f0f0ffb24242489ffffff01ffffff01ffffff013e3e3e5d2d2d2de52e2e2eff2b2b2bff151515ff141414ff212121ff363636ff3b3b3b95ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013636361b111111a3080808ff0c0c0cff181818ff0f0f0fff171717b545454525ffffff01ffffff017f7f7f05363636c7282828ff313131ff313131ff2b2b2bff151515ff171717ff161616ff0c0c0cfb3434346bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01303030350f0f0fc7121212d337373741ffffff01ffffff01ffffff01ffffff01ffffff016b6b6b0b3a3a3a7d2c2c2cf12f2f2fff2b2b2bff151515ff101010ff171717bb4646462dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01515151193535359b242424ff131313d72828284bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff014e4e4e2b59595905ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff000000000000ffff28000000400000008000000001002000000000000042000000000000000000000000000000000000ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0176767635666666914e4e4e457c7c7c09ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff018080801569696989545454696c6c6c0bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff018484840d70707061616161d5606060fb3d3d3ddf4e4e4e9172727213ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017070704d626262b35f5f5ffb464646f1454545a16a6a6a33ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017676760f67676753646464cf5e5e5eff656565ff626262ff414141ff404040ff444444e54b4b4b7b69696919ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01979797036c6c6c45676767a95d5d5dff616161ff626262ff484848ff424242ff3e3e3efd4e4e4e8958585831ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017e7e7e2b616161a75f5f5fef616161ff636363ff656565ff626262ff424242ff464646ff444444ff414141fd434343b961616153ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017777771969696981606060e7606060ff636363ff636363ff626262ff484848ff464646ff454545ff424242fd414141d95656566569696911ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01858585056e6e6e29656565995f5f5ff1616161ff636363ff636363ff636363ff656565ff626262ff424242ff464646ff464646ff464646ff444444ff3f3f3fff484848af5353534b86868607ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01797979216a6a6a6f616161ed5e5e5eff636363ff636363ff636363ff636363ff626262ff484848ff464646ff464646ff464646ff464646ff3e3e3eff474747d75151515762626213ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01838383036f6f6f755f5f5fd3606060ff626262ff636363ff636363ff636363ff636363ff656565ff626262ff424242ff464646ff464646ff464646ff464646ff454545ff434343ff404040e94e4e4e8d5f5f5f1bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff018f8f8f056b6b6b45616161c95f5f5ff7616161ff636363ff636363ff636363ff636363ff636363ff626262ff484848ff464646ff464646ff464646ff464646ff464646ff444444ff424242f1434343b16666662dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017070700f6969695f626262d35e5e5eff626262ff636363ff636363ff636363ff636363ff636363ff636363ff656565ff626262ff424242ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff404040ff444444f14d4d4d776a6a6a23ffffff01ffffff01ffffff01ffffff017b7b7b096c6c6c39636363c15f5f5ffb626262ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff626262ff484848ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff434343ff414141f54a4a4aa35b5b5b2d70707007ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0171717143676767a7616161f3616161ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff656565ff626262ff424242ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff444444ff414141f7474747cd54545447ffffff01ffffff015b5b5b096b6b6b99646464e1606060ff626262ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff626262ff484848ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff444444ff424242ff414141d552525277ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01404040b33b3b3bff5c5c5cff656565ff646464ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff656565ff626262ff424242ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff474747ff454545ff3a3a3aff313131ad34343407ffffff012e2e2e25383838ff535353ff656565ff656565ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff626262ff484848ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff474747ff464646ff3b3b3bff3a3a3ae9ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9313131ff363636ff484848ff636363ff676767ff636363ff636363ff636363ff636363ff636363ff636363ff656565ff626262ff424242ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff404040ff363636ff343434ff353535a537373705ffffff0135353521333333ff333333ff434343ff5c5c5cff686868ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff626262ff484848ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff484848ff414141ff393939ff313131ff3c3c3cdbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff363636ff323232ff353535ff4b4b4bff636363ff656565ff636363ff626262ff636363ff636363ff656565ff626262ff424242ff464646ff464646ff464646ff464646ff474747ff464646ff414141ff363636ff343434ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff333333ff313131ff484848ff5e5e5eff666666ff646464ff626262ff636363ff636363ff636363ff626262ff484848ff464646ff464646ff464646ff464646ff464646ff474747ff424242ff3a3a3aff343434ff353535ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff363636ff363636ff343434ff333333ff3d3d3dff555555ff686868ff656565ff626262ff636363ff656565ff626262ff424242ff464646ff464646ff464646ff484848ff444444ff393939ff353535ff353535ff363636ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff363636ff353535ff323232ff363636ff515151ff646464ff656565ff636363ff636363ff636363ff626262ff484848ff464646ff464646ff464646ff484848ff454545ff3d3d3dff353535ff343434ff363636ff363636ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff363636ff363636ff363636ff363636ff343434ff303030ff3f3f3fff575757ff666666ff656565ff646464ff626262ff424242ff464646ff474747ff454545ff3a3a3aff343434ff353535ff363636ff363636ff363636ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff363636ff363636ff363636ff363636ff303030ff373737ff535353ff636363ff656565ff636363ff626262ff484848ff464646ff474747ff454545ff3e3e3eff353535ff343434ff363636ff363636ff363636ff363636ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff363636ff363636ff363636ff363636ff363636ff363636ff333333ff333333ff484848ff606060ff696969ff626262ff434343ff474747ff3e3e3eff363636ff353535ff353535ff363636ff363636ff363636ff363636ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff363636ff363636ff363636ff363636ff353535ff343434ff333333ff3e3e3eff5d5d5dff686868ff626262ff484848ff474747ff424242ff373737ff353535ff353535ff363636ff363636ff363636ff363636ff363636ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff323232ff323232ff505050ff616161ff3d3d3dff373737ff343434ff353535ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff343434ff313131ff434343ff606060ff464646ff383838ff343434ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff3a3a3aff2b2b2bff1e1e1eff2d2d2dff383838ff373737ff353535ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff393939ff323232ff1c1c1cff262626ff373737ff383838ff353535ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff363636ff363636ff363636ff363636ff363636ff353535ff373737ff383838ff303030ff191919ff080808ff101010ff141414ff1a1a1aff303030ff383838ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff363636ff363636ff363636ff363636ff353535ff363636ff383838ff363636ff1d1d1dff0b0b0bff0c0c0cff141414ff181818ff292929ff373737ff373737ff363636ff363636ff363636ff363636ff363636ff363636ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff363636ff363636ff363636ff363636ff353535ff393939ff363636ff222222ff0c0c0cff0a0a0aff0c0c0cff121212ff171717ff151515ff161616ff212121ff353535ff393939ff363636ff363636ff363636ff363636ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff363636ff363636ff363636ff353535ff383838ff3a3a3aff262626ff121212ff0a0a0aff0c0c0cff0f0f0fff171717ff151515ff151515ff1e1e1eff2f2f2fff3a3a3aff363636ff363636ff363636ff363636ff363636ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff363636ff363636ff363636ff383838ff363636ff262626ff0d0d0dff090909ff0d0d0dff0d0d0dff0c0c0cff121212ff171717ff171717ff171717ff141414ff151515ff232323ff353535ff383838ff363636ff353535ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff363636ff353535ff383838ff383838ff292929ff131313ff080808ff0c0c0cff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff151515ff131313ff202020ff313131ff383838ff363636ff363636ff363636ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9343434ff353535ff363636ff3a3a3aff2e2e2eff131313ff0a0a0aff0b0b0bff0d0d0dff0d0d0dff0d0d0dff0c0c0cff121212ff171717ff171717ff171717ff171717ff161616ff141414ff1a1a1aff2a2a2aff393939ff373737ff363636ff363636ff363636a537373705ffffff0135353521363636ff363636ff363636ff3a3a3aff313131ff1c1c1cff0a0a0aff0a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff171717ff161616ff151515ff161616ff282828ff363636ff383838ff363636ff333333ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444a9353535ff383838ff313131ff151515ff080808ff0b0b0bff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff121212ff171717ff171717ff171717ff171717ff171717ff171717ff161616ff131313ff1b1b1bff2d2d2dff373737ff373737ff363636a537373705ffffff0134343421363636ff383838ff333333ff1e1e1eff090909ff0a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff171717ff171717ff171717ff171717ff131313ff171717ff2a2a2aff363636ff353535ff3d3d3ddbffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01444444af353535ff1e1e1eff0d0d0dff0a0a0aff0c0c0cff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff121212ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff151515ff151515ff222222ff333333ff353535ad30303007ffffff0134343423373737ff282828ff0d0d0dff0a0a0aff0c0c0cff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff151515ff141414ff1b1b1bff2e2e2eff3e3e3ee1ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013e3e3e6f0f0f0fd5040404ff0b0b0bff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff121212ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff101010ff0e0e0ee72f2f2f7347474703ffffff013b3b3b13141414cd050505f70a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff121212ff0c0c0cf12a2a2aa5ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff015f5f5f052020202b1a1a1aa1080808f1070707ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff121212ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff141414ff0c0c0cff212121af2a2a2a496d6d6d07ffffff01ffffff01ffffff01333333231d1d1d730b0b0beb060606ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff171717ff171717ff171717ff171717ff171717ff151515ff0e0e0eff181818d72626265546464615ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff014d4d4d29121212af080808ef0a0a0aff0c0c0cff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff121212ff171717ff171717ff171717ff171717ff171717ff171717ff141414ff121212f9141414b93b3b3b4fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0138383819151515890a0a0ae5080808ff0c0c0cff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff171717ff171717ff171717ff161616ff101010fb151515d72c2c2c614444440dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0133333311262626510f0f0fd7050505ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff121212ff171717ff171717ff171717ff171717ff171717ff101010ff141414e7242424733a3a3a19ffffff01ffffff01ffffff01878787097272725f4d4d4d736a6a6a11ffffff01ffffff01ffffff016060600524242445191919ad040404ff0a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff171717ff171717ff171717ff111111ff0e0e0efd242424873232322dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff015c5c5c0d2525255f090909d7080808fb0b0b0bff0d0d0dff0c0c0cff121212ff171717ff171717ff161616ff121212ff121212df2121218965656511ffffff01ffffff01ffffff018080800d6767674b646464d1606060ff454545ff464646df4f4f4f6165656517ffffff01ffffff01ffffff01ffffff012d2d2d4b101010b5060606fb0a0a0aff0d0d0dff0d0d0dff0f0f0fff171717ff171717ff161616ff131313ff101010ef2020209d4242422dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff012c2c2c2d1f1f1f83080808fb080808ff0d0d0dff121212ff171717ff141414ff0f0f0ff91e1e1eb12c2c2c354d4d4d09ffffff01ffffff01ffffff0178787825646464a75f5f5feb616161ff656565ff4a4a4aff414141ff424242f3414141bd69696937ffffff01ffffff01ffffff01ffffff0142424219171717710d0d0de3060606ff0c0c0cff0f0f0fff171717ff151515ff0d0d0dff171717c3292929575656560dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013737372d1212129d080808ef0d0d0dff121212f5191919bf2e2e2e3d70707003ffffff01ffffff018c8c8c037676762564646497606060ed606060ff636363ff636363ff656565ff4a4a4aff444444ff464646ff444444ff404040f74a4a4aad5555553162626207ffffff01ffffff01ffffff014040401125252589090909dd0a0a0aff121212ff141414c738383869ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff015b5b5b0b1f1f1f591d1d1daf292929673f3f3f19ffffff01ffffff01ffffff01ffffff016d6d6d715f5f5fcd606060ff626262ff636363ff636363ff636363ff656565ff4a4a4aff444444ff464646ff464646ff454545ff434343ff414141db4f4f4f857b7b7b11ffffff01ffffff01ffffff0153535307222222331d1d1da91b1b1b8d4141412365656503ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff017c7c7c0f6868685d636363cb5e5e5eff626262ff636363ff636363ff636363ff636363ff636363ff656565ff4a4a4aff444444ff464646ff464646ff464646ff464646ff464646ff404040ff454545e14c4c4c6b69696917ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0177777733626262a3606060f3616161ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff656565ff4a4a4aff444444ff464646ff464646ff464646ff464646ff464646ff464646ff444444ff424242f9454545b55d5d5d49ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff014b4b4b0f5e5e5e85626262ff626262ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff656565ff4a4a4aff444444ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff444444ff414141ff454545a16464641dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0132323225333333cf4e4e4eff646464ff666666ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff636363ff656565ff4a4a4aff444444ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff474747ff404040ff303030e35757573bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd313131ff363636ff515151ff636363ff656565ff636363ff636363ff636363ff636363ff636363ff636363ff656565ff4a4a4aff444444ff464646ff464646ff464646ff464646ff464646ff464646ff464646ff414141ff373737ff343434ff323232e159595939ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff343434ff333333ff3c3c3cff5b5b5bff686868ff636363ff626262ff636363ff636363ff636363ff656565ff4a4a4aff444444ff464646ff464646ff454545ff464646ff4c4c4cff454545ff393939ff353535ff353535ff353535ff323232e159595939ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff363636ff363636ff353535ff313131ff3f3f3fff5d5d5dff666666ff646464ff626262ff636363ff656565ff4a4a4aff444444ff454545ff474747ff4a4a4aff404040ff212121ff2f2f2fff373737ff373737ff353535ff363636ff323232e159595939ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff363636ff363636ff363636ff353535ff333333ff363636ff484848ff646464ff676767ff626262ff656565ff4a4a4aff444444ff4b4b4bff4a4a4aff262626ff0b0b0bff090909ff171717ff252525ff353535ff393939ff363636ff323232e159595939ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff363636ff363636ff363636ff363636ff363636ff363636ff323232ff363636ff4c4c4cff646464ff676767ff4d4d4dff484848ff2c2c2cff0b0b0bff020202ff040404ff0b0b0bff171717ff141414ff161616ff282828ff353535ff343434e359595939ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff363636ff363636ff363636ff363636ff363636ff363636ff363636ff343434ff323232ff3f3f3fff5f5f5fff3a3a3aff161616ff030303ff030303ff050505ff040404ff0b0b0bff171717ff171717ff161616ff151515ff1a1a1aff242424e55555553bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff363636ff363636ff363636ff363636ff363636ff363636ff353535ff363636ff383838ff2e2e2eff191919ff262626ff111111ff030303ff030303ff050505ff040404ff0b0b0bff171717ff171717ff151515ff111111f9121212cd272727557d7d7d09ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff363636ff363636ff363636ff363636ff363636ff363636ff383838ff373737ff242424ff0b0b0bff0a0a0aff393939ff393939ff222222ff080808ff020202ff030303ff0b0b0bff181818ff0f0f0fff151515f32424247935353525ffffff01ffffff01ffffff01a3a3a30fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff363636ff363636ff363636ff363636ff383838ff373737ff272727ff0c0c0cff090909ff0c0c0cff0e0e0eff373737ff363636ff3a3a3aff393939ff1e1e1eff080808ff080808ff0f0f0feb232323914040401dffffff01ffffff01ffffff01ffffff01ffffff018282825d626262c36d6d6d4d8d8d8d09ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff363636ff353535ff363636ff3a3a3aff2f2f2fff131313ff0b0b0bff0b0b0bff0d0d0dff0c0c0cff0e0e0eff373737ff363636ff353535ff363636ff393939ff303030ff1c1c1cc92626264d68686807ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01868686515e5e5eff646464e9696969957878781fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723363636cd363636ff373737ff383838ff313131ff161616ff090909ff0b0b0bff0d0d0dff0d0d0dff0d0d0dff0c0c0cff0e0e0eff373737ff363636ff363636ff363636ff353535ff353535ff3c3c3c8fffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0186868651616161ff676767ff646464ff656565f16a6a6a7d7f7f7f25ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0137373723353535cd393939ff373737ff1f1f1fff0d0d0dff0a0a0aff0c0c0cff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff0e0e0eff373737ff363636ff363636ff363636ff363636ff353535ff37373791ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0186868651616161ff676767ff666666ff676767ff686868f9555555cd55555511ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0134343425323232cf212121ff0e0e0eff090909ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff0e0e0eff383838ff363636ff363636ff363636ff363636ff353535ff37373791ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0186868651616161ff686868ff696969ff5f5f5fff3d3d3dff303030ff4848481dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01474747132323238f020202ff080808ff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff0c0c0cff2e2e2eff393939ff363636ff353535ff363636ff353535ff37373791ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0185858551666666ff676767ff494949ff353535ff323232ff353535ff4e4e4e1bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0130303045101010af080808f70a0a0aff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff0d0d0dff131313ff1c1c1cff303030ff373737ff363636ff353535ff37373791ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0181818151494949ff363636ff313131ff363636ff353535ff363636ff4e4e4e1bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff0141414113191919690f0f0fdb060606ff0c0c0cff0d0d0dff0d0d0dff0d0d0dff0d0d0dff0c0c0cff0d0d0dff171717ff151515ff161616ff222222ff363636ff383838ff37373791ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff014d4d4d53272727c1242424ff373737ff373737ff353535ff353535ff363636ff4e4e4e1bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01626262091c1c1c830b0b0bd7090909ff0c0c0cff0d0d0dff0d0d0dff0c0c0cff0d0d0dff171717ff171717ff171717ff141414ff151515ff202020ff35353595ffffff01ffffff01ffffff01ffffff017474740540404049343434af2a2a2aff262626ff101010ff191919ff2e2e2eff373737ff363636ff363636ff4e4e4e1bffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff015a5a5a073636362d141414a7080808f5080808ff0d0d0dff0c0c0cff0d0d0dff171717ff171717ff171717ff151515ff0e0e0efb1b1b1bbb3d3d3d29ffffff01ffffff01ffffff0151515119393939892a2a2ae92d2d2dff323232ff282828ff141414ff151515ff151515ff1f1f1fff343434ff393939ff4949491dffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff013636362f111111b5070707f30a0a0aff0d0d0dff171717ff141414ff111111f5111111c74343433d70707005ffffff01ffffff017c7c7c034e4e4e632a2a2af7292929ff323232ff313131ff323232ff282828ff141414ff171717ff171717ff151515ff0e0e0efd222222e153535315ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff012d2d2d151f1f1f590e0e0edb040404ff0f0f0fff171717e7262626673f3f3f1dffffff01ffffff01ffffff01ffffff01ffffff01444444293535358b2d2d2deb2b2b2bff313131ff323232ff282828ff141414ff171717ff121212ff0d0d0dff2222229d2626263dbebebe03ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01505050112626266f1d1d1d7f36363617ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01616161213333339d2c2c2ce92f2f2fff282828ff111111ff111111f7191919ab3c3c3c41ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff015151510b3b3b3b43383838c51f1f1fff141414d71e1e1e654f4f4f13ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff015858580b4d4d4d4159595909ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff01ffffff010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000`
