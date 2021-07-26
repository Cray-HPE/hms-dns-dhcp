// MIT License
// 
// (C) Copyright [2021] Hewlett Packard Enterprise Development LP
// 
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package dns_dhcp

import (
	"log"
	"testing"
	"net/http"
	"net/http/httptest"
	"os"
	"encoding/json"
	"io/ioutil"
	"github.com/Cray-HPE/hms-smd/pkg/sm"
)

var expSvcName = "DNSHelper"
var gotUA bool

func TestNewDHCPDNSHelper(t *testing.T) {
	url := "http://a/b/c"
	hlp := NewDHCPDNSHelper(url,nil)
	if (hlp.HTTPClient == nil) {
		t.Errorf("ERROR, New func didn't create HTTP client.")
	}
	if (hlp.HSMURL != url) {
		t.Errorf("ERROR, New func didn't create HSM url.")
	}
	hname,_ := os.Hostname()
	if (serviceName != hname) {
		t.Errorf("ERROR, New func didn't set service/host name.")
	}

	serviceName = ""
	hlp = NewDHCPDNSHelperInstance(url,nil,"XYZZY")
	if (serviceName != "XYZZY") {
		t.Errorf("ERROR, NewInstance func didn't set service/host name.")
	}
}

func hasUserAgentHeader(r *http.Request) bool {
    if (len(r.Header) == 0) {
        return false
    }

    alist,ok := r.Header["User-Agent"]
    if (!ok) {
        return false
    }

    for _,hdr := range(alist) {
        if (hdr == expSvcName) {
            return true
        }
    }
    return true
}

var ethComps = []sm.CompEthInterface{ {ID: "AAA",}, {ID: "BBB",},}
var ethCompsStatic = []sm.CompEthInterface{ {ID: "AAA",}, {ID: "BBB",},}

func handleUnk(w http.ResponseWriter, req *http.Request) {
	gotUA = hasUserAgentHeader(req)

	if (req.Method == "GET") {
		ba,baerr := json.Marshal(ethCompsStatic)
		if (baerr != nil) {
			log.Printf("GET: Can't marshal payload: %v",baerr)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type","application/json")
		w.Write(ba)
	} else if (req.Method == "POST") {
		var eee sm.CompEthInterface
		body,_ := ioutil.ReadAll(req.Body)
		err := json.Unmarshal(body,&eee)
		if (err != nil) {
			log.Printf("POST: Can't unmarshal req body: %v",err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		ethComps = append(ethComps,eee)
		w.WriteHeader(http.StatusCreated)
	} else if (req.Method == "PATCH") {
		var eee sm.CompEthInterface
		body,_ := ioutil.ReadAll(req.Body)
		err := json.Unmarshal(body,&eee)
		if (err != nil) {
			log.Printf("PATCH: Can't unmarshal req body: %v",err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		for ix,_ := range(ethComps) {
			if (ethComps[ix].ID == eee.ID) {
				ethComps[ix].Type = eee.Type //only patch Type field
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}

func TestGetUnknownComponents(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(handleUnk))
	defer srv.Close()

	hlp := NewDHCPDNSHelperInstance(srv.URL,nil,expSvcName)
	gotUA = false
	compList,err := hlp.GetUnknownComponents()
	if (err != nil) {
		t.Errorf("ERROR, GetUnknownComponents() error: %v",err)
	}
	if (len(compList) != 2) {
		t.Errorf("ERROR expecting 2 components, got %d",len(compList))
	}
	if (compList[0].ID != "AAA") {
		t.Errorf("ERROR comp 0 wrong name, exp 'AAA', got '%s'",
			compList[0].ID)
	}
	if (compList[1].ID != "BBB") {
		t.Errorf("ERROR comp 1 wrong name, exp 'BBB', got '%s'",
			compList[1].ID)
	}
	if (!gotUA) {
		t.Errorf("ERROR, no User-Agent header found.")
	}
}

func TestGetAllEthernetInterfaces(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(handleUnk))
	defer srv.Close()

	hlp := NewDHCPDNSHelperInstance(srv.URL,nil,expSvcName)
	gotUA = false
	compList,err := hlp.GetAllEthernetInterfaces()
	if (err != nil) {
		t.Errorf("ERROR, GetAllEthernetInterfaces() error: %v",err)
	}
	if (len(compList) != 2) {
		t.Errorf("ERROR expecting 2 components, got %d",len(compList))
	}
	if (compList[0].ID != "AAA") {
		t.Errorf("ERROR comp 0 wrong name, exp 'AAA', got '%s'",
			compList[0].ID)
	}
	if (compList[1].ID != "BBB") {
		t.Errorf("ERROR comp 1 wrong name, exp 'BBB', got '%s'",
			compList[1].ID)
	}
	if (!gotUA) {
		t.Errorf("ERROR, no User-Agent header found.")
	}
}

func TestAddNewEthernetInterface(t *testing.T) {
	var ethi sm.CompEthInterface

	srv := httptest.NewServer(http.HandlerFunc(handleUnk))
	defer srv.Close()

	hlp := NewDHCPDNSHelperInstance(srv.URL,nil,expSvcName)
	gotUA = false
	ethi.ID = "CCC"
	err := hlp.AddNewEthernetInterface(ethi,false)
	if (err != nil) {
		t.Errorf("ERROR, AddNewEthernetInterface() error: %v",err)
	}
	if (!gotUA) {
		t.Errorf("ERROR, no User-Agent header found.")
	}
	if (len(ethComps) != 3) {
		t.Errorf("ERROR, AddNewEthernetInterface() didn't add anything.")
	}
}

func TestPatchEthernetInterface(t *testing.T) {
	var ethi sm.CompEthInterface
	srv := httptest.NewServer(http.HandlerFunc(handleUnk))
	defer srv.Close()

	hlp := NewDHCPDNSHelperInstance(srv.URL,nil,expSvcName)
	gotUA = false
	ethi.ID = "BBB"
	ethi.Type = "None"
	err := hlp.PatchEthernetInterface(ethi)
	if (err != nil) {
		t.Errorf("ERROR, PatchEthernetInterface() error: %v",err)
	}
	if (!gotUA) {
		t.Errorf("ERROR, no User-Agent header found.")
	}

	for _,eee := range(ethComps) {
		if (eee.ID == "BBB") {
			if (eee.Type != "None") {
				t.Errorf("ERROR, PatchEthernetInterface() didn't patch.")
			}
		}
	}
}


