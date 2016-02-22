package analysis

import (
	"testing"

	"flag"

	"github.com/alkasir/alkasir/pkg/central/db"
)

func TestAnalysis(t *testing.T) {
	flag.Set("logtostderr", "true")
	flag.Set("v", "100")
	go samplesAnalyzer()

	sampleAnalysisC <- []db.Sample{
		db.Sample{
			Token:  "1",
			Type:   "NewClientToken",
			Origin: "Central",
			Host:   "youtube.com",
		},
		db.Sample{
			Token:  "1",
			Type:   "HTTPHeader",
			Origin: "Central",
			Host:   "youtube.com",
			Data: ytresponse,
		},
		db.Sample{
			Token:  "1",
			Type:   "HTTPHeader",
			Origin: "Client",
			Host:   "youtube.com",
			Data: ytresponse2,
		},
		// db.Sample{
		// 	Token:  "1",
		// 	Type:   "DNSQuery",
		// 	Origin: "Central",
		// 	Host:   "youtube.com",
		// },
		// db.Sample{
		// 	Token:  "1",
		// 	Type:   "DNSQuery",
		// 	Origin: "Client",
		// 	Host:   "youtube.com",
		// },
	}


}



var ytresponse = []byte(`{
  "url":"http://www.youtube.com/watch?v=wflisz67bpe&feature=share&bpctr=1345276397&has_verified=1",
  "response_header":{
    "Server":"gwiseguy/2.0",
    "Expires":"Tue, 27 Apr 1971 19:44:06 EST",
    "X-Frame-Options":"SAMEORIGIN",
    "X-Xss-Protection":"1; mode=block; report=https://www.google.com/appserve/security-bugs/log/youtube",
    "Alt-Svc":"quic=\":443\"; ma=604800; v=\"30,29,28,27,26,25\"",
    "Set-Cookie":"VISITOR_INFO1_LIVE=skWQY7rKRU8; path=/; domain=.youtube.com; expires=Sun, 16-Oct-2016 09:01:29 GMT; httponly",
    "X-Content-Type-Options":"nosniff",
    "P3p":"CP=\"This is not a P3P policy! See http://support.google.com/accounts/answer/151657?hl=sv for more info.\"",
    "Date":"Mon, 15 Feb 2016 21:08:29 GMT",
    "Strict-Transport-Security":"max-age=604800",
    "Cache-Control":"no-cache",
    "Content-Type":"text/html; charset=utf-8",
    "Alternate-Protocol":"443:quic,p=1"
  },
  "redirects":[
    {
      "url":"http://www.youtube.com/watch?v=wflisz67bpe&feature=share&bpctr=1345276397&has_verified=1",
      "header":{
        "Server":"gwiseguy/2.0",
        "Expires":"Tue, 27 Apr 1971 19:44:06 EST",
        "Date":"Mon, 15 Feb 2016 21:08:29 GMT",
        "X-Xss-Protection":"1; mode=block; report=https://www.google.com/appserve/security-bugs/log/youtube",
        "Cache-Control":"no-cache",
        "X-Content-Type-Options":"nosniff",
        "Content-Length":"0",
        "Set-Cookie":"VISITOR_INFO1_LIVE=N1kD78VK8j0; path=/; domain=.youtube.com; expires=Sun, 16-Oct-2016 09:01:29 GMT; httponly",
        "Location":"https://www.youtube.com/watch?v=wflisz67bpe&feature=share&bpctr=1345276397&has_verified=1",
        "P3p":"CP=\"This is not a P3P policy! See http://support.google.com/accounts/answer/151657?hl=en for more info.\"",
        "Content-Type":"text/html; charset=utf-8"
      },
      "status_code":301
    }
  ],
  "error":"",
  "status_code":200
}`)


var ytresponse2 = []byte(`{
  "url":"http://www.youtube.com/watch?v=wflisz67bpe&feature=share&bpctr=1345276397&has_verified=1",
  "response_header":{
    "Server":"gwiseguy/2.0",
    "Expires":"Tue, 27 Apr 1971 19:44:06 EST",
    "X-Frame-Options":"SAMEORIGIN",
    "X-Xss-Protection":"1; mode=block; report=https://www.google.com/appserve/security-bugs/log/youtube",
    "Alt-Svc":"quic=\":443\"; ma=604800; v=\"30,29,28,27,26,25\"",
    "Set-Cookie":"VISITOR_INFO1_LIVE=skWQY7rKRU8; path=/; domain=.youtube.com; expires=Sun, 16-Oct-2016 09:01:29 GMT; httponly",
    "X-Content-Type-Options":"nosniff",
    "P3p":"CP=\"This is not a P3P policy! See http://support.google.com/accounts/answer/151657?hl=sv for more info.\"",
    "Date":"Mon, 15 Feb 2016 21:08:29 GMT",
    "Strict-Transport-Security":"max-age=604800",
    "Cache-Control":"no-cache",
    "Content-Type":"text/html; charset=utf-8",
    "Alternate-Protocol":"443:quic,p=1"
  },
  "redirects":[
    {
      "url":"http://www.youtube.com/watch?v=wflisz67bpe&feature=share&bpctr=1345276397&has_verified=1",
      "header":{
        "Server":"gwiseguy/2.0",
        "Expires":"Tue, 27 Apr 1971 19:44:06 EST",
        "Date":"Mon, 15 Feb 2016 21:08:29 GMT",
        "X-Xss-Protection":"1; mode=block; report=https://www.google.com/appserve/security-bugs/log/youtube",
        "Cache-Control":"no-cache",
        "X-Content-Type-Options":"nosniff",
        "Content-Length":"0",
        "Set-Cookie":"VISITOR_INFO1_LIVE=N1kD78VK8j0; path=/; domain=.youtube.com; expires=Sun, 16-Oct-2016 09:01:29 GMT; httponly",
        "Location":"https://www.youtube.com/watch?v=wflisz67bpe&feature=share&bpctr=1345276397&has_verified=1",
        "P3p":"CP=\"This is not a P3P policy! See http://support.google.com/accounts/answer/151657?hl=en for more info.\"",
        "Content-Type":"text/html; charset=utf-8"
      },
      "status_code":301
    }
  ],
  "error":"",
  "status_code":413
}`)
